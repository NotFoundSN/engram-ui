//go:build windows

package installer

import (
	"fmt"
	"os"
	"path/filepath"
)

type windowsAutostart struct{}

func (w *windowsAutostart) Install(execPath string) (Result, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return Result{}, fmt.Errorf("APPDATA environment variable not set")
	}

	startupDir := WindowsStartupDir(appData)
	batPath := filepath.Join(startupDir, "engram-ui.bat")
	content := BuildWindowsStartupBat(execPath)

	if err := os.MkdirAll(startupDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create startup dir: %w", err)
	}
	if err := os.WriteFile(batPath, []byte(content), 0o644); err != nil {
		return Result{}, fmt.Errorf("write startup bat: %w", err)
	}

	return Result{
		Destination: batPath,
		Action:      ActionInstalled,
		Notes:       "engram-ui will start automatically on next login",
	}, nil
}

func (w *windowsAutostart) Remove() (Result, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return Result{}, fmt.Errorf("APPDATA environment variable not set")
	}

	startupDir := WindowsStartupDir(appData)
	batPath := filepath.Join(startupDir, "engram-ui.bat")

	if _, err := os.Stat(batPath); os.IsNotExist(err) {
		return Result{Action: ActionNotRegistered, Notes: "not currently registered"}, nil
	}

	if err := os.Remove(batPath); err != nil {
		return Result{}, fmt.Errorf("remove startup bat: %w", err)
	}

	return Result{Destination: batPath, Action: ActionRemoved}, nil
}

// NewAutostartManager returns the Windows Startup folder autostart manager.
func NewAutostartManager() AutostartManager {
	return &windowsAutostart{}
}
