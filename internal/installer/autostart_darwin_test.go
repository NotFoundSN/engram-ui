//go:build darwin

package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDarwinAutostart_Install(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr := &darwinAutostart{}
	execPath := "/usr/local/bin/engram-ui"

	result, err := mgr.Install(execPath)
	if err != nil {
		t.Fatalf("Install(): %v", err)
	}

	expectedPath := MacOSLaunchAgentPath(tmpHome)
	if result.Destination != expectedPath {
		t.Errorf("Install: Destination = %q, want %q", result.Destination, expectedPath)
	}
	if result.Action != ActionInstalled {
		t.Errorf("Install: Action = %q, want %q", result.Action, ActionInstalled)
	}

	// Verify file content.
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("ReadFile plist: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<key>RunAtLoad</key>") {
		t.Errorf("plist missing RunAtLoad, got:\n%s", content)
	}
	if !strings.Contains(content, "<true/>") {
		t.Errorf("plist RunAtLoad not true, got:\n%s", content)
	}
	if !strings.Contains(content, "<key>KeepAlive</key>") {
		t.Errorf("plist missing KeepAlive, got:\n%s", content)
	}
	if !strings.Contains(content, "<false/>") {
		t.Errorf("plist KeepAlive not false, got:\n%s", content)
	}
	if !strings.Contains(content, "<string>"+execPath+"</string>") {
		t.Errorf("plist missing execPath, got:\n%s", content)
	}

	// Second run: idempotent.
	_, err = mgr.Install(execPath)
	if err != nil {
		t.Fatalf("Install (second run): %v", err)
	}
}

func TestDarwinAutostart_Remove_NotRegistered(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	mgr := &darwinAutostart{}
	result, err := mgr.Remove()
	if err != nil {
		t.Fatalf("Remove (not registered): %v", err)
	}
	if result.Action != ActionNotRegistered {
		t.Errorf("Remove: Action = %q, want %q", result.Action, ActionNotRegistered)
	}
}

func TestDarwinAutostart_Remove_Registered(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	plistPath := MacOSLaunchAgentPath(tmpHome)
	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plistPath, []byte("<plist/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	mgr := &darwinAutostart{}
	result, err := mgr.Remove()
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if result.Action != ActionRemoved {
		t.Errorf("Remove: Action = %q, want %q", result.Action, ActionRemoved)
	}
	if _, statErr := os.Stat(plistPath); !os.IsNotExist(statErr) {
		t.Errorf("Remove: plist still exists at %q", plistPath)
	}
}
