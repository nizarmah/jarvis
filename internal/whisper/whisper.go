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

// ClientConfig is the configuration for a Client.
type ClientConfig struct {
	// Debug enables logging while transcribing.
	Debug bool
	// Model is the name of the model to use.
	Model string
	// Language is the language to use.
	Language string
	// OutputDir is the directory to output the results to.
	OutputDir string
	// Prompt is the prompt to use for the transcription.
	Prompt string
}

// Client is a client for the whisper service.
type Client struct {
	debug     bool
	model     string
	language  string
	outputDir string
	prompt    string
}

// NewClient creates a new Client.
func NewClient(ctx context.Context, cfg ClientConfig) (*Client, error) {
	// Ensure the whisper service is running.
	if err := ensureWhisperIsRunning(ctx, cfg.Debug); err != nil {
		return nil, err
	}

	if cfg.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if cfg.Language == "" {
		return nil, fmt.Errorf("language is required")
	}

	if cfg.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	if err := createDirIfNotExists(cfg.OutputDir); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	return &Client{
		debug:     cfg.Debug,
		model:     cfg.Model,
		language:  cfg.Language,
		outputDir: cfg.OutputDir,
		prompt:    cfg.Prompt,
	}, nil
}

// Transcribe transcribes the audio file and returns the transcription.
func (t *Client) Transcribe(ctx context.Context, filePath string) (string, error) {
	transcriptionPath, err := t.doTranscription(ctx, filePath)
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
func (t *Client) doTranscription(ctx context.Context, filePath string) (string, error) {
	args := []string{
		// Execute a command inside the service.
		"compose", "exec", "-T", "whisper",
		// Transcribe the audio file.
		"whisper", filePath,
		// Do not carry over the context.
		"--condition_on_previous_text", "False",
		// Specify the model to use.
		"--model", t.model,
		// Specify the language to use.
		"--language", t.language,
		// Output the results in text format.
		"--output_format", "txt",
		// Output the results to the specified directory.
		"--output_dir", t.outputDir,
	}

	if t.prompt != "" {
		args = append(args, "--initial_prompt", t.prompt)
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	if t.debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("transcription command failed: %w", err)
	}

	audioFilename := filepath.Base(filePath)
	transcriptionFilename := fmt.Sprintf("%s.txt", strings.TrimSuffix(audioFilename, filepath.Ext(filePath)))

	transcriptionPath := filepath.Join(t.outputDir, transcriptionFilename)

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

	if !slices.Contains(services, "whisper") {
		return fmt.Errorf("service %q is not running", "whisper")
	}

	return nil
}
