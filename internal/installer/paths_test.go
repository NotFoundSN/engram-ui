package installer

import (
	"path/filepath"
	"testing"
)

func TestClaudeSkillDir(t *testing.T) {
	cases := []struct {
		name    string
		homeDir string
		want    string
	}{
		{
			name:    "unix-style home",
			homeDir: "/home/user",
			want:    filepath.Join("/home/user", ".claude", "skills", "engram-conventions"),
		},
		{
			name:    "windows home",
			homeDir: `C:\Users\user`,
			want:    filepath.Join(`C:\Users\user`, ".claude", "skills", "engram-conventions"),
		},
		{
			name:    "macos home",
			homeDir: "/Users/user",
			want:    filepath.Join("/Users/user", ".claude", "skills", "engram-conventions"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClaudeSkillDir(tc.homeDir)
			if got != tc.want {
				t.Errorf("ClaudeSkillDir(%q) = %q, want %q", tc.homeDir, got, tc.want)
			}
		})
	}
}

func TestOpenCodeSkillDir(t *testing.T) {
	cases := []struct {
		name          string
		homeDir       string
		xdgConfigHome string
		want          string
	}{
		{
			name:          "linux with XDG_CONFIG_HOME set",
			homeDir:       "/home/user",
			xdgConfigHome: "/home/user/.config",
			want:          filepath.Join("/home/user/.config", "opencode", "skills", "engram-conventions"),
		},
		{
			name:          "linux without XDG — falls back to home/.config",
			homeDir:       "/home/user",
			xdgConfigHome: "",
			want:          filepath.Join("/home/user", ".config", "opencode", "skills", "engram-conventions"),
		},
		{
			name:          "macOS — always uses home/.config (not Library/Application Support)",
			homeDir:       "/Users/user",
			xdgConfigHome: "",
			want:          filepath.Join("/Users/user", ".config", "opencode", "skills", "engram-conventions"),
		},
		{
			name:          "windows — always uses home/.config (not APPDATA)",
			homeDir:       `C:\Users\user`,
			xdgConfigHome: "",
			want:          filepath.Join(`C:\Users\user`, ".config", "opencode", "skills", "engram-conventions"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := OpenCodeSkillDir(tc.homeDir, tc.xdgConfigHome)
			if got != tc.want {
				t.Errorf("OpenCodeSkillDir(%q, %q) = %q, want %q", tc.homeDir, tc.xdgConfigHome, got, tc.want)
			}
		})
	}
}

func TestWindowsStartupDir(t *testing.T) {
	cases := []struct {
		name    string
		appData string
		want    string
	}{
		{
			name:    "standard APPDATA",
			appData: `C:\Users\user\AppData\Roaming`,
			want:    filepath.Join(`C:\Users\user\AppData\Roaming`, "Microsoft", "Windows", "Start Menu", "Programs", "Startup"),
		},
		{
			name:    "custom APPDATA",
			appData: `D:\Custom\AppData`,
			want:    filepath.Join(`D:\Custom\AppData`, "Microsoft", "Windows", "Start Menu", "Programs", "Startup"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WindowsStartupDir(tc.appData)
			if got != tc.want {
				t.Errorf("WindowsStartupDir(%q) = %q, want %q", tc.appData, got, tc.want)
			}
		})
	}
}

func TestMacOSLaunchAgentPath(t *testing.T) {
	cases := []struct {
		name    string
		homeDir string
		want    string
	}{
		{
			name:    "standard macOS home",
			homeDir: "/Users/user",
			want:    filepath.Join("/Users/user", "Library", "LaunchAgents", "com.gentleman-programming.engram-ui.plist"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MacOSLaunchAgentPath(tc.homeDir)
			if got != tc.want {
				t.Errorf("MacOSLaunchAgentPath(%q) = %q, want %q", tc.homeDir, got, tc.want)
			}
		})
	}
}

func TestLinuxSystemdUnitPath(t *testing.T) {
	cases := []struct {
		name          string
		homeDir       string
		xdgConfigHome string
		want          string
	}{
		{
			name:          "default (no XDG override)",
			homeDir:       "/home/user",
			xdgConfigHome: "",
			want:          filepath.Join("/home/user", ".config", "systemd", "user", "engram-ui.service"),
		},
		{
			name:          "with XDG_CONFIG_HOME override",
			homeDir:       "/home/user",
			xdgConfigHome: "/home/user/.config",
			want:          filepath.Join("/home/user/.config", "systemd", "user", "engram-ui.service"),
		},
		{
			name:          "non-default XDG_CONFIG_HOME",
			homeDir:       "/home/user",
			xdgConfigHome: "/custom/config",
			want:          filepath.Join("/custom/config", "systemd", "user", "engram-ui.service"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := LinuxSystemdUnitPath(tc.homeDir, tc.xdgConfigHome)
			if got != tc.want {
				t.Errorf("LinuxSystemdUnitPath(%q, %q) = %q, want %q", tc.homeDir, tc.xdgConfigHome, got, tc.want)
			}
		})
	}
}
