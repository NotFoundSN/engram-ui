// Package cli provides the multi-subcommand dispatch for engram-ui.
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Gentleman-Programming/engram-ui/internal/tui"
)

// stdout and stderr are package-level writers, injectable for tests.
var stdout io.Writer = os.Stdout
var stderr io.Writer = os.Stderr

// runTUIFn invokes the interactive TUI. Injectable for tests.
var runTUIFn = tui.RunTUI

// isInteractiveFn reports whether stdin is a terminal. Injectable for tests.
var isInteractiveFn = isInteractive

// Dispatch routes os.Args[1:] to the appropriate subcommand handler.
// Returns the process exit code (0 = success, 1 = error, 2 = usage error).
//
// With zero arguments and an interactive stdin, dispatch launches the TUI
// installer. In non-interactive contexts (piped input, autostart, CI),
// dispatch prints help instead of auto-serving so scripted invocations
// behave predictably. Use `engram-ui serve` explicitly to run the daemon.
func Dispatch(args []string) int {
	if len(args) == 0 {
		if isInteractiveFn() {
			if err := runTUIFn(); err != nil {
				fmt.Fprintf(stderr, "engram-ui: TUI error: %v\n", err)
				return 1
			}
			return 0
		}
		printUsage(stdout)
		return 0
	}

	switch args[0] {
	case "serve":
		return cmdServe(args[1:])
	case "setup":
		return cmdSetup(args[1:])
	case "remove":
		return cmdRemove(args[1:])
	case "list":
		return cmdList(args[1:])
	case "version", "--version", "-v":
		return cmdVersion()
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	case "--no-tui":
		// Force-print help even when stdin is a TTY. Lets users discover the
		// CLI without entering the interactive installer.
		printUsage(stdout)
		return 0
	default:
		if strings.HasPrefix(args[0], "-") {
			// Flag-first invocation: implicit serve for backward compat.
			// e.g. "engram-ui --listen=:9000"
			return cmdServe(args)
		}
		fmt.Fprintf(stderr, "engram-ui: unknown subcommand %q\n", args[0])
		printUsage(stderr)
		return 2
	}
}

// isInteractive returns true when stdin is a terminal (character device).
// Non-interactive contexts include pipes, redirected files, and systemd-
// managed services.
func isInteractive() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// printUsage writes the usage text to w.
func printUsage(w io.Writer) {
	fmt.Fprintln(w, `engram-ui — web viewer for engram persistent memory

Usage:
  engram-ui                              launch interactive installer TUI (default)
  engram-ui serve [flags]                start the web UI daemon
  engram-ui setup <skill> [--tool=...]   install a skill for one or both tools
  engram-ui remove <skill> [--tool=...]  remove a skill from one or both tools
  engram-ui list [--json]                list available skills
  engram-ui version                      print version and exit
  engram-ui help                         print this help
  engram-ui --no-tui                     print help instead of launching the TUI

Serve flags:
  --engram=<url>     engram REST API base URL (default: http://localhost:7437)
  --listen=<addr>    address engram-ui listens on (default: :7438)
  --no-spawn         fail instead of auto-spawning 'engram serve'

Setup / Remove flags:
  --tool=<claude|opencode|both>   target tool (default: both)

Note: "autostart" is a reserved skill name that invokes OS autostart registration.
  --tool is ignored for autostart.`)
}
