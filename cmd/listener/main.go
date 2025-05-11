// Package main is the entry point for listener.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/nizarmah/jarvis/internal/env"
	"github.com/nizarmah/jarvis/internal/executor"
	"github.com/nizarmah/jarvis/internal/ffmpeg"
	"github.com/nizarmah/jarvis/internal/ollama"
	"github.com/nizarmah/jarvis/internal/whisper"
)

var (
	wakeUpWord = "jarvis"
)

var (
	promptTemplate = fmt.Sprintf(
		("You are Jarvis, a chat assistant. " +
			"You will receive messages, often auto-corrected. " +
			"Try to understand the user's intent and respond with a valid command. " +
			"If you cannot understand the user's intent, respond with: do_nothing " +
			"The valid commands are: %s " +
			"The user's message is: %%q " +
			"Command: "),
		strings.Join(executor.Commands, ", "),
	)
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

	// Initialize the executor client.
	executor, err := executor.NewClient(executor.ClientConfig{
		Address: e.ExecutorAddress,
		Debug:   e.ExecutorDebug,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the ollama client.
	interpreter := ollama.NewClient(ollama.ClientConfig{
		Debug:   e.OllamaDebug,
		Model:   e.OllamaModel,
		Timeout: e.OllamaTimeout,
		URL:     e.OllamaURL,
	})

	// Initialize the whisper client.
	transcriber, err := whisper.NewClient(ctx, whisper.ClientConfig{
		Debug:     e.TranscriberDebug,
		OutputDir: e.TranscriberOutputDir,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the recorder.
	recorder, err := ffmpeg.NewRecorder(ffmpeg.RecorderConfig{
		Debug:     e.RecorderDebug,
		OutputDir: e.RecorderOutputDir,
		OS:        runtime.GOOS,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the combiner.
	combiner, err := ffmpeg.NewCombiner(ffmpeg.CombinerConfig{
		Debug:      e.CombinerDebug,
		InputDir:   e.RecorderOutputDir,
		OutputDir:  e.CombinerOutputDir,
		OnCombined: createAudioProcessor(e, transcriber, interpreter, executor),
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
func createAudioProcessor(
	e *env.Env,
	transcriber *whisper.Client,
	interpreter *ollama.Client,
	executor *executor.Client,
) ffmpeg.OnCombinedFunc {
	return func(ctx context.Context, filePath string) error {
		// Transcribe the audio file.
		transcript, err := transcribeAudio(ctx, transcriber, filePath)
		if err != nil {
			return fmt.Errorf("failed to transcribe audio: %w", err)
		}

		// Ignore empty transcripts.
		if transcript == "" {
			return nil
		}

		if e.AudioProcessorDebug {
			log.Println(fmt.Sprintf("transcript: %s", transcript))
		}

		// Check if the transcript has the wake up word.
		if !hasWakeUpWord(transcript) {
			return nil
		}

		// Extract the command from the transcript.
		cmd, err := interpretCommand(ctx, interpreter, transcript)
		if err != nil {
			return fmt.Errorf("failed to extract command: %w", err)
		}

		// If the command is empty, do nothing.
		if cmd == "" {
			return nil
		}

		if e.AudioProcessorDebug {
			log.Println(fmt.Sprintf("command: %s", cmd))
		}

		// Execute the command.
		if err := executor.SendCommand(ctx, cmd); err != nil {
			return fmt.Errorf("failed to send command to executor: %w", err)
		}

		return nil
	}
}

// TranscribeAudio transcribes the audio file.
func transcribeAudio(
	ctx context.Context,
	transcriber *whisper.Client,
	filePath string,
) (string, error) {
	// Transcribe the audio file.
	untrimmed, err := transcriber.Transcribe(ctx, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Clean up the audio file.
	if err := os.Remove(filePath); err != nil {
		return "", fmt.Errorf("failed to cleanup audio file: %w", err)
	}

	// Trim the transcript.
	trimmed := strings.TrimSpace(untrimmed)

	// Cleanup punctuation.
	cleaned := strings.ReplaceAll(trimmed, ".", "")

	return strings.ToLower(cleaned), nil
}

// HasWakeUpWord checks if the wake up word is in the transcript.
func hasWakeUpWord(transcript string) bool {
	return strings.Contains(transcript, wakeUpWord)
}

// InterpretCommand interprets the command from the transcript.
func interpretCommand(
	ctx context.Context,
	interpreter *ollama.Client,
	transcript string,
) (string, error) {
	// Build a prompt to instruct LLM.
	prompt := fmt.Sprintf(promptTemplate, transcript)

	// Prompt the LLM.
	response, err := interpreter.Prompt(ctx, prompt)
	if err != nil {
		// Ignore the error if the context was cancelled.
		if errors.Is(err, context.DeadlineExceeded) {
			return "", nil
		}

		return "", fmt.Errorf("failed to prompt LLM: %w", err)
	}

	// Search for the command in the response.
	for _, command := range executor.Commands {
		if strings.Contains(response, command) {
			return command, nil
		}
	}

	return "", nil
}
