package render

import (
	"testing"
	"unicode/utf8"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxRunes  int
		wantLen   int  // expected rune count of result
		wantEqual bool // true if result should equal input (no truncation)
	}{
		{
			name:      "empty string",
			input:     "",
			maxRunes:  10,
			wantLen:   0,
			wantEqual: true,
		},
		{
			name:      "string shorter than maxRunes",
			input:     "hello",
			maxRunes:  10,
			wantLen:   5,
			wantEqual: true,
		},
		{
			name:      "string exactly maxRunes",
			input:     "hello",
			maxRunes:  5,
			wantLen:   5,
			wantEqual: true,
		},
		{
			name:      "string longer than maxRunes",
			input:     "hello world",
			maxRunes:  5,
			wantLen:   5,
			wantEqual: false,
		},
		{
			name:      "multi-byte runes (emoji) within limit",
			input:     "hi 🎉",
			maxRunes:  10,
			wantLen:   4,
			wantEqual: true,
		},
		{
			name:      "multi-byte runes (emoji) truncated",
			input:     "hello 🎉 world",
			maxRunes:  6,
			wantLen:   6,
			wantEqual: false,
		},
		{
			name:      "CJK characters truncated",
			input:     "こんにちは世界",
			maxRunes:  5,
			wantLen:   5,
			wantEqual: false,
		},
		{
			name:      "exactly 140 runes unchanged",
			input:     string(make([]rune, 140)),
			maxRunes:  140,
			wantLen:   140,
			wantEqual: true,
		},
		{
			name:      "141 runes truncated to 140",
			input:     string(make([]rune, 141)),
			maxRunes:  140,
			wantLen:   140,
			wantEqual: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Truncate(tc.input, tc.maxRunes)

			// Rune count check
			gotRunes := utf8.RuneCountInString(got)
			if gotRunes != tc.wantLen {
				t.Errorf("Truncate(%q, %d) rune count = %d, want %d", tc.input, tc.maxRunes, gotRunes, tc.wantLen)
			}

			// Must be valid UTF-8 (no split multi-byte sequences)
			if !utf8.ValidString(got) {
				t.Errorf("Truncate(%q, %d) produced invalid UTF-8", tc.input, tc.maxRunes)
			}

			// Equality check
			if tc.wantEqual && got != tc.input {
				t.Errorf("Truncate(%q, %d) = %q, want unchanged input", tc.input, tc.maxRunes, got)
			}
			if !tc.wantEqual && got == tc.input {
				t.Errorf("Truncate(%q, %d) = %q, want truncated result", tc.input, tc.maxRunes, got)
			}
		})
	}
}
