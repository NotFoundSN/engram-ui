package tui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Gentleman-Programming/engram-ui/internal/installer"
)

// DetectState returns the current state of an Item on the local filesystem.
//
// The home + xdg arguments are explicit (not read from env inside the function)
// so tests can use a tmpdir without races against the parent process env.
// Pass empty xdg to mean "use default XDG_CONFIG_HOME (= ~/.config)".
//
// State semantics per Item kind:
//
//   autostart                 → StateInstalled if the per-OS autostart file
//                                exists; StateNotInstalled otherwise. Never
//                                StateUnavailable (autostart works on all
//                                supported OSes).
//
//   skill:{name}:claude       → StateUnavailable if ~/.claude/ is missing,
//                                StateInstalled if the destination SKILL.md
//                                exists, StateNotInstalled otherwise.
//
//   skill:{name}:opencode     → StateUnavailable if the OpenCode config dir
//                                is missing, StateInstalled if SKILL.md
//                                exists, StateNotInstalled otherwise.
func DetectState(item Item, home, xdg string) State {
	switch {
	case item.ID == "autostart":
		return detectAutostart()
	case strings.HasPrefix(item.ID, "skill:") && strings.HasSuffix(item.ID, ":claude"):
		return detectClaudeSkill(home, item.SkillName)
	case strings.HasPrefix(item.ID, "skill:") && strings.HasSuffix(item.ID, ":opencode"):
		return detectOpenCodeSkill(home, xdg, item.SkillName)
	default:
		return StateUnknown
	}
}

func detectAutostart() State {
	appData := os.Getenv("APPDATA")
	home, err := os.UserHomeDir()
	if err != nil {
		return StateNotInstalled
	}
	xdg := os.Getenv("XDG_CONFIG_HOME")

	// Check each per-OS candidate path; if any exists, autostart is installed.
	candidates := []string{}
	if appData != "" {
		candidates = append(candidates, filepath.Join(installer.WindowsStartupDir(appData), "engram-ui.bat"))
	}
	candidates = append(candidates, installer.MacOSLaunchAgentPath(home))
	candidates = append(candidates, installer.LinuxSystemdUnitPath(home, xdg))

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return StateInstalled
		}
	}
	return StateNotInstalled
}

func detectClaudeSkill(home, skillName string) State {
	toolRoot := filepath.Join(home, ".claude")
	if _, err := os.Stat(toolRoot); os.IsNotExist(err) {
		return StateUnavailable
	}
	skillMD := filepath.Join(installer.ClaudeSkillDir(home, skillName), "SKILL.md")
	if _, err := os.Stat(skillMD); err == nil {
		return StateInstalled
	}
	return StateNotInstalled
}

func detectOpenCodeSkill(home, xdg, skillName string) State {
	base := xdg
	if base == "" {
		base = filepath.Join(home, ".config")
	}
	toolRoot := filepath.Join(base, "opencode")
	if _, err := os.Stat(toolRoot); os.IsNotExist(err) {
		return StateUnavailable
	}
	skillMD := filepath.Join(installer.OpenCodeSkillDir(home, xdg, skillName), "SKILL.md")
	if _, err := os.Stat(skillMD); err == nil {
		return StateInstalled
	}
	return StateNotInstalled
}
