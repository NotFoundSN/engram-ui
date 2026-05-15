package cli

import (
	"fmt"

	"github.com/Gentleman-Programming/engram-ui/internal/installer"
)

// Function variables for installer calls — allows test injection.
var installClaudeCodeFn = func() (string, error) {
	r, err := installer.InstallClaudeCodeSkill()
	return r.Destination, err
}

var installOpenCodeFn = func() (string, error) {
	r, err := installer.InstallOpenCodeSkill()
	return r.Destination, err
}

var installAutostartFn = func() (string, error) {
	r, err := installer.InstallAutostart()
	return r.Destination, err
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

// cmdSetup routes the setup subcommand arguments to the appropriate installer function.
func cmdSetup(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(stderr, "engram-ui setup: missing target\n")
		printSetupUsage()
		return 2
	}

	switch args[0] {
	case "claude-code":
		dest, err := installClaudeCodeFn()
		if err != nil {
			fmt.Fprintf(stderr, "engram-ui setup claude-code: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "installed: %s\n", dest)
		return 0

	case "opencode":
		dest, err := installOpenCodeFn()
		if err != nil {
			fmt.Fprintf(stderr, "engram-ui setup opencode: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "installed: %s\n", dest)
		return 0

	case "os-autostart":
		dest, err := installAutostartFn()
		if err != nil {
			fmt.Fprintf(stderr, "engram-ui setup os-autostart: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "registered: %s\n", dest)
		return 0

	case "remove-autostart":
		dest, err := removeAutostartFn()
		if err != nil {
			fmt.Fprintf(stderr, "engram-ui setup remove-autostart: %v\n", err)
			return 1
		}
		if dest == "" {
			fmt.Fprintln(stdout, "not currently registered (nothing to remove)")
		} else {
			fmt.Fprintf(stdout, "removed: %s\n", dest)
		}
		return 0

	default:
		fmt.Fprintf(stderr, "engram-ui setup: unknown target %q\n", args[0])
		printSetupUsage()
		return 2
	}
}

func printSetupUsage() {
	fmt.Fprintln(stderr, `Usage: engram-ui setup <target>

Targets:
  claude-code        install engram-conventions skill for Claude Code
  opencode           install engram-conventions skill for OpenCode
  os-autostart       register engram-ui as an OS autostart entry
  remove-autostart   remove the OS autostart entry`)
}
