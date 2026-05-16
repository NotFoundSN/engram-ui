package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Gentleman-Programming/engram-ui/internal/installer"
)

// defaultTestSkills is the catalog stub used by setupSeam so tests don't hit
// the real embedded FS and "brainstorm" is always a known skill.
var defaultTestSkills = []installer.Skill{
	{Name: "brainstorm", Description: "test brainstorm"},
	{Name: "debug", Description: "test debug"},
}

// setupSeam saves/restores stdout, stderr, catalog, and all install/remove fn vars.
func setupSeam(outBuf, errBuf *bytes.Buffer,
	claudeFn func(string) (string, error),
	openFn func(string) (string, error),
	autostartFn func() (string, error),
	fn func(),
) {
	origOut, origErr := stdout, stderr
	stdout, stderr = outBuf, errBuf

	origClaude := installClaudeCodeFn
	origOpen := installOpenCodeFn
	origAutostart := installAutostartFn
	origCatalog := loadCatalogFn

	// Stub catalog so every test with a valid skill name passes catalog check.
	loadCatalogFn = func() ([]installer.Skill, error) { return defaultTestSkills, nil }

	if claudeFn != nil {
		installClaudeCodeFn = claudeFn
	}
	if openFn != nil {
		installOpenCodeFn = openFn
	}
	if autostartFn != nil {
		installAutostartFn = autostartFn
	}

	defer func() {
		stdout, stderr = origOut, origErr
		installClaudeCodeFn = origClaude
		installOpenCodeFn = origOpen
		installAutostartFn = origAutostart
		loadCatalogFn = origCatalog
	}()

	fn()
}

// --- SCN-01: Happy path — setup both tools (default) ---

func TestCmdSetup_BothTools_Default(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var claudeCalled, openCalled bool

	setupSeam(&outBuf, &errBuf,
		func(n string) (string, error) { claudeCalled = true; return "/fake/claude/brainstorm", nil },
		func(n string) (string, error) { openCalled = true; return "/fake/opencode/brainstorm", nil },
		nil,
		func() {
			code := cmdSetup([]string{"brainstorm"})
			if code != 0 {
				t.Errorf("cmdSetup([brainstorm]) = %d, want 0", code)
			}
		},
	)

	if !claudeCalled {
		t.Error("installClaudeCodeFn not called")
	}
	if !openCalled {
		t.Error("installOpenCodeFn not called")
	}
	if outBuf.Len() != 0 {
		t.Errorf("stdout should be empty (success goes to stderr), got: %q", outBuf.String())
	}
	// Both legs should print [leg] prefix on stderr (REQ-1.5 / REQ-2.5).
	errOut := errBuf.String()
	if !strings.Contains(errOut, "[claude]") {
		t.Errorf("stderr missing [claude] prefix, got: %q", errOut)
	}
	if !strings.Contains(errOut, "[opencode]") {
		t.Errorf("stderr missing [opencode] prefix, got: %q", errOut)
	}
}

// --- SCN-02: Setup single tool — claude ---

func TestCmdSetup_ToolClaude(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var claudeCalled, openCalled bool

	setupSeam(&outBuf, &errBuf,
		func(n string) (string, error) { claudeCalled = true; return "/fake/claude/brainstorm", nil },
		func(n string) (string, error) { openCalled = true; return "/fake/opencode/brainstorm", nil },
		nil,
		func() {
			code := cmdSetup([]string{"brainstorm", "--tool=claude"})
			if code != 0 {
				t.Errorf("cmdSetup([brainstorm --tool=claude]) = %d, want 0", code)
			}
		},
	)

	if !claudeCalled {
		t.Error("installClaudeCodeFn should have been called")
	}
	if openCalled {
		t.Error("installOpenCodeFn should NOT have been called for --tool=claude")
	}
	if !strings.Contains(errBuf.String(), "[claude]") {
		t.Errorf("stderr should contain [claude] prefix (REQ-1.5), got: %q", errBuf.String())
	}
}

// --- SCN-03: Setup single tool — opencode ---

