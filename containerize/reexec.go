package containerize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// This is executed when we apply reexec hack
// The command to execute is passed as command arguments starting with Args[1]
func ExecuteCommand() {
	// Mount the new isolated proc namespace
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(0), ""); err != nil {
		fmt.Println("Error mounting /proc:", err)
		os.Exit(42)
	}

	// Hide the parent's hostname
	if err := syscall.Sethostname([]byte("container")); err != nil {
		fmt.Println("Error setting hostname:", err)
		os.Exit(42)
	}

	// Drop permissions to allow the command from changing the host filesystem, etc
	if err := syscall.Setregid(65534, 65534); err != nil {
		fmt.Printf("Could not drop permissions to the nobody group: %v\n", err)
		os.Exit(42)
	}
	if err := syscall.Setreuid(65534, 65534); err != nil {
		fmt.Printf("Could not drop permissions to the nobody user: %v\n", err)
		os.Exit(42)
	}

	// Find the binary to execute
	path := os.Args[1]
	if filepath.Base(path) == path {
		if lp, err := exec.LookPath(path); err != nil {
			fmt.Printf("Could not find the binary to execute for command '%s': %v\n", path, err)
		} else {
			path = lp
		}
	}

	// Replace the current process with the command we want to execute
	if err := syscall.Exec(path, os.Args[1:], os.Environ()); err != nil {
		fmt.Printf("Error starting the command (%v): %s\n", os.Args, err)
		os.Exit(42)
	}

	// This should never happen, but just to be sure
	os.Exit(125)
}