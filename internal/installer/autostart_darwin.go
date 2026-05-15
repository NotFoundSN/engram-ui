//go:build darwin

package installer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type darwinAutostart struct{}

const darwinPlistLabel = "com.gentleman-programming.engram-ui"

func (d *darwinAutostart) Install(execPath string) (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}
	plistPath := MacOSLaunchAgentPath(home)
	content := BuildLaunchAgentPlist(execPath, darwinPlistLabel)

	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return Result{}, fmt.Errorf("create LaunchAgents dir: %w", err)
	}
	if err := os.WriteFile(plistPath, []byte(content), 0o644); err != nil {
		return Result{}, fmt.Errorf("write plist: %w", err)
	}

	notes := ""
	// Best-effort activate.
	if out, err := exec.Command("launchctl", "load", plistPath).CombinedOutput(); err != nil {
		notes = fmt.Sprintf("launchctl load failed (%v): %s; run manually: launchctl load %s", err, out, plistPath)
		log.Printf("warning: %s", notes)
	}

	return Result{
		Destination: plistPath,
		Action:      ActionInstalled,
		Notes:       notes,
	}, nil
}

func (d *darwinAutostart) Remove() (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}
	plistPath := MacOSLaunchAgentPath(home)

	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return Result{Action: ActionNotRegistered, Notes: "not currently registered"}, nil
	}

	// Best-effort unload.
	if out, err := exec.Command("launchctl", "unload", plistPath).CombinedOutput(); err != nil {
		log.Printf("warning: launchctl unload failed (%v): %s", err, out)
	}

	if err := os.Remove(plistPath); err != nil {
		return Result{}, fmt.Errorf("remove plist: %w", err)
	}

	return Result{Destination: plistPath, Action: ActionRemoved}, nil
}

// NewAutostartManager returns the macOS LaunchAgent autostart manager.
func NewAutostartManager() AutostartManager {
	return &darwinAutostart{}
}
