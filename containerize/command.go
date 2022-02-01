package containerize

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"teleport-exec/cgroups"
	"teleport-exec/filestream"

	"github.com/docker/engine/pkg/reexec"
	"github.com/google/uuid"
	"go.uber.org/multierr"
)

//-------------------------------------------------------------------------------------------------
var defaultEnvironment = []string{
	"HOME=/root",
	"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	"TERM=xterm",
}

// Command represents the state of a single containerized command
type Command struct {
	CommandId string
	Command   string
	Args      []string
	executor  *exec.Cmd
	log       *ProcessLog

	mu sync.RWMutex

	cgroup *cgroups.Container

	done chan bool // Used to notify all Wait() calls that the process has died

	running  bool // Used to easily check the running status of the process without waitpid, etc
	isClosed bool // Used to prevent double-close
	started  bool // Used to prevent double-start
}

// NewCommand sets up a container for a given command and returns a Command instance associated with it
func NewCommand(command []string) (*Command, error) {
	c := Command{
		CommandId: uuid.NewString(),
		Command:   command[0],
		Args:      command[1:],
		running:   false,
		done:      make(chan bool),
	}

	// Set up the command execution
	cmd := reexec.Command(append([]string{"executeCommand"}, command...)...)
	c.executor = cmd
	cmd.SysProcAttr = c.sysProcAttr()
	cmd.Env = defaultEnvironment

	// Redirect command output to a command-specific log (so that we could read/stream it later)
	pl, err := NewProcessLog(c.CommandId)
	if err != nil {
		return nil, fmt.Errorf("failed to set up logging for command '%s': %w", c.CommandId, err)
	}
	c.log = pl
	cmd.Stdout = pl.fd
	cmd.Stderr = pl.fd

	return &c, nil
}

// Start executes the command in a container and starts a separate thread waiting for the command to finish
func (c *Command) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return errors.New("command is already running")
	}
	c.started = true

	// On Linux, pdeathsig (which we use to ensure that all containerized processed die with the server)
	// will kill the child process when the thread starting it dies, not when the process dies.
	// runtime.LockOSThread ensures that as long as the goroutine which called cmd.Start is executing,
	// the backing OS thread will still be around and the process won't be killed by the OS.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	log.Printf("Starting command '%s' with %d args: %v", c.Command, len(c.Args), c.Args)
	if err := c.executor.Start(); err != nil {
		return fmt.Errorf("failed to run command '%s': %v", c.Command, err)
	}
	c.running = true

	if err := c.setupLimits(); err != nil {
		return fmt.Errorf("failed to set up resource limits: %w", err)
	}

	// Start a separate thread waiting for the command to finish
	go c.waitForCompletion()

	return nil
}

//-------------------------------------------------------------------------------------------------
func (c *Command) waitForCompletion() {
	// Block until command is done and get the exit status
	_ = c.executor.Wait()

	// Now that the process is dead, update the running state accordingly
	c.mu.Lock()
	defer c.mu.Unlock()
	c.running = false

	// Let all log streams know that there is no reason to wait for more output anymore
	c.log.LogComplete()

	// Let all Wait() calls know we're done with the command
	close(c.done)
}

func (c *Command) pid() int {
	return c.executor.Process.Pid
}

// Sets up Cgroup-based limits for the process running the command
func (c *Command) setupLimits() error {
	log.Printf("Setting up resource limits for command '%s'", c.CommandId)
	var err error

	// Create a cgroup for this command
	c.cgroup, err = cgroups.NewContainer(c.CommandId)
	if err != nil {
		return fmt.Errorf("failed to setup cgroups for command %s: %w", c.CommandId, err)
	}

	// Add the process to the cgroup
	if err := c.cgroup.AddProcess(c.pid()); err != nil {
		return fmt.Errorf("failed to add process to cgroup: %w", err)
	}

	// Set up limits
	if err := c.cgroup.MemoryLimitBytes(10 * 1024 * 1024); err != nil { // 10Mb limit
		return fmt.Errorf("failed to set a memory limit: %w", err)
	}
	if err := c.cgroup.IoWeight(1); err != nil { // Put the lowest IO-weight on this command
		return fmt.Errorf("failed to set a io weight limit: %w", err)
	}
	if err := c.cgroup.CpuLimitPct(10); err != nil { // 10% of CPU maximum per command
		return fmt.Errorf("failed to set a CPU limit: %w", err)
	}
	return nil
}

