package installer

import "embed"

// skillFS contains the embedded engram-conventions skill payload.
// The path is relative to this file (internal/installer/), so it resolves
// to internal/installer/skills/engram-conventions/.
//
//go:embed all:skills/engram-conventions
var skillFS embed.FS

// skillEmbedRoot is the root path within skillFS for the engram-conventions skill.
const skillEmbedRoot = "skills/engram-conventions"
