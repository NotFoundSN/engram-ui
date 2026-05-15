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
)

// Result is returned from every installer entrypoint.
type Result struct {
	Destination string // absolute path written or removed
	Action      Action // human-readable verb
	Notes       string // optional next-step note (e.g. "log out and back in")
}

// ErrUnsupportedPlatform is returned on plan9/openbsd/etc.
var ErrUnsupportedPlatform = errors.New("platform not supported")

// evalSymlinks is a function variable so tests can stub it.
var evalSymlinks = evalSymlinksReal

// ResolveExecPath returns the absolute path to the current executable,
// with symlinks resolved on Linux (for systemd ExecStart correctness).
func ResolveExecPath() (string, error) {
	return evalSymlinks()
}
