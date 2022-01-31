package containerize

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"teleport-exec/filestream"
)

//-------------------------------------------------------------------------------------------------
type ProcessLog struct {
	command_id  string
	fd          *os.File
	readers     map[*filestream.FileStream]bool
	readersLock sync.Mutex
}

func NewProcessLog(command_id string) (*ProcessLog, error) {
	file_name := "/tmp/command-" + command_id + ".out"

	fd, err := os.Create(file_name)
	if err != nil {
		return nil, fmt.Errorf("failed to create a log file for command %s (%s): %w", command_id, file_name, err)
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

	// Close the log and delete the log file
	err = l.fd.Close()
	if err == nil {
		err = os.Remove(l.FileName())
	}

	// Close all readers
	for reader := range l.readers {
		err = reader.Close()
		if err != nil {
			return err
		}
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
