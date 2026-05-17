// Package render — time helpers for the UI.
//
// engram emits timestamps as RFC3339 strings, but tests and older fixtures
// sometimes ship date-only (`2026-05-16`) values. Both parse formats are
// supported here so callers can hand us whatever the API returned and we'll
// produce a human-readable relative or absolute label.
package render

import (
	"fmt"
	"strings"
	"time"
)

// parseISO tries the formats we accept from engram and from older mock data.
// Returns the zero Time and false if nothing matched — callers decide how to
// render the unknown case (most just omit the time UI).
func parseISO(iso string) (time.Time, bool) {
	iso = strings.TrimSpace(iso)
	if iso == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, iso); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// TimeAgo formats iso as a human-readable relative timestamp like "5m ago"
// or "3w ago". Returns "" when iso fails to parse — the caller can omit the
// time element instead of rendering a broken label. Future timestamps (clock
// skew, mock data) clamp to "just now".
func TimeAgo(iso string, now time.Time) string {
	t, ok := parseISO(iso)
	if !ok {
		return ""
	}
	d := now.Sub(t)
	if d < 0 {
		return "just now"
	}
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d/time.Minute))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d/time.Hour))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d/(24*time.Hour)))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d/(7*24*time.Hour)))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo ago", int(d/(30*24*time.Hour)))
	default:
		return fmt.Sprintf("%dy ago", int(d/(365*24*time.Hour)))
	}
}

// TimeAgoNow is the convenience wrapper most callers use — they don't need
// a custom `now` outside tests.
func TimeAgoNow(iso string) string {
	return TimeAgo(iso, time.Now())
}

// FormatDateTime produces a readable absolute timestamp for tooltips. Returns
// the raw input when parsing fails so we never erase information the caller
// gave us. Format: "Jan 2, 2006 15:04" (UTC).
func FormatDateTime(iso string) string {
	t, ok := parseISO(iso)
	if !ok {
		return iso
	}
	return t.UTC().Format("Jan 2, 2006 15:04 MST")
}

// FormatDate produces a date-only label like "May 16, 2026". Used inline in
// detail headers where the time-of-day is noise — the full timestamp is
// available via FormatDateTime in the tooltip.
func FormatDate(iso string) string {
	t, ok := parseISO(iso)
	if !ok {
		return iso
	}
	return t.UTC().Format("Jan 2, 2006")
}
