package ffmpeg

import (
	"fmt"
	"os"
)

// buildFfmpegArgs builds the arguments for the ffmpeg command.
func buildFfmpegArgs(configArgs, inputArgs, outputArgs []string) []string {
	args := configArgs

	// Append the input first, then the args, then the output.
	// This is necessary because ffmpeg expects a certain order of arguments.
	args = append(inputArgs, args...)
	args = append(args, outputArgs...)

	return args
}

// createDirIfNotExists creates the directory if it does not exist.
func createDirIfNotExists(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create dir %s: %w", dir, err)
	}

	return nil
}
