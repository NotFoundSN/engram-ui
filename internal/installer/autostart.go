package installer

// AutostartManager is the platform-agnostic surface for OS autostart install/remove.
// Implementations live in autostart_{windows,darwin,linux,other}.go behind build tags.
type AutostartManager interface {
	Install(execPath string) (Result, error)
	Remove() (Result, error)
}

// NewAutostartManager returns the per-OS implementation.
// Selected at compile time via build tags.
// Declared here (without body) — body provided by each autostart_*.go file.

// InstallAutostart resolves the current executable path and installs an OS autostart entry.
func InstallAutostart() (Result, error) {
	execPath, err := ResolveExecPath()
	if err != nil {
		return Result{}, err
	}
	return NewAutostartManager().Install(execPath)
}

// RemoveAutostart removes the OS autostart entry for engram-ui.
func RemoveAutostart() (Result, error) {
	return NewAutostartManager().Remove()
}
