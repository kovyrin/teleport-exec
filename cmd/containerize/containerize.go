package main

import (
	"context"
	"fmt"
	"log"
	"os"
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

	// Timeout after a while
	timeout_ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Start tailing the command output
	stream := cmd.NewLogStream(timeout_ctx)
	defer stream.Close()

	// Stream content until the stream is terminated
	for {
		content := stream.MoreBytes()
		if content == nil {
			log.Println("The stream has been stopped")
			break
		}
		fmt.Print(string(content))
	}
}
