package installer

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill describes one installable skill discovered from the embedded tree.
type Skill struct {
	Name        string // skill identifier (from YAML frontmatter `name`)
	Description string // human-readable summary (from YAML frontmatter `description`)
}

// skillFrontmatter is the minimal shape we need from SKILL.md frontmatter.
type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// LoadCatalog walks the embedded skills tree at skills/*/claude/SKILL.md
// and returns one Skill entry per discovered file, sorted ascending by Name.
// A skill is only catalogued when its claude/ variant exists — this is the
// canonical entry, the opencode/ variant is a sibling for cross-tool install.
//
// Skills that are not yet migrated to the claude/+opencode/ layout are
// silently skipped (no claude/SKILL.md → not discoverable).
func LoadCatalog() ([]Skill, error) {
	entries, err := fs.ReadDir(skillFS, skillsEmbedRoot)
	if err != nil {
		return nil, fmt.Errorf("read skills root: %w", err)
	}

	skills := make([]Skill, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		claudeSkill := path.Join(skillsEmbedRoot, e.Name(), "claude", "SKILL.md")
		data, err := skillFS.ReadFile(claudeSkill)
		if err != nil {
			// Skill not migrated to claude/+opencode/ layout — skip.
			continue
		}

		fm, err := parseFrontmatter(data)
		if err != nil {
			return nil, fmt.Errorf("parse frontmatter %s: %w", claudeSkill, err)
		}
		if fm.Name == "" {
			return nil, fmt.Errorf("frontmatter at %s missing required field 'name'", claudeSkill)
		}

		skills = append(skills, Skill{
			Name:        fm.Name,
			Description: strings.TrimSpace(fm.Description),
		})
	}

	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})
	return skills, nil
}

// parseFrontmatter extracts the YAML frontmatter block delimited by `---`
// markers at the top of a SKILL.md file and unmarshals it into skillFrontmatter.
// Returns an empty struct (not an error) when no frontmatter is present.
func parseFrontmatter(data []byte) (skillFrontmatter, error) {
	body := string(data)
	if !strings.HasPrefix(body, "---") {
		return skillFrontmatter{}, nil
	}

	// Find the closing `---` after the first line.
	rest := body[3:]
	// Trim leading newline after opening marker.
	if strings.HasPrefix(rest, "\r\n") {
		rest = rest[2:]
	} else if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	}

	endIdx := strings.Index(rest, "\n---")
	if endIdx < 0 {
		return skillFrontmatter{}, fmt.Errorf("frontmatter opening `---` has no closing marker")
	}
	yamlBlock := rest[:endIdx]

	var fm skillFrontmatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return skillFrontmatter{}, fmt.Errorf("yaml unmarshal: %w", err)
	}
	return fm, nil
}
