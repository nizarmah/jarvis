// Package main is the entry point for listener.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nizarmah/jarvis/internal/ffmpeg"
	"github.com/nizarmah/jarvis/internal/ollama"
	"github.com/nizarmah/jarvis/internal/whisper"
)

const (
	executorPort = "4242"
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

	log.Println("Available commands:")
	log.Println("\t- pause the YouTube video")
	log.Println("\t- play the YouTube video")

	log.Println("Press Ctrl+C to stop.")

	// Wait for Ctrl+C or kill from context.
	<-ctx.Done()
	log.Println("Context cancelled — exiting.")
}

// CreateAudioProcessor creates a function that transcribes the audio file and extracts the command.
func createAudioProcessor(transcriber *whisper.Transcriber, ollama *ollama.Client) ffmpeg.OnCombinedFunc {
	return func(ctx context.Context, filePath string) error {
		// Transcribe the audio file.
		transcript, err := transcribeAudio(ctx, transcriber, filePath)
		if err != nil {
			return fmt.Errorf("failed to transcribe audio: %w", err)
		}

		// Build a prompt to instruct LLM.
		prompt := fmt.Sprintf(promptTemplate, transcript)

		// Prompt the LLM.
		cmd, err := ollama.Prompt(ctx, prompt)
		if err != nil {
			return fmt.Errorf("failed to prompt LLM: %w", err)
		}

		// Execute the command.
		if err := executeCommand(ctx, cmd); err != nil {
			return fmt.Errorf("failed to execute command: %w", err)
		}

		if ollamaDebug {
			log.Println(fmt.Sprintf("transcript: %s, command: %s", transcript, cmd))
		}

		return nil
	}
}

// TranscribeAudio transcribes the audio file.
func transcribeAudio(ctx context.Context, transcriber *whisper.Transcriber, filePath string) (string, error) {
	// Transcribe the audio file.
	transcript, err := transcriber.Transcribe(ctx, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Clean up the audio file.
	if err := os.Remove(filePath); err != nil {
		return "", fmt.Errorf("failed to cleanup audio file: %w", err)
	}

	return transcript, nil
}

// ExecuteCommand sends the command to the executor.
func executeCommand(_ context.Context, command string) error {
	// If the command is "do_nothing", do nothing.
	if command == "do_nothing" {
		return nil
	}

	// Connect to the executor.
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", executorPort))
	if err != nil {
		return fmt.Errorf("failed to connect to executor: %w", err)
	}
	defer conn.Close()

	// Send the command to the executor.
	_, err = conn.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("failed to send command to executor: %w", err)
	}

	return nil
}
