package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the Bubbletea state for the engram-ui installer TUI.
type Model struct {
	home, xdg string
	items     []Item

	// current[id] is the on-disk State as detected at startup (or after apply).
	current map[string]State
	// desired[id] is set only for items the user has explicitly toggled away
	// from current. Empty when no changes are staged.
	desired map[string]State

	activeTab Tab
	// cursor[tab] is the row index within that tab. Persists across tab
	// switches so navigating away and back keeps your place.
	cursor map[Tab]int

	confirmingQuit bool
	applying       bool
	spinner        spinner.Model
	results        []Result
	width, height  int
}

// NewModel constructs the initial Model. home + xdg are explicit so callers
// (including tests) can control the filesystem root the TUI operates against.
func NewModel(home, xdg string) Model {
	items := BuildCatalog()
	current := make(map[string]State, len(items))
	for _, it := range items {
		current[it.ID] = DetectState(it, home, xdg)
	}
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	return Model{
		home:      home,
		xdg:       xdg,
		items:     items,
		current:   current,
		desired:   make(map[string]State),
		activeTab: TabServer,
		cursor:    map[Tab]int{TabServer: 0, TabSkillsClaude: 0, TabSkillsOpenCode: 0, TabReview: 0},
		spinner:   sp,
	}
}

// Init satisfies tea.Model. No async work needed at start.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case applyDoneMsg:
		m.applying = false
		m.results = msg.results
		// Refresh detected current state — actions changed the filesystem.
		for _, it := range m.items {
			m.current[it.ID] = DetectState(it, m.home, m.xdg)
		}
		m.desired = make(map[string]State)
		return m, nil

	case spinner.TickMsg:
		// Only advance the spinner while an apply is in flight. Otherwise the
		// tick is a stale message from a completed run — drop it so we don't
		// schedule a new tick forever.
		if !m.applying {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		return m.updateKey(msg)
	}
	return m, nil
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Quit confirmation overlay intercepts everything.
	if m.confirmingQuit {
		switch msg.String() {
		case "y", "Y":
			return m, tea.Quit
		case "n", "N", "esc":
			m.confirmingQuit = false
			return m, nil
		}
		return m, nil
	}

	// Block input while applying.
	if m.applying {
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c", "esc":
		if len(m.desired) > 0 && msg.String() != "ctrl+c" {
			m.confirmingQuit = true
			return m, nil
		}
		return m, tea.Quit

	case "tab", "right", "l":
		m.activeTab = nextTab(m.activeTab)
		return m, nil
	case "shift+tab", "left", "h":
		m.activeTab = prevTab(m.activeTab)
		return m, nil

	case "1":
		m.activeTab = TabServer
		return m, nil
	case "2":
		m.activeTab = TabSkillsClaude
		return m, nil
	case "3":
		m.activeTab = TabSkillsOpenCode
		return m, nil
	case "4":
		m.activeTab = TabReview
		return m, nil

	case "up", "k":
		m.moveCursor(-1)
		return m, nil
	case "down", "j":
		m.moveCursor(+1)
		return m, nil

	case " ", "space":
		m.toggleAtCursor()
		return m, nil

	case "enter":
		if m.activeTab == TabReview && len(m.desired) > 0 {
			m.applying = true
			return m, tea.Batch(m.applyCmd(), m.spinner.Tick)
		}
		return m, nil
	}

	return m, nil
}

// moveCursor advances the cursor within the active tab, skipping items
// whose CurrentState is Unavailable (per decision: skip-nav for disabled rows).
func (m Model) moveCursor(delta int) {
	rows := m.itemsForTab(m.activeTab)
	if len(rows) == 0 {
		return
	}
	cur := m.cursor[m.activeTab]
	for step := 0; step < len(rows); step++ {
		cur = (cur + delta + len(rows)) % len(rows)
		if m.current[rows[cur].ID] != StateUnavailable {
			m.cursor[m.activeTab] = cur
			return
		}
	}
	// All rows unavailable — leave cursor as-is.
}

