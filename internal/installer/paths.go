// Package installer path helpers — pure, no build tags, testable on all OSes.
package installer

import "path/filepath"

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
// {home}/Library/LaunchAgents/com.gentleman-programming.engram-ui.plist
func MacOSLaunchAgentPath(homeDir string) string {
	return filepath.Join(homeDir, "Library", "LaunchAgents", "com.gentleman-programming.engram-ui.plist")
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
