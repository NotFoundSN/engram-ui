package cli

import (
	"fmt"

	"github.com/NotFoundSN/engram-ui/internal/version"
)

// cmdVersion prints the version string and returns exit code 0.
func cmdVersion() int {
	fmt.Fprintf(stdout, "engram-ui %s\n", version.Current())
	return 0
}