func TestCmdSetup_ToolOpenCode(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var claudeCalled, openCalled bool

	setupSeam(&outBuf, &errBuf,
		func(n string) (string, error) { claudeCalled = true; return "/fake/claude/brainstorm", nil },
		func(n string) (string, error) { openCalled = true; return "/fake/opencode/brainstorm", nil },
		nil,
		func() {
			code := cmdSetup([]string{"brainstorm", "--tool=opencode"})
			if code != 0 {
				t.Errorf("cmdSetup([brainstorm --tool=opencode]) = %d, want 0", code)
			}
		},
	)

	if claudeCalled {
		t.Error("installClaudeCodeFn should NOT have been called for --tool=opencode")
	}
	if !openCalled {
		t.Error("installOpenCodeFn should have been called")
	}
	if !strings.Contains(errBuf.String(), "[opencode]") {
		t.Errorf("stderr should contain [opencode] prefix (REQ-2.5), got: %q", errBuf.String())
	}
}

// --- SCN-06: Setup autostart — no tool flag ---

func TestCmdSetup_Autostart_NoToolFlag(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var autostartCalled bool
	var claudeCalled, openCalled bool

	origAutostart := installAutostartFn
	origClaude := installClaudeCodeFn
	origOpen := installOpenCodeFn
	installAutostartFn = func() (string, error) { autostartCalled = true; return "/fake/autostart", nil }
	installClaudeCodeFn = func(n string) (string, error) { claudeCalled = true; return "", nil }
	installOpenCodeFn = func(n string) (string, error) { openCalled = true; return "", nil }
	origOut, origErr := stdout, stderr
	stdout, stderr = &outBuf, &errBuf
	defer func() {
		installAutostartFn = origAutostart
		installClaudeCodeFn = origClaude
		installOpenCodeFn = origOpen
		stdout, stderr = origOut, origErr
	}()

	code := cmdSetup([]string{"autostart"})
	if code != 0 {
		t.Errorf("cmdSetup([autostart]) = %d, want 0", code)
	}
	if !autostartCalled {
		t.Error("installAutostartFn should have been called")
	}
	if claudeCalled {
		t.Error("installClaudeCodeFn should NOT be called for autostart")
	}
	if openCalled {
		t.Error("installOpenCodeFn should NOT be called for autostart")
	}
	// No --tool note on stderr (flag not present), but "registered:" success goes to stderr (REQ-7.4).
	if !strings.Contains(errBuf.String(), "registered") {
		t.Errorf("stderr should contain autostart success message, got: %q", errBuf.String())
	}
	if outBuf.Len() != 0 {
		t.Errorf("stdout should be empty for autostart, got: %q", outBuf.String())
	}
}

// --- SCN-07: Setup autostart — tool flag present (note + proceed) ---

func TestCmdSetup_Autostart_WithToolFlag_NoteAndProceed(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	var autostartCalled bool

	origAutostart := installAutostartFn
	origOut, origErr := stdout, stderr
	installAutostartFn = func() (string, error) { autostartCalled = true; return "/fake/autostart", nil }
	stdout, stderr = &outBuf, &errBuf
	defer func() {
		installAutostartFn = origAutostart
		stdout, stderr = origOut, origErr
	}()

	code := cmdSetup([]string{"autostart", "--tool=claude"})
	if code != 0 {
		t.Errorf("cmdSetup([autostart --tool=claude]) = %d, want 0", code)
	}
	if !autostartCalled {
		t.Error("installAutostartFn should have been called")
	}
	// stderr must contain the note about --tool being ignored.
	if !strings.Contains(errBuf.String(), "--tool") || !strings.Contains(errBuf.String(), "ignored") {
		t.Errorf("stderr should contain --tool ignored note, got: %q", errBuf.String())
	}
}

// --- SCN-10: Setup unknown skill — exit 2 ---

