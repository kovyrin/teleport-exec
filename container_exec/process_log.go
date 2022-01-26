package container_exec

import (
	"errors"
	"log"
	"os"
)

//-------------------------------------------------------------------------------------------------
type ProcessLog struct {
	command_id string
	fd         *os.File
	readers    map[*LogStream]bool
}

func NewProcessLog(command_id string) *ProcessLog {
	file_name := "/tmp/command-" + command_id + ".out"

	l := &ProcessLog{
		command_id: command_id,
		readers:    make(map[*LogStream]bool),
	}

	var err error
	l.fd, err = os.Create(file_name)
	if err != nil {
		log.Fatalf("Failed to create a log file for command %s (%s): %v", command_id, file_name, err)
	}

	return l
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
func (l *ProcessLog) NewLogStream() *LogStream {
	file, err := os.Open(l.FileName())
	if err != nil {
		log.Fatalf("Failed to open log file '%s': %v", l.FileName(), err)
	}
	stream := NewLogStream(file)
	l.readers[stream] = true
	return stream
}

// Returns a new file reader for the log.
func (l *ProcessLog) CloseLogStream(stream *LogStream) error {
	if _, ok := l.readers[stream]; ok {
		delete(l.readers, stream)
		return stream.Close()
	}
	return errors.New("unknown file reader")
}
