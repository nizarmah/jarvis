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
	// ChunkNum is the number of chunks to record.
	ChunkNum int
	// ChunkSize is the size of the audio chunks in seconds.
	ChunkSize int
	// Debug enables logging during ffmpeg command execution.
	Debug bool
	// OutputDir is the directory for the audio chunks.
	OutputDir string
	// OS is the operating system for the recorder.
	OS string
}

// Recorder is a rolling recorder for audio chunks.
type Recorder struct {
	chunkNum  int
	chunkSize int
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
		chunkNum:  cfg.ChunkNum,
		chunkSize: cfg.ChunkSize,
		debug:     cfg.Debug,
		outputDir: cfg.OutputDir,
		os:        runtime.GOOS,
	}, nil
}

// Start starts a rolling recorder using ffmpeg.
func (r *Recorder) Start(ctx context.Context) error {
	args, err := r.buildRecorderArgs()
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
func (r *Recorder) buildRecorderArgs() ([]string, error) {
	inputArgs, err := inputDeviceArgs(r.os)
	if err != nil {
		return nil, err
	}

	chunkArgs := append(
		chunkFfmpegArgs,
		// Each segment/file is 2 seconds long
		"-segment_time", fmt.Sprintf("%d", r.chunkSize),
		// Only keep the last 6 files (chunk_0.wav to chunk_5.wav)
		"-segment_wrap", fmt.Sprintf("%d", r.chunkNum),
	)

	return buildFfmpegArgs(
		chunkArgs,
		inputArgs,
		[]string{fmt.Sprintf("%s/%s", r.outputDir, chunkPattern)},
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
