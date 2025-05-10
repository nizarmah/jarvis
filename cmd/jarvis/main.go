// Package main is the entry point for jarvis.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/nizarmah/jarvis/internal/ffmpeg"
)

const (
	ffmpegDebug     = false
	ffmpegPlatform  = ffmpeg.PlatformMac
	ffmpegOutputDir = "artifacts/audio_chunks"
)

func main() {
	// Context.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
	)
	defer cancel()

	// Initialize the recorder.
	rec, err := ffmpeg.NewRecorder(ffmpeg.Options{
		Debug:     ffmpegDebug,
		Platform:  ffmpegPlatform,
		OutputDir: ffmpegOutputDir,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Start the recorder in context so it is auto-stopped.
	if err := rec.Start(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Jarvis is listening...")
	log.Println("To use Jarvis, say 'Jarvis, <command>!'")

	log.Println("No available commands yet.")

	log.Println("Press Ctrl+C to stop.")

	// Wait for Ctrl+C or kill from context.
	<-ctx.Done()
	log.Println("Context cancelled — exiting.")
}
