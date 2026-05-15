package cli

import (
	"fmt"
	"runtime/debug"
)

// version is the binary version string, injected at build time by GoReleaser:
// -X github.com/Gentleman-Programming/engram-ui/internal/cli.version={{.Version}}
var version = "dev"

func init() {
	// If version is still "dev" at startup, try to get the version from
	// the module build info (populated by `go install ...@v1.2.3`).
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
}

// cmdVersion prints the version string and returns exit code 0.
func cmdVersion() int {
	fmt.Fprintf(stdout, "engram-ui %s\n", version)
	return 0
}
