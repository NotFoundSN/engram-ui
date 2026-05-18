package version

import (
	"testing"
)

func TestCurrent(t *testing.T) {
	// Test that Current() returns the version string
	got := Current()
	if got == "" {
		t.Error("Current() returned empty string, expected non-empty version")
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		name          string
		version       string
		wantMajor     int
		wantMinor     int
		wantPatch     int
		wantErr       bool
	}{
		{
			name:      "standard semver v1.2.3",
			version:   "v1.2.3",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 3,
			wantErr:   false,
		},
		{
			name:      "semver without v prefix 2.5.1",
			version:   "2.5.1",
			wantMajor: 2,
			wantMinor: 5,
			wantPatch: 1,
			wantErr:   false,
		},
		{
			name:      "dev version",
			version:   "dev",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   true,
		},
		{
			name:      "empty version",
			version:   "",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   true,
		},
		{
			name:      "invalid format",
			version:   "v1.2",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   true,
		},
		{
			name:      "go install version format",
			version:   "v0.1.0",
			wantMajor: 0,
			wantMinor: 1,
			wantPatch: 0,
			wantErr:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			major, minor, patch, err := Parse(tc.version)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got nil", tc.version)
				}
				return
			}
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tc.version, err)
				return
			}
			if major != tc.wantMajor || minor != tc.wantMinor || patch != tc.wantPatch {
				t.Errorf("Parse(%q) = (%d, %d, %d), want (%d, %d, %d)",
					tc.version, major, minor, patch, tc.wantMajor, tc.wantMinor, tc.wantPatch)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	cases := []struct {
		name string
		v1   string
		v2   string
		want int // negative: v1 < v2, 0: equal, positive: v1 > v2
	}{
		{
			name: "equal versions v1.2.3",
			v1:   "v1.2.3",
			v2:   "v1.2.3",
			want: 0,
		},
		{
			name: "equal without v prefix",
			v1:   "1.2.3",
			v2:   "v1.2.3",
			want: 0,
		},
		{
			name: "v1 less than v2",
			v1:   "v1.0.0",
			v2:   "v2.0.0",
			want: -1,
		},
		{
			name: "v1 greater than v2",
			v1:   "v2.5.0",
			v2:   "v2.4.9",
			want: 1,
		},
		{
			name: "minor version difference",
			v1:   "v1.5.0",
			v2:   "v1.4.0",
			want: 1,
		},
		{
			name: "patch version difference",
			v1:   "v1.0.1",
			v2:   "v1.0.0",
			want: 1,
		},
		{
			name: "dev is less than release",
			v1:   "dev",
			v2:   "v1.0.0",
			want: -1,
		},
		{
			name: "dev equal to dev",
			v1:   "dev",
			v2:   "dev",
			want: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Compare(tc.v1, tc.v2)
			if tc.want == 0 {
				if got != 0 {
					t.Errorf("Compare(%q, %q) = %d, want 0", tc.v1, tc.v2, got)
				}
			} else if tc.want < 0 {
				if got >= 0 {
					t.Errorf("Compare(%q, %q) = %d, want negative", tc.v1, tc.v2, got)
				}
			} else {
				if got <= 0 {
					t.Errorf("Compare(%q, %q) = %d, want positive", tc.v1, tc.v2, got)
				}
			}
		})
	}
}
