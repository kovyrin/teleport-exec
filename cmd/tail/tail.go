package main

import (
	"context"
	"io"
	"log"
	"os"
	"teleport-exec/file_stream"
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
	timeout_ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Print a message when the timeout is reached
	go func() {
		<-timeout_ctx.Done()
		log.Println("Timeout reached")
	}()

	// Start a new stream from the file
	stream, err := file_stream.NewFileStream(timeout_ctx, file_name)
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
			log.Fatalln("Error while reading the stream:", err)
			log.Println("The stream has been stopped")
			break
		}
		os.Stdout.Write(buffer[:read_bytes])
	}
}
