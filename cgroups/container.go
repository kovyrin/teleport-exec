package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Container represents a single command-specific cgroup
type Container struct {
	containerId string
}

// NewContainer returns an instance of a cgroups-based container for a given container id
func NewContainer(containerId string) (*Container, error) {
	cgroup := Container{containerId: containerId}

	if err := os.MkdirAll(cgroup.groupPath(), 0644); err != nil {
		return nil, fmt.Errorf("failed to create a cgroup for container %s: %w", containerId, err)
	}

	return &cgroup, nil
}

// Close removes all cgroup-based limits for a given container
func (c *Container) Close() error {
	return os.RemoveAll(c.groupPath())
}

// AddProcess adds a given process to the cgroup
func (c *Container) AddProcess(pid int) error {
	return c.writeControl("cgroup.procs", fmt.Sprintf("%d", pid))
}

// MemoryLimitBytes limits memory usage of the container down to a given size (in bytes)
func (c *Container) MemoryLimitBytes(limit uint) error {
	return c.writeControl("memory.max", strconv.Itoa(int(limit)))
}

// IoWeight sets the default IO-weight for the container
func (c *Container) IoWeight(limit uint) error {
	return c.writeControl("io.bfq.weight", fmt.Sprintf("default %d", limit))
}

// CpuLimitPct limits maximum CPU allocation for the container down to a given percentage of total CPU power
func (c *Container) CpuLimitPct(percent uint) error {
	return c.writeControl("cpu.max", fmt.Sprintf("%d 1000000", percent*10000))
}

//-------------------------------------------------------------------------------------------------
func (c *Container) groupPath() string {
	return filepath.Join(rootPath, c.containerId)
}

func (c *Container) controlPath(controlFile string) string {
	return filepath.Join(c.groupPath(), controlFile)
}

func (c *Container) writeControl(controlFile string, content string) error {
	return os.WriteFile(c.controlPath(controlFile), []byte(content), 0644)
}