func TestCmdSetup_UnknownSkill_Exit2(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	origOut, origErr := stdout, stderr
	stdout, stderr = &outBuf, &errBuf
	origCatalog := loadCatalogFn
	// Catalog does NOT contain "nonexistent".
	loadCatalogFn = func() ([]installer.Skill, error) {
		return []installer.Skill{{Name: "brainstorm", Description: "test"}}, nil
	}
	defer func() {
		stdout, stderr = origOut, origErr
		loadCatalogFn = origCatalog
	}()

	code := cmdSetup([]string{"nonexistent"})
	if code != 2 {
		t.Errorf("cmdSetup([nonexistent]) = %d, want 2 (usage error for unknown skill)", code)
	}

	if !strings.Contains(errBuf.String(), "nonexistent") {
		t.Errorf("stderr should identify unknown skill, got: %q", errBuf.String())
	}
}

// --- SCN-13: Partial failure — one leg fails → exit 1 ---

func TestCmdSetup_PartialFailure_BothTools_Exit1(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	setupSeam(&outBuf, &errBuf,
		func(n string) (string, error) { return "/fake/claude/brainstorm", nil },
		func(n string) (string, error) { return "", errors.New("permission denied") },
		nil,
		func() {
			code := cmdSetup([]string{"brainstorm", "--tool=both"})
			if code != 1 {
				t.Errorf("cmdSetup([brainstorm --tool=both] partial failure) = %d, want 1", code)
			}
		},
	)

	// Success goes to stderr (REQ-1.5), error also on stderr.
	if !strings.Contains(errBuf.String(), "[claude]") {
		t.Errorf("stderr should contain [claude] success line (REQ-1.5), got: %q", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "[opencode]") || !strings.Contains(errBuf.String(), "permission denied") {
		t.Errorf("stderr should contain [opencode] error, got: %q", errBuf.String())
	}
}

// --- SCN-14: Setup — missing skill argument — exit 2 ---

func TestCmdSetup_MissingSkill_Exit2(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	setupSeam(&outBuf, &errBuf, nil, nil, nil, func() {
		code := cmdSetup([]string{})
		if code != 2 {
			t.Errorf("cmdSetup([]) = %d, want 2", code)
		}
	})

	if errBuf.Len() == 0 {
		t.Error("stderr should contain usage error")
	}
}

// --- SCN-11: Invalid --tool value — exit 2 ---

func TestCmdSetup_InvalidToolValue_Exit2(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	setupSeam(&outBuf, &errBuf, nil, nil, nil, func() {
		code := cmdSetup([]string{"brainstorm", "--tool=foo"})
		if code != 2 {
			t.Errorf("cmdSetup([brainstorm --tool=foo]) = %d, want 2", code)
		}
	})

	if !strings.Contains(errBuf.String(), "foo") {
		t.Errorf("stderr should mention invalid tool value, got: %q", errBuf.String())
	}
}

// --- SCN-12: Empty --tool value — exit 2 ---

func TestCmdSetup_EmptyToolValue_Exit2(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	setupSeam(&outBuf, &errBuf, nil, nil, nil, func() {
		code := cmdSetup([]string{"brainstorm", "--tool="})
		if code != 2 {
			t.Errorf("cmdSetup([brainstorm --tool=]) = %d, want 2", code)
		}
	})

	if errBuf.Len() == 0 {
		t.Error("stderr should contain usage error")
	}
}

// --- AUX (REQ-1.5): Per-leg [leg] prefix on single-leg invocations ---

func TestCmdSetup_LegPrefix_SingleLeg(t *testing.T) {
	var outBuf, errBuf bytes.Buffer

	setupSeam(&outBuf, &errBuf,
		func(n string) (string, error) { return "/fake/claude/" + n, nil },
		nil,
		nil,
		func() {
			code := cmdSetup([]string{"brainstorm", "--tool=claude"})
			if code != 0 {
				t.Errorf("code = %d, want 0", code)
			}
		},
	)

	errOut := errBuf.String()
	if !strings.Contains(errOut, "[claude]") {
		t.Errorf("single-leg stderr should contain [claude] prefix (REQ-1.5), got: %q", errOut)
	}
}
