package container_exec

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//-------------------------------------------------------------------------------------------------
func TestNewCommand(t *testing.T) {
	Convey("When NewCommand is called", t, func() {
		command := []string{"echo", "banana"}
		cmd := NewCommand(command)

		Convey("It should persist the command value", func() {
			So(cmd.Command, ShouldResemble, command[0])
			So(cmd.Args, ShouldResemble, command[1:])
		})
	})
}

//-------------------------------------------------------------------------------------------------
func TestStart(t *testing.T) {
	Convey("When Start is called on a command", t, func() {
		command := []string{"true"}
		cmd := NewCommand(command)

		Convey("It should set up the executor", func() {
			cmd.Start()
			So(cmd.executor, ShouldNotBeNil)
		})

		Convey("It should redirect standard output and error streams", func() {
			cmd.Start()
			So(cmd.executor.Stdout, ShouldNotBeNil)
			So(cmd.executor.Stderr, ShouldNotBeNil)
		})

		Convey("It should actually run the command", func() {
			cmd.Start()
			cmd.Wait()
			So(cmd.executor.ProcessState.Exited(), ShouldBeTrue)
			So(cmd.executor.ProcessState.ExitCode(), ShouldEqual, 0)
			So(cmd.Failure(), ShouldBeNil)
		})

		Convey("When we fail to run a command", func() {
			command := []string{"banana"}
			cmd = NewCommand(command)
			cmd.Start()
			cmd.Wait()

			Convey("It should capture the error", func() {
				So(cmd.Failure(), ShouldNotBeNil)
			})
		})

		Convey("When a command fails", func() {
			command := []string{"/bin/sh", "-c", "banana"}
			cmd = NewCommand(command)
			cmd.Start()
			cmd.Wait()

			Convey("It should capture the exit code", func() {
				state := cmd.executor.ProcessState
				So(state.Exited(), ShouldBeTrue)
				So(state.ExitCode(), ShouldEqual, 127)
			})
		})

	})
}

//-------------------------------------------------------------------------------------------------
func TestRunning(t *testing.T) {
	Convey("Running()", t, func() {
		command := []string{"true"}
		cmd := NewCommand(command)

		Convey("When a command has finished", func() {
			cmd.Start()
			cmd.Wait()
			Convey("It should return false", func() {
				So(cmd.Running(), ShouldBeFalse)
			})
		})

		Convey("When a command is still running", func() {
			command = []string{"sleep", "1"}
			cmd.Start()
			Convey("It should return true", func() {
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
			cmd := NewCommand([]string{"sleep", "5"})
			cmd.Start()
			So(cmd.Running(), ShouldBeTrue)
			cmd.Close()
			So(cmd.Running(), ShouldBeFalse)
			So(cmd.executor.ProcessState.Success(), ShouldBeFalse)
		})
	})
}

func TestCommand_ResultCode(t *testing.T) {
	Convey("ResultCode()", t, func() {
		Convey("Returns 0 when a command succeeds", func() {
			cmd := NewCommand([]string{"true"})
			cmd.Start()
			cmd.Wait()
			So(cmd.Running(), ShouldBeFalse)
			code, err := cmd.ResultCode()
			So(err, ShouldBeNil)
			So(code, ShouldEqual, 0)
		})

		Convey("Returns the right code when a command fails", func() {
			cmd := NewCommand([]string{"false"})
			cmd.Start()
			cmd.Wait()
			So(cmd.Running(), ShouldBeFalse)
			code, err := cmd.ResultCode()
			So(err, ShouldBeNil)
			So(code, ShouldEqual, 1)
		})

		Convey("Returns an error if the command is still running", func() {
			cmd := NewCommand([]string{"sleep", "1"})
			cmd.Start()
			So(cmd.Running(), ShouldBeTrue)
			_, err := cmd.ResultCode()
			So(err, ShouldNotBeNil)
		})
	})
}
