package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"teleport-exec/filestream"
	"time"
)

//-------------------------------------------------------------------------------------------------
func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Need an argument!")
	}
	file_name := os.Args[1]
	log.Println("Tailing file:", file_name)

	// Timeout tailing after a while
	timeout_ctx, timeout_cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer timeout_cancel()

	// Stop the stream when cancelled via Ctrl+C
	interrupt_ctx, interrupt_cancel := signal.NotifyContext(timeout_ctx, os.Interrupt)
	defer interrupt_cancel()

	// Start a new stream from the file in a tail mode
	stream, err := filestream.New(interrupt_ctx, file_name, true)
	if err != nil {
		log.Fatalln("Failed to initialize a file stream:", err)
	}
	defer stream.Close()

	// Stream content until the stream is terminated
	buffer := make([]byte, 100)
	for {
		read_bytes, err := stream.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Error while reading the stream:", err)
			log.Println("The stream has been stopped")
			break
		}
		os.Stdout.Write(buffer[:read_bytes])
	}
}
