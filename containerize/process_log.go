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

//-------------------------------------------------------------------------------------------------
type ProcessLog struct {
	command_id  string
	fd          *os.File
	readers     map[*filestream.FileStream]bool
	readersLock sync.Mutex
	isClosed    bool
}

func NewProcessLog(command_id string) (*ProcessLog, error) {
	fd, err := os.CreateTemp("", "command-"+command_id+".out")
	if err != nil {
		return nil, fmt.Errorf("failed to create a log file for command '%s': %w", command_id, err)
	}

	l := &ProcessLog{
		command_id: command_id,
		readers:    make(map[*filestream.FileStream]bool),
		fd:         fd,
	}

	return l, nil
}

//-------------------------------------------------------------------------------------------------
func (l *ProcessLog) FileName() string {
	return l.fd.Name()
}

// Closes the log and deletes the log file
func (l *ProcessLog) Close() (err error) {
	l.readersLock.Lock()
	defer l.readersLock.Unlock()

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

//-------------------------------------------------------------------------------------------------
// Tells all readers to stop waiting for more content since the log is complete (command is done)
func (l *ProcessLog) LogComplete() {
	l.readersLock.Lock()
	defer l.readersLock.Unlock()

	for reader := range l.readers {
		reader.DisableTail()
	}
}

//-------------------------------------------------------------------------------------------------
// Returns a new file reader for the log.
func (l *ProcessLog) NewLogStream(ctx context.Context, tail bool) (stream *filestream.FileStream, err error) {
	l.readersLock.Lock()
	defer l.readersLock.Unlock()

	// Store the stream so that we could close it later
	stream, err = filestream.New(ctx, l.FileName(), tail)
	if err == nil {
		l.readers[stream] = true
	}

	return stream, err
}

// Closes the log stream, including all active readers and deletes the log file
func (l *ProcessLog) CloseLogStream(stream *filestream.FileStream) error {
	l.readersLock.Lock()
	defer l.readersLock.Unlock()

	if _, ok := l.readers[stream]; ok {
		delete(l.readers, stream)
		return stream.Close()
	}
	return errors.New("unknown file reader")
}
