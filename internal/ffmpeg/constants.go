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
	chunkDuration = 2
	// chunkFormat is the format of each ffmpeg chunk.
	chunkFormat = "wav"
	// chunkWrap is the number of chunks to keep.
	chunkWrap = 6
	// mergedFormat is the format of the merged ffmpeg file of X chunks.
	mergedFormat = "wav"
)

// Chunk patterns.
var (
	// chunkPattern is the pattern for the ffmpeg chunks.
	// We purposefully use `%%d` to escape the `%` character so the result is `chunk_%d.wav`
	chunkPattern = fmt.Sprintf("chunk_%%d.%s", chunkFormat)
	// chunkRegex is the regex for the ffmpeg chunks.
	chunkRegex = regexp.MustCompile(fmt.Sprintf(`chunk_(\d+)\.%s$`, chunkFormat))
	// mergedPattern is the pattern for the ffmpeg merged file of X chunks.
	// We purposefully use `%%d` to escape the `%` character so the result is `merged_%d.wav`
	mergedPattern = fmt.Sprintf("merged_%%d.%s", mergedFormat)
)
