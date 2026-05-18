// Package installer path helpers — pure, no build tags, testable on all OSes.
package installer

import (
	"path/filepath"
	"strings"
)

// StableBinaryPath returns the stable binary path for the given OS.
// Returns empty string for unsupported platforms.
func StableBinaryPath(homeDir, localAppData, goos string) string {
	switch goos {
	case "windows":
		return filepath.Join(localAppData, "engram-ui", "engram-ui.exe")
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "engram-ui", "engram-ui")
	case "linux":
		return filepath.Join(homeDir, ".local", "bin", "engram-ui")
	default:
		return ""
	}
}

// IsStableBinaryPath returns true if the given path is already a stable path.
// It checks against known stable path prefixes for each OS.
func IsStableBinaryPath(path, homeDir, localAppData, goos string) bool {
	if path == "" {
		return false
	}

	// Normalize path separators for comparison
	path = filepath.ToSlash(path)

	// Get the expected stable path for comparison
	expectedStable := filepath.ToSlash(StableBinaryPath(homeDir, localAppData, goos))
	if expectedStable != "" && path == expectedStable {
		return true
	}

	// Check against stable prefixes
	prefixes := StableBinaryPrefixes(goos, homeDir, localAppData)
	for _, prefix := range prefixes {
		prefix = filepath.ToSlash(prefix)
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// StableBinaryPrefixes returns the known stable path prefixes for the given OS.
// Prefixes are concrete paths (env vars and ~ are resolved using homeDir/localAppData)
// so they can be matched directly against absolute paths with strings.HasPrefix.
func StableBinaryPrefixes(goos, homeDir, localAppData string) []string {
	switch goos {
	case "windows":
		return []string{
			localAppData + `\engram-ui\`,
			`C:\Program Files\`,
			`C:\Program Files (x86)\`,
		}
	case "darwin":
		return []string{
			"/opt/homebrew/bin/",
			"/usr/local/bin/",
			homeDir + "/.local/bin/",
		}
	case "linux":
		return []string{
			"/usr/local/bin/",
			"/opt/homebrew/bin/",
			homeDir + "/.local/bin/",
		}
	default:
		return nil
	}
}

// ClaudeSkillDir returns the destination root for the named skill under
// Claude Code: {home}/.claude/skills/{name}
func ClaudeSkillDir(homeDir, name string) string {
	return filepath.Join(homeDir, ".claude", "skills", name)
}

// OpenCodeSkillDir returns: {base}/opencode/skills/{name}
// where base = xdgConfigHome if non-empty, else {homeDir}/.config.
// OpenCode uses ~/.config/opencode/ on ALL platforms (including macOS/Windows),
// ignoring os.UserConfigDir() which would return Library/Application Support on macOS
// and %APPDATA% on Windows. This mirrors engram's setup_test.go contract.
func OpenCodeSkillDir(homeDir, xdgConfigHome, name string) string {
	base := xdgConfigHome
	if base == "" {
		base = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(base, "opencode", "skills", name)
}

// WindowsStartupDir returns the Windows Startup folder path:
// {appData}/Microsoft/Windows/Start Menu/Programs/Startup
func WindowsStartupDir(appData string) string {
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
}

// MacOSLaunchAgentPath returns the plist destination path:
// {home}/Library/LaunchAgents/com.notfoundsn.engram-ui.plist
func MacOSLaunchAgentPath(homeDir string) string {
	return filepath.Join(homeDir, "Library", "LaunchAgents", "com.notfoundsn.engram-ui.plist")
}

// LinuxSystemdUnitPath returns: {base}/systemd/user/engram-ui.service
// where base = xdgConfigHome if non-empty, else {homeDir}/.config
func LinuxSystemdUnitPath(homeDir, xdgConfigHome string) string {
	base := xdgConfigHome
	if base == "" {
		base = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(base, "systemd", "user", "engram-ui.service")
}
