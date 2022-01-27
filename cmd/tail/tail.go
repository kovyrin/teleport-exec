package main

import (
	"context"
	"fmt"
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

	// Start a new stream from the file
	stream := file_stream.NewFileStream(file_name, timeout_ctx)
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
