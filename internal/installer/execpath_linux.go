//go:build linux

package installer

import (
	"os"
	"path/filepath"
)

func evalSymlinksReal() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		// Fall back to unresolved path on eval failure.
		return exe, nil
	}
	return resolved, nil
}
