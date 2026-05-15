// Package installer pure unit/plist/bat generators (no build tags — testable on all OSes).
package installer

import "fmt"

// BuildWindowsStartupBat returns a Windows startup batch file for engram-ui.
// The batch file runs engram-ui detached using `start /B` so no console window stays.
func BuildWindowsStartupBat(execPath string) string {
	return fmt.Sprintf("@echo off\nstart \"\" /B \"%s\" serve\n", execPath)
}

// BuildLaunchAgentPlist returns a macOS LaunchAgent plist for engram-ui.
func BuildLaunchAgentPlist(execPath, label string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>serve</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<false/>
</dict>
</plist>
`, label, execPath)
}

// BuildSystemdUnit returns a systemd user unit file for engram-ui.
func BuildSystemdUnit(execPath string) string {
	return fmt.Sprintf(`[Unit]
Description=engram-ui web viewer
After=network.target

[Service]
Type=simple
ExecStart=%s serve
Restart=on-failure

[Install]
WantedBy=default.target
`, execPath)
}
