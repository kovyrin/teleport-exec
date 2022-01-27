package container_exec

import (
	"bufio"
	"os"
	"testing"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

//-------------------------------------------------------------------------------------------------
func TestNewProcessLog(t *testing.T) {
	Convey("When creating a new process log", t, func() {
		id := uuid.NewString()
		pl := NewProcessLog(id)

		Convey("It should create a new log file", func() {
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
		pl := NewProcessLog(id)

		Convey("It should delete the log file", func() {
			_, err := os.Stat(pl.FileName())
			So(err, ShouldBeNil)
			pl.Close()
			_, err = os.Stat(pl.FileName())
			So(err, ShouldNotBeNil)
		})
	})
}

//-------------------------------------------------------------------------------------------------
func TestProcessNewLogStream(t *testing.T) {
	Convey("When NewLogStream() is called", t, func() {
		id := uuid.NewString()
		pl := NewProcessLog(id)
		stream := pl.NewLogStream()

		Convey("It should return a file reader for the log", func() {
			So(stream, ShouldNotBeNil)
		})

		Convey("It should record the reader to be closed later", func() {
			So(pl.readers, ShouldContainKey, stream)
		})

		Convey("The returned reader could be used to consume the content from the file", func() {
			pl.fd.WriteString("banana")
			s := bufio.NewScanner(stream.reader)
			s.Scan()
			So(s.Text(), ShouldResemble, "banana")
		})

		Reset(func() {
			pl.Close()
		})
	})
}

// //-------------------------------------------------------------------------------------------------
func TestProcessCloseLogStream(t *testing.T) {
	Convey("When CloseReader() is called", t, func() {
		id := uuid.NewString()
		pl := NewProcessLog(id)
		stream := pl.NewLogStream()

		Convey("When called with a valid reader", func() {
			res := pl.CloseLogStream(stream)

			Convey("It should return no error", func() {
				So(res, ShouldBeNil)
			})

			Convey("It should delete the reader from the readers list", func() {
				So(pl.readers, ShouldNotContainKey, stream)
			})
		})

		Convey("When called with an unknown reader", func() {
			stream := NewLogStream("/etc/hosts")
			res := pl.CloseLogStream(stream)

			Convey("It should return an error", func() {
				So(res, ShouldNotBeNil)
			})
		})

		Reset(func() {
			pl.Close()
		})
	})
}
