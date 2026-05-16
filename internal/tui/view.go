package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	stylePage = lipgloss.NewStyle().Padding(1, 2)

	styleTab       = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("242"))
	styleTabActive = lipgloss.NewStyle().Padding(0, 2).Bold(true).Background(lipgloss.Color("12")).Foreground(lipgloss.Color("15"))

	styleRow         = lipgloss.NewStyle().Padding(0, 1)
	styleRowSelected = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("237"))
	styleDisabled    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleStaged      = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

	styleHelp   = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).MarginTop(1)
	styleHeader = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).MarginBottom(1)
	styleErrMsg = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleOKMsg  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

// View renders the entire TUI for the current Model state.
func (m Model) View() string {
	if m.confirmingQuit {
		return m.viewQuitConfirm()
	}

	var b strings.Builder
	b.WriteString(m.viewTabBar())
	b.WriteString("\n\n")

	switch m.activeTab {
	case TabServer, TabSkillsClaude, TabSkillsOpenCode:
		b.WriteString(m.viewItemList())
	case TabReview:
		b.WriteString(m.viewReview())
	}

	b.WriteString("\n")
	b.WriteString(m.viewFooter())

	return stylePage.Render(b.String())
}

func (m Model) viewTabBar() string {
	tabs := []Tab{TabServer, TabSkillsClaude, TabSkillsOpenCode, TabReview}
	parts := make([]string, 0, len(tabs))
	for i, t := range tabs {
		label := fmt.Sprintf("%d %s", i+1, t.String())
		if t == TabReview && len(m.desired) > 0 {
			label = fmt.Sprintf("%s (%d staged)", label, len(m.desired))
		}
		if t == m.activeTab {
			parts = append(parts, styleTabActive.Render(label))
		} else {
			parts = append(parts, styleTab.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m Model) viewItemList() string {
	rows := m.itemsForTab(m.activeTab)
	if len(rows) == 0 {
		return styleDisabled.Render("(no items on this tab)")
	}

	cur := m.cursor[m.activeTab]
	var b strings.Builder
	b.WriteString(styleHeader.Render(m.activeTab.String()))
	b.WriteString("\n")

	for i, it := range rows {
		line := m.renderRow(it, i == cur)
		b.WriteString(line)
		b.WriteString("\n")
	}

	if len(m.results) > 0 && m.activeTab != TabReview {
		// quietly note that results are available on the Review tab
		b.WriteString("\n")
		b.WriteString(styleHelp.Render(fmt.Sprintf("Last apply produced %d results — see Review tab.", len(m.results))))
	}

	return b.String()
}

func (m Model) renderRow(it Item, selected bool) string {
	state := m.current[it.ID]
	staged, hasStaged := m.desired[it.ID]

	glyph := fmtState(state)
	if hasStaged && staged != state {
		// Show the target state with an arrow.
		glyph = fmt.Sprintf("%s → %s", fmtState(state), fmtState(staged))
	}

	label := it.Label
	hint := ""
	if state == StateUnavailable {
		switch {
		case strings.HasSuffix(it.ID, ":claude"):
			hint = " — ~/.claude/ not detected (install Claude Code first)"
		case strings.HasSuffix(it.ID, ":opencode"):
			hint = " — ~/.config/opencode/ not detected (install OpenCode first)"
		default:
			hint = " — unavailable"
		}
	}

	body := fmt.Sprintf("%s %s%s", glyph, label, hint)

	switch {
	case state == StateUnavailable:
		body = styleDisabled.Render(body)
	case hasStaged && staged != state:
		body = styleStaged.Render(body)
	}

	style := styleRow
	if selected {
		style = styleRowSelected
	}
	return style.Render(body)
}

func (m Model) viewReview() string {
	var b strings.Builder
	b.WriteString(styleHeader.Render("Review pending changes"))
	b.WriteString("\n")

	staged := m.stagedChanges()

	if len(staged) == 0 && len(m.results) == 0 {
		b.WriteString(styleDisabled.Render("No changes staged. Toggle items on the Server / Skills tabs, then return here."))
		return b.String()
	}

	if len(staged) > 0 {
		b.WriteString(fmt.Sprintf("%-40s  %-15s  %-15s\n", "Component", "Before", "After"))
		b.WriteString(strings.Repeat("─", 76))
		b.WriteString("\n")
		for _, it := range staged {
			before := fmtStateName(m.current[it.ID])
			after := fmtStateName(m.desired[it.ID])
			b.WriteString(fmt.Sprintf("%-40s  %-15s  %-15s\n", it.ID, before, after))
		}
		b.WriteString("\n")
		if m.applying {
			b.WriteString(fmt.Sprintf("%s applying changes…", m.spinner.View()))
		} else {
			b.WriteString(styleHelp.Render("Press Enter to apply all changes."))
		}
	}

	if len(m.results) > 0 {
		b.WriteString("\n\n")
		b.WriteString(styleHeader.Render("Last apply results"))
		b.WriteString("\n")
		for _, r := range m.results {
			if r.OK {
				b.WriteString(styleOKMsg.Render(fmt.Sprintf("✓ %s — %s", r.ItemID, r.Action)))
			} else {
				b.WriteString(styleErrMsg.Render(fmt.Sprintf("✗ %s — %v", r.ItemID, r.Err)))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m Model) viewFooter() string {
	hints := []string{
		"tab/shift+tab cycle tabs",
		"↑/↓ nav",
		"space toggle",
	}
	if m.activeTab == TabReview && len(m.desired) > 0 {
		hints = append(hints, "enter apply")
	}
	hints = append(hints, "q quit")
	return styleHelp.Render(strings.Join(hints, "  ·  "))
}

func (m Model) viewQuitConfirm() string {
	msg := fmt.Sprintf(
		"You have %d staged change(s) that have not been applied.\n\nQuit anyway? (y/n)",
		len(m.desired),
	)
	return stylePage.Render(styleHeader.Render("Discard staged changes?") + "\n\n" + msg)
}
