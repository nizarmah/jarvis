// Package ffmpeg provides a rolling recorder for audio chunks.
package ffmpeg

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
)

// RecorderConfig is the configuration for the recorder.
type RecorderConfig struct {
	// Debug enables logging during ffmpeg command execution.
	Debug bool
	// OutputDir is the directory for the audio chunks.
	OutputDir string
	// OS is the operating system for the recorder.
	OS string
}

// Recorder is a rolling recorder for audio chunks.
type Recorder struct {
	debug     bool
	outputDir string
	os        string
}

// NewRecorder initializes the recorder.
func NewRecorder(cfg RecorderConfig) (*Recorder, error) {
	if cfg.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	if err := createDirIfNotExists(cfg.OutputDir); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		break
	default:
		return nil, fmt.Errorf("unsupported platform: %q", runtime.GOOS)
	}

	return &Recorder{
		debug:     cfg.Debug,
		outputDir: cfg.OutputDir,
		os:        runtime.GOOS,
	}, nil
}

// Start starts a rolling recorder using ffmpeg.
func (r *Recorder) Start(ctx context.Context) error {
	args, err := buildRecorderArgs(r.os, r.outputDir)
	if err != nil {
		return fmt.Errorf("failed to build args: %w", err)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if r.debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	log.Println("recorder started")

	return cmd.Start()
}

// buildRecorderArgs builds the arguments for the ffmpeg command.
func buildRecorderArgs(os string, outputDir string) ([]string, error) {
	inputArgs, err := inputDeviceArgs(os)
	if err != nil {
		return nil, err
	}

	return buildFfmpegArgs(
		chunkFfmpegArgs,
		inputArgs,
		[]string{fmt.Sprintf("%s/%s", outputDir, chunkPattern)},
	), nil
}

// inputDeviceArgs returns the input device arguments for the platform.
func inputDeviceArgs(os string) ([]string, error) {
	switch os {
	case "darwin":
		// AVFoundation on macOS; ":0" = default microphone.
		return []string{"-f", "avfoundation", "-i", ":0"}, nil

	case "linux":
		// ALSA on Linux; "default" = system default microphone.
		return []string{"-f", "alsa", "-i", "default"}, nil

	case "windows":
		// DirectShow on Windows; adjust "Microphone" if needed.
		return []string{"-f", "dshow", "-i", "audio=Microphone"}, nil

	default:
		return nil, fmt.Errorf("unsupported platform: %q", os)
	}
}
