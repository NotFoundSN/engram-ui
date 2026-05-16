package installer

import "embed"

// skillFS contains the embedded skills tree. The path is relative to this
// file (internal/installer/), so it resolves to internal/installer/skills/.
//
// Each skill lives at skills/{name}/{tool}/* where {tool} is "claude" or
// "opencode". The catalog (catalog.go) discovers skills by walking
// skills/*/claude/SKILL.md and parsing the YAML frontmatter.
//
//go:embed all:skills
var skillFS embed.FS

// skillsEmbedRoot is the root path within skillFS for all skills.
const skillsEmbedRoot = "skills"
