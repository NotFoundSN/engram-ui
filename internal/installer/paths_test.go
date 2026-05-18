package installer

import (
	"path/filepath"
	"testing"
)

func TestClaudeSkillDir(t *testing.T) {
	cases := []struct {
		name     string
		homeDir  string
		skill    string
		want     string
	}{
		{
			name:    "unix-style home with brainstorm",
			homeDir: "/home/user",
			skill:   "brainstorm",
			want:    filepath.Join("/home/user", ".claude", "skills", "brainstorm"),
		},
		{
			name:    "windows home with debug",
			homeDir: `C:\Users\user`,
			skill:   "debug",
			want:    filepath.Join(`C:\Users\user`, ".claude", "skills", "debug"),
		},
		{
			name:    "macos home with engram-conventions",
			homeDir: "/Users/user",
			skill:   "engram-conventions",
			want:    filepath.Join("/Users/user", ".claude", "skills", "engram-conventions"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClaudeSkillDir(tc.homeDir, tc.skill)
			if got != tc.want {
				t.Errorf("ClaudeSkillDir(%q, %q) = %q, want %q", tc.homeDir, tc.skill, got, tc.want)
			}
		})
	}
}

func TestOpenCodeSkillDir(t *testing.T) {
	cases := []struct {
		name          string
		homeDir       string
		xdgConfigHome string
		skill         string
		want          string
	}{
		{
			name:          "linux with XDG_CONFIG_HOME set",
			homeDir:       "/home/user",
			xdgConfigHome: "/home/user/.config",
			skill:         "brainstorm",
			want:          filepath.Join("/home/user/.config", "opencode", "skills", "brainstorm"),
		},
		{
			name:          "linux without XDG — falls back to home/.config",
			homeDir:       "/home/user",
			xdgConfigHome: "",
			skill:         "debug",
			want:          filepath.Join("/home/user", ".config", "opencode", "skills", "debug"),
		},
		{
			name:          "macOS — always uses home/.config (not Library/Application Support)",
			homeDir:       "/Users/user",
			xdgConfigHome: "",
			skill:         "engram-conventions",
			want:          filepath.Join("/Users/user", ".config", "opencode", "skills", "engram-conventions"),
		},
		{
			name:          "windows — always uses home/.config (not APPDATA)",
			homeDir:       `C:\Users\user`,
			xdgConfigHome: "",
			skill:         "brainstorm",
			want:          filepath.Join(`C:\Users\user`, ".config", "opencode", "skills", "brainstorm"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := OpenCodeSkillDir(tc.homeDir, tc.xdgConfigHome, tc.skill)
			if got != tc.want {
				t.Errorf("OpenCodeSkillDir(%q, %q, %q) = %q, want %q", tc.homeDir, tc.xdgConfigHome, tc.skill, got, tc.want)
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
			want:    filepath.Join("/Users/user", "Library", "LaunchAgents", "com.notfoundsn.engram-ui.plist"),
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

func TestStableBinaryPath(t *testing.T) {
	cases := []struct {
		name         string
		homeDir      string
		localAppData string
		goos         string
		want         string
	}{
		{
			name:         "Windows with LOCALAPPDATA",
			homeDir:      `C:\Users\user`,
			localAppData: `C:\Users\user\AppData\Local`,
			goos:         "windows",
			want:         filepath.Join(`C:\Users\user\AppData\Local`, "engram-ui", "engram-ui.exe"),
		},
		{
			name:         "macOS",
			homeDir:      "/Users/user",
			localAppData: "",
			goos:         "darwin",
			want:         filepath.Join("/Users/user", "Library", "Application Support", "engram-ui", "engram-ui"),
		},
		{
			name:         "Linux",
			homeDir:      "/home/user",
			localAppData: "",
			goos:         "linux",
			want:         filepath.Join("/home/user", ".local", "bin", "engram-ui"),
		},
		{
			name:         "unsupported platform",
			homeDir:      "/home/user",
			localAppData: "",
			goos:         "plan9",
			want:         "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StableBinaryPath(tc.homeDir, tc.localAppData, tc.goos)
			if got != tc.want {
				t.Errorf("StableBinaryPath(%q, %q, %q) = %q, want %q", tc.homeDir, tc.localAppData, tc.goos, got, tc.want)
			}
		})
	}
}

func TestIsStableBinaryPath(t *testing.T) {
	cases := []struct {
		name         string
		path         string
		homeDir      string
		localAppData string
		goos         string
		want         bool
	}{
		{
			name:         "Windows stable path",
			path:         `C:\Users\user\AppData\Local\engram-ui\engram-ui.exe`,
			homeDir:      `C:\Users\user`,
			localAppData: `C:\Users\user\AppData\Local`,
			goos:         "windows",
			want:         true,
		},
		{
			name:         "Windows transient path",
			path:         `C:\Users\user\Downloads\engram-ui.exe`,
			homeDir:      `C:\Users\user`,
			localAppData: `C:\Users\user\AppData\Local`,
			goos:         "windows",
			want:         false,
		},
		{
			name:         "macOS stable path",
			path:         "/Users/user/Library/Application Support/engram-ui/engram-ui",
			homeDir:      "/Users/user",
			localAppData: "",
			goos:         "darwin",
			want:         true,
		},
		{
			name:         "macOS homebrew path",
			path:         "/opt/homebrew/bin/engram-ui",
			homeDir:      "/Users/user",
			localAppData: "",
			goos:         "darwin",
			want:         true,
		},
		{
			name:         "Linux stable path",
			path:         "/home/user/.local/bin/engram-ui",
			homeDir:      "/home/user",
			localAppData: "",
			goos:         "linux",
			want:         true,
		},
		{
			name:         "Linux system path",
			path:         "/usr/local/bin/engram-ui",
			homeDir:      "/home/user",
			localAppData: "",
			goos:         "linux",
			want:         true,
		},
		{
			name:         "Linux transient path",
			path:         "/home/user/Downloads/engram-ui",
			homeDir:      "/home/user",
			localAppData: "",
			goos:         "linux",
			want:         false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsStableBinaryPath(tc.path, tc.homeDir, tc.localAppData, tc.goos)
			if got != tc.want {
				t.Errorf("IsStableBinaryPath(%q, %q, %q, %q) = %v, want %v", tc.path, tc.homeDir, tc.localAppData, tc.goos, got, tc.want)
			}
		})
	}
}

func TestStableBinaryPrefixes(t *testing.T) {
	cases := []struct {
		name string
		goos string
		want int // number of expected prefixes
	}{
		{
			name: "Windows prefixes",
			goos: "windows",
			want: 3, // LOCALAPPDATA, ProgramFiles, C:\Program Files
		},
		{
			name: "macOS prefixes",
			goos: "darwin",
			want: 3, // homebrew, usr/local/bin, ~/.local/bin
		},
		{
			name: "Linux prefixes",
			goos: "linux",
			want: 3, // /usr/local/bin, /opt/homebrew/bin, ~/.local/bin
		},
		{
			name: "unsupported platform",
			goos: "plan9",
			want: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StableBinaryPrefixes(tc.goos)
			if len(got) != tc.want {
				t.Errorf("StableBinaryPrefixes(%q) returned %d prefixes, want %d", tc.goos, len(got), tc.want)
			}
		})
	}
}
