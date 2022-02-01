package containerize

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"teleport-exec/filestream"

	"go.uber.org/multierr"
)

// ProcessLog represents a process output log file for a given command
type ProcessLog struct {
	commandId string
	fd        *os.File
	readers   map[*filestream.FileStream]bool
	mu        sync.Mutex
	isClosed  bool
}

func NewProcessLog(commandId string) (*ProcessLog, error) {
	fd, err := os.CreateTemp("", "command-"+commandId+".out")
	if err != nil {
		return nil, fmt.Errorf("failed to create a log file for command '%s': %w", commandId, err)
	}

	l := &ProcessLog{
		commandId: commandId,
		readers:   make(map[*filestream.FileStream]bool),
		fd:        fd,
	}

	return l, nil
}

// FileName returns the name of the log file used by the process
func (l *ProcessLog) FileName() string {
	return l.fd.Name()
}

// Close closed the log and deletes the log file
func (l *ProcessLog) Close() (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Prevent double-close
	if l.isClosed {
		return nil
	}
	l.isClosed = true

	// Close the log file and delete the file
	err = multierr.Append(err, l.fd.Close())
	err = multierr.Append(err, os.Remove(l.FileName()))

	// Close all active readers
	for reader := range l.readers {
		err = multierr.Append(err, reader.Close())
	}

	return err
}

// LogComplete tells all readers to stop waiting for more content since the log is complete (command is done)
func (l *ProcessLog) LogComplete() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for reader := range l.readers {
		reader.DisableTail()
	}
}

// NewLogStream returns a new file reader for the log.
func (l *ProcessLog) NewLogStream(ctx context.Context, tail bool) (stream *filestream.FileStream, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Store the stream so that we could close it later
	stream, err = filestream.New(ctx, l.FileName(), tail)
	if err == nil {
		l.readers[stream] = true
	}

	return stream, err
}

// CloseLogStream closes a given log stream, removing it from the list of readers
func (l *ProcessLog) CloseLogStream(stream *filestream.FileStream) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.readers[stream]; ok {
		delete(l.readers, stream)
		return stream.Close()
	}
	return errors.New("unknown file reader")
}
