package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"teleport-exec/cgroups"
	"teleport-exec/containerize"
	"time"
)

func main() {
	controller := containerize.NewController()
	defer controller.Close()

	if len(os.Args) < 2 {
		fmt.Println("Usage: containerize.go command args...")
		os.Exit(1)
	}

	// Prepare cgroups support
	cgroups.Setup()
	defer cgroups.TearDown()

	// Try to start the command
	cmd, err := controller.StartCommand(os.Args[1:])
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}

	// Timeout tailing after a while
	timeout_ctx, timeout_cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer timeout_cancel()

	// Stop the stream when cancelled via Ctrl+C
	interrupt_ctx, interrupt_cancel := signal.NotifyContext(timeout_ctx, os.Interrupt)
	defer interrupt_cancel()

	// Start tailing the command output
	stream, err := cmd.NewLogStream(interrupt_ctx)
	if err != nil {
		log.Println("Failed to start a log stream:", err)
		os.Exit(1)
	}
	defer stream.Close()

	// Stream content until the stream is terminated
	fmt.Println(strings.Repeat("-", 80))
	io.Copy(os.Stdout, stream)
	fmt.Println(strings.Repeat("-", 80))

	// Stop the command, wait for it to stop and exit status to become available
	if err := cmd.Close(); err != nil {
		fmt.Println("Failed to close the command: %w", err)
		os.Exit(1)
	}
	exit_code, _ := cmd.ResultCode()
	exit_status, _ := cmd.ResultDescription()

	fmt.Printf("Exit code: %d\n", exit_code)
	fmt.Printf("Status: %s\n", exit_status)
}
