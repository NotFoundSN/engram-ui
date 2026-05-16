package tui

import (
	"testing"
)

func TestBuildCatalog_ContainsAutostart(t *testing.T) {
	items := BuildCatalog()

	found := false
	for _, it := range items {
		if it.ID == "autostart" {
			found = true
			if it.Tab != TabServer {
				t.Errorf("autostart Tab = %v, want TabServer", it.Tab)
			}
		}
	}
	if !found {
		t.Errorf("BuildCatalog: autostart item missing (got %d items)", len(items))
	}
}

func TestBuildCatalog_OneSkillItemPerToolPerSkill(t *testing.T) {
	// For each skill in the installer catalog × {claude, opencode} we expect
	// exactly one Item in the TUI catalog.
	items := BuildCatalog()

	// Expected combinations (3 skills × 2 tools = 6).
	want := map[string]Tab{
		"skill:brainstorm:claude":          TabSkillsClaude,
		"skill:brainstorm:opencode":        TabSkillsOpenCode,
		"skill:debug:claude":               TabSkillsClaude,
		"skill:debug:opencode":             TabSkillsOpenCode,
		"skill:engram-conventions:claude":  TabSkillsClaude,
		"skill:engram-conventions:opencode": TabSkillsOpenCode,
	}

	got := make(map[string]Tab)
	for _, it := range items {
		if _, ok := want[it.ID]; ok {
			got[it.ID] = it.Tab
		}
	}

	for id, wantTab := range want {
		if gotTab, ok := got[id]; !ok {
			t.Errorf("BuildCatalog: expected item %q missing", id)
		} else if gotTab != wantTab {
			t.Errorf("BuildCatalog: item %q Tab = %v, want %v", id, gotTab, wantTab)
		}
	}
}

func TestBuildCatalog_TotalCount(t *testing.T) {
	// 1 server item (autostart) + 3 skills × 2 tools = 7 items.
	items := BuildCatalog()
	if len(items) != 7 {
		t.Errorf("BuildCatalog: got %d items, want 7", len(items))
	}
}

func TestBuildCatalog_LabelsNonEmpty(t *testing.T) {
	items := BuildCatalog()
	for _, it := range items {
		if it.Label == "" {
			t.Errorf("item %q has empty Label", it.ID)
		}
	}
}
