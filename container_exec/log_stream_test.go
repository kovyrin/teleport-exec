package container_exec

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLogStream_NextLine(t *testing.T) {
	Convey("NextLine()", t, func() {
		f, _ := os.Open("/etc/hosts")
		stream := NewLogStream(f)

		Convey("Should return the next line when it is not the end of the file", func() {
			line, eof := stream.NextLine()
			So(line, ShouldNotResemble, "")
			So(eof, ShouldBeFalse)
		})

		Convey("Should return an empty string when we're at the end of the file", func() {
			for stream.scanner.Scan() {
				// Reach the end of the stream
			}
			line, eof := stream.NextLine()
			So(line, ShouldResemble, "")
			So(eof, ShouldBeTrue)
		})

		Reset(func() {
			stream.Close()
		})
	})
}
