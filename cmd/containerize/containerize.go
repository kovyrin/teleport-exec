package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"teleport-exec/container_exec"
	"time"
)

func main() {
	controller := container_exec.NewController()
	defer controller.Close()

	if len(os.Args) < 2 {
		fmt.Println("Usage: containerize.go command args...")
		os.Exit(1)
	}
	cmd := controller.StartCommand(os.Args[1:])

	// Timeout tailing after a while
	timeout_ctx, timeout_cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer timeout_cancel()

	// Stop the stream when cancelled via Ctrl+C
	interrupt_ctx, interrupt_cancel := signal.NotifyContext(timeout_ctx, os.Interrupt)
	defer interrupt_cancel()

	// Start tailing the command output
	stream, err := cmd.NewLogStream(interrupt_ctx)
	if err != nil {
		log.Fatalln("Failed to start a log stream:", err)
	}
	defer stream.Close()

	// Stream content until the stream is terminated
	buffer := make([]byte, 100)
	for {
		read_bytes, err := stream.Read(buffer)
		if err == io.EOF {
			log.Println("The stream has been stopped")
			break
		}
		if err != nil {
			log.Fatalln("Error reading from the command stream:", err)
		}
		os.Stdout.Write(buffer[:read_bytes])
	}
}
