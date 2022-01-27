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

func (l *LogStream) MoreBytes(output []byte) (read_bytes int, eof bool) {
	read_bytes, err := l.reader.Read(output)
	if err == io.EOF {
		return read_bytes, true
	}
	if err != nil {
		log.Fatalf("Failed to read data from file '%s': %v", l.fileName, err)
	}
	return read_bytes, false
}

func (log *LogStream) Close() error {
	return log.reader.Close()
}
