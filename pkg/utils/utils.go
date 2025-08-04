package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// get project root absolute dir
func GetProjectRoot() (string, error) {
	// program cnt, filepath, line number, ok
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get project root: %v", ok)
	}

	dir := filepath.Dir(path)
	for {
		if _, err := os.Stat(filepath.Join(dir, "config.json")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("project root (with go.mod or config.json) not found")
		}
		dir = parent
	}
}
