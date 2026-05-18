package installer

import (
	"os"
	"runtime"
)

// AutostartManager is the platform-agnostic surface for OS autostart install/remove.
// Implementations live in autostart_{windows,darwin,linux,other}.go behind build tags.
type AutostartManager interface {
	Install(execPath string) (Result, error)
	Remove() (Result, error)
}

// NewAutostartManager returns the per-OS implementation.
// Selected at compile time via build tags.
// Declared here (without body) — body provided by each autostart_*.go file.

// InstallAutostart resolves the current executable path, installs or validates a stable
// user-scoped binary, then installs an OS autostart entry pointing to that stable binary.
// If the stable binary installation fails or a downgrade is detected, it returns an error
// without registering the autostart entry.
func InstallAutostart() (Result, error) {
	execPath, err := ResolveExecPath()
	if err != nil {
		return Result{}, err
	}

	// Get home directory for stable path resolution
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}

	// Get LOCALAPPDATA on Windows
	localAppData := ""
	if runtime.GOOS == "windows" {
		localAppData = os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return Result{}, os.ErrNotExist
		}
	}

	// Ensure stable binary before registering autostart
	stableResult, err := EnsureStableBinary(execPath, homeDir, localAppData)
	if err != nil {
		// Return the result with error details (includes downgrade info)
		return stableResult, err
	}

	// Use the stable path for autostart registration
	stablePath := stableResult.Destination
	if stablePath == "" {
		stablePath = StableBinaryPath(homeDir, localAppData, runtime.GOOS)
	}

	// Register autostart entry with stable path
	mgrResult, err := NewAutostartManager().Install(stablePath)
	if err != nil {
		return mgrResult, err
	}

	// Preserve stable install notes if we skipped or performed special actions
	if stableResult.Notes != "" && mgrResult.Notes == "" {
		mgrResult.Notes = stableResult.Notes
	}
	mgrResult.SourceVersion = stableResult.SourceVersion
	mgrResult.InstalledVersion = stableResult.InstalledVersion

	return mgrResult, nil
}

// RemoveAutostart removes the OS autostart entry for engram-ui.
func RemoveAutostart() (Result, error) {
	return NewAutostartManager().Remove()
}
