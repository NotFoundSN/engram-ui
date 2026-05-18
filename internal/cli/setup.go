package cli

import (
	"fmt"

	"github.com/NotFoundSN/engram-ui/internal/installer"
)

// autostartName is the reserved skill name that routes to the OS autostart
// installer rather than a regular skill install. Documented reservation:
// if a future skill is named "autostart" it will be unreachable via setup/remove.
const autostartName = "autostart"

// Function variables for installer calls — allows test injection.

var installClaudeCodeFn = func(name string) (string, error) {
	r, err := installer.InstallClaudeCodeSkill(name)
	return r.Destination, err
}

var installOpenCodeFn = func(name string) (string, error) {
	r, err := installer.InstallOpenCodeSkill(name)
	return r.Destination, err
}

var installAutostartFn = func() (installer.Result, error) {
	return installer.InstallAutostart()
}

var removeAutostartFn = func() (string, error) {
	r, err := installer.RemoveAutostart()
	if err != nil {
		return r.Destination, err
	}
	if r.Action == installer.ActionNotRegistered {
		return "", nil // signal "not registered" to caller
	}
	return r.Destination, nil
}

var uninstallClaudeCodeFn = func(name string) (string, error) {
	r, err := installer.UninstallClaudeCodeSkill(name)
	if err != nil {
		return r.Destination, err
	}
	if r.Action == installer.ActionNotRegistered {
		return "", nil // sentinel: empty dest = "nothing to remove"
	}
	return r.Destination, nil
}

var uninstallOpenCodeFn = func(name string) (string, error) {
	r, err := installer.UninstallOpenCodeSkill(name)
	if err != nil {
		return r.Destination, err
	}
	if r.Action == installer.ActionNotRegistered {
		return "", nil // sentinel: empty dest = "nothing to remove"
	}
	return r.Destination, nil
}

// checkSkillInCatalog returns nil when name is present in the loaded catalog,
// or a descriptive error when not found. Used by cmdSetup and cmdRemove to
// map "unknown skill" to exit 2 (usage error) per REQ-1.7 / REQ-6.3.
// Note: loadCatalogFn is defined in list.go and shared across this package.
func checkSkillInCatalog(name string) error {
	skills, err := loadCatalogFn()
	if err != nil {
		return err
	}
	for _, s := range skills {
		if s.Name == name {
			return nil
		}
	}
	return fmt.Errorf("not found in catalog")
}

// cmdSetup implements the `setup` subcommand.
// Usage: engram-ui setup <skill> [--tool=<claude|opencode|both>]
func cmdSetup(args []string) int {
	name, targets, usageErr, err := parseToolFlag("setup", args)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		if usageErr {
			printSetupUsage()
			return 2
		}
		return 1
	}

	// autostart special case — tool flag is ignored.
	if name == autostartName {
		// If user passed --tool explicitly, targets will have been narrowed from
		// default ["claude","opencode"]. We detect this by seeing if the original
		// args contained a --tool= flag. We can infer it from targets length:
		// default (omitted) → ["claude","opencode"] (len 2), explicit → len 1.
		// But wait: --tool=both also gives len 2. We need a different approach.
		// Check if any arg starts with "--tool=".
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
		result, err := installAutostartFn()
		if err != nil {
			// Handle downgrade block with special messaging
			if result.Action == installer.ActionBlockedDowngrade {
				fmt.Fprintf(stderr, "engram-ui setup autostart: downgrade blocked\n")
				fmt.Fprintf(stderr, "  installed: %s\n", result.InstalledVersion)
				fmt.Fprintf(stderr, "  source:    %s\n", result.SourceVersion)
				fmt.Fprintf(stderr, "Run setup from the newer version to upgrade.\n")
				return 1
			}
			fmt.Fprintf(stderr, "engram-ui setup autostart: %v\n", err)
			return 1
		}

		// Handle success cases
		switch result.Action {
		case installer.ActionSkipped:
			if result.Notes != "" {
				fmt.Fprintf(stderr, "autostart %s: %s\n", result.Action, result.Notes)
			} else {
				fmt.Fprintf(stderr, "autostart %s\n", result.Action)
			}
		default:
			fmt.Fprintf(stderr, "registered: %s\n", result.Destination)
		}
		return 0
	}

	// Validate skill exists in catalog before attempting install.
	if err := checkSkillInCatalog(name); err != nil {
		fmt.Fprintf(stderr, "engram-ui setup: unknown skill %q — %v\n", name, err)
		return 2
	}

	// Regular skill path — iterate legs, no fail-fast.
	failed := false
	for _, leg := range targets {
		var dest string
		var legErr error
		switch leg {
		case "claude":
			dest, legErr = installClaudeCodeFn(name)
		case "opencode":
			dest, legErr = installOpenCodeFn(name)
		}
		if legErr != nil {
			fmt.Fprintf(stderr, "engram-ui setup %s [%s]: %v\n", name, leg, legErr)
			failed = true
			continue
		}
		fmt.Fprintf(stderr, "installed [%s]: %s\n", leg, dest)
	}

	if failed {
		return 1
	}
	return 0
}

func printSetupUsage() {
	fmt.Fprintln(stderr, `Usage: engram-ui setup <skill> [--tool=<claude|opencode|both>]

Arguments:
  <skill>              name of the skill to install (or "autostart")

Options:
  --tool=<value>       target tool: claude, opencode, or both (default: both)

The --tool flag is ignored when skill is "autostart".`)
}
