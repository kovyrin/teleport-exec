package containerize

import (
	"context"
	"os"
	"teleport-exec/filestream"
	"testing"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

//-------------------------------------------------------------------------------------------------
func TestNewProcessLog(t *testing.T) {
	Convey("When creating a new process log", t, func() {
		id := uuid.NewString()
		pl, err := NewProcessLog(id)

		Convey("It should create a new log file", func() {
			So(err, ShouldBeNil)

			// Make sure the log file exists
			_, err := os.Stat(pl.FileName())
			So(err, ShouldBeNil)
		})

		Reset(func() {
			pl.Close()
		})
	})
}

//-------------------------------------------------------------------------------------------------
func TestProcessLogClose(t *testing.T) {
	Convey("When Close() is called", t, func() {
		id := uuid.NewString()
		pl, err := NewProcessLog(id)
		So(err, ShouldBeNil)

		Convey("It should delete the log file", func() {
			// Check that the file exists
			_, stat_err := os.Stat(pl.FileName())
			So(stat_err, ShouldBeNil)

			// Close the stream and delete the file
			So(pl.Close(), ShouldBeNil)

			// Make sure the file is gone
			_, stat_err = os.Stat(pl.FileName())
			So(stat_err, ShouldNotBeNil)
		})

		Convey("Should not blow up when called multiple times", func() {
			So(pl.Close(), ShouldBeNil)
			So(pl.Close(), ShouldBeNil)
		})
	})
}

//-------------------------------------------------------------------------------------------------
func TestProcessNewLogStream(t *testing.T) {
	ctx := context.Background()

	Convey("When NewLogStream() is called", t, func() {
		id := uuid.NewString()
		pl, _ := NewProcessLog(id)

		Convey("For a command that is still running", func() {
			stream, err := pl.NewLogStream(ctx, true)

			Convey("It should return a file reader for the log", func() {
				So(stream, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("It should record the reader to be closed later", func() {
				So(pl.readers, ShouldContainKey, stream)
			})

			Convey("The returned reader could be used to consume the content from the file", func() {
				pl.fd.WriteString("banana")
				buffer := make([]byte, 100)
				read_bytes, err := stream.Read(buffer)
				So(string(buffer[:read_bytes]), ShouldResemble, "banana")
				So(err, ShouldBeNil)
			})
		})

		Convey("For a command that has finished", func() {
			stream, err := pl.NewLogStream(ctx, false)

			Convey("It should return a file reader for the log", func() {
				So(stream, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("It should record the reader to be closed later", func() {
				So(pl.readers, ShouldContainKey, stream)
			})

			Convey("The returned reader could be used to consume the content from the file", func() {
				pl.fd.WriteString("banana")
				buffer := make([]byte, 100)
				read_bytes, err := stream.Read(buffer)
				So(string(buffer[:read_bytes]), ShouldResemble, "banana")
				So(err, ShouldBeNil)
			})
		})

		Reset(func() {
			pl.Close()
		})
	})
}

// //-------------------------------------------------------------------------------------------------
func TestProcessCloseLogStream(t *testing.T) {
	ctx := context.Background()

	Convey("When CloseReader() is called", t, func() {
		id := uuid.NewString()
		pl, _ := NewProcessLog(id)

		Convey("For a command that is still running (tail mode)", func() {
			stream, _ := pl.NewLogStream(ctx, true)
			err := pl.CloseLogStream(stream)

			Convey("When called with a valid reader", func() {
				Convey("It should return no error", func() {
					So(err, ShouldBeNil)
				})

				Convey("It should delete the reader from the readers list", func() {
					So(pl.readers, ShouldNotContainKey, stream)
				})
			})

			Convey("When called with an unknown reader", func() {
				stream, _ := filestream.New(ctx, "/etc/passwd", true)
				err := pl.CloseLogStream(stream)

				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("For a command that has finished", func() {
			stream, _ := pl.NewLogStream(ctx, false)
			err := pl.CloseLogStream(stream)

			Convey("When called with a valid reader", func() {
				Convey("It should return no error", func() {
					So(err, ShouldBeNil)
				})

				Convey("It should delete the reader from the readers list", func() {
					So(pl.readers, ShouldNotContainKey, stream)
				})
			})

			Convey("When called with an unknown reader", func() {
				stream, _ := filestream.New(ctx, "/etc/passwd", true)
				err := pl.CloseLogStream(stream)

				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Reset(func() {
			pl.Close()
		})
	})
}
