// Package render converts engram observation content into safe HTML.
package render

import (
	"bytes"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var (
	md = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		// WithHardWraps intentionally OMITTED: source single-newlines are
		// soft wraps (CommonMark default). Explicit hard breaks via
		// trailing two spaces still produce <br>.
	)
	policy = buildPolicy()
)

// buildPolicy starts from bluemonday's UGCPolicy and additionally allows the
// GFM task-list checkboxes goldmark emits (`<input type="checkbox" disabled
// [checked]>`). UGCPolicy strips <input> entirely by default.
func buildPolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowElements("input")
	p.AllowAttrs("type").Matching(bluemonday.SpaceSeparatedTokens).OnElements("input")
	p.AllowAttrs("checked", "disabled").OnElements("input")
	return p
}

// Markdown converts the given markdown input to sanitized HTML.
//
// The output is always run through bluemonday. If goldmark fails (it
// generally does not), the raw input is sanitized as a plain string so
// we never emit untrusted HTML.
func Markdown(input string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		return policy.Sanitize(input)
	}
	return policy.Sanitize(buf.String())
}
