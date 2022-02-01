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
	fileName string
	reader   *os.File
	watcher  *fsnotify.Watcher

	tailEnabled bool      // Controls what happens when we hit the EOF: wait for more (true) or stop (false)
	logComplete chan bool // Used to notify the reader when the log is complete

	mu       sync.RWMutex    // Protects access to the stream data
	done     chan bool       // A channel for letting the reader know that the stream is closed
	isClosed bool            // Set to true when Close() is first called
	ctx      context.Context // Used to abort streaming when a client disconnects, etc
}

// New creates a FileStream instance for a given file
func New(ctx context.Context, fileName string, tail bool) (*FileStream, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file '%s': %w", fileName, err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to start a watcher: %w", err)
	}

	err = watcher.Add(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to add file '%s' to the watcher: %w", fileName, err)
	}

	return &FileStream{
		fileName:    fileName,
		tailEnabled: tail,
		reader:      file,
		watcher:     watcher,
		ctx:         ctx,
		done:        make(chan bool),
		logComplete: make(chan bool),
	}, nil
}

// TailEnabled returns true if the stream is in a tailing mode (blocking at the end of a file, waiting for more content)
func (s *FileStream) TailEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tailEnabled
}

// DisableTail disables tail mode for the stream,
///notifying any waiting readers that no more content will be added to the file
func (s *FileStream) DisableTail() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tailEnabled {
		s.tailEnabled = false
		close(s.logComplete) // Tell the reader to stop waiting for more content
	}
}

// Close stops any active watchers and closes the underlying file stream
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
	_ = s.watcher.Close()
	return s.reader.Close()
}

//-------------------------------------------------------------------------------------------------
// Blocks until there is a write event on the file, or we're asked to stop waiting and/or reading
func (s *FileStream) waitForChanges() {
	for {
		select {
		// Some write occurred, and we should see if we could read something
		case event := <-s.watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				return
			}
		// We were asked to stop tailing the file
		case <-s.logComplete:
			return
		// We were asked to stop reading the file
		case <-s.done:
			return
		// The context has been cancelled and we should stop
		case <-s.ctx.Done():
			return
		}
	}
}

// Returns true if we are done reading the stream or were asked to stop
func (s *FileStream) shouldStop() bool {
	select {
	case <-s.done:
		return true
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}

// Read implements the io.Reader interface.
// Returns the data available in the stream and blocks for more data if the stream is empty.
func (s *FileStream) Read(buffer []byte) (int, error) {
	for {
		if s.shouldStop() {
			return 0, io.EOF
		}
		readBytes, err := s.reader.Read(buffer)

		// How we handle the EOF depends on the current tail mode
		if err == io.EOF {
			if s.TailEnabled() {
				err = nil
			} else {
				_ = s.Close()
			}
		}

		// If there was an error (including an IO in non-tailing mode)
		// or if we have read something, return it to the caller
		if err != nil || readBytes > 0 {
			return readBytes, err
		}

		// Wait until we have more data to read or if we're stopped by context or a Close() call
		s.waitForChanges()
	}
}
