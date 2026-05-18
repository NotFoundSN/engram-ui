package tui

import (
	"fmt"
	"strings"

	"github.com/NotFoundSN/engram-ui/internal/installer"
)

// Result captures the outcome of applying one staged change.
type Result struct {
	ItemID string // Item.ID this result belongs to
	OK     bool   // true on success, false on failure
	Err    error  // non-nil when OK=false
	Action string // human-readable verb: "installed", "removed", "no change", ...
}

// Apply executes every staged change in desired against the current state
// derived from the items + filesystem. It implements continue+summary
// semantics: a failure on one item does NOT abort processing of the others.
//
// Empty desired = no work; the returned slice is empty.
//
// The home + xdg arguments allow tests to operate against a tmpdir without
// mutating the real home directory.
func Apply(items []Item, desired map[string]State, home, xdg string) []Result {
	if len(desired) == 0 {
		return nil
	}

	itemByID := make(map[string]Item, len(items))
	for _, it := range items {
		itemByID[it.ID] = it
	}

	results := make([]Result, 0, len(desired))
	for id, want := range desired {
		it, known := itemByID[id]
		if !known {
			results = append(results, Result{
				ItemID: id,
				OK:     false,
				Err:    fmt.Errorf("unknown item id %q", id),
			})
			continue
		}

		current := DetectState(it, home, xdg)
		if current == want {
			results = append(results, Result{ItemID: id, OK: true, Action: "no change"})
			continue
		}

		results = append(results, applyOne(it, want))
	}

	return results
}

func applyOne(item Item, want State) Result {
	switch {
	case item.ID == "autostart":
		return applyAutostart(item, want)
	case strings.HasSuffix(item.ID, ":claude"):
		return applyClaudeSkill(item, want)
	case strings.HasSuffix(item.ID, ":opencode"):
		return applyOpenCodeSkill(item, want)
	default:
		return Result{ItemID: item.ID, OK: false, Err: fmt.Errorf("no handler for item %q", item.ID)}
	}
}

func applyAutostart(item Item, want State) Result {
	switch want {
	case StateInstalled:
		r, err := installer.InstallAutostart()
		if err != nil {
			return Result{ItemID: item.ID, OK: false, Err: err}
		}
		return Result{ItemID: item.ID, OK: true, Action: string(r.Action)}
	case StateNotInstalled:
		r, err := installer.RemoveAutostart()
		if err != nil {
			return Result{ItemID: item.ID, OK: false, Err: err}
		}
		return Result{ItemID: item.ID, OK: true, Action: string(r.Action)}
	default:
		return Result{ItemID: item.ID, OK: false, Err: fmt.Errorf("invalid desired state %v for autostart", want)}
	}
}

func applyClaudeSkill(item Item, want State) Result {
	switch want {
	case StateInstalled:
		r, err := installer.InstallClaudeCodeSkill(item.SkillName)
		if err != nil {
			return Result{ItemID: item.ID, OK: false, Err: err}
		}
		return Result{ItemID: item.ID, OK: true, Action: string(r.Action)}
	case StateNotInstalled:
		r, err := installer.UninstallClaudeCodeSkill(item.SkillName)
		if err != nil {
			return Result{ItemID: item.ID, OK: false, Err: err}
		}
		return Result{ItemID: item.ID, OK: true, Action: string(r.Action)}
	default:
		return Result{ItemID: item.ID, OK: false, Err: fmt.Errorf("invalid desired state %v for skill", want)}
	}
}

func applyOpenCodeSkill(item Item, want State) Result {
	switch want {
	case StateInstalled:
		r, err := installer.InstallOpenCodeSkill(item.SkillName)
		if err != nil {
			return Result{ItemID: item.ID, OK: false, Err: err}
		}
		return Result{ItemID: item.ID, OK: true, Action: string(r.Action)}
	case StateNotInstalled:
		r, err := installer.UninstallOpenCodeSkill(item.SkillName)
		if err != nil {
			return Result{ItemID: item.ID, OK: false, Err: err}
		}
		return Result{ItemID: item.ID, OK: true, Action: string(r.Action)}
	default:
		return Result{ItemID: item.ID, OK: false, Err: fmt.Errorf("invalid desired state %v for skill", want)}
	}
}
