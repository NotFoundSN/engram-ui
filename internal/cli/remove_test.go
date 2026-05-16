package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// removeSeam saves/restores stdout, stderr, and uninstall fn vars.
func removeSeam(outBuf, errBuf *bytes.Buffer,
	claudeFn func(string) (string, error),
	openFn func(string) (string, error),
	autostartFn func() (string, error),
	fn func(),
) {
	origOut, origErr := stdout, stderr
	stdout, stderr = outBuf, errBuf

	origClaude := uninstallClaudeCodeFn
	origOpen := uninstallOpenCodeFn
	origAutostart := removeAutostartFn

	if claudeFn != nil {
		uninstallClaudeCodeFn = claudeFn
	}
	if openFn != nil {
		uninstallOpenCodeFn = openFn
	}
	if autostartFn != nil {
		removeAutostartFn = autostartFn
	}

	defer func() {
		stdout, stderr = origOut, origErr
		uninstallClaudeCodeFn = origClaude
		uninstallOpenCodeFn = origOpen
		removeAutostartFn = origAutostart
	}()

	fn()
}

// --- SCN-04: Remove both tools (default) ---

func TestCmdRemove_BothTools_Default(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var claudeCalled, openCalled bool

	removeSeam(&outBuf, &errBuf,
		func(n string) (string, error) { claudeCalled = true; return "/fake/claude/brainstorm", nil },
		func(n string) (string, error) { openCalled = true; return "/fake/opencode/brainstorm", nil },
		nil,
		func() {
			code := cmdRemove([]string{"brainstorm"})
			if code != 0 {
				t.Errorf("cmdRemove([brainstorm]) = %d, want 0", code)
			}
		},
	)

	if !claudeCalled {
		t.Error("uninstallClaudeCodeFn not called")
	}
	if !openCalled {
		t.Error("uninstallOpenCodeFn not called")
	}
	if outBuf.Len() != 0 {
		t.Errorf("stdout should be empty (success goes to stderr), got: %q", outBuf.String())
	}
	// Success messages go to stderr (REQ-2.5 / REQ-7.4).
	errOut := errBuf.String()
	if !strings.Contains(errOut, "[claude]") {
		t.Errorf("stderr missing [claude] prefix, got: %q", errOut)
	}
	if !strings.Contains(errOut, "[opencode]") {
		t.Errorf("stderr missing [opencode] prefix, got: %q", errOut)
	}
}

// --- SCN-05: Remove — explicit both ---

func TestCmdRemove_ExplicitBoth(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var claudeCalled, openCalled bool

	removeSeam(&outBuf, &errBuf,
		func(n string) (string, error) { claudeCalled = true; return "/fake/claude/brainstorm", nil },
		func(n string) (string, error) { openCalled = true; return "/fake/opencode/brainstorm", nil },
		nil,
		func() {
			code := cmdRemove([]string{"brainstorm", "--tool=both"})
			if code != 0 {
				t.Errorf("cmdRemove([brainstorm --tool=both]) = %d, want 0", code)
			}
		},
	)

	if !claudeCalled || !openCalled {
		t.Error("both uninstall fns should be called for --tool=both")
	}
}

// --- Single tool: claude ---

func TestCmdRemove_ToolClaude(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var claudeCalled, openCalled bool

	removeSeam(&outBuf, &errBuf,
		func(n string) (string, error) { claudeCalled = true; return "/fake/claude/brainstorm", nil },
		func(n string) (string, error) { openCalled = true; return "/fake/opencode/brainstorm", nil },
		nil,
		func() {
			code := cmdRemove([]string{"brainstorm", "--tool=claude"})
			if code != 0 {
				t.Errorf("cmdRemove([brainstorm --tool=claude]) = %d, want 0", code)
			}
		},
	)

	if !claudeCalled {
		t.Error("uninstallClaudeCodeFn should be called")
	}
	if openCalled {
		t.Error("uninstallOpenCodeFn should NOT be called for --tool=claude")
	}
	if !strings.Contains(errBuf.String(), "[claude]") {
		t.Errorf("stderr missing [claude] prefix (REQ-2.5), got: %q", errBuf.String())
	}
}

// --- Single tool: opencode ---

func TestCmdRemove_ToolOpenCode(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var claudeCalled, openCalled bool

	removeSeam(&outBuf, &errBuf,
		func(n string) (string, error) { claudeCalled = true; return "/fake/claude/brainstorm", nil },
		func(n string) (string, error) { openCalled = true; return "/fake/opencode/brainstorm", nil },
		nil,
		func() {
			code := cmdRemove([]string{"brainstorm", "--tool=opencode"})
			if code != 0 {
				t.Errorf("cmdRemove([brainstorm --tool=opencode]) = %d, want 0", code)
			}
		},
	)

	if claudeCalled {
		t.Error("uninstallClaudeCodeFn should NOT be called for --tool=opencode")
	}
	if !openCalled {
		t.Error("uninstallOpenCodeFn should be called")
	}
}

