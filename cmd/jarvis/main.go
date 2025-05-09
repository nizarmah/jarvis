// Package main is the entry point for jarvis.
package main

import (
	"log"

	"github.com/nizarmah/jarvis/internal/ffmpeg"
)

const (
	ffmpegPlatform  = ffmpeg.PlatformMac
	ffmpegOutputDir = "artifacts/audio_chunks"
)

func main() {
	rec, err := ffmpeg.New(ffmpegPlatform, ffmpegOutputDir)
	if err != nil {
		log.Fatal(err)
	}

	if err := rec.Start(); err != nil {
		log.Fatal(err)
	}
	defer rec.Stop()

	// Block forever or until interrupt
	select {}
}
