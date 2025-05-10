package whisper

import (
	"fmt"
	"os"
)

// createDirIfNotExists creates the directory if it does not exist.
func createDirIfNotExists(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create dir %s: %w", dir, err)
	}

	return nil
}
