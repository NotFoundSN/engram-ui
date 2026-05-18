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
		name         string
		goos         string
		homeDir      string
		localAppData string
		wantContains []string // substrings that MUST appear in at least one prefix
		wantLen      int
	}{
		{
			name:         "Windows prefixes resolve LOCALAPPDATA",
			goos:         "windows",
			homeDir:      `C:\Users\user`,
			localAppData: `C:\Users\user\AppData\Local`,
			wantContains: []string{
				`C:\Users\user\AppData\Local\engram-ui\`,
				`C:\Program Files\`,
				`C:\Program Files (x86)\`,
			},
			wantLen: 3,
		},
		{
			name:         "macOS prefixes resolve home",
			goos:         "darwin",
			homeDir:      "/Users/user",
			localAppData: "",
			wantContains: []string{
				"/opt/homebrew/bin/",
				"/usr/local/bin/",
				"/Users/user/.local/bin/",
			},
			wantLen: 3,
		},
		{
			name:         "Linux prefixes resolve home",
			goos:         "linux",
			homeDir:      "/home/user",
			localAppData: "",
			wantContains: []string{
				"/usr/local/bin/",
				"/opt/homebrew/bin/",
				"/home/user/.local/bin/",
			},
			wantLen: 3,
		},
		{
			name:         "unsupported platform",
			goos:         "plan9",
			homeDir:      "/home/user",
			localAppData: "",
			wantContains: nil,
			wantLen:      0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StableBinaryPrefixes(tc.goos, tc.homeDir, tc.localAppData)
			if len(got) != tc.wantLen {
				t.Errorf("StableBinaryPrefixes(%q, %q, %q) returned %d prefixes, want %d (got %v)", tc.goos, tc.homeDir, tc.localAppData, len(got), tc.wantLen, got)
			}
			for _, needle := range tc.wantContains {
				found := false
				for _, prefix := range got {
					if prefix == needle {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("StableBinaryPrefixes(%q, %q, %q) missing expected prefix %q (got %v)", tc.goos, tc.homeDir, tc.localAppData, needle, got)
				}
			}
		})
	}
}

func TestIsStableBinaryPath_PrefixMatch(t *testing.T) {
	// Regression: prefix list must catch paths under stable dirs that are NOT
	// the canonical exact path (e.g. renamed binary, symlink target).
	cases := []struct {
		name         string
		path         string
		homeDir      string
		localAppData string
		goos         string
		want         bool
	}{
		{
			name:    "Linux renamed binary under ~/.local/bin",
			path:    "/home/user/.local/bin/engram-ui-dev",
			homeDir: "/home/user",
			goos:    "linux",
			want:    true,
		},
		{
			name:    "macOS renamed binary under ~/.local/bin",
			path:    "/Users/user/.local/bin/engram-ui-dev",
			homeDir: "/Users/user",
			goos:    "darwin",
			want:    true,
		},
		{
			name:         "Windows renamed binary under LOCALAPPDATA/engram-ui",
			path:         `C:\Users\user\AppData\Local\engram-ui\engram-ui-dev.exe`,
			homeDir:      `C:\Users\user`,
			localAppData: `C:\Users\user\AppData\Local`,
			goos:         "windows",
			want:         true,
		},
		{
			name:    "Linux unrelated path under home",
			path:    "/home/user/Downloads/engram-ui",
			homeDir: "/home/user",
			goos:    "linux",
			want:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsStableBinaryPath(tc.path, tc.homeDir, tc.localAppData, tc.goos)
			if got != tc.want {
				t.Errorf("IsStableBinaryPath(%q, ...) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
