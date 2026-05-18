package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/NotFoundSN/engram-ui/internal/installer"
)

func withListSeam(outBuf, errBuf *bytes.Buffer, catalogFn func() ([]installer.Skill, error), fn func()) {
	origOut, origErr := stdout, stderr
	stdout, stderr = outBuf, errBuf
	origCatalog := loadCatalogFn
	loadCatalogFn = catalogFn
	defer func() {
		stdout, stderr = origOut, origErr
		loadCatalogFn = origCatalog
	}()
	fn()
}

var sampleSkills = []installer.Skill{
	{Name: "brainstorm", Description: "AI brainstorming tool"},
	{Name: "debug", Description: "debugging helper"},
}

// --- SCN-08: list — plain text ---

func TestCmdList_PlainText_OneLinePerSkill(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	withListSeam(&outBuf, &errBuf, func() ([]installer.Skill, error) { return sampleSkills, nil }, func() {
		code := cmdList([]string{})
		if code != 0 {
			t.Errorf("cmdList([]) = %d, want 0", code)
		}
	})

	out := outBuf.String()
	if !strings.Contains(out, "brainstorm — AI brainstorming tool") {
		t.Errorf("stdout missing brainstorm line, got: %q", out)
	}
	if !strings.Contains(out, "debug — debugging helper") {
		t.Errorf("stdout missing debug line, got: %q", out)
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr should be empty, got: %q", errBuf.String())
	}
}

// --- AUX (REQ-3.2): list plain-text emits "name — " for empty description ---

func TestCmdList_PlainText_EmptyDescription_EmitsNameAndDash(t *testing.T) {
	// REQ-3.2 contract: every skill produces a line in the form `name — description`.
	// With empty description, the trailing " — " is preserved (spec-literal).
	// Catalog data today has no empty-description skills; this guards future drift.
	mixed := []installer.Skill{
		{Name: "noop", Description: ""},
		{Name: "brainstorm", Description: "AI brainstorming tool"},
	}
	var outBuf, errBuf bytes.Buffer
	withListSeam(&outBuf, &errBuf, func() ([]installer.Skill, error) { return mixed, nil }, func() {
		code := cmdList([]string{})
		if code != 0 {
			t.Errorf("cmdList([]) = %d, want 0", code)
		}
	})

	out := outBuf.String()
	if !strings.Contains(out, "noop — \n") {
		t.Errorf("empty-description line should be 'noop — \\n', got: %q", out)
	}
	if !strings.Contains(out, "brainstorm — AI brainstorming tool") {
		t.Errorf("non-empty description line missing, got: %q", out)
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr should be empty, got: %q", errBuf.String())
	}
}

// --- SCN-09: list --json — machine-readable ---

func TestCmdList_JSON_ValidArray(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	withListSeam(&outBuf, &errBuf, func() ([]installer.Skill, error) { return sampleSkills, nil }, func() {
		code := cmdList([]string{"--json"})
		if code != 0 {
			t.Errorf("cmdList([--json]) = %d, want 0", code)
		}
	})

	if errBuf.Len() != 0 {
		t.Errorf("stderr should be empty, got: %q", errBuf.String())
	}

	var entries []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &entries); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", err, outBuf.String())
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "brainstorm" || entries[0].Description != "AI brainstorming tool" {
		t.Errorf("entries[0] = %+v, want brainstorm/AI brainstorming tool", entries[0])
	}
	if entries[1].Name != "debug" || entries[1].Description != "debugging helper" {
		t.Errorf("entries[1] = %+v, want debug/debugging helper", entries[1])
	}
}

// --- SCN-16: list --json stdout purity under error ---

func TestCmdList_JSON_StdoutPurity_CatalogError(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	withListSeam(&outBuf, &errBuf, func() ([]installer.Skill, error) {
		return nil, errors.New("catalog unavailable")
	}, func() {
		code := cmdList([]string{"--json"})
		if code != 1 {
			t.Errorf("cmdList([--json] catalog error) = %d, want 1", code)
		}
	})

	// stdout must be empty (purity rule)
	if outBuf.Len() != 0 {
		t.Errorf("stdout must be empty on catalog error, got: %q", outBuf.String())
	}
	if !strings.Contains(errBuf.String(), "catalog unavailable") {
		t.Errorf("stderr should mention catalog error, got: %q", errBuf.String())
	}
}

// --- AUX (REQ-3.3): list --json empty catalog emits [] ---

func TestCmdList_JSON_EmptyCatalog(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	withListSeam(&outBuf, &errBuf, func() ([]installer.Skill, error) { return []installer.Skill{}, nil }, func() {
		code := cmdList([]string{"--json"})
		if code != 0 {
			t.Errorf("cmdList([--json] empty) = %d, want 0", code)
		}
	})

	// should emit [] followed by newline
	trimmed := strings.TrimSpace(outBuf.String())
	if trimmed != "[]" {
		t.Errorf("empty catalog JSON should be [], got: %q", outBuf.String())
	}
	if errBuf.Len() != 0 {
		t.Errorf("stderr should be empty, got: %q", errBuf.String())
	}
}

// --- AUX (REQ-3.2): list plain-text empty catalog emits zero lines ---

func TestCmdList_PlainText_EmptyCatalog(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	withListSeam(&outBuf, &errBuf, func() ([]installer.Skill, error) { return []installer.Skill{}, nil }, func() {
		code := cmdList([]string{})
		if code != 0 {
			t.Errorf("cmdList([]) empty = %d, want 0", code)
		}
	})

	if outBuf.Len() != 0 {
		t.Errorf("empty catalog plain-text should emit zero lines, got: %q", outBuf.String())
	}
}

// --- AUX (REQ-6.3): list unexpected positional arg → exit 2 ---

func TestCmdList_UnexpectedArg_Exit2(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	withListSeam(&outBuf, &errBuf, func() ([]installer.Skill, error) { return sampleSkills, nil }, func() {
		code := cmdList([]string{"unexpected"})
		if code != 2 {
			t.Errorf("cmdList([unexpected]) = %d, want 2", code)
		}
	})

	if !strings.Contains(errBuf.String(), "unexpected") {
		t.Errorf("stderr should mention unexpected arg, got: %q", errBuf.String())
	}
}

// --- AUX (REQ-3.5 + REQ-6.2): list plain-text catalog error → stderr + exit 1 ---

func TestCmdList_PlainText_CatalogError_Exit1(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	withListSeam(&outBuf, &errBuf, func() ([]installer.Skill, error) {
		return nil, errors.New("disk error")
	}, func() {
		code := cmdList([]string{})
		if code != 1 {
			t.Errorf("cmdList([]) catalog error = %d, want 1", code)
		}
	})

	if outBuf.Len() != 0 {
		t.Errorf("stdout must be empty on error, got: %q", outBuf.String())
	}
	if !strings.Contains(errBuf.String(), "disk error") {
		t.Errorf("stderr should contain error, got: %q", errBuf.String())
	}
}
