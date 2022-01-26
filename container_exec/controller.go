package container_exec

import (
	"errors"
)

type CommandStatus struct {
	CommandId string
	Command   string
	Running   bool
}

type ContainerExecController struct {
	commands map[string]*Command
}

func NewContainerExecController() *ContainerExecController {
	controller := ContainerExecController{}
	controller.commands = make(map[string]*Command)
	return &controller
}

//-------------------------------------------------------------------------------------------------
func (c *ContainerExecController) StartCommand(command string) *Command {
	cmd := NewCommand(command)
	cmd.Start()
	c.commands[cmd.CommandId] = &cmd
	return &cmd
}

//-------------------------------------------------------------------------------------------------
func (c *ContainerExecController) FindCommand(command_id string) (*Command, error) {
	cmd, ok := c.commands[command_id]
	if !ok {
		return nil, errors.New("Unknown command: " + command_id)
	}
	return cmd, nil
}

//-------------------------------------------------------------------------------------------------
func (c *ContainerExecController) FinishCommand(command_id string) error {
	cmd, err := c.FindCommand(command_id)
	if err != nil {
		return err
	}

	delete(c.commands, command_id)
	cmd.Close()
	return nil
}

//-------------------------------------------------------------------------------------------------
func (c *ContainerExecController) Close() {
	var uuids []string
	for id := range c.commands {
		uuids = append(uuids, id)
	}

	for _, id := range uuids {
		c.FinishCommand(id)
	}
}

//-------------------------------------------------------------------------------------------------
func (c *ContainerExecController) Commands() (commands []CommandStatus) {
	for command_id, cmd := range c.commands {
		commands = append(commands, CommandStatus{
			CommandId: command_id,
			Command:   cmd.Command,
			Running:   cmd.Running(),
		})
	}
	return commands
}
