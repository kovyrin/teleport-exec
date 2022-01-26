package container_exec

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//-------------------------------------------------------------------------------------------------
func TestNewCommand(t *testing.T) {
	Convey("When NewCommand is called", t, func() {
		command := `echo "banana"`
		cmd := NewCommand(command)

		Convey("It should persist the command value", func() {
			So(cmd.Command, ShouldEqual, command)
		})
	})
}

//-------------------------------------------------------------------------------------------------
func TestStart(t *testing.T) {
	Convey("When Start is called on a command", t, func() {
		command := `true`
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

		Convey("It should wrap the actual command with a shell", func() {
			cmd.Start()
			So(cmd.executor.Args, ShouldResemble, []string{"/bin/sh", "-c", command})
		})

		Convey("It should actually run the command", func() {
			cmd.Start()
			cmd.Wait()
			So(cmd.executor.ProcessState.Exited(), ShouldBeTrue)
			So(cmd.executor.ProcessState.ExitCode(), ShouldEqual, 0)
		})

		Convey("When a command fails", func() {
			command = "banana"
			cmd = NewCommand(command)
			cmd.Start()
			cmd.Wait()

			Convey("It should capture the exit code", func() {
				So(cmd.executor.ProcessState.Exited(), ShouldBeTrue)
				So(cmd.executor.ProcessState.ExitCode(), ShouldEqual, 127)
			})
		})
	})
}

//-------------------------------------------------------------------------------------------------
func TestRunning(t *testing.T) {
	Convey("Running()", t, func() {
		command := "true"
		cmd := NewCommand(command)

		Convey("When a command has finished", func() {
			cmd.Start()
			cmd.Wait()
			Convey("It should return false", func() {
				So(cmd.Running(), ShouldBeFalse)
			})
		})

		Convey("When a command is still running", func() {
			command = "sleep 1"
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
			cmd := NewCommand(`sleep 5`)
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
			cmd := NewCommand(`true`)
			cmd.Start()
			cmd.Wait()
			So(cmd.Running(), ShouldBeFalse)
			code := cmd.ResultCode()
			So(code, ShouldNotBeNil)
			So(*code, ShouldEqual, 0)
		})

		Convey("Returns the right code when a command fails", func() {
			cmd := NewCommand(`false`)
			cmd.Start()
			cmd.Wait()
			So(cmd.Running(), ShouldBeFalse)
			code := cmd.ResultCode()
			So(code, ShouldNotBeNil)
			So(*code, ShouldEqual, 1)
		})

		Convey("Returns nil if the command is still running", func() {
			cmd := NewCommand(`sleep 1`)
			cmd.Start()
			So(cmd.Running(), ShouldBeTrue)
			code := cmd.ResultCode()
			So(code, ShouldBeNil)
		})
	})
}
