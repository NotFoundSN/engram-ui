package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallClaudeCodeSkill(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	result, err := InstallClaudeCodeSkill()
	if err != nil {
		t.Fatalf("InstallClaudeCodeSkill(): %v", err)
	}

	// Verify SKILL.md exists at expected path.
	expectedDir := ClaudeSkillDir(tmpHome)
	skillMD := filepath.Join(expectedDir, "SKILL.md")
	if _, statErr := os.Stat(skillMD); statErr != nil {
		t.Errorf("InstallClaudeCodeSkill: SKILL.md not found at %q: %v", skillMD, statErr)
	}

	// Verify destination in result.
	if result.Destination != expectedDir {
		t.Errorf("InstallClaudeCodeSkill: Destination = %q, want %q", result.Destination, expectedDir)
	}

	// Verify action is ActionInstalled on first run.
	if result.Action != ActionInstalled {
		t.Errorf("InstallClaudeCodeSkill: Action = %q, want %q", result.Action, ActionInstalled)
	}

	// Second run: should report ActionOverwritten.
	result2, err := InstallClaudeCodeSkill()
	if err != nil {
		t.Fatalf("InstallClaudeCodeSkill (second run): %v", err)
	}
	if result2.Action != ActionOverwritten {
		t.Errorf("InstallClaudeCodeSkill second run: Action = %q, want %q", result2.Action, ActionOverwritten)
	}
}

func TestInstallOpenCodeSkill(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "") // clear any real env override

	result, err := InstallOpenCodeSkill()
	if err != nil {
		t.Fatalf("InstallOpenCodeSkill(): %v", err)
	}

	// Expected: {tmpHome}/.config/opencode/skills/engram-conventions/
	expectedDir := OpenCodeSkillDir(tmpHome, "")
	skillMD := filepath.Join(expectedDir, "SKILL.md")
	if _, statErr := os.Stat(skillMD); statErr != nil {
		t.Errorf("InstallOpenCodeSkill: SKILL.md not found at %q: %v", skillMD, statErr)
	}

	if result.Destination != expectedDir {
		t.Errorf("InstallOpenCodeSkill: Destination = %q, want %q", result.Destination, expectedDir)
	}
}

func TestInstallAutostart(t *testing.T) {
	// InstallAutostart uses os.Executable() internally — we can't easily stub that.
	// But we can at least test that the function calls through to NewAutostartManager.
	// On Windows, this will write a .bat to a temp APPDATA dir.
	tmpAppData := t.TempDir()
	t.Setenv("APPDATA", tmpAppData)

	// Stub ResolveExecPath to return a predictable path.
	orig := evalSymlinks
	defer func() { evalSymlinks = orig }()
	evalSymlinks = func() (string, error) {
		return `C:\fake\engram-ui.exe`, nil
	}

	result, err := InstallAutostart()
	if err != nil {
		t.Fatalf("InstallAutostart(): %v", err)
	}
	if result.Action != ActionInstalled {
		t.Errorf("InstallAutostart: Action = %q, want ActionInstalled", result.Action)
	}
}

func TestRemoveAutostart_NotRegistered(t *testing.T) {
	tmpAppData := t.TempDir()
	t.Setenv("APPDATA", tmpAppData)

	result, err := RemoveAutostart()
	if err != nil {
		t.Fatalf("RemoveAutostart(): %v", err)
	}
	if result.Action != ActionNotRegistered {
		t.Errorf("RemoveAutostart: Action = %q, want ActionNotRegistered", result.Action)
	}
}

func TestResolveExecPath(t *testing.T) {
	// Stub evalSymlinks to return a predictable path.
	orig := evalSymlinks
	defer func() { evalSymlinks = orig }()

	evalSymlinks = func() (string, error) {
		return "/fake/path/engram-ui", nil
	}

	path, err := ResolveExecPath()
	if err != nil {
		t.Fatalf("ResolveExecPath(): %v", err)
	}
	if path != "/fake/path/engram-ui" {
		t.Errorf("ResolveExecPath() = %q, want %q", path, "/fake/path/engram-ui")
	}
}
