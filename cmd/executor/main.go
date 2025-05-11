// Package main is the entry point for executor.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-vgo/robotgo"
	"github.com/nizarmah/jarvis/internal/server"
)

const (
	commandDebug = false
)

const (
	messageDebug = false
	serverDebug  = false
	serverPort   = "4242"
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

// createMessageHandler creates a message handler.
func createMessageHandler() server.OnMessageFunc {
	return func(ctx context.Context, msg string) error {
		msg = strings.TrimSpace(strings.ToLower(msg))
		if messageDebug {
			log.Printf("received message: %q", msg)
		}

		if err := handleCommand(ctx, msg); err != nil {
			log.Println(fmt.Sprintf("error handling command: %v", err))
		}

		return nil
	}
}

// handleCommand handles the command.
func handleCommand(_ context.Context, msg string) error {
	switch msg {
	case "pause_video":
		return pauseVideo()

	case "play_video":
		return playVideo()

	default:
		log.Printf("unsupported command: %s", msg)
	}

	return nil
}

// pauseVideo pauses the video.
func pauseVideo() error {
	if err := robotgo.KeyTap("k"); err != nil {
		return fmt.Errorf("failed to pause video: %w", err)
	}

	if commandDebug {
		log.Println("paused video")
	}

	return nil
}

// playVideo plays the video.
func playVideo() error {
	if err := robotgo.KeyTap("k"); err != nil {
		return fmt.Errorf("failed to play video: %w", err)
	}

	if commandDebug {
		log.Println("played video")
	}

	return nil
}
