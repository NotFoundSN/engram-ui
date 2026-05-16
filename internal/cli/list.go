package cli

import (
	"encoding/json"
	"fmt"

	"github.com/Gentleman-Programming/engram-ui/internal/installer"
)

// loadCatalogFn is the injectable catalog loader for tests.
var loadCatalogFn = installer.LoadCatalog

// listEntry is the JSON representation of a skill in the list output.
type listEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// cmdList implements the `list` subcommand.
// Usage: engram-ui list [--json]
func cmdList(args []string) int {
	// Parse --json flag manually (only one flag to handle).
	asJSON := false
	for _, arg := range args {
		switch arg {
		case "--json":
			asJSON = true
		default:
			if len(arg) > 0 && arg[0] == '-' {
				fmt.Fprintf(stderr, "engram-ui list: unknown flag %q\n", arg)
				printListUsage()
				return 2
			}
			// Positional argument — not expected.
			fmt.Fprintf(stderr, "engram-ui list: unexpected argument %q\n", arg)
			printListUsage()
			return 2
		}
	}

	skills, err := loadCatalogFn()
	if err != nil {
		fmt.Fprintf(stderr, "engram-ui list: %v\n", err)
		return 1
	}

	if asJSON {
		entries := make([]listEntry, len(skills))
		for i, s := range skills {
			entries[i] = listEntry{Name: s.Name, Description: s.Description}
		}
		enc := json.NewEncoder(stdout)
		if err := enc.Encode(entries); err != nil {
			fmt.Fprintf(stderr, "engram-ui list: json encode: %v\n", err)
			return 1
		}
		return 0
	}

	for _, s := range skills {
		// REQ-3.2: always emit name — description format; empty description
		// produces a trailing " — " which is spec-literal compliant.
		fmt.Fprintf(stdout, "%s — %s\n", s.Name, s.Description)
	}
	return 0
}

func printListUsage() {
	fmt.Fprintln(stderr, `Usage: engram-ui list [--json]

Options:
  --json    emit machine-readable JSON array to stdout`)
}
