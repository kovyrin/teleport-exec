package cgroups

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

const rootPath = "/sys/fs/cgroup/containerize"

//-------------------------------------------------------------------------------------------------
// Sets up cgroups support before we could use it for per-process limits
func Setup() error {
	log.Println("Setting up cgroups for containerized command execution...")

	// Make sure we have cgroup2 enabled
	mounted, err := cgroup2Mounted()
	if err != nil {
		return fmt.Errorf("failed to check cgroup2 mounts: %w", err)
	}
	if !mounted {
		return errors.New("cgroup2 is not enabled on this system")
	}

	// Create a new cgroup to be shared by all containers
	if err := os.MkdirAll(rootPath, 0644); err != nil {
		return fmt.Errorf("failed to create a root cgroup for our containers %s: %w", rootPath, err)
	}

	// Enable controllers we need
	subtree_control := rootPath + "/cgroup.subtree_control"
	if err := os.WriteFile(subtree_control, []byte("+memory +io +cpu"), 0644); err != nil {
		return fmt.Errorf("failed to enable cgroup controllers for the root cgroup: %w", err)
	}

	return nil
}

//-------------------------------------------------------------------------------------------------
func cgroup2Mounted() (bool, error) {
	mounts_file, err := os.Open("/proc/mounts")
	if err != nil {
		return false, fmt.Errorf("failed to open the mounts list: %w", err)
	}
	defer mounts_file.Close()

	scanner := bufio.NewScanner(mounts_file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if fields[0] == "cgroup2" {
			return true, nil
		}
	}
	return false, nil
}

//-------------------------------------------------------------------------------------------------
// Cleans up our cgroups before process shutdown
func TearDown() error {
	log.Println("Cleaning up cgroups...")
	return os.RemoveAll(rootPath)
}
