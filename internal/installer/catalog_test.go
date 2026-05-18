package installer

import (
	"io/fs"
	"path"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// skillFrontmatterWithTriggers includes triggers for testing bilingual coverage.
type skillFrontmatterWithTriggers struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Triggers    []string `yaml:"triggers"`
}

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

// loadOpenCodeSkillFrontmatter reads and parses the OpenCode variant SKILL.md
// frontmatter including triggers for testing.
func loadOpenCodeSkillFrontmatter(skillName string) (*skillFrontmatterWithTriggers, error) {
	skillPath := path.Join(skillsEmbedRoot, skillName, "opencode", "SKILL.md")
	data, err := skillFS.ReadFile(skillPath)
	if err != nil {
		return nil, err
	}

	body := string(data)
	if !strings.HasPrefix(body, "---") {
		return &skillFrontmatterWithTriggers{}, nil
	}

	rest := body[3:]
	if strings.HasPrefix(rest, "\r\n") {
		rest = rest[2:]
	} else if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	}

	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return nil, fs.ErrInvalid
	}
	yamlBlock := rest[:endIdx]

	var fm skillFrontmatterWithTriggers
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil, err
	}
	return &fm, nil
}

func TestEngramConventionsOpenCode_SpanishTriggersPresent(t *testing.T) {
	fm, err := loadOpenCodeSkillFrontmatter("engram-conventions")
	if err != nil {
		t.Fatalf("loadOpenCodeSkillFrontmatter(): %v", err)
	}

	requiredSpanish := []string{
		"guardar en engram",
		"guardar en memoria",
		"buscar en engram",
		"buscar memoria",
		"resumen de sesion",
		"cerrar sesion",
		"clave de tema",
		"tipos de observacion",
		"que tipo usar",
	}

	triggerSet := make(map[string]bool)
	for _, tr := range fm.Triggers {
		triggerSet[tr] = true
	}

	for _, req := range requiredSpanish {
		if !triggerSet[req] {
			t.Errorf("Missing required Spanish trigger: %q", req)
		}
	}
}

func TestEngramConventionsOpenCode_EnglishTriggersPreserved(t *testing.T) {
	fm, err := loadOpenCodeSkillFrontmatter("engram-conventions")
	if err != nil {
		t.Fatalf("loadOpenCodeSkillFrontmatter(): %v", err)
	}

	requiredEnglish := []string{
		"mem_save",
		"mem_search",
		"mem_context",
		"mem_session_summary",
		"mem_judge",
		"mem_update",
		"save to engram",
		"engram memory",
		"observation save",
		"topic_key",
	}

	triggerSet := make(map[string]bool)
	for _, tr := range fm.Triggers {
		triggerSet[tr] = true
	}

	for _, req := range requiredEnglish {
		if !triggerSet[req] {
			t.Errorf("Missing required English trigger: %q", req)
		}
	}
}

func TestEngramConventionsOpenCode_NoBareVerbTriggers(t *testing.T) {
	fm, err := loadOpenCodeSkillFrontmatter("engram-conventions")
	if err != nil {
		t.Fatalf("loadOpenCodeSkillFrontmatter(): %v", err)
	}

	bareVerbs := []string{"guardar", "buscar", "save", "search"}
	bareSet := make(map[string]bool)
	for _, v := range bareVerbs {
		bareSet[v] = true
	}

	for _, tr := range fm.Triggers {
		if bareSet[tr] {
			t.Errorf("Bare verb trigger found (should be memory-domain bound): %q", tr)
		}
	}
}

func TestEngramConventionsOpenCode_TriggerQualityRules(t *testing.T) {
	fm, err := loadOpenCodeSkillFrontmatter("engram-conventions")
	if err != nil {
		t.Fatalf("loadOpenCodeSkillFrontmatter(): %v", err)
	}

	memoryDomainWords := []string{"engram", "memoria", "sesion", "tema", "observacion", "tipo", "memory", "observation"}
	memorySet := make(map[string]bool)
	for _, w := range memoryDomainWords {
		memorySet[w] = true
	}

	// Technical identifiers (mem_*, topic_key) are exceptions to word count rules
	isTechnicalID := func(tr string) bool {
		return strings.HasPrefix(tr, "mem_") || tr == "topic_key"
	}

	for _, tr := range fm.Triggers {
		// Skip quality rules for technical identifiers (they are single-word by design)
		if isTechnicalID(tr) {
			continue
		}

		// Length check: 2-4 words for natural language triggers
		words := strings.Fields(tr)
		if len(words) < 2 || len(words) > 4 {
			t.Errorf("Trigger %q has %d words; expected 2-4 words", tr, len(words))
		}

		// Memory-domain bound check
		hasMemoryWord := false
		for _, w := range words {
			// Normalize for comparison (lowercase, strip punctuation)
			clean := strings.ToLower(strings.TrimSuffix(w, "_"))
			if memorySet[clean] {
				hasMemoryWord = true
				break
			}
		}
		if !hasMemoryWord {
			t.Errorf("Trigger %q is not memory-domain bound", tr)
		}
	}
}
