package server

import "testing"

func TestBuildSourceURL(t *testing.T) {
	cases := []struct {
		name           string
		project        string
		typ            string
		q              string
		sort           string
		sortExplicit   bool
		topicKeyPrefix string
		expected       string
	}{
		{
			name:         "no filters",
			project:      "alpha",
			sort:         "date_desc",
			sortExplicit: false,
			expected:     "/p/alpha",
		},
		{
			name:         "type only",
			project:      "alpha",
			typ:          "decision",
			sort:         "date_desc",
			sortExplicit: false,
			expected:     "/p/alpha?type=decision",
		},
		{
			name:         "q only",
			project:      "alpha",
			q:            "auth",
			sort:         "date_desc",
			sortExplicit: false,
			expected:     "/p/alpha?q=auth",
		},
		{
			name:         "explicit sort only",
			project:      "alpha",
			sort:         "date_asc",
			sortExplicit: true,
			expected:     "/p/alpha?sort=date_asc",
		},
		{
			name:         "type and explicit sort — matches spec scenario 5 pre-escape",
			project:      "alpha",
			typ:          "decision",
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
			sort:         "date_desc",
			sortExplicit: true,
			expected:     "/p/alpha?sort=date_desc",
		},
		{
			name:         "special chars in q are url-escaped",
			project:      "alpha",
			q:            "auth&bug",
			sort:         "date_desc",
			sortExplicit: false,
			expected:     "/p/alpha?q=auth%26bug",
		},
		{
			name:           "topic_key_prefix only",
			project:        "alpha",
			sort:           "date_desc",
			sortExplicit:   false,
			topicKeyPrefix: "sdd/auth/",
			expected:       "/p/alpha?topic_key_prefix=sdd%2Fauth%2F",
		},
		{
			name:           "topic_key_prefix combined with type and explicit sort",
			project:        "alpha",
			typ:            "spec",
			sort:           "date_asc",
			sortExplicit:   true,
			topicKeyPrefix: "sdd/auth/",
			expected:       "/p/alpha?type=spec&sort=date_asc&topic_key_prefix=sdd%2Fauth%2F",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := buildSourceURL(tc.project, tc.typ, tc.q, tc.sort, tc.sortExplicit, tc.topicKeyPrefix)
			if got != tc.expected {
				t.Errorf("buildSourceURL(%q,%q,%q,%q,%v,%q) = %q; want %q",
					tc.project, tc.typ, tc.q, tc.sort, tc.sortExplicit, tc.topicKeyPrefix, got, tc.expected)
			}
		})
	}
}
