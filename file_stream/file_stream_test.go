package file_stream

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileStream_MoreBytes(t *testing.T) {
	ctx := context.Background()

	Convey("file_stream.MoreBytes()", t, func() {
		file_name := "/tmp/tail_test.log"
		f, _ := os.Create(file_name)
		init_content := "hello, world!\n"
		f.WriteString(init_content)

		stream := NewFileStream(file_name, ctx)

		Convey("Should return a full buffer when possible", func() {
			f.WriteString("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
			content := stream.MoreBytes()
			So(len(content), ShouldEqual, 100)
		})

		Convey("Should return available bytes when reaches the end", func() {
			content := stream.MoreBytes()
			So(string(content), ShouldResemble, init_content)
		})

		Convey("Should return more data when more data is added to the file", func() {
			// Read the last 5 bytes and reach the end of the stream
			stream.reader.Seek(-5, io.SeekEnd)
			content := stream.MoreBytes()
			So(len(content), ShouldEqual, 5)

			// Add more data
			more_content := "banana"
			f.WriteString(more_content)

			// Should be able to consume it now
			content = stream.MoreBytes()
			So(string(content), ShouldResemble, more_content)
		})

		Convey("Should block and wait for more data when reaches the end", func() {
			// read all data
			stream.MoreBytes()

			// Write some content a second later
			go func() {
				time.Sleep(time.Second)
				f.WriteString("yo")
			}()

			content := stream.MoreBytes()
			So(string(content), ShouldResemble, "yo")
		})

		Convey("Should return nil when cancelled via the context", func() {
			// Create a stream that times out after a second
			timeout_ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			stream := NewFileStream(file_name, timeout_ctx)

			// Read all data
			stream.MoreBytes()

			// Try to read more and block
			content := stream.MoreBytes()

			// Should return nil after being cancelled
			So(content, ShouldBeNil)
		})

		Convey("Should return nil when cancelled via the Done channel", func() {
			// Read all data
			stream.MoreBytes()

			// Wait a second and then send a done signal
			go func() {
				time.Sleep(time.Second)
				stream.Done <- true
			}()

			// Try to read more and block
			content := stream.MoreBytes()

			// Should return nil after being cancelled
			So(content, ShouldBeNil)
		})

		Reset(func() {
			stream.Close()
		})
	})
}
