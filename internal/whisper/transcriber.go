// Package whisper provides a transcriber for audio files.
package whisper

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// TranscriberConfig is the configuration for a Transcriber.
type TranscriberConfig struct {
	// Debug enables logging while transcribing.
	Debug bool
	// OutputDir is the directory to output the results to.
	OutputDir string
}

// Transcriber is a transcriber for audio files.
type Transcriber struct {
	debug     bool
	outputDir string
}

// NewTranscriber creates a new Transcriber.
func NewTranscriber(ctx context.Context, cfg TranscriberConfig) (*Transcriber, error) {
	// Ensure the whisper service is running.
	if err := ensureWhisperIsRunning(ctx, cfg.Debug); err != nil {
		return nil, err
	}

	if cfg.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	if err := createDirIfNotExists(cfg.OutputDir); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	return &Transcriber{
		debug:     cfg.Debug,
		outputDir: cfg.OutputDir,
	}, nil
}

// Transcribe transcribes the audio file and returns the transcription.
func (t *Transcriber) Transcribe(ctx context.Context, filePath string) (string, error) {
	transcriptionPath, err := doTranscription(ctx, filePath, t.outputDir, t.debug)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio file: %w", err)
	}

	// Read the transcription file.
	untrimmedTranscript, err := os.ReadFile(transcriptionPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcription file: %w", err)
	}

	// Clean up the transcription file.
	if err := os.Remove(transcriptionPath); err != nil {
		return "", fmt.Errorf("failed to delete transcription file: %w", err)
	}

	transcript := strings.TrimSpace(string(untrimmedTranscript))

	return transcript, nil
}

// doTranscription transcribes the audio using the whisper service and returns the transcription file path.
func doTranscription(ctx context.Context, filePath, outputDir string, debug bool) (string, error) {
	args := []string{
		// Execute a command inside the service.
		"compose", "exec", "-T", transcriberService,
		// Transcribe the audio file.
		"whisper", filePath,
		// Specify the model to use.
		"--model", transcriberModel,
		// Specify the language to use.
		"--language", transcriberLanguage,
		// Output the results in text format.
		"--output_format", "txt",
		// Output the results to the specified directory.
		"--output_dir", outputDir,
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	if debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("transcription command failed: %w", err)
	}

	audioFilename := filepath.Base(filePath)
	transcriptionFilename := fmt.Sprintf("%s.txt", strings.TrimSuffix(audioFilename, filepath.Ext(filePath)))

	transcriptionPath := filepath.Join(outputDir, transcriptionFilename)

	return transcriptionPath, nil
}

func ensureWhisperIsRunning(ctx context.Context, debug bool) error {
	// Run docker compose ps, return which services are running.
	args := []string{"compose", "ps", "--status=running", "--services"}
	cmd := exec.CommandContext(ctx, "docker", args...)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check running containers: %w", err)
	}

	services := strings.Split(string(output), "\n")

	if debug {
		log.Println(fmt.Sprintf("services running: %s", services))
	}

	if !slices.Contains(services, transcriberService) {
		return fmt.Errorf("service %q is not running", transcriberService)
	}

	return nil
}
