//go:build linux

package installer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type linuxAutostart struct{}

func (l *linuxAutostart) Install(execPath string) (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}
	xdg := os.Getenv("XDG_CONFIG_HOME")
	unitPath := LinuxSystemdUnitPath(home, xdg)

	content := BuildSystemdUnit(execPath)
	if err := os.MkdirAll(filepath.Dir(unitPath), 0o755); err != nil {
		return Result{}, fmt.Errorf("create systemd user dir: %w", err)
	}
	if err := os.WriteFile(unitPath, []byte(content), 0o644); err != nil {
		return Result{}, fmt.Errorf("write unit file: %w", err)
	}

	notes := ""
	// Best-effort daemon-reload and enable.
	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		notes = fmt.Sprintf("systemctl daemon-reload failed (%v): %s; run manually", err, out)
		log.Printf("warning: %s", notes)
	} else {
		if out2, err2 := exec.Command("systemctl", "--user", "enable", "--now", "engram-ui.service").CombinedOutput(); err2 != nil {
			notes = fmt.Sprintf("systemctl enable failed (%v): %s; run manually: systemctl --user enable --now engram-ui.service", err2, out2)
			log.Printf("warning: %s", notes)
		}
	}

	return Result{
		Destination: unitPath,
		Action:      ActionInstalled,
		Notes:       notes,
	}, nil
}

func (l *linuxAutostart) Remove() (Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Result{}, err
	}
	xdg := os.Getenv("XDG_CONFIG_HOME")
	unitPath := LinuxSystemdUnitPath(home, xdg)

	if _, err := os.Stat(unitPath); os.IsNotExist(err) {
		return Result{Action: ActionNotRegistered, Notes: "not currently registered"}, nil
	}

	// Best-effort disable.
	if out, err := exec.Command("systemctl", "--user", "disable", "engram-ui.service").CombinedOutput(); err != nil {
		log.Printf("warning: systemctl disable failed (%v): %s", err, out)
	}

	if err := os.Remove(unitPath); err != nil {
		return Result{}, fmt.Errorf("remove unit file: %w", err)
	}

	return Result{Destination: unitPath, Action: ActionRemoved}, nil
}

// NewAutostartManager returns the Linux systemd-user autostart manager.
func NewAutostartManager() AutostartManager {
	return &linuxAutostart{}
}
