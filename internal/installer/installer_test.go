package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallClaudeCodeSkill_Brainstorm(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	result, err := InstallClaudeCodeSkill("brainstorm")
	if err != nil {
		t.Fatalf("InstallClaudeCodeSkill(brainstorm): %v", err)
	}

	expectedDir := ClaudeSkillDir(tmpHome, "brainstorm")
	skillMD := filepath.Join(expectedDir, "SKILL.md")
	if _, statErr := os.Stat(skillMD); statErr != nil {
		t.Errorf("InstallClaudeCodeSkill: SKILL.md not found at %q: %v", skillMD, statErr)
	}

	if result.Destination != expectedDir {
		t.Errorf("InstallClaudeCodeSkill: Destination = %q, want %q", result.Destination, expectedDir)
	}

	if result.Action != ActionInstalled {
		t.Errorf("InstallClaudeCodeSkill: Action = %q, want %q", result.Action, ActionInstalled)
	}

	// Verify the installed file is the CLAUDE variant (contains <HARD-GATE> tag, not blockquote).
	data, err := os.ReadFile(skillMD)
	if err != nil {
		t.Fatalf("read installed SKILL.md: %v", err)
	}
	if !strings.Contains(string(data), "<HARD-GATE>") {
		t.Errorf("InstallClaudeCodeSkill: expected <HARD-GATE> tag in Claude variant, not found")
	}

	// Second run: should report ActionOverwritten.
	result2, err := InstallClaudeCodeSkill("brainstorm")
	if err != nil {
		t.Fatalf("InstallClaudeCodeSkill (second run): %v", err)
	}
	if result2.Action != ActionOverwritten {
		t.Errorf("InstallClaudeCodeSkill second run: Action = %q, want %q", result2.Action, ActionOverwritten)
	}
}

func TestInstallOpenCodeSkill_Brainstorm(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	result, err := InstallOpenCodeSkill("brainstorm")
	if err != nil {
		t.Fatalf("InstallOpenCodeSkill(brainstorm): %v", err)
	}

	expectedDir := OpenCodeSkillDir(tmpHome, "", "brainstorm")
	skillMD := filepath.Join(expectedDir, "SKILL.md")
	if _, statErr := os.Stat(skillMD); statErr != nil {
		t.Errorf("InstallOpenCodeSkill: SKILL.md not found at %q: %v", skillMD, statErr)
	}

	if result.Destination != expectedDir {
		t.Errorf("InstallOpenCodeSkill: Destination = %q, want %q", result.Destination, expectedDir)
	}

	// Verify the installed file is the OPENCODE variant (blockquote, no <HARD-GATE> XML tag).
	data, err := os.ReadFile(skillMD)
	if err != nil {
		t.Fatalf("read installed SKILL.md: %v", err)
	}
	body := string(data)
	if strings.Contains(body, "<HARD-GATE>") {
		t.Errorf("InstallOpenCodeSkill: <HARD-GATE> XML tag should NOT appear in OpenCode variant")
	}
	if !strings.Contains(body, "HARD GATE — MANDATORY") {
		t.Errorf("InstallOpenCodeSkill: expected blockquote-style HARD GATE in OpenCode variant")
	}
}

func TestInstallClaudeCodeSkill_EngramConventions_MultiFile(t *testing.T) {
	// engram-conventions is a multi-file skill: SKILL.md + 4 root .md + workflows/.
	// This test verifies the recursive CopySkill walks nested directories
	// (specifically workflows/sdd.md should land at the destination).
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	result, err := InstallClaudeCodeSkill("engram-conventions")
	if err != nil {
		t.Fatalf("InstallClaudeCodeSkill(engram-conventions): %v", err)
	}

	expectedDir := ClaudeSkillDir(tmpHome, "engram-conventions")
	if result.Destination != expectedDir {
		t.Errorf("Destination = %q, want %q", result.Destination, expectedDir)
	}

	// Spot-check files at top level + nested workflows/.
	wantPaths := []string{
		"SKILL.md",
		"types.md",
		"topic-keys.md",
		"lifecycle.md",
		"multi-repo.md",
		filepath.Join("workflows", "sdd.md"),
		filepath.Join("workflows", "ad-hoc.md"),
	}
	for _, rel := range wantPaths {
		p := filepath.Join(expectedDir, rel)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s installed at %q, missing: %v", rel, p, err)
		}
	}
}

func TestInstallAutostart(t *testing.T) {
	// Create temp directories for both APPDATA and LOCALAPPDATA
	tmpDir := t.TempDir()
	tmpAppData := filepath.Join(tmpDir, "AppData", "Roaming")
	tmpLocalAppData := filepath.Join(tmpDir, "AppData", "Local")
	if err := os.MkdirAll(tmpAppData, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tmpLocalAppData, 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("APPDATA", tmpAppData)
	t.Setenv("LOCALAPPDATA", tmpLocalAppData)

	// Create a mock "source" binary that will be copied to stable path
	sourceDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(sourceDir, "engram-ui.exe")
	if err := os.WriteFile(sourcePath, []byte("mock binary"), 0755); err != nil {
		t.Fatal(err)
	}

	orig := evalSymlinks
	defer func() { evalSymlinks = orig }()
	evalSymlinks = func() (string, error) {
		return sourcePath, nil
	}

	result, err := InstallAutostart()
	if err != nil {
		t.Fatalf("InstallAutostart(): %v", err)
	}
	// Action should reflect the autostart installation (not the stable binary install)
	if result.Action != ActionInstalled {
		t.Errorf("InstallAutostart: Action = %q, want ActionInstalled", result.Action)
	}

	// Verify stable binary was created
	stablePath := StableBinaryPath(tmpDir, tmpLocalAppData, "windows")
	if _, err := os.Stat(stablePath); os.IsNotExist(err) {
		t.Errorf("Stable binary was not created at %q", stablePath)
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
