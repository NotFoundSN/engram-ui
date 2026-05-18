package installer

import (
	"strings"
	"testing"
)

func TestBuildWindowsStartupBat(t *testing.T) {
	cases := []struct {
		name     string
		execPath string
		wantIn   []string
	}{
		{
			name:     "simple path",
			execPath: `C:\Users\user\engram-ui.exe`,
			wantIn: []string{
				"@echo off",
				`start "" /B "C:\Users\user\engram-ui.exe" serve`,
			},
		},
		{
			name:     "path with spaces",
			execPath: `C:\Program Files\engram-ui\engram-ui.exe`,
			wantIn: []string{
				"@echo off",
				`start "" /B "C:\Program Files\engram-ui\engram-ui.exe" serve`,
			},
		},
		{
			name:     "path with backslashes",
			execPath: `C:\tools\bin\engram-ui.exe`,
			wantIn: []string{
				"@echo off",
				`"C:\tools\bin\engram-ui.exe" serve`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildWindowsStartupBat(tc.execPath)
			for _, want := range tc.wantIn {
				if !strings.Contains(got, want) {
					t.Errorf("BuildWindowsStartupBat(%q): missing %q in output:\n%s", tc.execPath, want, got)
				}
			}
		})
	}
}

func TestBuildLaunchAgentPlist(t *testing.T) {
	cases := []struct {
		name     string
		execPath string
		label    string
		wantIn   []string
	}{
		{
			name:     "basic",
			execPath: "/usr/local/bin/engram-ui",
			label:    "com.notfoundsn.engram-ui",
			wantIn: []string{
				"<key>Label</key>",
				"<string>com.notfoundsn.engram-ui</string>",
				"<key>ProgramArguments</key>",
				"<string>/usr/local/bin/engram-ui</string>",
				"<string>serve</string>",
				"<key>RunAtLoad</key>",
				"<true/>",
				"<key>KeepAlive</key>",
				"<false/>",
			},
		},
		{
			name:     "label with dots",
			execPath: "/opt/homebrew/bin/engram-ui",
			label:    "com.acme.engram-ui",
			wantIn: []string{
				"<string>com.acme.engram-ui</string>",
				"<string>/opt/homebrew/bin/engram-ui</string>",
				"<true/>",
				"<false/>",
			},
		},
		{
			name:     "exec path with spaces",
			execPath: "/home/user/my apps/engram-ui",
			label:    "com.notfoundsn.engram-ui",
			wantIn: []string{
				"<string>/home/user/my apps/engram-ui</string>",
				"<true/>",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildLaunchAgentPlist(tc.execPath, tc.label)
			for _, want := range tc.wantIn {
				if !strings.Contains(got, want) {
					t.Errorf("BuildLaunchAgentPlist(%q, %q): missing %q in output:\n%s", tc.execPath, tc.label, want, got)
				}
			}
		})
	}
}

func TestBuildSystemdUnit(t *testing.T) {
	cases := []struct {
		name          string
		execPath      string
		wantExecStart string
		wantIn        []string
	}{
		{
			name:          "basic path",
			execPath:      "/usr/local/bin/engram-ui",
			wantExecStart: "ExecStart=/usr/local/bin/engram-ui serve",
			wantIn:        []string{"[Unit]", "[Service]", "Type=simple", "WantedBy=default.target", "[Install]"},
		},
		{
			name:          "path with spaces",
			execPath:      "/home/user/my apps/engram-ui",
			wantExecStart: `ExecStart=/home/user/my apps/engram-ui serve`,
			wantIn:        []string{"[Service]", "Type=simple", "WantedBy=default.target"},
		},
		{
			name:          "path with unicode",
			execPath:      "/home/用户/engram-ui",
			wantExecStart: "ExecStart=/home/用户/engram-ui serve",
			wantIn:        []string{"[Service]", "WantedBy=default.target"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildSystemdUnit(tc.execPath)
			if !strings.Contains(got, tc.wantExecStart) {
				t.Errorf("BuildSystemdUnit(%q):\ngot:\n%s\nwant ExecStart line: %q", tc.execPath, got, tc.wantExecStart)
			}
			for _, want := range tc.wantIn {
				if !strings.Contains(got, want) {
					t.Errorf("BuildSystemdUnit(%q): missing %q in output:\n%s", tc.execPath, want, got)
				}
			}
		})
	}
}
