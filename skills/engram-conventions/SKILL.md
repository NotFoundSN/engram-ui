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

---

## Compatibility

Designed against the Agent Skills spec. Works with:
- **Claude Code** — loads from `~/.claude/skills/` (global) or `.claude/skills/`
  (project-local).
- **opencode** — loads from its configured skills directory.

**Installation**: keep the canonical copy at `skills/engram-conventions/` and
symlink or copy into each tool's skills location. A single source of truth
avoids drift between tool copies.

**Version compatibility**: relies only on standard MCP tool names (`mem_save`,
`mem_search`, etc.) and the Agent Skills spec `SKILL.md` loading convention.
No version-specific feature flags required.
