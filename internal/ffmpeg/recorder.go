// Package ffmpeg provides a rolling recorder for audio chunks.
package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// RecorderConfig is the configuration for the recorder.
type RecorderConfig struct {
	// Debug enables logging during ffmpeg command execution.
	Debug bool
	// Platform is the platform for the recorder.
	Platform platform
	// OutputDir is the directory for the audio chunks.
	OutputDir string
}

// Recorder is a rolling recorder for audio chunks.
type Recorder struct {
	debug     bool
	outputDir string
	platform  platform
}

// NewRecorder initializes the recorder.
func NewRecorder(cfg RecorderConfig) (*Recorder, error) {
	switch cfg.Platform {
	case PlatformMac, PlatformLinux, PlatformWindows:
		return &Recorder{
			debug:     cfg.Debug,
			outputDir: cfg.OutputDir,
			platform:  cfg.Platform,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported platform: %q", cfg.Platform)
	}
}

// Start starts a rolling recorder using ffmpeg.
func (r *Recorder) Start(ctx context.Context) error {
	if err := createDirIfNotExists(r.outputDir); err != nil {
		return fmt.Errorf("failed to start recorder: %w", err)
	}

	args, err := buildArgs(r.platform, r.outputDir)
	if err != nil {
		return fmt.Errorf("failed to build args: %w", err)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if r.debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Start()
}

// buildArgs builds the arguments for the ffmpeg command.
func buildArgs(p platform, outputDir string) ([]string, error) {
	inputArgs, err := inputDeviceArgs(p)
	if err != nil {
		return nil, err
	}

	commonArgs := []string{
		// Use 16-bit signed little-endian PCM audio (raw, uncompressed)
		"-acodec", "pcm_s16le",
		// Sample rate: 16 kHz (recommended for Whisper)
		"-ar", "16000",
		// Mono audio (1 channel)
		"-ac", "1",
		// Enable segmenting the output into separate files
		"-f", "segment",
		// Each segment/file is 2 seconds long
		"-segment_time", fmt.Sprintf("%d", chunkDuration),
		// Use the WAV container format for each file
		"-segment_format", chunkFormat,
		// Only keep the last 6 files (chunk_0.wav to chunk_5.wav)
		"-segment_wrap", fmt.Sprintf("%d", chunkWrap),
		// Restart timestamps at 0 for each segment (avoids time drift)
		"-reset_timestamps", "1",
		// Output pattern (chunk_0.wav, chunk_1.wav, ...)
		fmt.Sprintf("%s/%s", outputDir, chunkPattern),
	}

	return append(inputArgs, commonArgs...), nil
}

// inputDeviceArgs returns the input device arguments for the platform.
func inputDeviceArgs(p platform) ([]string, error) {
	switch p {
	case PlatformMac:
		// AVFoundation on macOS; ":0" = default microphone.
		return []string{"-f", "avfoundation", "-i", ":0"}, nil

	case PlatformLinux:
		// ALSA on Linux; "default" = system default microphone.
		return []string{"-f", "alsa", "-i", "default"}, nil

	case PlatformWindows:
		// DirectShow on Windows; adjust "Microphone" if needed.
		return []string{"-f", "dshow", "-i", "audio=Microphone"}, nil

	default:
		return nil, fmt.Errorf("unsupported platform: %q", p)
	}
}
