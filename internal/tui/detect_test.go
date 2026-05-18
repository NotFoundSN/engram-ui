package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NotFoundSN/engram-ui/internal/installer"
)

func TestDetectState_EmptyHome_AllSkillsNotInstalled(t *testing.T) {
	tmpHome := t.TempDir()

	items := BuildCatalog()
	for _, it := range items {
		// Skip autostart — its detection is OS-specific and the empty-home
		// case is not portable across OSes (uses APPDATA on Windows, plist
		// path on macOS, systemd unit on Linux). Covered separately.
		if it.ID == "autostart" {
			continue
		}
		got := DetectState(it, tmpHome, "")
		if got == StateInstalled {
			t.Errorf("item %q: with empty home, got Installed, want NotInstalled or Unavailable", it.ID)
		}
	}
}

func TestDetectState_AfterInstall_ClaudeSkillReportsInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Install brainstorm for Claude Code via the installer package directly.
	if _, err := installer.InstallClaudeCodeSkill("brainstorm"); err != nil {
		t.Fatalf("InstallClaudeCodeSkill: %v", err)
	}

	// Now detect should report it as Installed.
	items := BuildCatalog()
	var found *Item
	for i := range items {
		if items[i].ID == "skill:brainstorm:claude" {
			found = &items[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("catalog missing skill:brainstorm:claude")
	}

	got := DetectState(*found, tmpHome, "")
	if got != StateInstalled {
		t.Errorf("DetectState after install: got %v, want StateInstalled", got)
	}
}

func TestDetectState_AfterInstall_OpenCodeSkillReportsInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	if _, err := installer.InstallOpenCodeSkill("debug"); err != nil {
		t.Fatalf("InstallOpenCodeSkill: %v", err)
	}

	items := BuildCatalog()
	var found *Item
	for i := range items {
		if items[i].ID == "skill:debug:opencode" {
			found = &items[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("catalog missing skill:debug:opencode")
	}

	got := DetectState(*found, tmpHome, "")
	if got != StateInstalled {
		t.Errorf("DetectState after install: got %v, want StateInstalled", got)
	}
}

func TestDetectState_ToolDirMissing_ReportsUnavailable(t *testing.T) {
	// When ~/.claude/ does not exist, Claude Code skill items should be
	// reported as Unavailable (not just NotInstalled). This drives the UX:
	// disabled rows with a hint "Claude Code not detected".
	tmpHome := t.TempDir()
	// Do NOT create ~/.claude/. The skill item should report Unavailable.

	items := BuildCatalog()
	var found *Item
	for i := range items {
		if items[i].ID == "skill:brainstorm:claude" {
			found = &items[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("catalog missing skill:brainstorm:claude")
	}

	got := DetectState(*found, tmpHome, "")
	if got != StateUnavailable {
		t.Errorf("DetectState with missing ~/.claude/: got %v, want StateUnavailable", got)
	}
}

func TestDetectState_ToolDirPresent_ReportsNotInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	// Create ~/.claude/ but no skill subfolder.
	if err := os.MkdirAll(filepath.Join(tmpHome, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	items := BuildCatalog()
	var found *Item
	for i := range items {
		if items[i].ID == "skill:brainstorm:claude" {
			found = &items[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("catalog missing skill:brainstorm:claude")
	}

	got := DetectState(*found, tmpHome, "")
	if got != StateNotInstalled {
		t.Errorf("DetectState with empty ~/.claude/: got %v, want StateNotInstalled", got)
	}
}
