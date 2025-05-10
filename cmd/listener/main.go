// Package main is the entry point for listener.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nizarmah/jarvis/internal/ffmpeg"
	"github.com/nizarmah/jarvis/internal/ollama"
	"github.com/nizarmah/jarvis/internal/whisper"
)

const (
	recorderDebug     = false
	combinerDebug     = false
	ffmpegPlatform    = ffmpeg.PlatformMac
	ffmpegChunksDir   = "artifacts/audio/chunks"
	ffmpegCombinedDir = "artifacts/audio/combined"
)

const (
	transcriberDebug = false
	whisperOutputDir = "artifacts/audio/transcripts"
)

const (
	ollamaDebug = false
	// Llama3 is good at following instructions and extracting commands.
	// Also, it runs well with Whisper and FFmpeg on my 8GB laptop.
	ollamaModel = "llama3"
	// Timeout after 5 seconds to avoid hallucinations.
	ollamaTimeout = 5 * time.Second
	ollamaURL     = "http://localhost:11434"
)

const (
	promptTemplate = `You are Jarvis, a voice assistant that listens to noisy voice transcripts.
	Your job is to detect whether the user is calling you, and whether they are giving you a valid command.

	The transcript may contain errors or mispronunciations.
	Users might call you "Jarvis", "Jarmis", "Jarvez", "Germous", or similar variations.

	Only respond with a valid command if BOTH of the following are true:
	1. The assistant (you) was clearly addressed — even with a mispronounced name.
	2. One of the following commands was clearly intended:
		- pause_video
		- play_video
		- skip_ad

	If both are true, respond with the correct command (exactly as written above).
	If at least one is false, respond with:
		- do_nothing

	Respond with ONE WORD only.

	Transcript: %q

	Command:`
)

func main() {
	// Context.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
	)
	defer cancel()

	// Initialize the ollama client.
	ollama := ollama.NewClient(ollama.ClientConfig{
		Debug:   ollamaDebug,
		Model:   ollamaModel,
		Timeout: ollamaTimeout,
		URL:     ollamaURL,
	})

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
		Debug:      combinerDebug,
		InputDir:   ffmpegChunksDir,
		OutputDir:  ffmpegCombinedDir,
		OnCombined: createAudioProcessor(transcriber, ollama),
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

func createAudioProcessor(transcriber *whisper.Transcriber, ollama *ollama.Client) ffmpeg.OnCombinedFunc {
	return func(ctx context.Context, filePath string) error {
		// Transcribe the audio file.
		transcript, err := transcribeAudio(ctx, transcriber, filePath)
		if err != nil {
			return err
		}

		// Build a prompt to instruct LLM.
		prompt := fmt.Sprintf(promptTemplate, transcript)

		log.Println(fmt.Sprintf("prompt: %s", prompt))

		// Prompt the LLM.
		cmd, err := ollama.Prompt(ctx, prompt)
		if err != nil {
			return err
		}

		log.Println(fmt.Sprintf("command: %s", cmd))

		if ollamaDebug {
			log.Println(fmt.Sprintf("transcript: %s, command: %s", transcript, cmd))
		}

		return nil
	}
}

func transcribeAudio(ctx context.Context, transcriber *whisper.Transcriber, filePath string) (string, error) {
	// Transcribe the audio file.
	transcript, err := transcriber.Transcribe(ctx, filePath)
	if err != nil {
		return "", fmt.Errorf("process audio failed during transcription: %w", err)
	}

	// Clean up the audio file.
	if err := os.Remove(filePath); err != nil {
		return "", fmt.Errorf("process audio failed during cleanup: %w", err)
	}

	return transcript, nil
}
