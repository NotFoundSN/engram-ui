package render

import (
	"strings"
	"testing"
)

func TestMarkdown_TaskListCheckedRendersCheckbox(t *testing.T) {
	input := "- [x] done item\n- [ ] pending item\n"
	out := Markdown(input)

	if !strings.Contains(out, `<input`) {
		t.Errorf("expected <input> element for task list, got:\n%s", out)
	}
	if !strings.Contains(out, `type="checkbox"`) {
		t.Errorf("expected type=\"checkbox\" attribute, got:\n%s", out)
	}
	if !strings.Contains(out, `checked`) {
		t.Errorf("expected `checked` attribute on done task, got:\n%s", out)
	}
	if !strings.Contains(out, `disabled`) {
		t.Errorf("expected `disabled` attribute (task lists are not interactive), got:\n%s", out)
	}
	// Source text after the checkbox must remain.
	if !strings.Contains(out, "done item") {
		t.Errorf("expected 'done item' text in output, got:\n%s", out)
	}
	if !strings.Contains(out, "pending item") {
		t.Errorf("expected 'pending item' text in output, got:\n%s", out)
	}
}

func TestMarkdown_SoftWrapDoesNotProduceBr(t *testing.T) {
	// Markdown source uses single newlines as soft wraps. CommonMark default
	// behavior joins them into a single paragraph without <br>. The previous
	// WithHardWraps option mangled this.
	input := "Line one of paragraph\nLine two of paragraph\nLine three of paragraph\n"
	out := Markdown(input)

	if strings.Contains(out, "<br") {
		t.Errorf("soft-wrapped paragraph must NOT contain <br>, got:\n%s", out)
	}
	// All three lines must still appear (joined into single paragraph).
	for _, want := range []string{"Line one", "Line two", "Line three"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}

func TestMarkdown_ExplicitHardBreakStillRespected(t *testing.T) {
	// Trailing two-space + newline is the CommonMark hard-break marker.
	// Removing WithHardWraps must not break this explicit signal.
	input := "Line A  \nLine B\n"
	out := Markdown(input)

	if !strings.Contains(out, "<br") {
		t.Errorf("explicit `  \\n` hard break must produce <br>, got:\n%s", out)
	}
}

func TestMarkdown_DoubleNewlineMakesNewParagraph(t *testing.T) {
	input := "Paragraph one.\n\nParagraph two.\n"
	out := Markdown(input)

	// Two distinct <p> elements expected.
	count := strings.Count(out, "<p>")
	if count < 2 {
		t.Errorf("expected at least 2 <p> tags for blank-line-separated paragraphs, got %d in:\n%s", count, out)
	}
}
