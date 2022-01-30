package containerize

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewContainerExecController(t *testing.T) {
	Convey("Generates a new controller", t, func() {
		controller := NewController()

		Convey("With an empty command set", func() {
			So(controller.commands, ShouldBeEmpty)
		})
	})
}

func TestContainerExecController_StartCommand(t *testing.T) {
	Convey("StartCommand()", t, func() {
		c := NewController()

		Convey("Adds a new command to the command set", func() {
			So(len(c.commands), ShouldEqual, 0)
			c.StartCommand([]string{"echo", "banana"})
			So(len(c.commands), ShouldEqual, 1)
		})

		Convey("Uses a unique command id as the key for the command set", func() {
			cmd1, _ := c.StartCommand([]string{"echo", "banana1"})
			cmd2, _ := c.StartCommand([]string{"echo", "banana2"})
			So(c.commands[cmd1.CommandId], ShouldEqual, cmd1)
			So(c.commands[cmd2.CommandId], ShouldEqual, cmd2)
			So(cmd1.CommandId, ShouldNotResemble, cmd2.CommandId)
		})
	})
}

func TestContainerExecController_Close(t *testing.T) {
	Convey("Close()", t, func() {
		c := NewController()

		Convey("Should close all commands", func() {
			cmd1, _ := c.StartCommand([]string{"echo", "banana1"})
			cmd1_id := cmd1.CommandId

			cmd2, _ := c.StartCommand([]string{"echo", "banana2"})
			cmd2_id := cmd2.CommandId

			So(c.commands, ShouldContainKey, cmd1_id)
			So(c.commands, ShouldContainKey, cmd2_id)
			c.Close()
			So(c.commands, ShouldNotContainKey, cmd1_id)
			So(c.commands, ShouldNotContainKey, cmd2_id)
		})
	})
}
