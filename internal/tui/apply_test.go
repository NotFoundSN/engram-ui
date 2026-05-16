package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApply_InstallsStagedSkill(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")
	// Create ~/.claude/ so the skill is Available (not Unavailable).
	if err := os.MkdirAll(filepath.Join(tmpHome, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	items := BuildCatalog()
	desired := map[string]State{
		"skill:brainstorm:claude": StateInstalled,
	}

	results := Apply(items, desired, tmpHome, "")
	if len(results) != 1 {
		t.Fatalf("Apply: got %d results, want 1", len(results))
	}

	r := results[0]
	if r.ItemID != "skill:brainstorm:claude" {
		t.Errorf("result ItemID = %q, want skill:brainstorm:claude", r.ItemID)
	}
	if !r.OK {
		t.Errorf("result OK = false; err = %v", r.Err)
	}

	// Verify install really happened on disk.
	wantPath := filepath.Join(tmpHome, ".claude", "skills", "brainstorm", "SKILL.md")
	if _, err := os.Stat(wantPath); err != nil {
		t.Errorf("expected installed SKILL.md at %q, missing: %v", wantPath, err)
	}
}

func TestApply_NoChangeWhenDesiredEqualsCurrent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	items := BuildCatalog()
	// Empty desired map = no changes staged.
	results := Apply(items, map[string]State{}, tmpHome, "")
	if len(results) != 0 {
		t.Errorf("Apply with empty desired: got %d results, want 0", len(results))
	}
}

func TestApply_ContinuesAfterError(t *testing.T) {
	// Stage two items: one valid, one with unknown ID (will fail).
	// Apply must return BOTH results (continue-on-error semantics).
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")
	if err := os.MkdirAll(filepath.Join(tmpHome, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	items := BuildCatalog()
	desired := map[string]State{
		"skill:brainstorm:claude": StateInstalled,
		"bogus-item-id":           StateInstalled, // not in catalog → must fail but not abort
	}

	results := Apply(items, desired, tmpHome, "")
	if len(results) != 2 {
		t.Fatalf("Apply: got %d results, want 2 (one ok, one err)", len(results))
	}

	var ok, errCount int
	for _, r := range results {
		if r.OK {
			ok++
		} else {
			errCount++
		}
	}
	if ok != 1 || errCount != 1 {
		t.Errorf("results: ok=%d err=%d, want ok=1 err=1", ok, errCount)
	}
}

func TestApply_UninstallsStagedSkill(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")
	if err := os.MkdirAll(filepath.Join(tmpHome, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// First install brainstorm.
	items := BuildCatalog()
	preInstall := map[string]State{"skill:brainstorm:claude": StateInstalled}
	if results := Apply(items, preInstall, tmpHome, ""); len(results) != 1 || !results[0].OK {
		t.Fatalf("setup install failed: %+v", results)
	}

	wantPath := filepath.Join(tmpHome, ".claude", "skills", "brainstorm", "SKILL.md")
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("setup install verification failed: %v", err)
	}

	// Now stage uninstall.
	desired := map[string]State{"skill:brainstorm:claude": StateNotInstalled}
	results := Apply(items, desired, tmpHome, "")
	if len(results) != 1 || !results[0].OK {
		t.Fatalf("uninstall Apply failed: %+v", results)
	}

	if _, err := os.Stat(wantPath); !os.IsNotExist(err) {
		t.Errorf("expected SKILL.md removed at %q, but Stat err = %v", wantPath, err)
	}
}
