// Package installer provides skill installation and OS autostart registration.
package installer

import "errors"

// Action describes what the installer did.
type Action string

const (
	ActionInstalled           Action = "installed"
	ActionOverwritten         Action = "installed (overwritten)"
	ActionRemoved             Action = "removed"
	ActionNotRegistered       Action = "not registered"
	ActionUnsupportedPlatform Action = "unsupported platform"
	ActionSkipped             Action = "skipped"
	ActionBlockedDowngrade    Action = "blocked (downgrade)"
)

// Result is returned from every installer entrypoint.
type Result struct {
	Destination      string // absolute path written or removed
	Action           Action // human-readable verb
	Notes            string // optional next-step note (e.g. "log out and back in")
	SourceVersion    string // version of the source binary (for debugging)
	InstalledVersion string // version of the installed binary (for debugging)
}

// ErrUnsupportedPlatform is returned on plan9/openbsd/etc.
var ErrUnsupportedPlatform = errors.New("platform not supported")

// ErrDowngradeBlocked is returned when attempting to install an older version over a newer one.
var ErrDowngradeBlocked = errors.New("stable binary downgrade blocked")

// evalSymlinks is a function variable so tests can stub it.
var evalSymlinks = evalSymlinksReal

// ResolveExecPath returns the absolute path to the current executable,
// with symlinks resolved on Linux (for systemd ExecStart correctness).
func ResolveExecPath() (string, error) {
	return evalSymlinks()
}
