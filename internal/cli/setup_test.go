package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestCmdSetup_NoArgs(t *testing.T) {
	var buf bytes.Buffer
	origStderr := stderr
	stderr = &buf
	defer func() { stderr = origStderr }()

	code := cmdSetup([]string{})
	if code != 2 {
		t.Errorf("cmdSetup([]) = %d, want 2", code)
	}
}

func TestCmdSetup_UnknownTarget(t *testing.T) {
	var buf bytes.Buffer
	origStderr := stderr
	stderr = &buf
	defer func() { stderr = origStderr }()

	code := cmdSetup([]string{"foo"})
	if code != 2 {
		t.Errorf("cmdSetup([foo]) = %d, want 2", code)
	}
	if !strings.Contains(buf.String(), "unknown") {
		t.Errorf("cmdSetup([foo]) stderr %q should mention 'unknown'", buf.String())
	}
}

func TestCmdSetup_ClaudeCode(t *testing.T) {
	// Inject stub installer.
	origInstallClaude := installClaudeCodeFn
	defer func() { installClaudeCodeFn = origInstallClaude }()

	called := false
	installClaudeCodeFn = func() (string, error) {
		called = true
		return "/fake/.claude/skills/engram-conventions", nil
	}

	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := cmdSetup([]string{"claude-code"})
	if code != 0 {
		t.Errorf("cmdSetup([claude-code]) = %d, want 0", code)
	}
	if !called {
		t.Error("cmdSetup([claude-code]): installClaudeCodeFn not called")
	}
	if !strings.Contains(buf.String(), "/fake/.claude/skills/engram-conventions") {
		t.Errorf("cmdSetup([claude-code]) stdout %q should contain destination path", buf.String())
	}
}

func TestCmdSetup_OpenCode(t *testing.T) {
	origInstallOpenCode := installOpenCodeFn
	defer func() { installOpenCodeFn = origInstallOpenCode }()

	called := false
	installOpenCodeFn = func() (string, error) {
		called = true
		return "/fake/.config/opencode/skills/engram-conventions", nil
	}

	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := cmdSetup([]string{"opencode"})
	if code != 0 {
		t.Errorf("cmdSetup([opencode]) = %d, want 0", code)
	}
	if !called {
		t.Error("cmdSetup([opencode]): installOpenCodeFn not called")
	}
}

func TestCmdSetup_OsAutostart(t *testing.T) {
	origInstallAutostart := installAutostartFn
	defer func() { installAutostartFn = origInstallAutostart }()

	called := false
	installAutostartFn = func() (string, error) {
		called = true
		return "/fake/startup/engram-ui.bat", nil
	}

	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := cmdSetup([]string{"os-autostart"})
	if code != 0 {
		t.Errorf("cmdSetup([os-autostart]) = %d, want 0", code)
	}
	if !called {
		t.Error("cmdSetup([os-autostart]): installAutostartFn not called")
	}
}

func TestCmdSetup_ClaudeCode_Error(t *testing.T) {
	origInstallClaude := installClaudeCodeFn
	defer func() { installClaudeCodeFn = origInstallClaude }()

	installClaudeCodeFn = func() (string, error) {
		return "", fmt.Errorf("permission denied")
	}

	var buf bytes.Buffer
	origStderr := stderr
	stderr = &buf
	defer func() { stderr = origStderr }()

	code := cmdSetup([]string{"claude-code"})
	if code != 1 {
		t.Errorf("cmdSetup([claude-code] with error) = %d, want 1", code)
	}
}

func TestCmdSetup_OsAutostart_Error(t *testing.T) {
	origInstallAutostart := installAutostartFn
	defer func() { installAutostartFn = origInstallAutostart }()

	installAutostartFn = func() (string, error) {
		return "", fmt.Errorf("unsupported platform")
	}

	var buf bytes.Buffer
	origStderr := stderr
	stderr = &buf
	defer func() { stderr = origStderr }()

	code := cmdSetup([]string{"os-autostart"})
	if code != 1 {
		t.Errorf("cmdSetup([os-autostart] with error) = %d, want 1", code)
	}
}

func TestCmdSetup_RemoveAutostart_NotRegistered(t *testing.T) {
	origRemoveAutostart := removeAutostartFn
	defer func() { removeAutostartFn = origRemoveAutostart }()

	removeAutostartFn = func() (string, error) {
		return "", nil // empty dest = not registered
	}

	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := cmdSetup([]string{"remove-autostart"})
	if code != 0 {
		t.Errorf("cmdSetup([remove-autostart] not-registered) = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "not currently registered") {
		t.Errorf("expected 'not currently registered' in output, got: %q", buf.String())
	}
}

func TestCmdSetup_RemoveAutostart(t *testing.T) {
	origRemoveAutostart := removeAutostartFn
	defer func() { removeAutostartFn = origRemoveAutostart }()

	called := false
	removeAutostartFn = func() (string, error) {
		called = true
		return "removed", nil
	}

	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := cmdSetup([]string{"remove-autostart"})
	if code != 0 {
		t.Errorf("cmdSetup([remove-autostart]) = %d, want 0", code)
	}
	if !called {
		t.Error("cmdSetup([remove-autostart]): removeAutostartFn not called")
	}
}
