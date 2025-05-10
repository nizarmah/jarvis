package ffmpeg

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
)

// OnCombinedFunc is the callback for post-processing the combined file.
type OnCombinedFunc func(ctx context.Context, filePath string) error

// CombinerConfig is the configuration for the combiner.
type CombinerConfig struct {
	// Debug enables logging while watching and combining chunks.
	Debug bool
	// InputDir is the directory for the audio chunks.
	InputDir string
	// OutputDir is the directory for the output file.
	OutputDir string
	// OnCombined is the callback for post-processing the combined file.
	OnCombined OnCombinedFunc
}

// Combiner is a combiner for audio chunks.
type Combiner struct {
	debug     bool
	inputDir  string
	outputDir string

	onCombined OnCombinedFunc
	watcher    *fsnotify.Watcher
}

// NewCombiner initializes the combiner.
func NewCombiner(cfg CombinerConfig) (*Combiner, error) {
	if cfg.InputDir == "" {
		return nil, fmt.Errorf("input directory is required")
	}

	if err := createDirIfNotExists(cfg.InputDir); err != nil {
		return nil, fmt.Errorf("failed to create input dir: %w", err)
	}

	if cfg.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	if err := createDirIfNotExists(cfg.OutputDir); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	if cfg.OnCombined == nil {
		return nil, fmt.Errorf("on combined callback is required")
	}

	return &Combiner{
		debug:      cfg.Debug,
		inputDir:   cfg.InputDir,
		outputDir:  cfg.OutputDir,
		onCombined: cfg.OnCombined,
	}, nil
}

// Start starts the combiner by watching for file changes in the input directory.
func (c *Combiner) Start(ctx context.Context) error {
	if c.watcher != nil {
		return fmt.Errorf("combiner already started")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Store the watcher for future use.
	c.watcher = watcher

	// Watch the input directory, to combine the audio chunks.
	if err := watcher.Add(c.inputDir); err != nil {
		return fmt.Errorf("failed to watch %s: %w", c.inputDir, err)
	}

	go c.runWatcher(ctx)

	if c.debug {
		log.Println("combiner started")
	}

	return nil
}

// Stop stops the combiner by closing the watcher.
func (c *Combiner) stop(reason string) error {
	if c.watcher == nil {
		return fmt.Errorf("watcher not started")
	}

	if err := c.watcher.Close(); err != nil {
		return fmt.Errorf("failed to close watcher: %w", err)
	}

	c.watcher = nil

	if c.debug {
		log.Println(fmt.Sprintf("combiner stopped: %s", reason))
	}

	return nil
}

// RunWatcher watches for file changes and calls the handler when a new chunk is available.
func (c *Combiner) runWatcher(ctx context.Context) {
	for {
		select {
		// Clean up the goroutine when necessary.
		case <-ctx.Done():
			c.stop("context cancelled")
			return

		// Handle the error from the watcher.
		case err, ok := <-c.watcher.Errors:
			c.stop(fmt.Sprintf("watcher error (ok: %v): %s", ok, err))
			return

		// Handle the event from the watcher.
		case event, ok := <-c.watcher.Events:
			if !ok {
				c.stop("watcher closed")
				return
			}
			if err := c.handleWatcherEvent(ctx, event); err != nil {
				c.stop(fmt.Sprintf("failed to handle watcher event: %s", err))
				return
			}
		}
	}
}

// HandleWatcherEvent filters the events from the watcher and only handles chunk-related events.
func (c *Combiner) handleWatcherEvent(ctx context.Context, event fsnotify.Event) error {
	// Only handle when a chunk is written to a file, that's when a chunk is complete.
	// Don't handle when a chunk is created, because sometimes it doesn't have data yet.
	if event.Op&(fsnotify.Write) == 0 {
		return nil
	}

	// Only handle files that match our chunk pattern.
	if !chunkRegex.MatchString(event.Name) {
		return nil
	}

	if c.debug {
		log.Println(fmt.Sprintf("chunk event: %s", event.Name))
	}

	return c.handleChunk(ctx, event.Name)
}

// HandleChunk gets the current and previous chunks and combines them.
func (c *Combiner) handleChunk(ctx context.Context, filename string) error {
	// Extract the chunk index from the filename.
	filenameParts := chunkRegex.FindStringSubmatch(filename)
	if len(filenameParts) != 2 {
		return fmt.Errorf("invalid filename (%s) parts: %v", filename, filenameParts)
	}

	// Convert the chunk index to an integer.
	chunkIndex, err := strconv.Atoi(filenameParts[1])
	if err != nil {
		return fmt.Errorf("invalid filename (%s) index: %w", filename, err)
	}

	// Get the previous chunk index, considering the chunk wrap.
	// eg. current chunk `chunk_0.wav` has previous chunk `chunk_5.wav`
	previousChunkIndex := chunkIndex - 1
	if previousChunkIndex < 0 {
		previousChunkIndex = chunkWrap - 1
	}

	// Get the current and previous chunk filenames.
	currChunk := fmt.Sprintf(chunkPattern, chunkIndex)
	prevChunk := fmt.Sprintf(chunkPattern, previousChunkIndex)

	// Get the current and previous chunk paths.
	currChunkPath := fmt.Sprintf("%s/%s", c.inputDir, currChunk)
	prevChunkPath := fmt.Sprintf("%s/%s", c.inputDir, prevChunk)

	// Create the combined filename and path.
	combined := fmt.Sprintf(combinedPattern, time.Now().UnixNano())
	combinedPath := filepath.Join(c.outputDir, combined)

	if c.debug {
		log.Println(fmt.Sprintf("handling chunk: %s, previous chunk: %s", currChunk, prevChunk))
	}

	return c.combineChunks(ctx, currChunkPath, prevChunkPath, combinedPath)
}

// CombineChunks combines two chunks into a single file.
func (c *Combiner) combineChunks(ctx context.Context, currChunkPath, prevChunkPath, combinedPath string) error {
	// Concatenate the curr and prev chunks into a single file.
	// If the previous chunk doesn't exist, fallback to using only the current chunk.
	// Eg. curr: `chunk_0.wav`, prev: `chunk_5.wav`, but it's the first chunk, so prev doesn't exist yet.
	input := fmt.Sprintf("concat:%s|%s", prevChunkPath, currChunkPath)
	if _, err := os.Stat(prevChunkPath); os.IsNotExist(err) {
		input = currChunkPath
	}

	args := buildFfmpegArgs(
		combinedFfmpegArgs,
		[]string{"-i", input},
		[]string{"-c", "copy", combinedPath},
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if c.debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to combine chunks: %w", err)
	}

	if c.debug {
		log.Println(fmt.Sprintf("combined chunks: %s -> %s", input, combinedPath))
	}

	if err := c.onCombined(ctx, combinedPath); err != nil {
		return fmt.Errorf("failed to post-process combined file: %w", err)
	}

	return nil
}
