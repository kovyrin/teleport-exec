package filestream

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileStream_MoreBytes(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()
	buffer := make([]byte, 100)

	Convey("filestream.MoreBytes()", t, func() {
		file_name := "/tmp/tail_test.log"
		f, _ := os.Create(file_name)
		init_content := "hello, world!\n"
		f.WriteString(init_content)

		tail_enabled := true
		stream, _ := New(ctx, file_name, tail_enabled)

		Convey("Should return a full buffer when possible", func() {
			f.WriteString(strings.Repeat("x", len(buffer)))
			read_bytes, err := stream.Read(buffer)
			So(read_bytes, ShouldEqual, len(buffer))
			So(err, ShouldBeNil)
		})

		Convey("Should return available bytes when reaches the end", func() {
			read_bytes, err := stream.Read(buffer)
			So(err, ShouldBeNil)
			So(read_bytes, ShouldEqual, len(init_content))
			So(string(buffer[:read_bytes]), ShouldResemble, init_content)
		})

		Convey("Should return more data when more data is added to the file", func() {
			// Read the last 5 bytes and reach the end of the stream
			stream.reader.Seek(-5, io.SeekEnd)
			read_bytes, err := stream.Read(buffer)
			So(read_bytes, ShouldEqual, 5)
			So(err, ShouldBeNil)

			// Add more data
			more_content := "banana"
			f.WriteString(more_content)

			// Should be able to consume it now
			read_bytes, err = stream.Read(buffer)
			So(read_bytes, ShouldEqual, len(more_content))
			So(err, ShouldBeNil)
			So(string(buffer[:read_bytes]), ShouldResemble, more_content)
		})

		Convey("When running in a tail mode", func() {
			Convey("Should block and wait for more data when reaches the end", func() {
				// read all data
				stream.Read(buffer)

				// Write some content a second later
				go func() {
					time.Sleep(time.Second)
					f.WriteString("yo")
				}()

				read_bytes, err := stream.Read(buffer)
				So(read_bytes, ShouldEqual, 2)
				So(err, ShouldBeNil)
				So(string(buffer[:read_bytes]), ShouldResemble, "yo")
			})
		})

		Convey("When running in a non-tail mode", func() {
			stream, _ := New(ctx, file_name, false)

			Convey("Should return an EOF when reaches the end", func() {
				// read all data
				stream.Read(buffer)

				// Write some content a second later
				wg.Add(1)
				go func() {
					defer wg.Done()
					time.Sleep(time.Second)
					f.WriteString("yo")
				}()

				read_bytes, err := stream.Read(buffer)
				So(read_bytes, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)

				// Wait for the async write operation to complete
				wg.Wait()
			})
		})

		Convey("When tailing is disabled while reading", func() {
			stream.DisableTail()

			Convey("Should return an EOF when reaches the end", func() {
				// read all data
				stream.Read(buffer)

				// Write some content a second later
				wg.Add(1)
				go func() {
					defer wg.Done()
					time.Sleep(time.Second)
					f.WriteString("yo")
				}()

				read_bytes, err := stream.Read(buffer)
				So(read_bytes, ShouldEqual, 0)
				So(err, ShouldEqual, io.EOF)

				// Wait for the async write operation to complete
				wg.Wait()
			})
		})

		Convey("Should return an EOF when cancelled via the context", func() {
			// Create a stream that times out after a second
			timeout_ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			stream, _ := New(timeout_ctx, file_name, tail_enabled)

			// Read all data
			stream.Read(buffer)

			// Try to read more and block
			_, err := stream.Read(buffer)

			// Should return an EOF after being cancelled
			So(err, ShouldEqual, io.EOF)
		})

		Convey("Should return an EOF when stopped by closing", func() {
			// Read all data
			stream.Read(buffer)

			// Stop reading by closing the stream in a second
			go func() {
				time.Sleep(time.Second)
				stream.Close()
			}()

			// Try to read more and block
			_, err := stream.Read(buffer)

			// Should return an EOF after being stopped
			So(err, ShouldEqual, io.EOF)
		})

		Convey("Should return an EOF without reading any data if called after closing", func() {
			stream.Close()
			read_bytes, err := stream.Read(buffer)
			So(err, ShouldEqual, io.EOF)
			So(read_bytes, ShouldEqual, 0)
		})

		Reset(func() {
			stream.Close()
			os.Remove(file_name)
		})
	})

	Convey("filestream.Close()", t, func() {
		stream, _ := New(ctx, "/etc/hosts", true)

		Convey("Should not blow up when called twice and return no errors", func() {
			So(stream.Close(), ShouldBeNil)
			So(stream.Close(), ShouldBeNil)
		})
	})
}
