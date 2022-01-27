package container_exec

import (
	"io"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLogStream_MoreBytes(t *testing.T) {
	Convey("NextLine()", t, func() {
		file_name := "/tmp/tail_test.log"
		f, _ := os.Create(file_name)
		f.WriteString("hello, world!\n")

		stream := NewLogStream(file_name)

		Convey("Should return a full buffer when possible", func() {
			buffer := make([]byte, 6)
			bytes := stream.MoreBytes(buffer)
			So(bytes, ShouldEqual, 6)
			So(string(buffer[:bytes]), ShouldResemble, "hello,")
		})

		Convey("Should return available bytes when reaches the end", func() {
			buffer := make([]byte, 10)
			stream.reader.Seek(-5, io.SeekEnd)
			bytes := stream.MoreBytes(buffer)
			So(bytes, ShouldEqual, 5)
			So(string(buffer[:bytes]), ShouldResemble, "rld!\n")
		})

		Convey("Should return more data when more data is added to the file", func() {
			// Read the last 5 bytes and reach the end of the stream
			buffer := make([]byte, 10)
			stream.reader.Seek(-5, io.SeekEnd)
			bytes := stream.MoreBytes(buffer)
			So(bytes, ShouldEqual, 5)

			// Add more data
			more_data := "banana"
			f.WriteString(more_data)

			// Should be able to consume it now
			buffer = make([]byte, 10)
			bytes = stream.MoreBytes(buffer)
			So(string(buffer[:bytes]), ShouldResemble, more_data)
		})

		Reset(func() {
			stream.Close()
		})
	})
}
