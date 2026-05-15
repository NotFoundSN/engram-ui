package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCmdVersion(t *testing.T) {
	var buf bytes.Buffer
	origStdout := stdout
	stdout = &buf
	defer func() { stdout = origStdout }()

	code := cmdVersion()
	if code != 0 {
		t.Errorf("cmdVersion() = %d, want 0", code)
	}

	got := buf.String()
	if !strings.HasPrefix(got, "engram-ui ") {
		t.Errorf("cmdVersion() output %q should start with 'engram-ui '", got)
	}
	// Default version is "dev" in tests (no ldflag injection).
	if !strings.Contains(got, "dev") && !strings.Contains(got, "v") {
		t.Errorf("cmdVersion() output %q should contain version string", got)
	}
}
