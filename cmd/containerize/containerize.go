package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
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

	// Try to start the command
	cmd, err := controller.StartCommand(os.Args[1:])
	if err != nil {
		log.Fatalln("Error:", err)
	}

	// Timeout tailing after a while
	timeout_ctx, timeout_cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	fmt.Println(strings.Repeat("-", 80))
	for {
		read_bytes, err := stream.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("Error reading from the command stream:", err)
		}
		os.Stdout.Write(buffer[:read_bytes])
	}
	fmt.Println(strings.Repeat("-", 80))

	// Stop the command, wait for it to stop and exit status to become available
	cmd.Close()
	exit_code, _ := cmd.ResultCode()
	exit_status, _ := cmd.ResultDescription()

	fmt.Printf("Exit code: %d\n", exit_code)
	fmt.Printf("Status: %s\n", exit_status)
}
