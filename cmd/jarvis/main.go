// Package main is the entry point for jarvis.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nizarmah/jarvis/internal/ffmpeg"
	"github.com/nizarmah/jarvis/internal/whisper"
)

const (
	transcriberDebug = false
	recorderDebug    = false
	combinerDebug    = false
)

const (
	ffmpegPlatform    = ffmpeg.PlatformMac
	ffmpegChunksDir   = "artifacts/audio/chunks"
	ffmpegCombinedDir = "artifacts/audio/combined"
)

const (
	whisperOutputDir = "artifacts/audio/transcripts"
)

func main() {
	// Context.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
	)
	defer cancel()

	// Initialize the transcriber.
	transcriber, err := whisper.NewTranscriber(ctx, whisper.TranscriberConfig{
		Debug:     transcriberDebug,
		OutputDir: whisperOutputDir,
	})
	if err != nil {
		log.Fatal(err)
	}

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
		Debug:     combinerDebug,
		InputDir:  ffmpegChunksDir,
		OutputDir: ffmpegCombinedDir,
		OnCombined: func(ctx context.Context, filePath string) error {
			return processAudio(ctx, transcriber, filePath)
		},
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

func processAudio(ctx context.Context, transcriber *whisper.Transcriber, filePath string) error {
	// Transcribe the audio file.
	transcription, err := transcriber.Transcribe(ctx, filePath)
	if err != nil {
		return fmt.Errorf("process audio failed during transcription: %w", err)
	}

	// Clean up the audio file.
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("process audio failed during cleanup: %w", err)
	}

	log.Println(fmt.Sprintf("processed audio: %s", transcription))

	return nil
}
