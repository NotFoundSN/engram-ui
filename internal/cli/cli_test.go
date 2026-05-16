package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gentleman-Programming/engram-ui/internal/installer"
)

func TestDispatch_NoArgs_TTY_LaunchesTUI(t *testing.T) {
	called := false
	origRun := runTUIFn
	defer func() { runTUIFn = origRun }()
	runTUIFn = func() error {
		called = true
		return nil
	}

	origTTY := isInteractiveFn
	defer func() { isInteractiveFn = origTTY }()
	isInteractiveFn = func() bool { return true }

	code := Dispatch([]string{})
	if code != 0 {
		t.Errorf("Dispatch([]) interactive = %d, want 0", code)
	}
	if !called {
		t.Error("Dispatch([]) interactive: expected runTUIFn to be called")
	}
}

func TestDispatch_NoArgs_NonTTY_PrintsHelp(t *testing.T) {
	origRun := runTUIFn
	defer func() { runTUIFn = origRun }()
	runTUIFn = func() error {
		t.Error("non-interactive: runTUIFn must NOT be called")
		return nil
	}

	origTTY := isInteractiveFn
	defer func() { isInteractiveFn = origTTY }()
	isInteractiveFn = func() bool { return false }

	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := Dispatch([]string{})
	if code != 0 {
		t.Errorf("Dispatch([]) non-interactive = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "engram-ui") {
		t.Errorf("Dispatch([]) non-interactive: stdout should contain help, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "serve") {
		t.Errorf("Dispatch([]) non-interactive: help should mention 'serve' subcommand")
	}
}

func TestDispatch_NoArgs_TTY_TUIError_Returns1(t *testing.T) {
	origRun := runTUIFn
	defer func() { runTUIFn = origRun }()
	runTUIFn = func() error { return fmt.Errorf("boom") }

	origTTY := isInteractiveFn
	defer func() { isInteractiveFn = origTTY }()
	isInteractiveFn = func() bool { return true }

	var buf bytes.Buffer
	origStderr := stderr
	stderr = &buf
	defer func() { stderr = origStderr }()

	code := Dispatch([]string{})
	if code != 1 {
		t.Errorf("Dispatch([]) interactive TUI error = %d, want 1", code)
	}
	if !strings.Contains(buf.String(), "boom") {
		t.Errorf("Dispatch([]) TUI error: stderr should contain error message, got %q", buf.String())
	}
}

func TestDispatch_NoTUIFlag_TTY_PrintsHelp(t *testing.T) {
	// --no-tui must short-circuit the TUI even when stdin is a terminal.
	origRun := runTUIFn
	defer func() { runTUIFn = origRun }()
	runTUIFn = func() error {
		t.Error("--no-tui: runTUIFn must NOT be called")
		return nil
	}

	origTTY := isInteractiveFn
	defer func() { isInteractiveFn = origTTY }()
	isInteractiveFn = func() bool { return true }

	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := Dispatch([]string{"--no-tui"})
	if code != 0 {
		t.Errorf("Dispatch([--no-tui]) TTY = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "engram-ui") {
		t.Errorf("Dispatch([--no-tui]) TTY: stdout should contain help, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "serve") {
		t.Errorf("Dispatch([--no-tui]) TTY: help should mention 'serve'")
	}
}

func TestDispatch_NoTUIFlag_NonTTY_PrintsHelp(t *testing.T) {
	// --no-tui is also valid in non-TTY contexts (idempotent: behaves like no-args non-TTY).
	origRun := runTUIFn
	defer func() { runTUIFn = origRun }()
	runTUIFn = func() error {
		t.Error("--no-tui non-TTY: runTUIFn must NOT be called")
		return nil
	}

	origTTY := isInteractiveFn
	defer func() { isInteractiveFn = origTTY }()
	isInteractiveFn = func() bool { return false }

	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := Dispatch([]string{"--no-tui"})
	if code != 0 {
		t.Errorf("Dispatch([--no-tui]) non-TTY = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "engram-ui") {
		t.Errorf("Dispatch([--no-tui]) non-TTY: stdout should contain help, got %q", buf.String())
	}
}

func TestDispatch_NoTUIFlag_HelpMentionsIt(t *testing.T) {
	// Help text should advertise --no-tui so users can discover it.
	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := Dispatch([]string{"help"})
	if code != 0 {
		t.Errorf("Dispatch([help]) = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "--no-tui") {
		t.Errorf("Dispatch([help]) output missing '--no-tui' flag, got: %q", buf.String())
	}
}

func TestDispatch_Version(t *testing.T) {
	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := Dispatch([]string{"version"})
	if code != 0 {
		t.Errorf("Dispatch([version]) = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "engram-ui") {
		t.Errorf("Dispatch([version]) output %q does not contain 'engram-ui'", buf.String())
	}
}

func TestDispatch_VersionFlag(t *testing.T) {
	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := Dispatch([]string{"--version"})
	if code != 0 {
		t.Errorf("Dispatch([--version]) = %d, want 0", code)
	}
}

func TestDispatch_ShortVersionFlag(t *testing.T) {
	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := Dispatch([]string{"-v"})
	if code != 0 {
		t.Errorf("Dispatch([-v]) = %d, want 0", code)
	}
}

func TestDispatch_Help(t *testing.T) {
	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := Dispatch([]string{"help"})
	if code != 0 {
		t.Errorf("Dispatch([help]) = %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "serve") {
		t.Errorf("Dispatch([help]) output missing 'serve'")
	}
}

func TestDispatch_UnknownSubcommand(t *testing.T) {
	var buf bytes.Buffer
	origStderr := stderr
	stderr = &buf
	defer func() { stderr = origStderr }()

	code := Dispatch([]string{"foo"})
	if code != 2 {
		t.Errorf("Dispatch([foo]) = %d, want 2", code)
	}
	if !strings.Contains(buf.String(), "foo") || !strings.Contains(buf.String(), "unknown") {
		t.Errorf("Dispatch([foo]) stderr %q should mention unknown subcommand", buf.String())
	}
}

func TestDispatch_Setup_RoutesToCmdSetup(t *testing.T) {
	// "setup" with a valid skill routes to cmdSetup.
	// Stub catalog + install fns to avoid real FS calls.
	origCatalog := loadCatalogFn
	defer func() { loadCatalogFn = origCatalog }()
	loadCatalogFn = func() ([]installer.Skill, error) {
		return []installer.Skill{{Name: "brainstorm", Description: "test"}}, nil
	}

	origClaude := installClaudeCodeFn
	origOpen := installOpenCodeFn
	defer func() { installClaudeCodeFn = origClaude; installOpenCodeFn = origOpen }()
	installClaudeCodeFn = func(n string) (string, error) { return "/fake/claude/" + n, nil }
	installOpenCodeFn = func(n string) (string, error) { return "/fake/open/" + n, nil }

	var outBuf bytes.Buffer
	origOut := stdout
	stdout = &outBuf
	defer func() { stdout = origOut }()

	code := Dispatch([]string{"setup", "brainstorm"})
	if code != 0 {
		t.Errorf("Dispatch([setup brainstorm]) = %d, want 0", code)
	}
}

func TestDispatch_Remove_RoutesToCmdRemove(t *testing.T) {
	// "remove" routes to cmdRemove — use stubs to avoid real FS ops.
	origClaude := uninstallClaudeCodeFn
	origOpen := uninstallOpenCodeFn
	defer func() { uninstallClaudeCodeFn = origClaude; uninstallOpenCodeFn = origOpen }()
	uninstallClaudeCodeFn = func(n string) (string, error) { return "/fake/claude/" + n, nil }
	uninstallOpenCodeFn = func(n string) (string, error) { return "/fake/open/" + n, nil }

	var outBuf bytes.Buffer
	origOut := stdout
	stdout = &outBuf
	defer func() { stdout = origOut }()

	code := Dispatch([]string{"remove", "brainstorm"})
	if code != 0 {
		t.Errorf("Dispatch([remove brainstorm]) = %d, want 0", code)
	}
}

func TestDispatch_List_RoutesToCmdList(t *testing.T) {
	// "list" routes to cmdList — stub catalog to avoid real FS.
	origCatalog := loadCatalogFn
	defer func() { loadCatalogFn = origCatalog }()
	loadCatalogFn = func() ([]installer.Skill, error) {
		return []installer.Skill{{Name: "brainstorm", Description: "test"}}, nil
	}

	var outBuf bytes.Buffer
	origOut := stdout
	stdout = &outBuf
	defer func() { stdout = origOut }()

	code := Dispatch([]string{"list"})
	if code != 0 {
		t.Errorf("Dispatch([list]) = %d, want 0", code)
	}
	if !strings.Contains(outBuf.String(), "brainstorm") {
		t.Errorf("Dispatch([list]) output missing skill, got: %q", outBuf.String())
	}
}

func TestDispatch_PrintUsage_MentionsNewVerbs(t *testing.T) {
	var outBuf bytes.Buffer
	origOut := stdout
	stdout = &outBuf
	defer func() { stdout = origOut }()

	code := Dispatch([]string{"help"})
	if code != 0 {
		t.Errorf("Dispatch([help]) = %d, want 0", code)
	}
	out := outBuf.String()
	for _, verb := range []string{"setup", "remove", "list"} {
		if !strings.Contains(out, verb) {
			t.Errorf("Dispatch([help]) output missing verb %q", verb)
		}
	}
}

func TestDispatch_FlagFirstImplicitServe(t *testing.T) {
	// --listen=... starts with '-', should route to cmdServe (not return 2).
	// We inject alreadyRunningCheck to make it return 0 immediately.
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer stub.Close()

	origCheck := alreadyRunningCheck
	defer func() { alreadyRunningCheck = origCheck }()
	alreadyRunningCheck = func(_ string) bool { return true }

	code := Dispatch([]string{"--listen=:9001"})
	if code != 0 {
		t.Errorf("Dispatch([--listen=:9001]) = %d, want 0 (implicit serve)", code)
	}
}
