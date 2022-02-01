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

	"github.com/docker/engine/pkg/reexec"
)

//-------------------------------------------------------------------------------------------------
func init() {
	reexec.Register("executeCommand", containerize.ExecuteCommand)
	if reexec.Init() {
		os.Exit(0)
	}
}

//-------------------------------------------------------------------------------------------------
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: containerize.go command args...")
		os.Exit(1)
	}

	// Prepare cgroups support
	if err := cgroups.Setup(); err != nil {
		fmt.Println("Error while setting up cgroups:", err)
		os.Exit(1)
	}

	// Set up the controller for running containers
	controller := containerize.NewController()

	// Clean things up when we're done
	defer func() {
		if err := cgroups.TearDown(); err != nil {
			fmt.Println("Error cleaning up cgroups:", err)
		}
		if err := controller.Close(); err != nil {
			fmt.Println("Error cleaning up controllers:", err)
		}
	}()

	// Try to start the command
	cmd, err := controller.StartCommand(os.Args[1:])
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}

	// Timeout tailing after a while
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer timeoutCancel()

	// Stop the stream when cancelled via Ctrl+C
	interruptCtx, interruptCancel := signal.NotifyContext(timeoutCtx, os.Interrupt)
	defer interruptCancel()

	// Start tailing the command output
	stream, err := cmd.NewLogStream(interruptCtx)
	if err != nil {
		log.Println("Failed to start a log stream:", err)
		os.Exit(1)
	}
	defer stream.Close()

	// Stream content until the stream is terminated
	fmt.Println(strings.Repeat("-", 80))
	_, err = io.Copy(os.Stdout, stream)
	if err != nil {
		fmt.Println("Error streaming command output:", err)
		os.Exit(1)
	}
	fmt.Println(strings.Repeat("-", 80))

	// Stop the command, wait for it to stop and exit status to become available
	if err := cmd.Close(); err != nil {
		fmt.Println("Failed to close the command: %w", err)
		os.Exit(1)
	}
	exitCode, _ := cmd.ResultCode()
	exitStatus, _ := cmd.ResultDescription()

	fmt.Printf("Exit code: %d\n", exitCode)
	fmt.Printf("Status: %s\n", exitStatus)
}
