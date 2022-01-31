package containerize

import (
	"errors"
)

type CommandStatus struct {
	CommandId string
	Command   string
	Args      []string
	Running   bool
}

type Controller struct {
	commands map[string]*Command
}

func NewController() *Controller {
	controller := Controller{}
	controller.commands = make(map[string]*Command)
	return &controller
}

//-------------------------------------------------------------------------------------------------
func (c *Controller) StartCommand(command []string) (*Command, error) {
	cmd, err := NewCommand(command)
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	c.commands[cmd.CommandId] = cmd
	return cmd, nil
}

//-------------------------------------------------------------------------------------------------
func (c *Controller) FindCommand(command_id string) (*Command, error) {
	cmd, ok := c.commands[command_id]
	if !ok {
		return nil, errors.New("Unknown command: " + command_id)
	}
	return cmd, nil
}

//-------------------------------------------------------------------------------------------------
func (c *Controller) FinishCommand(command_id string) error {
	cmd, err := c.FindCommand(command_id)
	if err != nil {
		return err
	}

	delete(c.commands, command_id)
	cmd.Close()
	return nil
}

//-------------------------------------------------------------------------------------------------
func (c *Controller) Close() {
	var uuids []string
	for id := range c.commands {
		uuids = append(uuids, id)
	}

	for _, id := range uuids {
		c.FinishCommand(id)
	}
}

//-------------------------------------------------------------------------------------------------
func (c *Controller) Commands() (commands []CommandStatus) {
	for command_id, cmd := range c.commands {
		commands = append(commands, CommandStatus{
			CommandId: command_id,
			Command:   cmd.Command,
			Args:      cmd.Args,
			Running:   cmd.Running(),
		})
	}
	return commands
}
