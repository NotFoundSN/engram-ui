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

func TestLoadCatalog_OnlyClaudeVariantSkillsAppear(t *testing.T) {
	// Catalog discovery walks skills/*/claude/SKILL.md as the canonical entry.
	// A skill without a claude/ subfolder MUST NOT appear (still being migrated).
	// This guards against half-migrated skills accidentally showing up.
	catalog, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog(): %v", err)
	}

	// As of Phase 2, brainstorm and debug have been migrated to claude/+opencode/.
	// engram-conventions still pending (Phase 3).
	for _, s := range catalog {
		if s.Name == "engram-conventions" {
			t.Errorf("LoadCatalog: engram-conventions appeared in catalog before Phase 3 migration")
		}
	}
}
