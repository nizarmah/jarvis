// Package main is the entry point for jarvis.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/nizarmah/jarvis/internal/ffmpeg"
)

const (
	recorderDebug = false
	combinerDebug = false
)

const (
	ffmpegPlatform    = ffmpeg.PlatformMac
	ffmpegChunksDir   = "artifacts/audio_chunks"
	ffmpegCombinedDir = "artifacts/audio_combined"
)

func main() {
	// Context.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
	)
	defer cancel()

	// Initialize the recorder.
	recorder, err := ffmpeg.NewRecorder(ffmpeg.RecorderConfig{
		Debug:     recorderDebug,
		Platform:  ffmpegPlatform,
		OutputDir: ffmpegChunksDir,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the combiner.
	combiner, err := ffmpeg.NewCombiner(ffmpeg.CombinerConfig{
		Debug:      combinerDebug,
		InputDir:   ffmpegChunksDir,
		OutputDir:  ffmpegCombinedDir,
		OnCombined: onCombined,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Start the combiner in context so it is auto-stopped.
	// We start the combiner first so it can start watching the chunks dir.
	if err := combiner.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// Start the recorder in context so it is auto-stopped.
	if err := recorder.Start(ctx); err != nil {
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

func onCombined(_ context.Context, filename string) error {
	log.Println(fmt.Sprintf("on combined: %s", filename))
	return nil
}
