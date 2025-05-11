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
	"github.com/nizarmah/jarvis/internal/env"
	"github.com/nizarmah/jarvis/internal/server"
)

func main() {
	// Initialize the env.
	e, err := env.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Context.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
	)
	defer cancel()

	// Initialize the server.
	server := server.NewTCPServer(server.TCPServerConfig{
		Address:   e.ExecutorAddress,
		Debug:     e.ExecutorDebug,
		OnMessage: createMessageHandler(e),
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
func createMessageHandler(e *env.Env) server.OnMessageFunc {
	return func(ctx context.Context, msg string) error {
		msg = strings.TrimSpace(strings.ToLower(msg))
		if e.MessageHandlerDebug {
			log.Printf("received message: %q", msg)
		}

		if err := handleCommand(ctx, e, msg); err != nil {
			log.Println(fmt.Sprintf("error handling command: %v", err))
		}

		return nil
	}
}

// handleCommand handles the command.
func handleCommand(_ context.Context, e *env.Env, msg string) error {
	switch msg {
	case "pause_video":
		return pauseVideo(e.CommandDebug)

	case "play_video":
		return playVideo(e.CommandDebug)

	default:
		log.Printf("unsupported command: %s", msg)
	}

	return nil
}

// pauseVideo pauses the video.
func pauseVideo(debug bool) error {
	if err := robotgo.KeyTap("k"); err != nil {
		return fmt.Errorf("failed to pause video: %w", err)
	}

	if debug {
		log.Println("paused video")
	}

	return nil
}

// playVideo plays the video.
func playVideo(debug bool) error {
	if err := robotgo.KeyTap("k"); err != nil {
		return fmt.Errorf("failed to play video: %w", err)
	}

	if debug {
		log.Println("played video")
	}

	return nil
}
