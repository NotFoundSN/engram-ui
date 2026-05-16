package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// newTestModel returns a Model rooted at a tmp home so tests do not touch
// the developer's real home dir. APPDATA is also pointed at a subdir of
// tmp so detectAutostart() does not see the developer's real Windows
// autostart entry (if any).
func newTestModel(t *testing.T) (Model, string) {
	t.Helper()
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".config", "opencode"), 0o755); err != nil {
		t.Fatalf("mkdir opencode: %v", err)
	}
	t.Setenv("APPDATA", filepath.Join(tmp, "AppData"))
	return NewModel(tmp, ""), tmp
}

func keyMsg(s string) tea.KeyMsg {
	// Map a few special names to tea.KeyMsg; otherwise treat as runes.
	switch s {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestNewModel_InitialStateIsServerTab(t *testing.T) {
	m, _ := newTestModel(t)
	if m.activeTab != TabServer {
		t.Errorf("initial activeTab = %v, want TabServer", m.activeTab)
	}
	if len(m.desired) != 0 {
		t.Errorf("initial desired = %v, want empty", m.desired)
	}
}

func TestModel_TabKeyCyclesForward(t *testing.T) {
	m, _ := newTestModel(t)
	model, _ := m.Update(keyMsg("tab"))
	m = model.(Model)
	if m.activeTab != TabSkillsClaude {
		t.Errorf("after tab: activeTab = %v, want TabSkillsClaude", m.activeTab)
	}
}

func TestModel_ShiftTabCyclesBackward(t *testing.T) {
	m, _ := newTestModel(t)
	model, _ := m.Update(keyMsg("shift+tab"))
	m = model.(Model)
	if m.activeTab != TabReview {
		t.Errorf("after shift+tab from Server: activeTab = %v, want TabReview", m.activeTab)
	}
}

func TestModel_NumberKeysJumpToTab(t *testing.T) {
	m, _ := newTestModel(t)
	model, _ := m.Update(keyMsg("3"))
	m = model.(Model)
	if m.activeTab != TabSkillsOpenCode {
		t.Errorf("after '3': activeTab = %v, want TabSkillsOpenCode", m.activeTab)
	}
}

func TestModel_SpaceTogglesAndStagesChange(t *testing.T) {
	m, _ := newTestModel(t)
	// Server tab is active, row 0 is autostart. Initial current is NotInstalled
	// (tmp home has no autostart artifacts).
	model, _ := m.Update(keyMsg(" "))
	m = model.(Model)
	if len(m.desired) != 1 {
		t.Fatalf("after space: desired has %d entries, want 1", len(m.desired))
	}
	if m.desired["autostart"] != StateInstalled {
		t.Errorf("after space: desired[autostart] = %v, want StateInstalled", m.desired["autostart"])
	}
}

func TestModel_DoubleSpaceUnstages(t *testing.T) {
	m, _ := newTestModel(t)
	// Toggle: NotInstalled -> Installed (staged).
	model, _ := m.Update(keyMsg(" "))
	m = model.(Model)
	// Toggle again: should remove from desired (back to current=NotInstalled).
	model, _ = m.Update(keyMsg(" "))
	m = model.(Model)
	if len(m.desired) != 0 {
		t.Errorf("after double space: desired = %v, want empty", m.desired)
	}
}

func TestModel_CursorDownInSkillsTab(t *testing.T) {
	m, _ := newTestModel(t)
	// Switch to SkillsClaude (3 rows).
	model, _ := m.Update(keyMsg("2"))
	m = model.(Model)
	if got := m.cursor[TabSkillsClaude]; got != 0 {
		t.Fatalf("initial cursor = %d, want 0", got)
	}
	model, _ = m.Update(keyMsg("down"))
	m = model.(Model)
	if got := m.cursor[TabSkillsClaude]; got != 1 {
		t.Errorf("after down: cursor = %d, want 1", got)
	}
}

func TestModel_QuitWithoutStagedQuitsImmediately(t *testing.T) {
	m, _ := newTestModel(t)
	_, cmd := m.Update(keyMsg("q"))
	if cmd == nil {
		t.Fatal("after q with no staged changes: expected tea.Quit cmd, got nil")
	}
}

func TestModel_QuitWithStagedShowsConfirm(t *testing.T) {
	m, _ := newTestModel(t)
	// Stage a change.
	model, _ := m.Update(keyMsg(" "))
	m = model.(Model)
	if len(m.desired) == 0 {
		t.Fatal("setup: expected staged change after space")
	}
	// Now try to quit — should show confirm, not quit.
	model, cmd := m.Update(keyMsg("q"))
	m = model.(Model)
	if cmd != nil {
		t.Errorf("with staged change, q produced cmd %v, want nil (confirm overlay)", cmd)
	}
	if !m.confirmingQuit {
		t.Error("with staged change, q should set confirmingQuit=true")
	}
}

func TestModel_ConfirmYesQuits(t *testing.T) {
	m, _ := newTestModel(t)
	model, _ := m.Update(keyMsg(" "))
	m = model.(Model)
	model, _ = m.Update(keyMsg("q"))
	m = model.(Model)
	// Now press y.
	_, cmd := m.Update(keyMsg("y"))
	if cmd == nil {
		t.Error("confirm y: expected tea.Quit cmd, got nil")
	}
}

func TestModel_ConfirmNoCancels(t *testing.T) {
	m, _ := newTestModel(t)
	model, _ := m.Update(keyMsg(" "))
	m = model.(Model)
	model, _ = m.Update(keyMsg("q"))
	m = model.(Model)
	// Now press n.
	model, _ = m.Update(keyMsg("n"))
	m = model.(Model)
	if m.confirmingQuit {
		t.Error("confirm n: expected confirmingQuit=false")
	}
	if len(m.desired) == 0 {
		t.Error("confirm n: staged change should remain")
	}
}

func TestModel_StagedChangesReturnsOnlyDiff(t *testing.T) {
	m, _ := newTestModel(t)
	// Stage autostart.
	model, _ := m.Update(keyMsg(" "))
	m = model.(Model)
	staged := m.stagedChanges()
	if len(staged) != 1 || staged[0].ID != "autostart" {
		t.Errorf("stagedChanges = %v, want one autostart entry", staged)
	}
}

func TestModel_ReviewTabRendersStaged(t *testing.T) {
	m, _ := newTestModel(t)
	model, _ := m.Update(keyMsg(" "))
	m = model.(Model)
	model, _ = m.Update(keyMsg("4")) // jump to Review
	m = model.(Model)
	out := m.View()
	if !contains(out, "autostart") {
		t.Errorf("Review View() missing 'autostart' entry; got:\n%s", out)
	}
	if !contains(out, "Press Enter to apply") {
		t.Errorf("Review View() missing apply hint; got:\n%s", out)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
