package container_exec

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
func (s *Command) Start() error {
	if s.executor != nil {
		return errors.New("command is already running")
	}

	// Lock the state while we're changing stuff around here
	s.commandMutex.Lock()
	defer s.commandMutex.Unlock()

	log.Printf("Starting command '%s' with %d args: %v", s.Command, len(s.Args), s.Args)

	// Set up the command execution
	s.executor = exec.Command(s.Command, s.Args...)
	s.executor.SysProcAttr = s.sysProcAttr()

	// Redirect command output to a command-specific log (so that we could read/stream it later)
	pl, err := NewProcessLog(s.CommandId)
	if err != nil {
		return err
	}
	s.log = pl
	s.executor.Stdout = pl.fd
	s.executor.Stderr = pl.fd

	// On Linux, pdeathsig will kill the child process when the thread dies,
	// not when the process dies. runtime.LockOSThread ensures that as long
	// as this function is executing that OS thread will still be around
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err = s.executor.Start()
	if err != nil {
		s.failure = err
		return fmt.Errorf("failed to run command '%s': %v", s.Command, err)
	}
	s.running = true

	go func() {
		s.executor.Wait()

		// Now that the process is dead, update the running state accordingly
		s.commandMutex.Lock()
		s.running = false
		s.commandMutex.Unlock()

		// Let all log streams know that there is no reason to wait for more output anymore
		s.log.LogComplete()
	}()

	return nil
}

//-------------------------------------------------------------------------------------------------
func (s *Command) sysProcAttr() *syscall.SysProcAttr {
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
func (s *Command) Running() bool {
	s.commandMutex.RLock()
	running := s.running
	s.commandMutex.RUnlock()
	return running
}

// Waits for the process to finish (without relying on waitpid, which destroys process exit status info)
func (s *Command) Wait() {
	for s.Running() {
		time.Sleep(100 * time.Millisecond)
	}
}

//-------------------------------------------------------------------------------------------------
// Terminates the command if it is running (including all sub-processes)
func (s *Command) Kill() {
	if s.Running() {
		s.commandMutex.Lock()
		defer s.commandMutex.Unlock()

		// Kill the whole process group
		killPg(s.executor.Process.Pid)
	}
}

func killPg(pgid int) error {
	if pgid > 0 {
		pgid = -pgid
	}

	return syscall.Kill(pgid, syscall.SIGKILL)
}

//-------------------------------------------------------------------------------------------------
func (s *Command) ResultCode() (int, error) {
	if s.Running() {
		return 0, errors.New("no result code available for a running command")
	}

	s.commandMutex.RLock()
	defer s.commandMutex.RUnlock()
	return s.executor.ProcessState.ExitCode(), nil
}

func (s *Command) ResultDescription() (string, error) {
	if s.Running() {
		return "", errors.New("no result description available for a running command")
	}

	s.commandMutex.RLock()
	defer s.commandMutex.RUnlock()
	return s.executor.ProcessState.String(), nil
}

//-------------------------------------------------------------------------------------------------
func (s *Command) Failure() error {
	s.commandMutex.RLock()
	defer s.commandMutex.RUnlock()
	return s.failure
}

//-------------------------------------------------------------------------------------------------
// Closes log files, releases other resources used by the command
func (s *Command) Close() {
	// Make sure the process has stopped
	if s.Running() {
		s.Kill()
	}

	// Make sure we release all the resources associated with the command
	s.Wait()

	// Close the log stream
	s.log.Close()
}

//-------------------------------------------------------------------------------------------------
func (s *Command) NewLogStream(ctx context.Context) (*filestream.FileStream, error) {
	return s.log.NewLogStream(ctx, s.Running())
}

func (s *Command) CloseLogStream(log *filestream.FileStream) error {
	return s.log.CloseLogStream(log)
}
