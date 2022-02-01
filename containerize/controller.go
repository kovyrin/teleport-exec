package containerize

import (
	"errors"
	"go.uber.org/multierr"
	"sync"
)

// CommandStatus represents a snapshot of status information for a single command
type CommandStatus struct {
	CommandId string
	Command   string
	Args      []string
	Running   bool
}

// Controller represents a container management controller state
type Controller struct {
	commands map[string]*Command
	mu       sync.RWMutex
}

// NewController sets up a new controller for managing containers
func NewController() *Controller {
	controller := Controller{}
	controller.commands = make(map[string]*Command)
	return &controller
}

// StartCommand creates a new container and runs a given command inside
func (c *Controller) StartCommand(command []string) (*Command, error) {
	cmd, err := NewCommand(command)
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.commands[cmd.CommandId] = cmd

	return cmd, nil
}

// FindCommand returns an instance of a container for a given commandId
func (c *Controller) FindCommand(commandId string) (*Command, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cmd, ok := c.commands[commandId]
	if !ok {
		return nil, errors.New("Unknown command: " + commandId)
	}
	return cmd, nil
}

// FinishCommand terminates
func (c *Controller) FinishCommand(commandId string) error {
	cmd, err := c.FindCommand(commandId)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.commands, commandId)
	return cmd.Close()
}

// Close shuts down the container controller and kills all running commands
func (c *Controller) Close() (err error) {
	var uuids []string
	c.mu.RLock()
	for id := range c.commands {
		uuids = append(uuids, id)
	}
	c.mu.RUnlock()

	for _, id := range uuids {
		err = multierr.Append(err, c.FinishCommand(id))
	}
	return err
}

// Commands returns a read-only snapshot of the current state of all containers registered with the controller
func (c *Controller) Commands() (commands []CommandStatus) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for commandId, cmd := range c.commands {
		commands = append(commands, CommandStatus{
			CommandId: commandId,
			Command:   cmd.Command,
			Args:      cmd.Args,
			Running:   cmd.Running(),
		})
	}
	return commands
}
