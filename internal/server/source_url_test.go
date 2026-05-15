package server

import "testing"

func TestBuildSourceURL(t *testing.T) {
	cases := []struct {
		name         string
		project      string
		typ          string
		q            string
		sort         string
		sortExplicit bool
		expected     string
	}{
		{
			name:         "no filters",
			project:      "alpha",
			typ:          "",
			q:            "",
			sort:         "date_desc",
			sortExplicit: false,
			expected:     "/p/alpha",
		},
		{
			name:         "type only",
			project:      "alpha",
			typ:          "decision",
			q:            "",
			sort:         "date_desc",
			sortExplicit: false,
			expected:     "/p/alpha?type=decision",
		},
		{
			name:         "q only",
			project:      "alpha",
			typ:          "",
			q:            "auth",
			sort:         "date_desc",
			sortExplicit: false,
			expected:     "/p/alpha?q=auth",
		},
		{
			name:         "explicit sort only",
			project:      "alpha",
			typ:          "",
			q:            "",
			sort:         "date_asc",
			sortExplicit: true,
			expected:     "/p/alpha?sort=date_asc",
		},
		{
			name:         "type and explicit sort — matches spec scenario 5 pre-escape",
			project:      "alpha",
			typ:          "decision",
			q:            "",
			sort:         "date_asc",
			sortExplicit: true,
			expected:     "/p/alpha?type=decision&sort=date_asc",
		},
		{
			name:         "all three params",
			project:      "alpha",
			typ:          "decision",
			q:            "auth",
			sort:         "date_asc",
			sortExplicit: true,
			expected:     "/p/alpha?type=decision&q=auth&sort=date_asc",
		},
		{
			name:         "explicit date_desc — default value but user typed it",
			project:      "alpha",
			typ:          "",
			q:            "",
			sort:         "date_desc",
			sortExplicit: true,
			expected:     "/p/alpha?sort=date_desc",
		},
		{
			name:         "special chars in q are url-escaped",
			project:      "alpha",
			typ:          "",
			q:            "auth&bug",
			sort:         "date_desc",
			sortExplicit: false,
			expected:     "/p/alpha?q=auth%26bug",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := buildSourceURL(tc.project, tc.typ, tc.q, tc.sort, tc.sortExplicit)
			if got != tc.expected {
				t.Errorf("buildSourceURL(%q,%q,%q,%q,%v) = %q; want %q",
					tc.project, tc.typ, tc.q, tc.sort, tc.sortExplicit, got, tc.expected)
			}
		})
	}
}
