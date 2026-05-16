package cli

import (
	"fmt"
)

// parseToolFlag parses a setup/remove verb's argv tail. It expects exactly
// one positional argument (the skill or autostart name) plus an optional
// --tool=<claude|opencode|both> flag (default "both"). targets is the
// expanded slice of legs to act on: ["claude"], ["opencode"], or ["claude","opencode"].
//
// Errors are returned as a plain error; the caller decides whether to map to
// exit 2 (usage) and how to format the message. usageErr is true when the
// problem is a usage error (missing positional, unknown flag, bad value);
// false would be reserved for unexpected internal errors (none today).
func parseToolFlag(verb string, args []string) (name string, targets []string, usageErr bool, err error) {
	// Manually parse args to implement --tool=value form only (no space-separated).
	// We still use flag.NewFlagSet for proper flag handling but we intercept
	// to enforce the allow-list.
	tool := "both"

	positionals := make([]string, 0, len(args))
	for _, arg := range args {
		if len(arg) >= 7 && arg[:7] == "--tool=" {
			tool = arg[7:]
			continue
		}
		// Reject any other flags (flags that start with - but aren't --tool=)
		if len(arg) > 0 && arg[0] == '-' {
			return "", nil, true, fmt.Errorf("engram-ui %s: unknown flag %q", verb, arg)
		}
		positionals = append(positionals, arg)
	}

	// Validate tool value.
	switch tool {
	case "claude":
		targets = []string{"claude"}
	case "opencode":
		targets = []string{"opencode"}
	case "both":
		targets = []string{"claude", "opencode"}
	default:
		return "", nil, true, fmt.Errorf("engram-ui %s: invalid --tool value %q (must be claude, opencode, or both)", verb, tool)
	}

	// Require exactly one positional.
	if len(positionals) == 0 {
		return "", nil, true, fmt.Errorf("engram-ui %s: missing skill name", verb)
	}
	if len(positionals) > 1 {
		return "", nil, true, fmt.Errorf("engram-ui %s: unexpected argument %q", verb, positionals[1])
	}

	return positionals[0], targets, false, nil
}
