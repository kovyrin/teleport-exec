package cgroups

import (
	"fmt"
	"os"
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
	return rootPath + "/" + c.container_id
}

func (c *Container) controlPath(control_file string) string {
	return c.groupPath() + "/" + control_file
}

//-------------------------------------------------------------------------------------------------
func (c *Container) AddProcess(pid int) error {
	return retryingWriteFile(c.controlPath("cgroup.procs"), fmt.Sprintf("%d", pid), 0644)
}

//-------------------------------------------------------------------------------------------------
func (c *Container) MemoryLimitBytes(limit uint) error {
	return retryingWriteFile(c.controlPath("memory.max"), fmt.Sprintf("%d", limit), 0644)
}

func (c *Container) IoWeight(limit uint) error {
	return retryingWriteFile(c.controlPath("io.bfq.weight"), fmt.Sprintf("default %d", limit), 0644)
}

func (c *Container) CpuLimitPct(percent uint) error {
	return retryingWriteFile(c.controlPath("cpu.max"), fmt.Sprintf("%d 1000000", percent*10000), 0644)
}
