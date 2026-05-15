// Package render converts engram observation content into safe HTML.
package render

import (
	"bytes"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	md = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps()),
	)
	policy = bluemonday.UGCPolicy()
)

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
