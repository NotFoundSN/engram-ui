package render

// Truncate returns s clipped to at most maxRunes Unicode code points.
// If s has more runes than maxRunes, the result has exactly maxRunes runes
// (no ellipsis appended — callers decide).
// Rune-safe: never splits a multi-byte UTF-8 sequence.
func Truncate(s string, maxRunes int) string {
	count := 0
	for i := range s {
		if count == maxRunes {
			return s[:i]
		}
		count++
	}
	return s
}
