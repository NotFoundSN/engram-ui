//go:build linux

package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxAutostart_Install(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	mgr := &linuxAutostart{}
	execPath := "/usr/local/bin/engram-ui"

	result, err := mgr.Install(execPath)
	if err != nil {
		t.Fatalf("Install(): %v", err)
	}

	expectedPath := LinuxSystemdUnitPath(tmpHome, "")
	if result.Destination != expectedPath {
		t.Errorf("Install: Destination = %q, want %q", result.Destination, expectedPath)
	}
	if result.Action != ActionInstalled {
		t.Errorf("Install: Action = %q, want %q", result.Action, ActionInstalled)
	}

	// Verify file content.
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("ReadFile unit: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "ExecStart="+execPath+" serve") {
		t.Errorf("unit file missing ExecStart line, got:\n%s", content)
	}
	if !strings.Contains(content, "WantedBy=default.target") {
		t.Errorf("unit file missing WantedBy, got:\n%s", content)
	}

	// Second run: idempotent.
	result2, err := mgr.Install(execPath)
	if err != nil {
		t.Fatalf("Install (second run): %v", err)
	}
	if result2.Action != ActionInstalled {
		t.Errorf("Install second run: Action = %q, want ActionInstalled", result2.Action)
	}
}

func TestLinuxAutostart_Remove_NotRegistered(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	mgr := &linuxAutostart{}
	result, err := mgr.Remove()
	if err != nil {
		t.Fatalf("Remove (not registered): %v", err)
	}
	if result.Action != ActionNotRegistered {
		t.Errorf("Remove: Action = %q, want %q", result.Action, ActionNotRegistered)
	}
}

func TestLinuxAutostart_Remove_Registered(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	// Write a unit file manually.
	unitPath := LinuxSystemdUnitPath(tmpHome, "")
	if err := os.MkdirAll(filepath.Dir(unitPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unitPath, []byte("[Unit]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mgr := &linuxAutostart{}
	result, err := mgr.Remove()
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if result.Action != ActionRemoved {
		t.Errorf("Remove: Action = %q, want %q", result.Action, ActionRemoved)
	}
	if _, statErr := os.Stat(unitPath); !os.IsNotExist(statErr) {
		t.Errorf("Remove: unit file still exists at %q", unitPath)
	}
}
