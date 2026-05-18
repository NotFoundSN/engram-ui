// Package version provides shared version information and parsing utilities.
package version

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
)

// version is the binary version string, injected at build time by GoReleaser:
// -X github.com/NotFoundSN/engram-ui/internal/version.version={{.Version}}
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

// Current returns the current version string.
func Current() string {
	return version
}

// Parse extracts major, minor, and patch version numbers from a version string.
// Supports formats like "v1.2.3" or "1.2.3". Returns an error for non-semver formats.
func Parse(v string) (major, minor, patch int, err error) {
	// Trim 'v' prefix if present
	v = strings.TrimPrefix(v, "v")

	// Handle dev/unknown versions
	if v == "dev" || v == "" {
		return 0, 0, 0, fmt.Errorf("cannot parse dev or empty version")
	}

	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %q", v)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %w", err)
	}

	return major, minor, patch, nil
}

// Compare compares two version strings.
// Returns:
//   - negative if v1 < v2
//   - 0 if v1 == v2
//   - positive if v1 > v2
//
// Handles "dev" versions specially: dev is considered less than any release version.
func Compare(v1, v2 string) int {
	// Handle dev versions
	isDev1 := v1 == "dev" || v1 == ""
	isDev2 := v2 == "dev" || v2 == ""

	if isDev1 && isDev2 {
		return 0
	}
	if isDev1 {
		return -1
	}
	if isDev2 {
		return 1
	}

	// Parse both versions
	maj1, min1, pat1, err1 := Parse(v1)
	maj2, min2, pat2, err2 := Parse(v2)

	// If either fails to parse, fall back to string comparison
	if err1 != nil || err2 != nil {
		return strings.Compare(v1, v2)
	}

	// Compare major
	if maj1 != maj2 {
		return maj1 - maj2
	}
	// Compare minor
	if min1 != min2 {
		return min1 - min2
	}
	// Compare patch
	return pat1 - pat2
}
