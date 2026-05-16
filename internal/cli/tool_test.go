package cli

import (
	"testing"
)

func TestParseToolFlag(t *testing.T) {
	tests := []struct {
		name        string
		verb        string
		args        []string
		wantSkill   string
		wantTargets []string
		wantUsageErr bool
		wantErr     bool
	}{
		{
			name:        "valid claude",
			verb:        "setup",
			args:        []string{"brainstorm", "--tool=claude"},
			wantSkill:   "brainstorm",
			wantTargets: []string{"claude"},
		},
		{
			name:        "valid opencode",
			verb:        "setup",
			args:        []string{"brainstorm", "--tool=opencode"},
			wantSkill:   "brainstorm",
			wantTargets: []string{"opencode"},
		},
		{
			name:        "valid both explicit",
			verb:        "setup",
			args:        []string{"brainstorm", "--tool=both"},
			wantSkill:   "brainstorm",
			wantTargets: []string{"claude", "opencode"},
		},
		{
			name:        "omitted flag defaults to both",
			verb:        "setup",
			args:        []string{"brainstorm"},
			wantSkill:   "brainstorm",
			wantTargets: []string{"claude", "opencode"},
		},
		{
			name:         "invalid value exits usageErr",
			verb:         "setup",
			args:         []string{"brainstorm", "--tool=foo"},
			wantUsageErr: true,
			wantErr:      true,
		},
		{
			name:         "empty value exits usageErr",
			verb:         "setup",
			args:         []string{"brainstorm", "--tool="},
			wantUsageErr: true,
			wantErr:      true,
		},
		{
			name:         "missing positional exits usageErr",
			verb:         "setup",
			args:         []string{},
			wantUsageErr: true,
			wantErr:      true,
		},
		{
			name:         "extra positional rejected",
			verb:         "setup",
			args:         []string{"brainstorm", "extra"},
			wantUsageErr: true,
			wantErr:      true,
		},
		{
			name:        "remove verb works",
			verb:        "remove",
			args:        []string{"debug", "--tool=claude"},
			wantSkill:   "debug",
			wantTargets: []string{"claude"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name, targets, usageErr, err := parseToolFlag(tc.verb, tc.args)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseToolFlag(%q, %v) expected error, got nil", tc.verb, tc.args)
				}
				if usageErr != tc.wantUsageErr {
					t.Errorf("parseToolFlag(%q, %v) usageErr = %v, want %v", tc.verb, tc.args, usageErr, tc.wantUsageErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseToolFlag(%q, %v) unexpected error: %v", tc.verb, tc.args, err)
			}
			if name != tc.wantSkill {
				t.Errorf("parseToolFlag(%q, %v) name = %q, want %q", tc.verb, tc.args, name, tc.wantSkill)
			}
			if len(targets) != len(tc.wantTargets) {
				t.Fatalf("parseToolFlag(%q, %v) targets = %v, want %v", tc.verb, tc.args, targets, tc.wantTargets)
			}
			for i, tgt := range targets {
				if tgt != tc.wantTargets[i] {
					t.Errorf("parseToolFlag(%q, %v) targets[%d] = %q, want %q", tc.verb, tc.args, i, tgt, tc.wantTargets[i])
				}
			}
		})
	}
}
