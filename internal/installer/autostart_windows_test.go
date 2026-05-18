//go:build windows

package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsAutostart_Install(t *testing.T) {
	tmpAppData := t.TempDir()
	t.Setenv("APPDATA", tmpAppData)

	mgr := &windowsAutostart{}
	execPath := `C:\Users\user\engram-ui.exe`

	result, err := mgr.Install(execPath)
	if err != nil {
		t.Fatalf("Install(): %v", err)
	}

	expectedDir := WindowsStartupDir(tmpAppData)
	expectedPath := filepath.Join(expectedDir, "engram-ui.bat")

	if result.Destination != expectedPath {
		t.Errorf("Install: Destination = %q, want %q", result.Destination, expectedPath)
	}
	if result.Action != ActionInstalled {
		t.Errorf("Install: Action = %q, want %q", result.Action, ActionInstalled)
	}

	// Verify file content.
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("ReadFile bat: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "@echo off") {
		t.Errorf("bat missing @echo off, got:\n%s", content)
	}
	if !strings.Contains(content, `"`+execPath+`" serve`) {
		t.Errorf("bat missing execPath+serve, got:\n%s", content)
	}

	// Second run: idempotent.
	_, err = mgr.Install(execPath)
	if err != nil {
		t.Fatalf("Install (second run): %v", err)
	}
}

func TestWindowsAutostart_Remove_NotRegistered(t *testing.T) {
	tmpAppData := t.TempDir()
	t.Setenv("APPDATA", tmpAppData)

	mgr := &windowsAutostart{}
	result, err := mgr.Remove()
	if err != nil {
		t.Fatalf("Remove (not registered): %v", err)
	}
	if result.Action != ActionNotRegistered {
		t.Errorf("Remove: Action = %q, want %q", result.Action, ActionNotRegistered)
	}
}

func TestWindowsAutostart_Remove_Registered(t *testing.T) {
	tmpAppData := t.TempDir()
	t.Setenv("APPDATA", tmpAppData)

	// Write a bat file manually.
	startupDir := WindowsStartupDir(tmpAppData)
	batPath := filepath.Join(startupDir, "engram-ui.bat")
	if err := os.MkdirAll(startupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(batPath, []byte("@echo off\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mgr := &windowsAutostart{}
	result, err := mgr.Remove()
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if result.Action != ActionRemoved {
		t.Errorf("Remove: Action = %q, want %q", result.Action, ActionRemoved)
	}
	if _, statErr := os.Stat(batPath); !os.IsNotExist(statErr) {
		t.Errorf("Remove: bat file still exists at %q", batPath)
	}
}

// TestInstallAutostart_UsesStablePath verifies that autostart entries use the stable path.
// Note: This test uses a mock to avoid actual file system operations.
func TestInstallAutostart_UsesStablePath(t *testing.T) {
	// Create a temp directory structure
	tmpDir := t.TempDir()
	localAppData := tmpDir
	t.Setenv("LOCALAPPDATA", localAppData)
	t.Setenv("APPDATA", tmpDir)

	// Create a mock source binary (we'll use a stable path directly to test the entry content)
	stablePath := StableBinaryPath(tmpDir, localAppData, "windows")
	if err := os.MkdirAll(filepath.Dir(stablePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stablePath, []byte("mock binary"), 0755); err != nil {
		t.Fatal(err)
	}

	// Test that the manager uses the provided exec path in the entry
	mgr := &windowsAutostart{}
	result, err := mgr.Install(stablePath)
	if err != nil {
		t.Fatalf("Install(): %v", err)
	}

	// Verify the bat file was created
	expectedDir := WindowsStartupDir(tmpDir)
	batPath := filepath.Join(expectedDir, "engram-ui.bat")
	data, err := os.ReadFile(batPath)
	if err != nil {
		t.Fatalf("ReadFile bat: %v", err)
	}

	content := string(data)
	// Verify the stable path is in the content
	if !strings.Contains(content, stablePath) {
		t.Errorf("bat should contain stable path %q, got:\n%s", stablePath, content)
	}

	if result.Action != ActionInstalled {
		t.Errorf("Install: Action = %q, want %q", result.Action, ActionInstalled)
	}
}
