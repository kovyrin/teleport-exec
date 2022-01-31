package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

//-------------------------------------------------------------------------------------------------
type Container struct {
	container_id string
}

//-------------------------------------------------------------------------------------------------
func NewContainer(container_id string) (*Container, error) {
	cgroup := Container{container_id: container_id}

	if err := os.MkdirAll(cgroup.groupPath(), 0644); err != nil {
		return nil, fmt.Errorf("failed to create a cgroup for container %s: %w", container_id, err)
	}

	return &cgroup, nil
}

//-------------------------------------------------------------------------------------------------
func (c *Container) Close() error {
	return os.RemoveAll(c.groupPath())
}

//-------------------------------------------------------------------------------------------------
func (c *Container) groupPath() string {
	return filepath.Join(rootPath, c.container_id)
}

func (c *Container) controlPath(control_file string) string {
	return filepath.Join(c.groupPath(), control_file)
}

func (c *Container) writeControl(control_file string, content string) error {
	return os.WriteFile(c.controlPath(control_file), []byte(content), 0644)
}

//-------------------------------------------------------------------------------------------------
func (c *Container) AddProcess(pid int) error {
	return c.writeControl("cgroup.procs", fmt.Sprintf("%d", pid))
}

//------------------------------------------------------------------------------------------------
func (c *Container) MemoryLimitBytes(limit uint) error {
	return c.writeControl("memory.max", strconv.Itoa(int(limit)))
}

func (c *Container) IoWeight(limit uint) error {
	return c.writeControl("io.bfq.weight", fmt.Sprintf("default %d", limit))
}

func (c *Container) CpuLimitPct(percent uint) error {
	return c.writeControl("cpu.max", fmt.Sprintf("%d 1000000", percent*10000))
}
