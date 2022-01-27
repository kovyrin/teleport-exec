package file_stream

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type FileStream struct {
	file_name string
	reader    *os.File
	watcher   *fsnotify.Watcher

	// Used for get command completion notifications
	Done chan bool

	// Used to abort streaming when a client disconnects, etc
	ctx context.Context
}

//-------------------------------------------------------------------------------------------------
func NewFileStream(file_name string, ctx context.Context) *FileStream {
	file, err := os.Open(file_name)
	if err != nil {
		log.Fatalf("Failed to open file '%s': %v", file_name, err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to start a watcher: %v", err)
	}

	err = watcher.Add(file_name)
	if err != nil {
		log.Fatalf("Failed to add file '%s' to the watcher: %v", file_name, err)
	}

	return &FileStream{
		file_name: file_name,
		reader:    file,
		watcher:   watcher,
		ctx:       ctx,
		Done:      make(chan bool),
	}
}

//-------------------------------------------------------------------------------------------------
func (s *FileStream) Close() error {
	err := s.watcher.Close()
	if err == nil {
		err = s.reader.Close()
	}
	return err
}

//-------------------------------------------------------------------------------------------------
func (s *FileStream) WaitForChanges() (changed bool) {
	result := make(chan bool)
	go func() {
		for {
			select {
			case event := <-s.watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					result <- true
					return
				}
			case <-s.ctx.Done():
				result <- false
				return
			}
		}
	}()

	return <-result
}

//-------------------------------------------------------------------------------------------------
func (s *FileStream) readBlock(buffer []byte) int {
	read_bytes, err := s.reader.Read(buffer)
	if err == io.EOF || read_bytes == 0 {
		return 0
	}

	if err != nil {
		log.Fatalf("Failed to read data from file '%s': %v", s.file_name, err)
	}

	return read_bytes
}

//-------------------------------------------------------------------------------------------------
func (s *FileStream) MoreBytes() []byte {
	buffer := make([]byte, 100)
	for {
		read_bytes := s.readBlock(buffer)
		if read_bytes > 0 {
			return buffer[:read_bytes]
		}

		// If we stopped before any content is available, we need to abort
		if !s.WaitForChanges() {
			return nil
		}
	}
}
