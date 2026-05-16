package cli

import (
	"fmt"
)

// cmdRemove implements the `remove` subcommand.
// Usage: engram-ui remove <skill> [--tool=<claude|opencode|both>]
func cmdRemove(args []string) int {
	name, targets, usageErr, err := parseToolFlag("remove", args)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		if usageErr {
			printRemoveUsage()
			return 2
		}
		return 1
	}

	// autostart special case — tool flag is ignored.
	if name == autostartName {
		toolExplicit := false
		for _, a := range args {
			if len(a) >= 7 && a[:7] == "--tool=" {
				toolExplicit = true
				break
			}
		}
		if toolExplicit {
			fmt.Fprintf(stderr, "note: --tool flag is ignored for autostart\n")
		}
		dest, err := removeAutostartFn()
		if err != nil {
			fmt.Fprintf(stderr, "engram-ui remove autostart: %v\n", err)
			return 1
		}
		if dest == "" {
			fmt.Fprintln(stderr, "not currently registered (nothing to remove)")
		} else {
			fmt.Fprintf(stderr, "removed: %s\n", dest)
		}
		return 0
	}

	// Regular skill path — iterate legs, no fail-fast.
	failed := false
	for _, leg := range targets {
		var dest string
		var legErr error
		switch leg {
		case "claude":
			dest, legErr = uninstallClaudeCodeFn(name)
		case "opencode":
			dest, legErr = uninstallOpenCodeFn(name)
		}
		if legErr != nil {
			fmt.Fprintf(stderr, "engram-ui remove %s [%s]: %v\n", name, leg, legErr)
			failed = true
			continue
		}
		if dest == "" {
			// Not-registered sentinel.
			fmt.Fprintf(stderr, "not registered [%s]: %s\n", leg, name)
		} else {
			fmt.Fprintf(stderr, "removed [%s]: %s\n", leg, dest)
		}
	}

	if failed {
		return 1
	}
	return 0
}

func printRemoveUsage() {
	fmt.Fprintln(stderr, `Usage: engram-ui remove <skill> [--tool=<claude|opencode|both>]

Arguments:
  <skill>              name of the skill to remove (or "autostart")

Options:
  --tool=<value>       target tool: claude, opencode, or both (default: both)

The --tool flag is ignored when skill is "autostart".`)
}
