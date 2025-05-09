// Package ffmpeg provides a rolling recorder for audio chunks.
package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
)

type platform string

const (
	// PlatformMac is the platform for macOS.
	PlatformMac platform = "mac"
	// PlatformLinux is the platform for Linux.
	PlatformLinux platform = "linux"
	// PlatformWindows is the platform for Windows.
	PlatformWindows platform = "windows"
)

// Recorder is a rolling recorder for audio chunks.
type Recorder struct {
	platform  platform
	outputDir string

	cmd *exec.Cmd
}

// New initializes the recorder.
func New(p platform, outputDir string) (*Recorder, error) {
	switch p {
	case PlatformMac, PlatformLinux, PlatformWindows:
		return &Recorder{platform: p, outputDir: outputDir}, nil

	default:
		return nil, fmt.Errorf("unsupported platform: %q", p)
	}
}

// Start starts a rolling recorder using ffmpeg.
func (r *Recorder) Start() error {
	if err := createOutputDirIfNotExists(r.outputDir); err != nil {
		return fmt.Errorf("failed to start recorder: %w", err)
	}

	args, err := buildArgs(r.platform, r.outputDir)
	if err != nil {
		return err
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Store the command for future use.
	r.cmd = cmd

	return cmd.Start()
}

// Stop stops the rolling recorder.
func (r *Recorder) Stop() error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	return r.cmd.Process.Kill()
}

func createOutputDirIfNotExists(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	return nil
}

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
		"-segment_time", "2",
		// Use the WAV container format for each file
		"-segment_format", "wav",
		// Only keep the last 6 files (chunk_0.wav to chunk_5.wav)
		"-segment_wrap", "6",
		// Restart timestamps at 0 for each segment (avoids time drift)
		"-reset_timestamps", "1",
		// Output pattern (chunk_0.wav, chunk_1.wav, ...)
		fmt.Sprintf("%s/chunk_%%d.wav", outputDir),
	}

	return append(inputArgs, commonArgs...), nil
}

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