// --- SCN-07 (remove variant): Remove autostart — no tool flag ---

func TestCmdRemove_Autostart_NoToolFlag(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var autostartCalled bool

	removeSeam(&outBuf, &errBuf, nil, nil,
		func() (string, error) { autostartCalled = true; return "/fake/autostart", nil },
		func() {
			code := cmdRemove([]string{"autostart"})
			if code != 0 {
				t.Errorf("cmdRemove([autostart]) = %d, want 0", code)
			}
		},
	)

	if !autostartCalled {
		t.Error("removeAutostartFn should have been called")
	}
	// No --tool note on stderr (flag not present), but "removed:" success goes to stderr (REQ-7.4).
	if !strings.Contains(errBuf.String(), "removed") {
		t.Errorf("stderr should contain autostart removal success, got: %q", errBuf.String())
	}
	if outBuf.Len() != 0 {
		t.Errorf("stdout should be empty for autostart, got: %q", outBuf.String())
	}
}

// --- Remove autostart with --tool flag — note + proceed ---

func TestCmdRemove_Autostart_WithToolFlag_NoteAndProceed(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var autostartCalled bool

	removeSeam(&outBuf, &errBuf, nil, nil,
		func() (string, error) { autostartCalled = true; return "/fake/autostart", nil },
		func() {
			code := cmdRemove([]string{"autostart", "--tool=opencode"})
			if code != 0 {
				t.Errorf("cmdRemove([autostart --tool=opencode]) = %d, want 0", code)
			}
		},
	)

	if !autostartCalled {
		t.Error("removeAutostartFn should have been called")
	}
	if !strings.Contains(errBuf.String(), "--tool") || !strings.Contains(errBuf.String(), "ignored") {
		t.Errorf("stderr should mention --tool ignored, got: %q", errBuf.String())
	}
}

// --- Not-registered sentinel → stdout "not registered" line, exit 0 ---

func TestCmdRemove_NotRegistered_Sentinel(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	removeSeam(&outBuf, &errBuf,
		func(n string) (string, error) { return "", nil }, // empty dest = not registered
		func(n string) (string, error) { return "", nil },
		nil,
		func() {
			code := cmdRemove([]string{"brainstorm"})
			if code != 0 {
				t.Errorf("cmdRemove([brainstorm] not-registered) = %d, want 0", code)
			}
		},
	)

	errOut := errBuf.String()
	if !strings.Contains(errOut, "not registered") {
		t.Errorf("stderr should mention 'not registered' (REQ-2.5), got: %q", errOut)
	}
}

// --- SCN-13 (remove variant): Partial failure — one leg fails → exit 1 ---

func TestCmdRemove_PartialFailure_Exit1(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	removeSeam(&outBuf, &errBuf,
		func(n string) (string, error) { return "/fake/claude/brainstorm", nil },
		func(n string) (string, error) { return "", errors.New("permission denied") },
		nil,
		func() {
			code := cmdRemove([]string{"brainstorm", "--tool=both"})
			if code != 1 {
				t.Errorf("cmdRemove([brainstorm --tool=both] partial failure) = %d, want 1", code)
			}
		},
	)

	if !strings.Contains(errBuf.String(), "[claude]") {
		t.Errorf("stderr should have [claude] success (REQ-2.5), got: %q", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "[opencode]") || !strings.Contains(errBuf.String(), "permission denied") {
		t.Errorf("stderr should have [opencode] error, got: %q", errBuf.String())
	}
}

// --- SCN-15: Remove — missing skill argument — exit 2 ---

func TestCmdRemove_MissingSkill_Exit2(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	removeSeam(&outBuf, &errBuf, nil, nil, nil, func() {
		code := cmdRemove([]string{})
		if code != 2 {
			t.Errorf("cmdRemove([]) = %d, want 2", code)
		}
	})

	if errBuf.Len() == 0 {
		t.Error("stderr should contain usage error")
	}
}

// --- Invalid --tool value — exit 2 ---

func TestCmdRemove_InvalidToolValue_Exit2(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	removeSeam(&outBuf, &errBuf, nil, nil, nil, func() {
		code := cmdRemove([]string{"brainstorm", "--tool=invalid"})
		if code != 2 {
			t.Errorf("cmdRemove([brainstorm --tool=invalid]) = %d, want 2", code)
		}
	})

	if !strings.Contains(errBuf.String(), "invalid") {
		t.Errorf("stderr should mention invalid tool value, got: %q", errBuf.String())
	}
}
