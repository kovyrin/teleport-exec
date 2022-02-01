package containerize

import (
	"teleport-exec/cgroups"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	cgroups.Setup()
}

//-------------------------------------------------------------------------------------------------
func TestNewCommand(t *testing.T) {
	Convey("When NewCommand is called", t, func() {
		command := []string{"echo", "banana"}
		cmd, _ := NewCommand(command)

		Convey("It should persist the command value", func() {
			So(cmd.Command, ShouldResemble, command[0])
			So(cmd.Args, ShouldResemble, command[1:])
		})

		Convey("It should set up the executor", func() {
			So(cmd.executor, ShouldNotBeNil)
		})
	})
}

//-------------------------------------------------------------------------------------------------
func TestStart(t *testing.T) {
	Convey("When Start is called on a command", t, func() {
		command := []string{"true"}
		cmd, _ := NewCommand(command)

		Convey("It should redirect standard output and error streams", func() {
			cmd.Start()
			So(cmd.executor.Stdout, ShouldNotBeNil)
			So(cmd.executor.Stderr, ShouldNotBeNil)
		})

		Convey("It should actually run the command", func() {
			err := cmd.Start()
			So(err, ShouldBeNil)
			cmd.Wait()
			So(cmd.executor.ProcessState.Exited(), ShouldBeTrue)
			So(cmd.executor.ProcessState.ExitCode(), ShouldEqual, 0)
		})

		Convey("When we fail to run a command", func() {
			command := []string{"banana"}
			cmd, _ = NewCommand(command)
			err := cmd.Start()

			Convey("It should capture the error", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Should still allow a Wait() call", func() {
				cmd.Wait() // Should just exit and not block or blow up
			})
		})

		Convey("When a command fails", func() {
			command := []string{"/bin/sh", "-c", "banana"}
			cmd, _ = NewCommand(command)
			cmd.Start()
			cmd.Wait()

			Convey("It should capture the exit code", func() {
				state := cmd.executor.ProcessState
				So(state.Exited(), ShouldBeTrue)
				So(state.ExitCode(), ShouldEqual, 127)
			})
		})

		Convey("Should return an error when called multiple times", func() {
			So(cmd.Start(), ShouldBeNil)
			So(cmd.Start(), ShouldNotBeNil)
		})

		Convey("Should not allow multiple starts even after the first one is done", func() {
			So(cmd.Start(), ShouldBeNil)
			cmd.Wait()
			So(cmd.Start(), ShouldNotBeNil)
		})
	})
}

//-------------------------------------------------------------------------------------------------
func TestRunning(t *testing.T) {
	Convey("Running()", t, func() {
		cmd, _ := NewCommand([]string{"sleep", "1"})
		err := cmd.Start()
		So(err, ShouldBeNil)

		Convey("When a command has finished", func() {
			cmd.Wait()
			Convey("It should return false", func() {
				So(cmd.Running(), ShouldBeFalse)
			})
		})

		Convey("When a command is still running", func() {
			Convey("It should return true until the command is done", func() {
				So(cmd.Running(), ShouldBeTrue)
				cmd.Wait()
				So(cmd.Running(), ShouldBeFalse)
			})
		})
	})
}

//-------------------------------------------------------------------------------------------------
func TestCommandClose(t *testing.T) {
	Convey("Close()", t, func() {
		Convey("It should kill the process if needed", func() {
			cmd, _ := NewCommand([]string{"sleep", "5"})
			cmd.Start()
			So(cmd.Running(), ShouldBeTrue)
			cmd.Close()
			So(cmd.Running(), ShouldBeFalse)
			So(cmd.executor.ProcessState.Success(), ShouldBeFalse)
		})

		Convey("It should not blow up when called multiple times", func() {
			cmd, _ := NewCommand([]string{"sleep", "5"})
			cmd.Start()
			So(cmd.Close(), ShouldBeNil)
			So(cmd.Close(), ShouldBeNil)
		})
	})
}

//-------------------------------------------------------------------------------------------------
func TestCommand_ResultCode(t *testing.T) {
	Convey("ResultCode()", t, func() {
		Convey("Returns 0 when a command succeeds", func() {
			cmd, _ := NewCommand([]string{"true"})
			cmd.Start()
			cmd.Wait()
			So(cmd.Running(), ShouldBeFalse)
			code, err := cmd.ResultCode()
			So(err, ShouldBeNil)
			So(code, ShouldEqual, 0)
		})

		Convey("Returns the right code when a command fails", func() {
			cmd, _ := NewCommand([]string{"false"})
			cmd.Start()
			cmd.Wait()
			So(cmd.Running(), ShouldBeFalse)
			code, err := cmd.ResultCode()
			So(err, ShouldBeNil)
			So(code, ShouldEqual, 1)
		})

		Convey("Returns an error if the command is still running", func() {
			cmd, _ := NewCommand([]string{"sleep", "1"})
			cmd.Start()
			So(cmd.Running(), ShouldBeTrue)
			_, err := cmd.ResultCode()
			So(err, ShouldNotBeNil)
		})
	})
}
