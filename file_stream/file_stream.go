package file_stream

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/fsnotify/fsnotify"
)

type FileStream struct {
	file_name string
	reader    *os.File
	watcher   *fsnotify.Watcher

	// Used to abort streaming when a client disconnects, etc
	ctx context.Context
}

//-------------------------------------------------------------------------------------------------
func New(ctx context.Context, file_name string) (*FileStream, error) {
	file, err := os.Open(file_name)
	if err != nil {
		return nil, fmt.Errorf("failed to open file '%s': %w", file_name, err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to start a watcher: %w", err)
	}

	err = watcher.Add(file_name)
	if err != nil {
		return nil, fmt.Errorf("failed to add file '%s' to the watcher: %w", file_name, err)
	}

	return &FileStream{
		file_name: file_name,
		reader:    file,
		watcher:   watcher,
		ctx:       ctx,
	}, nil
}

//-------------------------------------------------------------------------------------------------
// Closes the fsnotify watcher and the underlying file stream
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
func (s *FileStream) readBlock(buffer []byte) (int, error) {
	read_bytes, err := s.reader.Read(buffer)
	if err == io.EOF || read_bytes == 0 {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("failed to read data from file '%s': %w", s.file_name, err)
	}

	return read_bytes, nil
}

//-------------------------------------------------------------------------------------------------
// Implements the io.Reader interface
// Returns the data available in the stream and blocks for more data if the stream is empty.
func (s *FileStream) Read(buffer []byte) (int, error) {
	for {
		read_bytes, err := s.readBlock(buffer)
		if err != nil || read_bytes > 0 {
			return read_bytes, err
		}

		// If we stopped before any content is available, we need to abort
		if !s.WaitForChanges() {
			return 0, io.EOF
		}
	}
}
