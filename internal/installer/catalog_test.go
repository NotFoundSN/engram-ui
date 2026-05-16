package installer

import (
	"strings"
	"testing"
)

func TestLoadCatalog_FindsBrainstorm(t *testing.T) {
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog(): %v", err)
	}

	var found *Skill
	for i := range catalog {
		if catalog[i].Name == "brainstorm" {
			found = &catalog[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("LoadCatalog: brainstorm skill not found in catalog, got %d skills", len(catalog))
	}

	if found.Description == "" {
		t.Errorf("LoadCatalog: brainstorm description is empty")
	}
	if !strings.Contains(found.Description, "creative or design work") {
		t.Errorf("LoadCatalog: brainstorm description should mention 'creative or design work', got %q", found.Description)
	}
}

func TestLoadCatalog_NameFromFrontmatterNotFolder(t *testing.T) {
	// The catalog Name must come from the YAML frontmatter `name:` field,
	// which authoritatively matches Agent Skills spec — not from the folder
	// name (even though they happen to align today, the contract is YAML).
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog(): %v", err)
	}

	for _, s := range catalog {
		if s.Name == "" {
			t.Errorf("LoadCatalog: skill with empty Name found (entry: %+v)", s)
		}
	}
}

func TestLoadCatalog_SortedByName(t *testing.T) {
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog(): %v", err)
	}

	for i := 1; i < len(catalog); i++ {
		if catalog[i-1].Name >= catalog[i].Name {
			t.Errorf("LoadCatalog: catalog not sorted ascending — %q >= %q at index %d",
				catalog[i-1].Name, catalog[i].Name, i)
		}
	}
}

func TestLoadCatalog_FindsDebug(t *testing.T) {
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog(): %v", err)
	}

	var found *Skill
	for i := range catalog {
		if catalog[i].Name == "debug" {
			found = &catalog[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("LoadCatalog: debug skill not found in catalog, got %d skills", len(catalog))
	}

	if found.Description == "" {
		t.Errorf("LoadCatalog: debug description is empty")
	}
	if !strings.Contains(found.Description, "bug") && !strings.Contains(found.Description, "fix") {
		t.Errorf("LoadCatalog: debug description should mention 'bug' or 'fix', got %q", found.Description)
	}
}

func TestLoadCatalog_FindsEngramConventions(t *testing.T) {
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog(): %v", err)
	}

	var found *Skill
	for i := range catalog {
		if catalog[i].Name == "engram-conventions" {
			found = &catalog[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("LoadCatalog: engram-conventions skill not found, got %d skills", len(catalog))
	}

	if found.Description == "" {
		t.Errorf("LoadCatalog: engram-conventions description is empty")
	}
	if !strings.Contains(found.Description, "engram") && !strings.Contains(found.Description, "observation") {
		t.Errorf("LoadCatalog: engram-conventions description should mention 'engram' or 'observation', got %q", found.Description)
	}
}

func TestLoadCatalog_AllThreeSkillsMigrated(t *testing.T) {
	// After Phase 3, all three pre-existing skills are migrated to the
	// per-tool layout and must appear in the catalog.
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog(): %v", err)
	}

	want := map[string]bool{
		"brainstorm":         false,
		"debug":              false,
		"engram-conventions": false,
	}
	for _, s := range catalog {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, present := range want {
		if !present {
			t.Errorf("LoadCatalog: expected skill %q not present in catalog", name)
		}
	}
}
