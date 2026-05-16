// Package tui implements the interactive installer TUI for engram-ui.
//
// The package is split into three concerns:
//   - items.go — static catalog of installable components (this file)
//   - detect.go — read filesystem to determine CurrentState per item
//   - apply.go — execute install/uninstall actions in continue+summary mode
//
// The Bubbletea wiring lives in program.go (Commit B). The CLI entry that
// dispatches the engram-ui no-args invocation to this TUI lives in
// internal/cli/ (Commit C).
package tui

import (
	"fmt"

	"github.com/Gentleman-Programming/engram-ui/internal/installer"
)

// State describes the lifecycle position of one installable Item.
type State int

const (
	// StateUnknown is the zero value; reserved so a forgotten initialisation
	// does not silently look like NotInstalled.
	StateUnknown State = iota
	// StateInstalled means the destination artifact exists on disk.
	StateInstalled
	// StateNotInstalled means the destination's parent (the tool's skills
	// directory) exists but no artifact for this item is present.
	StateNotInstalled
	// StateUnavailable means the prerequisite tool (Claude Code, OpenCode)
	// is not detected on this system. The item is shown but disabled.
	StateUnavailable
)

// String returns a short human-readable label for the State.
func (s State) String() string {
	switch s {
	case StateInstalled:
		return "installed"
	case StateNotInstalled:
		return "not installed"
	case StateUnavailable:
		return "unavailable"
	default:
		return "unknown"
	}
}

// Tab identifies which TUI tab the Item belongs to.
type Tab int

const (
	TabServer Tab = iota
	TabSkillsClaude
	TabSkillsOpenCode
	TabReview
)

// String returns the tab label rendered in the TUI header.
func (t Tab) String() string {
	switch t {
	case TabServer:
		return "Server"
	case TabSkillsClaude:
		return "Skills · Claude Code"
	case TabSkillsOpenCode:
		return "Skills · OpenCode"
	case TabReview:
		return "Review"
	default:
		return ""
	}
}

// Item is one installable component visible in the TUI.
type Item struct {
	// ID is a stable identifier. Formats:
	//   - "autostart"                       — server autostart entry
	//   - "skill:{name}:claude"             — skill installed for Claude Code
	//   - "skill:{name}:opencode"           — skill installed for OpenCode
	ID string

	// Tab is the TUI tab the item is rendered under.
	Tab Tab

	// Label is the human-readable row title.
	Label string

	// Description is the longer subtitle / hint.
	Description string

	// SkillName is set for skill items; empty for the autostart row.
	SkillName string
}

// BuildCatalog returns the full set of installable items: the autostart row
// plus one row per (skill × tool) combination discovered from the installer's
// embedded skill catalog. Order: autostart first, then skills sorted by name
// with claude variant before opencode variant.
func BuildCatalog() []Item {
	skills, err := installer.LoadCatalog()
	if err != nil {
		// LoadCatalog failure is exceptional (embed FS is statically compiled
		// in) — return only the autostart row so the TUI still loads.
		skills = nil
	}

	items := make([]Item, 0, 1+len(skills)*2)
	items = append(items, Item{
		ID:          "autostart",
		Tab:         TabServer,
		Label:       "Run engram-ui automatically on login",
		Description: "Registers engram-ui serve as an OS autostart entry (Win/Mac/Linux).",
	})

	for _, s := range skills {
		items = append(items, Item{
			ID:          fmt.Sprintf("skill:%s:claude", s.Name),
			Tab:         TabSkillsClaude,
			Label:       s.Name,
			Description: s.Description,
			SkillName:   s.Name,
		})
		items = append(items, Item{
			ID:          fmt.Sprintf("skill:%s:opencode", s.Name),
			Tab:         TabSkillsOpenCode,
			Label:       s.Name,
			Description: s.Description,
			SkillName:   s.Name,
		})
	}

	return items
}
