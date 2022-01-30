package container_exec

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
		return nil, fmt.Errorf("Failed to create a log file for command %s (%s): %w", command_id, file_name, err)
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
	// Close all readers
	for reader := range l.readers {
		err = reader.Close()
		if err != nil {
			return err
		}
	}

	// Close the log and delete the log file
	err = l.fd.Close()
	if err == nil {
		err = os.Remove(l.FileName())
	}

	return err
}

//-------------------------------------------------------------------------------------------------
// Returns a new file reader for the log.
func (l *ProcessLog) NewLogStream(ctx context.Context) (*filestream.FileStream, error) {
	l.readersLock.Lock()
	defer l.readersLock.Unlock()

	stream, err := filestream.New(ctx, l.FileName())
	if err != nil {
		return nil, err
	}

	l.readers[stream] = true
	return stream, nil
}

// Returns a new file reader for the log.
func (l *ProcessLog) CloseLogStream(stream *filestream.FileStream) error {
	l.readersLock.Lock()
	defer l.readersLock.Unlock()

	if _, ok := l.readers[stream]; ok {
		delete(l.readers, stream)
		return stream.Close()
	}
	return errors.New("unknown file reader")
}
