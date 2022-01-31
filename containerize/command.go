package containerize

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"teleport-exec/filestream"
	"time"

	"github.com/google/uuid"
)

//-------------------------------------------------------------------------------------------------
type Command struct {
	CommandId    string
	Command      string
	Args         []string
	executor     *exec.Cmd
	log          *ProcessLog
	running      bool
	failure      error
	commandMutex sync.RWMutex
}

//-------------------------------------------------------------------------------------------------
func NewCommand(command []string) Command {
	return Command{
		CommandId: uuid.NewString(),
		Command:   command[0],
		Args:      command[1:],
		running:   false,
	}
}

//-------------------------------------------------------------------------------------------------
// Starts the command in a separate thread
func (c *Command) Start() error {
	if c.executor != nil {
		return errors.New("command is already running")
	}

	// Lock the state while we're changing stuff around here
	c.commandMutex.Lock()
	defer c.commandMutex.Unlock()

	log.Printf("Starting command '%s' with %d args: %v", c.Command, len(c.Args), c.Args)

	// Set up the command execution
	c.executor = exec.Command(c.Command, c.Args...)
	c.executor.SysProcAttr = c.sysProcAttr()

	// Redirect command output to a command-specific log (so that we could read/stream it later)
	pl, err := NewProcessLog(c.CommandId)
	if err != nil {
		return err
	}
	c.log = pl
	c.executor.Stdout = pl.fd
	c.executor.Stderr = pl.fd

	// On Linux, pdeathsig will kill the child process when the thread dies,
	// not when the process dies. runtime.LockOSThread ensures that as long
	// as this function is executing that OS thread will still be around
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err = c.executor.Start()
	if err != nil {
		c.failure = err
		return fmt.Errorf("failed to run command '%s': %v", c.Command, err)
	}
	c.running = true

	go func() {
		c.executor.Wait()

		// Now that the process is dead, update the running state accordingly
		c.commandMutex.Lock()
		c.running = false
		c.commandMutex.Unlock()

		// Let all log streams know that there is no reason to wait for more output anymore
		c.log.LogComplete()
	}()

	return nil
}

//-------------------------------------------------------------------------------------------------
func (c *Command) sysProcAttr() *syscall.SysProcAttr {
	attr := syscall.SysProcAttr{
		// Use the same process group for all processed spawned by the command
		// This is needed to make sure we can kill the whole tree at once
		Setpgid: true,

		// Make sure all processes die when the server is killed
		Pdeathsig: syscall.SIGKILL,

		// Isolate the process as needed
		Cloneflags: syscall.CLONE_NEWUTS | // New UTS IPC namespace (isolated hostname, etc)
			syscall.CLONE_NEWPID | // Isolated PID namespace
			syscall.CLONE_NEWNET, // Isolated network environment

		// This requires more magic if we want to really chroot somewhere (by itself it does not do much)
		// syscall.CLONE_NEWNS | // Isolated mount namespace

		// syscall.CLONE_NEWUSER | // Isolated user namespace
		// Whatever uid/gid we use for the server will be mapped into root within the container
		// UidMappings: []syscall.SysProcIDMap{
		// 	{
		// 		ContainerID: 0,
		// 		HostID:      os.Getuid(),
		// 		Size:        1,
		// 	},
		// },
		// GidMappings: []syscall.SysProcIDMap{
		// 	{
		// 		ContainerID: 0,
		// 		HostID:      os.Getgid(),
		// 		Size:        1,
		// 	},
		// },
	}

	return &attr
}

//-------------------------------------------------------------------------------------------------
// Returns true if the command is still running
func (c *Command) Running() bool {
	c.commandMutex.RLock()
	running := c.running
	c.commandMutex.RUnlock()
	return running
}

// Waits for the process to finish (without relying on waitpid, which destroys process exit status info)
func (c *Command) Wait() {
	for c.Running() {
		time.Sleep(100 * time.Millisecond)
	}
}

//-------------------------------------------------------------------------------------------------
// Terminates the command if it is running (including all sub-processes)
func (c *Command) Kill() {
	if c.Running() {
		c.commandMutex.Lock()
		defer c.commandMutex.Unlock()

		// Kill the whole process group
		killPg(c.executor.Process.Pid)
	}
}

func killPg(pgid int) error {
	if pgid > 0 {
		pgid = -pgid
	}

	return syscall.Kill(pgid, syscall.SIGKILL)
}

//-------------------------------------------------------------------------------------------------
func (c *Command) ResultCode() (int, error) {
	if c.Running() {
		return 0, errors.New("no result code available for a running command")
	}

	c.commandMutex.RLock()
	defer c.commandMutex.RUnlock()
	return c.executor.ProcessState.ExitCode(), nil
}

func (c *Command) ResultDescription() (string, error) {
	if c.Running() {
		return "", errors.New("no result description available for a running command")
	}

	c.commandMutex.RLock()
	defer c.commandMutex.RUnlock()
	return c.executor.ProcessState.String(), nil
}

//-------------------------------------------------------------------------------------------------
func (c *Command) Failure() error {
	c.commandMutex.RLock()
	defer c.commandMutex.RUnlock()
	return c.failure
}

//-------------------------------------------------------------------------------------------------
// Closes log files, releases other resources used by the command
func (c *Command) Close() {
	// Make sure the process has stopped
	if c.Running() {
		c.Kill()
	}

	// Make sure we release all the resources associated with the command
	c.Wait()

	// Close the log stream
	c.log.Close()
}

//-------------------------------------------------------------------------------------------------
func (c *Command) NewLogStream(ctx context.Context) (*filestream.FileStream, error) {
	return c.log.NewLogStream(ctx, c.Running())
}

func (c *Command) CloseLogStream(log *filestream.FileStream) error {
	return c.log.CloseLogStream(log)
}