// toggleAtCursor flips the desired state of the row under the cursor.
// On the Review tab toggle is a no-op (Review shows the diff, doesn't host
// staging actions). Unavailable rows are non-toggleable.
func (m *Model) toggleAtCursor() {
	if m.activeTab == TabReview {
		return
	}
	rows := m.itemsForTab(m.activeTab)
	if len(rows) == 0 {
		return
	}
	cur := m.cursor[m.activeTab]
	if cur >= len(rows) {
		return
	}
	it := rows[cur]
	if m.current[it.ID] == StateUnavailable {
		return
	}

	currentDesired := m.effectiveDesired(it.ID)
	// Toggle: Installed <-> NotInstalled.
	var next State
	if currentDesired == StateInstalled {
		next = StateNotInstalled
	} else {
		next = StateInstalled
	}

	// If toggling back to current, drop the entry (no change staged).
	if next == m.current[it.ID] {
		delete(m.desired, it.ID)
	} else {
		m.desired[it.ID] = next
	}
}

// effectiveDesired returns the staged desired state for an item, or its current
// state when nothing is staged.
func (m Model) effectiveDesired(id string) State {
	if v, ok := m.desired[id]; ok {
		return v
	}
	return m.current[id]
}

// itemsForTab returns the catalog rows belonging to a tab.
func (m Model) itemsForTab(tab Tab) []Item {
	out := make([]Item, 0, len(m.items))
	for _, it := range m.items {
		if it.Tab == tab {
			out = append(out, it)
		}
	}
	return out
}

// stagedChanges returns the subset of items whose desired state differs from
// current, in catalog order.
func (m Model) stagedChanges() []Item {
	out := make([]Item, 0, len(m.desired))
	for _, it := range m.items {
		if want, ok := m.desired[it.ID]; ok && want != m.current[it.ID] {
			out = append(out, it)
		}
	}
	return out
}

// applyCmd returns a tea.Cmd that executes Apply asynchronously and emits an
// applyDoneMsg when finished.
func (m Model) applyCmd() tea.Cmd {
	desired := make(map[string]State, len(m.desired))
	for k, v := range m.desired {
		desired[k] = v
	}
	home, xdg := m.home, m.xdg
	items := m.items
	return func() tea.Msg {
		results := Apply(items, desired, home, xdg)
		return applyDoneMsg{results: results}
	}
}

type applyDoneMsg struct {
	results []Result
}

// nextTab and prevTab cycle through the 4 tabs (Review last).
func nextTab(t Tab) Tab {
	switch t {
	case TabServer:
		return TabSkillsClaude
	case TabSkillsClaude:
		return TabSkillsOpenCode
	case TabSkillsOpenCode:
		return TabReview
	case TabReview:
		return TabServer
	}
	return TabServer
}

func prevTab(t Tab) Tab {
	switch t {
	case TabServer:
		return TabReview
	case TabSkillsClaude:
		return TabServer
	case TabSkillsOpenCode:
		return TabSkillsClaude
	case TabReview:
		return TabSkillsOpenCode
	}
	return TabServer
}

// fmtState renders a one-character glyph for the row checkbox column.
func fmtState(s State) string {
	switch s {
	case StateInstalled:
		return "[x]"
	case StateNotInstalled:
		return "[ ]"
	case StateUnavailable:
		return "[-]"
	default:
		return "[?]"
	}
}

// fmtStateName is the longer English label for the Review tab columns.
func fmtStateName(s State) string {
	return s.String()
}

// RunTUI launches the Bubbletea program. Returns an error if Bubbletea reports
// one (e.g. terminal not interactive — caller should fall back to help text).
func RunTUI() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("UserHomeDir: %w", err)
	}
	xdg := os.Getenv("XDG_CONFIG_HOME")

	m := NewModel(home, xdg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
