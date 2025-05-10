// Package main is the entry point for executor.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/nizarmah/jarvis/internal/server"
)

const (
	serverDebug = false
	serverPort  = "4242"
)

func main() {
	// Context.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
	)
	defer cancel()

	// Initialize the server.
	server := server.NewTCPServer(server.TCPServerConfig{
		Debug:     serverDebug,
		Port:      serverPort,
		OnMessage: createMessageHandler(),
	})

	// Start the server.
	if err := server.Start(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}

	log.Println("Jarvis is ready to execute commands...")

	log.Println("Press Ctrl+C to stop.")

	// Wait for Ctrl+C or kill from context.
	<-ctx.Done()
	log.Println("Context cancelled — exiting.")
}

func createMessageHandler() server.OnMessageFunc {
	return func(_ context.Context, msg string) error {
		log.Println(fmt.Sprintf("received message: %s", msg))
		return nil
	}
}
