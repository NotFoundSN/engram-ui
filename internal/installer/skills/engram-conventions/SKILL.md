---
name: engram-conventions
description: >
  Canonical conventions for saving observations to engram memory. Load this
  whenever you are about to call mem_save, mem_search, mem_context,
  mem_session_summary, mem_judge, mem_update, or any engram tool. Covers
  observation type taxonomy, topic_key namespacing, multi-repo setup via
  .engram/config.json, session lifecycle, conflict resolution, and workflow
  cookbooks for SDD, Superpowers, and ad-hoc saves.
triggers:
  - mem_save
  - mem_search
  - mem_context
  - mem_session_summary
  - mem_judge
  - mem_update
  - save to engram
  - engram memory
  - observation save
  - topic_key
---

# engram-conventions

Portable, opinionated guidance on top of engram's open schema. Defines a
canonical taxonomy of 14 observation types, a `<namespace>/<feature>[/<phase>]`
convention for `topic_key`, multi-repo handling, full lifecycle operations, and
per-workflow cookbooks for SDD, Superpowers, and ad-hoc saves.

Zero changes to engram's HTTP API or MCP server. Discipline lives in this skill,
not in the API.

---

## Quick Rules

1. **type** — always pick from the canonical list. See `types.md`.
2. **topic_key** — always use the namespace convention. See `topic-keys.md`.
3. **multi-repo** — if this looks like a multi-repo product, check/create
   `.engram/config.json` before saving. See `multi-repo.md`.
4. **reads** — use the same namespace convention in queries; bare keyword
   search is noisy in busy projects.
5. **session end** — always call `mem_session_summary` before saying "done".
   See `lifecycle.md`.
6. **shareable URLs** — when the user is likely to want to read or validate
   saved content (spec/design/proposal/plan/report just generated, interactive
   review mode, substantial content), SHOULD emit
   `Review: http://localhost:7438/observations/{id}` (`{id}` from the
   `mem_save` response) so they can review in engram-ui. MAY skip for trivial
   saves or autonomous chains. See `lifecycle.md` for full format,
   `ENGRAM_UI_URL` override, fallback, and situational guidance.

---

## Decision Table

| Situation | Read |
|-----------|------|
| Choosing a `type` for any save | `types.md` |
| Setting a `topic_key` | `topic-keys.md` |
| Multiple repos that look like one product | `multi-repo.md` |
| Reading back past work (search, context, timeline) | `lifecycle.md` |
| Resolving a conflict (`mem_judge`) | `lifecycle.md` |
| Updating an existing observation | `lifecycle.md` |
| Saving output from an SDD phase | `workflows/sdd.md` |
| Saving output from a Superpowers skill | `workflows/superpowers.md` |
| Saving a decision, bugfix, discovery, pattern, config, or preference | `workflows/ad-hoc.md` |
| Surfacing a saved memory's URL to the user | `lifecycle.md#surfacing-memories-to-the-user-via-url` |
| Choosing between a `.md` file and an engram save | `lifecycle.md#when-to-use-engram-vs-a-standalone-md-file` |

---

## Compatibility

Designed against the Agent Skills spec. Works with:
- **Claude Code** — loads `SKILL.md` from `~/.claude/skills/` (global) or
  `.claude/skills/` (project-local).
- **OpenCode** — loads from several locations (any of these works):
  - `~/.config/opencode/skills/` (global native — on Windows, this is the
    literal path `%USERPROFILE%\.config\opencode\skills`, NOT `%APPDATA%`)
  - `~/.claude/skills/` (cross-tool — OpenCode also reads Claude Code's
    skills directory)
  - `~/.agents/skills/` (Agent Skills spec global)
  - Project-local equivalents: `.opencode/skills/`, `.claude/skills/`,
    `.agents/skills/`
  - Override the global root via `$OPENCODE_CONFIG_DIR`

> **Tip**: running `engram-ui setup claude-code` also makes the skill
> available to OpenCode via `~/.claude/skills/` (cross-tool path). One
> install command can cover both agents.

**Installation**: run `engram-ui setup claude-code` or
`engram-ui setup opencode`. The canonical skill source lives in this
repository at `internal/installer/skills/engram-conventions/` and is
embedded into the `engram-ui` binary at build time via `//go:embed`. The
setup commands copy the embedded payload into the target agent's skills
directory (`~/.claude/skills/engram-conventions/` for Claude Code, etc.).
Re-run the setup command after upgrading `engram-ui` to refresh the local
copy.

**Version compatibility**: relies only on standard MCP tool names (`mem_save`,
`mem_search`, etc.) and the Agent Skills spec `SKILL.md` loading convention.
No version-specific feature flags required.
