package server

import (
	"strings"
	"testing"
)

func TestValidateFrom_Table(t *testing.T) {
	// Build a 2049-rune string starting with '/'
	overLength := "/" + strings.Repeat("a", 2048)
	// Build a 2048-rune string starting with '/'
	atLength := "/" + strings.Repeat("a", 2047)

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "root",
			input:    "/",
			expected: "/",
		},
		{
			name:     "simple path",
			input:    "/p/alpha",
			expected: "/p/alpha",
		},
		{
			name:     "path with query",
			input:    "/p/alpha?type=decision",
			expected: "/p/alpha?type=decision",
		},
		{
			name:     "path with multi-query",
			input:    "/p/alpha?type=decision&q=auth&sort=date_asc",
			expected: "/p/alpha?type=decision&q=auth&sort=date_asc",
		},
		{
			name:     "no leading slash",
			input:    "p/alpha",
			expected: "",
		},
		{
			name:     "http scheme",
			input:    "http://evil.example.com",
			expected: "",
		},
		{
			name:     "https scheme",
			input:    "https://evil.example.com",
			expected: "",
		},
		{
			name:     "javascript scheme",
			input:    "javascript:alert(1)",
			expected: "",
		},
		{
			name:     "data scheme",
			input:    "data:text/html,...",
			expected: "",
		},
		{
			name:     "protocol-relative",
			input:    "//evil.example.com",
			expected: "",
		},
		{
			name:     "protocol-relative root",
			input:    "//",
			expected: "",
		},
		{
			name:     "traversal absolute",
			input:    "/p/foo/../../etc/passwd",
			expected: "",
		},
		{
			name:     "traversal at start",
			input:    "/../foo",
			expected: "",
		},
		{
			name:     "traversal at end",
			input:    "/foo/..",
			expected: "",
		},
		{
			name:     "traversal in segment",
			input:    "/foo..bar",
			expected: "",
		},
		{
			name:     "over-length",
			input:    overLength,
			expected: "",
		},
		{
			name:     "at-length boundary",
			input:    atLength,
			expected: atLength,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := validateFrom(tc.input)
			if got != tc.expected {
				t.Errorf("validateFrom(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}
