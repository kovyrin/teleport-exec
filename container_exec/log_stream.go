package container_exec

import (
	"io"
	"log"
	"os"
)

type LogStream struct {
	fileName string
	reader   *os.File
}

func NewLogStream(file_name string) *LogStream {
	file, err := os.Open(file_name)
	if err != nil {
		log.Fatalf("Failed to open log file '%s': %v", file_name, err)
	}

	return &LogStream{
		fileName: file_name,
		reader:   file,
	}
}

func (l *LogStream) MoreBytes(output []byte) (read_bytes int) {
	// Get current position
	current_pos, err := l.reader.Seek(0, io.SeekCurrent)
	if err != nil {
		log.Fatalf("Failed to find current file position for '%s': %v", l.fileName, err)
	}

	// Get file size
	stat, err := os.Stat(l.fileName)
	if err != nil {
		log.Fatalf("Failed to check file status for file '%s': %v", l.fileName, err)
	}

	// If there is nothing to read, just return an eof
	if stat.Size() == current_pos {
		return 0
	}

	// Otherwise, read up to max_bytes of data
	read_bytes, err = l.reader.Read(output)
	if err != nil && err != io.EOF {
		log.Fatalf("Failed to read data from file '%s': %v", l.fileName, err)
	}
	return read_bytes
}

func (log *LogStream) Close() error {
	return log.reader.Close()
}
