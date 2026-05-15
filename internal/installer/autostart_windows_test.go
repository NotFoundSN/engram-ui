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