//-------------------------------------------------------------------------------------------------
func (c *Command) sysProcAttr() *syscall.SysProcAttr {
	nobodyUid := 65534
	nobodyGid := 65534

	nobody, err := user.Lookup("nobody")
	if err == nil && nobody != nil {
		uid, err := strconv.Atoi(nobody.Uid)
		if err != nil {
			nobodyUid = uid
		}

		gid, err := strconv.Atoi(nobody.Gid)
		if err != nil {
			nobodyGid = gid
		}
	}

	attr := syscall.SysProcAttr{
		// Use the same process group for all processed spawned by the command
		// This is needed to make sure we can kill the whole tree at once
		Setpgid: true,

		// Make sure all processes die when the server is killed
		Pdeathsig: syscall.SIGKILL,

		// Isolate the process as needed
		Cloneflags: syscall.CLONE_NEWUTS | // New UTS IPC namespace (isolated hostname, etc)
			syscall.CLONE_NEWPID | // Isolated PID namespace
			syscall.CLONE_NEWIPC | // Isolated IPC namespace
			syscall.CLONE_NEWNET | // Isolated network environment
			syscall.CLONE_NEWNS | // Isolated mount namespace
			syscall.CLONE_NEWUSER, // Isolated user namespace

		// Whatever uid/gid we use for the server will be mapped into root within the container
		// This is needed to allow the re-executed binary to mount /proc, etc. before dropping privileges
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
			{
				ContainerID: 65534,
				HostID:      nobodyUid,
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
			{
				ContainerID: 65534,
				HostID:      nobodyGid,
				Size:        1,
			},
		},
	}

	return &attr
}

// Running returns true if the command is still running
func (c *Command) Running() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// Wait blocks until the process finishes (without relying on waitpid, which destroys process exit status info)
func (c *Command) Wait() {
	if !c.Running() {
		return
	}
	<-c.done
}

// Kill terminates the command process (including all sub-processes) if it is running
func (c *Command) Kill() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	// Kill the whole process group
	return syscall.Kill(-c.pid(), syscall.SIGKILL)
}

// ResultCode returns the process status code for the command (only if it has finished)
func (c *Command) ResultCode() (int, error) {
	if c.Running() {
		return 0, errors.New("no result code available for a running command")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.executor.ProcessState.ExitCode(), nil
}

// ResultDescription returns a string description of the process results (exit code, etc)
func (c *Command) ResultDescription() (string, error) {
	if c.Running() {
		return "", errors.New("no result description available for a running command")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.executor.ProcessState.String(), nil
}

// Close terminates the container, closes log files, releases other resources used by the command
func (c *Command) Close() (err error) {
	// Prevent double-close
	c.mu.Lock()
	if c.isClosed {
		return nil
	}
	c.isClosed = true
	c.mu.Unlock()

	// Make sure the process has stopped, and we have a result status
	err = multierr.Append(err, c.Kill())
	c.Wait()

	// Close the log stream
	err = multierr.Append(err, c.log.Close())

	// Cleanup cgroups
	err = multierr.Append(err, c.cgroup.Close())

	return err
}

// NewLogStream returns a filestream.FileStream object associated with the output from this command
func (c *Command) NewLogStream(ctx context.Context) (*filestream.FileStream, error) {
	return c.log.NewLogStream(ctx, c.Running())
}

// CloseLogStream closes a given command output stream
func (c *Command) CloseLogStream(log *filestream.FileStream) error {
	return c.log.CloseLogStream(log)
}
