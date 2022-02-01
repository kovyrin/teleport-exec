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
		log.Println("Need an argument!")
		os.Exit(1)
	}
	fileName := os.Args[1]
	log.Println("Tailing file:", fileName)

	// Timeout tailing after a while
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer timeoutCancel()

	// Stop the stream when cancelled via Ctrl+C
	interruptCtx, interruptCancel := signal.NotifyContext(timeoutCtx, os.Interrupt)
	defer interruptCancel()

	// Start a new stream from the file in a tail mode
	stream, err := filestream.New(interruptCtx, fileName, true)
	if err != nil {
		log.Println("Failed to initialize a file stream:", err)
		os.Exit(1)
	}

	// Stream content until the stream is terminated
	if _, err := io.Copy(os.Stdout, stream); err != nil {
		log.Println("Failed to stream content:", err)
		os.Exit(1)
	}

	// Done here
	if err := stream.Close(); err != nil {
		log.Println("Failed to close the stream:", err)
	}
}
