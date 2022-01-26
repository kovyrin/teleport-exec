package container_exec

import (
	"bufio"
	"os"
)

type LogStream struct {
	reader  *os.File
	scanner *bufio.Scanner
}

func NewLogStream(f *os.File) *LogStream {
	return &LogStream{
		reader:  f,
		scanner: bufio.NewScanner(f),
	}
}

func (log *LogStream) NextLine() (log_line string, eof bool) {
	if !log.scanner.Scan() {
		return "", true
	}
	return log.scanner.Text(), false
}

func (log *LogStream) AllLines() (lines []string) {
	for {
		line, eof := log.NextLine()
		if eof {
			break
		}
		lines = append(lines, line)
	}
	return lines
}

func (log *LogStream) Close() error {
	return log.reader.Close()
}
