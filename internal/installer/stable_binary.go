package installer

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/NotFoundSN/engram-ui/internal/version"
)

// EnsureStableBinary ensures the binary at sourcePath is installed at the stable path.
// It handles:
//   - Skipping if source is already at a stable path
//   - Skipping if stable binary exists and is byte-identical
//   - Version comparison to block downgrades
//   - Copying source to stable path with proper permissions
//
// Returns a Result with Action set to ActionSkipped, ActionInstalled, ActionOverwritten, or ActionBlockedDowngrade.
func EnsureStableBinary(sourcePath string, homeDir, localAppData string) (Result, error) {
	goos := runtime.GOOS

	// Get the stable path
	stablePath := StableBinaryPath(homeDir, localAppData, goos)
	if stablePath == "" {
		return Result{}, ErrUnsupportedPlatform
	}

	// Check if source is already at stable path
	if IsStableBinaryPath(sourcePath, homeDir, localAppData, goos) {
		return Result{
			Destination:      sourcePath,
			Action:           ActionSkipped,
			Notes:            "already at stable path",
			SourceVersion:    version.Current(),
			InstalledVersion: version.Current(),
		}, nil
	}

	// Check if stable binary exists
	stableExists := false
	if _, err := os.Stat(stablePath); err == nil {
		stableExists = true
	}

	// If stable exists, check if it's byte-identical
	if stableExists {
		identical, err := filesAreIdentical(sourcePath, stablePath)
		if err != nil {
			return Result{}, fmt.Errorf("comparing binaries: %w", err)
		}
		if identical {
			return Result{
				Destination:      stablePath,
				Action:           ActionSkipped,
				Notes:            "stable binary is identical",
				SourceVersion:    version.Current(),
				InstalledVersion: version.Current(),
			}, nil
		}

		// Check versions before overwriting
		installedVersion, err := resolveBinaryVersion(stablePath)
		if err == nil && installedVersion != "" {
			sourceVersion := version.Current()
			if version.Compare(sourceVersion, installedVersion) < 0 {
				// Source is older - block downgrade
				return Result{
					Destination:      stablePath,
					Action:           ActionBlockedDowngrade,
					Notes:            fmt.Sprintf("downgrade blocked: installed v%s, source v%s", installedVersion, sourceVersion),
					SourceVersion:    sourceVersion,
					InstalledVersion: installedVersion,
				}, ErrDowngradeBlocked
			}
		} else if err != nil {
			// Log parse failure but proceed with overwrite (downgrade guard bypassed)
			log.Printf("warning: failed to parse installed binary version at %s: %v", stablePath, err)
		}
	}

	// Ensure parent directory exists
	stableDir := filepath.Dir(stablePath)
	if err := os.MkdirAll(stableDir, 0755); err != nil {
		return Result{}, fmt.Errorf("creating stable directory: %w", err)
	}

	// Copy source to stable path
	if err := copyFile(sourcePath, stablePath); err != nil {
		// Check for Windows file-in-use error
		if runtime.GOOS == "windows" && strings.Contains(err.Error(), "being used by another process") {
			return Result{}, fmt.Errorf("stable binary is in use: close any running engram-ui instances and retry: %w", err)
		}
		return Result{}, fmt.Errorf("copying to stable path: %w", err)
	}

	action := ActionInstalled
	if stableExists {
		action = ActionOverwritten
	}

	return Result{
		Destination:      stablePath,
		Action:           action,
		Notes:            "stable binary installed",
		SourceVersion:    version.Current(),
		InstalledVersion: version.Current(),
	}, nil
}

// resolveBinaryVersion is a package-level variable that can be stubbed in tests.
// It resolves the version of a binary by executing it.
var resolveBinaryVersion = resolveBinaryVersionReal

// resolveBinaryVersionReal runs the binary with version subcommand and parses the output.
// Returns empty string if version cannot be determined.
func resolveBinaryVersionReal(binaryPath string) (string, error) {
	// Try "version" subcommand first (matches engram-ui's interface)
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse "engram-ui v1.2.3" or "engram-ui 1.2.3" format
	outputStr := strings.TrimSpace(string(output))
	parts := strings.Fields(outputStr)
	if len(parts) >= 2 {
		return parts[len(parts)-1], nil
	}

	return "", fmt.Errorf("unable to parse version from: %s", outputStr)
}

// filesAreIdentical compares two files by size and content.
func filesAreIdentical(path1, path2 string) (bool, error) {
	info1, err := os.Stat(path1)
	if err != nil {
		return false, err
	}

	info2, err := os.Stat(path2)
	if err != nil {
		return false, err
	}

	// Quick check: different sizes means different content
	if info1.Size() != info2.Size() {
		return false, nil
	}

	// Open and compare content
	f1, err := os.Open(path1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(path2)
	if err != nil {
		return false, err
	}
	defer f2.Close()

	// Compare in chunks
	const chunkSize = 64 * 1024
	buf1 := make([]byte, chunkSize)
	buf2 := make([]byte, chunkSize)

	for {
		n1, err1 := f1.Read(buf1)
		n2, err2 := f2.Read(buf2)

		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		}

		if err1 != nil && err1 != io.EOF {
			return false, err1
		}
		if err2 != nil && err2 != io.EOF {
			return false, err2
		}
	}
}

// copyFile copies a file from src to dst, setting executable permissions on Unix.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Get source file permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy content
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Sync to ensure data is written
	if err := destFile.Sync(); err != nil {
		return err
	}

	// On Unix, ensure the file is executable
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dst, 0755); err != nil {
			return err
		}
	}

	return nil
}
