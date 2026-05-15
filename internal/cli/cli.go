// Package cli provides the multi-subcommand dispatch for engram-ui.
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// stdout and stderr are package-level writers, injectable for tests.
var stdout io.Writer = os.Stdout
var stderr io.Writer = os.Stderr

// Dispatch routes os.Args[1:] to the appropriate subcommand handler.
// Returns the process exit code (0 = success, 1 = error, 2 = usage error).
func Dispatch(args []string) int {
	if len(args) == 0 {
		// No subcommand: default to serve (backward compat with v2).
		return cmdServe(args)
	}

	switch args[0] {
	case "serve":
		return cmdServe(args[1:])
	case "setup":
		return cmdSetup(args[1:])
	case "version", "--version", "-v":
		return cmdVersion()
	case "help", "--help", "-h":
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

// printUsage writes the usage text to w.
func printUsage(w io.Writer) {
	fmt.Fprintln(w, `engram-ui — web viewer for engram persistent memory

Usage:
  engram-ui [serve] [flags]    start the web UI (default)
  engram-ui setup <target>     install skills or configure OS autostart
  engram-ui version            print version and exit
  engram-ui help               print this help

Serve flags:
  --engram=<url>     engram REST API base URL (default: http://localhost:7437)
  --listen=<addr>    address engram-ui listens on (default: :7438)
  --no-spawn         fail instead of auto-spawning 'engram serve'

Setup targets:
  claude-code        install engram-conventions skill for Claude Code
  opencode           install engram-conventions skill for OpenCode
  os-autostart       register engram-ui as an OS autostart entry
  remove-autostart   remove the OS autostart entry`)
}
