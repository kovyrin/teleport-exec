package filestream

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type FileStream struct {
	file_name string
	reader    *os.File
	watcher   *fsnotify.Watcher

	mu       sync.Mutex      // Protects access to the stream data
	done     chan bool       // A channel for letting the reader know that the stream is closed
	isClosed bool            // Set to true when Close() is first called
	ctx      context.Context // Used to abort streaming when a client disconnects, etc
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
		done:      make(chan bool),
	}, nil
}

//-------------------------------------------------------------------------------------------------
// Closes the fsnotify watcher and the underlying file stream
func (s *FileStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prevent double-close
	if s.isClosed {
		return nil
	}
	s.isClosed = true

	// Send a "quit" message to the reader goroutine
	close(s.done)

	// Close both the watcher and the file reader,
	// Return possible errors from the file reader since the watcher always returns nil
	s.watcher.Close()
	return s.reader.Close()
}

//-------------------------------------------------------------------------------------------------
func (s *FileStream) waitForChanges() (changed bool) {
	result := make(chan bool)
	go func() {
		for {
			select {
			case event := <-s.watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					result <- true
					return
				}
			case <-s.done:
				result <- false
				return
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
		// Do not attempt to read if we're already closed
		select {
		case <-s.done:
			return 0, io.EOF
		default:
		}

		// Get some data if possible
		read_bytes, err := s.readBlock(buffer)
		if err != nil || read_bytes > 0 {
			return read_bytes, err
		}

		// Wait until we have more data to read or if we're stopped by context or a Close() call
		if !s.waitForChanges() {
			return 0, io.EOF
		}
	}
}
