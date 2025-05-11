package ffmpeg

import (
	"fmt"
	"regexp"
)

// Platform is the platform/os for the ffmpeg.
type platform string

// Platform constants.
const (
	// PlatformMac is the platform for macOS.
	PlatformMac platform = "mac"
	// PlatformLinux is the platform for Linux.
	PlatformLinux platform = "linux"
	// PlatformWindows is the platform for Windows.
	PlatformWindows platform = "windows"
)

// Chunk constants.
const (
	// chunkDuration is the duration of each ffmpeg chunk.
	chunkDuration = 3
	// chunkFormat is the format of each ffmpeg chunk.
	// We use aac because ffmpeg `concat:` only supports mp3 and aac, and in our case aac is better.
	// We can use ffmpeg `filter_complex` but it's slower, and `-f concat` requires a file list `list.txt`.
	chunkFormat = "aac"
	// chunkWrap is the number of chunks to keep.
	chunkWrap = 3
)

// Chunk variables.
var (
	// chunkPattern is the pattern for the ffmpeg chunks.
	// We purposefully use `%%d` to escape the `%` character so the result is `chunk_%d.wav`
	chunkPattern = fmt.Sprintf("chunk_%%d.%s", chunkFormat)
	// chunkRegex is the regex for the ffmpeg chunks.
	chunkRegex = regexp.MustCompile(fmt.Sprintf(`chunk_(\d+)\.%s$`, chunkFormat))
	// chunkFfmpegArgs is the ffmpeg args for the chunks.
	chunkFfmpegArgs = []string{
		// Use 16-bit signed little-endian PCM audio (raw, uncompressed)
		"-acodec", "aac",
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
	}
)

// Combined constants.
const (
	// combinedFormat is the format of the combined ffmpeg file of X chunks.
	// We use wav because Whisper requires wav files to transcribe the audio.
	combinedFormat = "wav"
)

// Combined variables.
var (
	// combinedPattern is the pattern for the ffmpeg combined file of X chunks.
	// We purposefully use `%%d` to escape the `%` character so the result is `combined_%d.wav`
	combinedPattern = fmt.Sprintf("combined_%%d.%s", combinedFormat)
	// combinedFfmpegArgs is the ffmpeg args for the combined file.
	combinedFfmpegArgs = []string{
		// Use 16-bit signed little-endian PCM audio (raw, uncompressed)
		"-acodec", "pcm_s16le",
		// Sample rate: 16 kHz (recommended for Whisper)
		"-ar", "16000",
		// Mono audio (1 channel)
		"-ac", "1",
	}
)
