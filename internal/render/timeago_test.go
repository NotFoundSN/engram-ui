package render

import (
	"testing"
	"time"
)

func TestTimeAgo(t *testing.T) {
	now := time.Date(2026, 5, 16, 14, 30, 0, 0, time.UTC)
	ago := func(d time.Duration) string {
		return now.Add(-d).Format(time.RFC3339)
	}

	cases := []struct {
		name string
		iso  string
		want string
	}{
		{"unparseable returns empty", "not-a-date", ""},
		{"empty returns empty", "", ""},
		{"future clamps to just now", now.Add(2 * time.Minute).Format(time.RFC3339), "just now"},
		{"under a minute", ago(20 * time.Second), "just now"},
		{"minutes", ago(5 * time.Minute), "5m ago"},
		{"hours", ago(3 * time.Hour), "3h ago"},
		{"days", ago(3 * 24 * time.Hour), "3d ago"},
		{"weeks", ago(14 * 24 * time.Hour), "2w ago"},
		{"months", ago(75 * 24 * time.Hour), "2mo ago"},
		{"years", ago(400 * 24 * time.Hour), "1y ago"},
		{"date-only input", "2026-05-14", "2d ago"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TimeAgo(tc.iso, now)
			if got != tc.want {
				t.Errorf("TimeAgo(%q) = %q, want %q", tc.iso, got, tc.want)
			}
		})
	}
}

func TestFormatDateTime(t *testing.T) {
	cases := []struct {
		name string
		iso  string
		want string
	}{
		{"RFC3339 in UTC", "2026-05-16T14:30:00Z", "May 16, 2026 14:30 UTC"},
		{"date-only renders at midnight", "2026-05-16", "May 16, 2026 00:00 UTC"},
		{"unparseable preserved verbatim", "not-a-date", "not-a-date"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatDateTime(tc.iso)
			if got != tc.want {
				t.Errorf("FormatDateTime(%q) = %q, want %q", tc.iso, got, tc.want)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	cases := []struct {
		name string
		iso  string
		want string
	}{
		{"RFC3339", "2026-05-16T14:30:00Z", "May 16, 2026"},
		{"date-only", "2026-05-16", "May 16, 2026"},
		{"unparseable preserved verbatim", "not-a-date", "not-a-date"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatDate(tc.iso)
			if got != tc.want {
				t.Errorf("FormatDate(%q) = %q, want %q", tc.iso, got, tc.want)
			}
		})
	}
}
