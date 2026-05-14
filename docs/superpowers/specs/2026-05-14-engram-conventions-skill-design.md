# Design Spec: engram-conventions Skill

**Date**: 2026-05-14
**Project**: engram-ui
**Status**: draft

---

## 1. Overview

Engram stores observations with two free-form fields that are critical for
filtering and grouping: `type` and `topic_key`. The MCP server enforces no
constraints on either field — any string is accepted. This flexibility is
intentional (engram is a generic memory store), but it means that without a
shared convention, every agent and every workflow produces a different shape.
One agent might save `type: "note"`, another `type: "architectural-decision"`,
another `type: "arch"`. One uses `topic_key: "auth"`, another uses
`topic_key: "AUTH-2024"`, a third omits it entirely. The result is a blob that
is hard to query, impossible to group, and unusable by a UI layer.

The `engram-conventions` skill solves this by providing a portable, opinionated
layer of guidance on top of engram's open schema. It defines a canonical
taxonomy of 14 observation types, a `<namespace>/<feature>[/<phase>]`
convention for `topic_key`, handling for multi-repo products, and per-workflow
cookbooks covering SDD, Superpowers, and ad-hoc saves. The skill is pure
guidance — it requires zero changes to engram's HTTP API or MCP server. Any
agent that reads this skill can produce well-shaped observations that
engram-ui's filter, group, and timeline views can consume reliably.

---

## 2. Goals

- Establish a canonical observation `type` taxonomy of 14 types with clear
  "use when / don't use when" rules so agents pick the right type consistently.
- Establish a `topic_key` namespace convention (`<namespace>/<feature>[/<phase>]`)
  that is lowercase kebab-case, max 4 levels deep, and maps predictably to
  workflow phases.
- Handle multi-repo products via `.engram/config.json` so observations from
  related repos land in the same engram project and can be queried together.
- Cover the full observation lifecycle: writes (`mem_save`), reads
  (`mem_search`, `mem_get_observation`, `mem_context`), session summaries
  (`mem_session_summary`), conflict resolution (`mem_judge`), and updates
  (`mem_update`).
- Provide a workflow cookbook for three contexts: SDD (Gentle.AI
  agent-teams-lite), Superpowers (brainstorming/writing-plans), and ad-hoc
  saves not tied to any workflow.
- Be portable across tools: designed against the Agent Skills spec so it works
  with Claude Code, opencode, or any tool that loads `SKILL.md` files.
- Be non-invasive: zero changes to engram's source code. Discipline lives in
  the skill, not the API.

---

## 3. Non-goals

- **Not modifying engram's HTTP API** — no new endpoints, no schema changes.
- **Not modifying engram's MCP server** — no enum enforcement on `type`, no
  `Required` constraint added to any field. Engram remains flexible.
- **Not replacing the existing `engram:memory` skill** — that skill governs
  WHEN to save (proactive triggers, session lifecycle). This skill is
  complementary: it governs HOW to shape the save (which `type`, which
  `topic_key`, what content structure).
- **Not implementing UI changes in engram-ui** — the UI layer that consumes
  these conventions is a separate scope and tracked separately.

---

## 4. Architecture — File Layout

```
skills/engram-conventions/
├── SKILL.md              # entry point — always loaded per Agent Skills spec
├── types.md              # 14 canonical observation types reference
├── topic-keys.md         # topic_key namespacing convention
├── multi-repo.md         # .engram/config.json strategy + signal detection
├── lifecycle.md          # reads, session summary, conflicts, updates
└── workflows/
    ├── sdd.md            # SDD (Gentle.AI agent-teams-lite) phase mapping
    ├── superpowers.md    # Superpowers brainstorming/writing-plans mapping
    └── ad-hoc.md         # Manual saves without workflow
```

**Rationale**: `SKILL.md` is always loaded (per Agent Skills spec — frontmatter
`description` with aggressive trigger keywords causes auto-loading whenever the
agent is about to call any engram tool). The subfiles are read on-demand: when
`SKILL.md`'s decision table matches the current situation, the agent reads the
relevant subfile only. This is progressive disclosure — an agent handling a
simple bugfix does not load the entire SDD workflow mapping. Context cost stays
proportional to what the agent actually needs.

All subfiles use short, scannable sections (headers + tables + examples) so the
agent can parse the relevant rule in one pass without reading the whole file.

---

## 5. SKILL.md Entry Point

### Frontmatter

```yaml
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
```

The `description` doubles as the trigger signal. Tools that use semantic
matching (like opencode) will load this skill whenever the agent's intent
includes any of those keywords.

### Body — 5-rule TL;DR

```markdown
## Quick Rules

1. **type** — always pick from the canonical list. See `types.md`.
2. **topic_key** — always use the namespace convention. See `topic-keys.md`.
3. **multi-repo** — if this looks like a multi-repo product, check/create
   `.engram/config.json` before saving. See `multi-repo.md`.
4. **reads** — use the same namespace convention in queries; bare keyword
   search is noisy in busy projects.
5. **session end** — always call `mem_session_summary` before saying "done".
   See `lifecycle.md`.
```

### Decision Table

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

### Compatibility Note

Designed against the Agent Skills spec. Works with:
- **Claude Code** — loads from `~/.claude/skills/` (global) or
  `.claude/skills/` (project-local).
- **opencode** — loads from its configured skills directory.

Installation: keep the canonical copy at `skills/engram-conventions/` and
symlink or copy into each tool's skills location.

---

## 6. Core Conventions — Types

### Taxonomy Table

| Type | When to use | Example title |
|------|-------------|---------------|
| `exploration` | Pre-commitment investigation or comparison of approaches | "Compared SSE vs WebSocket for live updates" |
| `proposal` | Intent + scope before specs are written | "Proposal: auth refactor with JWT rotation" |
| `spec` | Requirements + scenarios for a change | "Spec: auth refactor requirements" |
| `design` | Technical approach + per-feature architecture decisions | "Design: token rotation strategy" |
| `plan` | Implementation roadmap (steps, order, dependencies) | "Plan: auth refactor implementation" |
| `tasks` | Checklist breakdown of work items | "Tasks: auth refactor (12 items)" |
| `report` | Verify or archive output; status snapshots | "Verify report: auth refactor — 2 CRITICAL" |
| `decision` | Non-architectural tactical choice (tool, workflow, format) | "Decision: httpOnly cookies over localStorage" |
| `architecture` | System-wide architectural decision with explicit tradeoffs | "Architecture: hex layout for auth domain" |
| `bugfix` | Bug identified + root cause + fix applied | "Fixed N+1 in UserList query" |
| `pattern` | Convention established (naming, structure, approach) | "Pattern: prefix integration tests with `_e2e`" |
| `discovery` | Non-obvious codebase finding or surprising behavior | "Discovery: FTS5 strips digits, breaks search" |
| `config` | Configuration or environment change | "Config: switched test DB to SQLite in-memory" |
| `preference` | User constraint or preference learned | "Preference: always use httpOnly cookies" |

**Fallback rule**: if no type fits cleanly, use `discovery` and make the
mismatch explicit in the title (e.g., "Discovery: [process note] team deploys
Fridays only"). The skill deliberately keeps the list as guidance, not an
enforced enum, so engram remains flexible. Discipline lives here, not in the
API.

**`decision` vs `architecture`**: use `architecture` when the choice affects
the whole system or multiple domains and involves explicit tradeoffs
(e.g., "hexagonal layout for auth domain"). Use `decision` for tactical,
local choices (e.g., "cookie storage format for session token").

### `types.md` Planned Content

Each type gets a mini-section with:
- **Use when** / **Don't use when**
- **Title shape** (verb + what, short, searchable)
- **Content shape** (What / Why / Where / Learned)
- **Example `mem_save` call** (complete, copy-pasteable)

Plus a decision-aid tree at the end: a flowchart in text form (is this a
finding? → discovery/bugfix. Is this a choice? → decision/architecture. Is
this workflow output? → see workflows/) that an agent can follow when unsure.

---

## 7. Core Conventions — topic_key

### Convention

```
<namespace>/<feature>[/<phase>][/<sub>]
```

### Rules

- **Lowercase kebab-case only** — no uppercase, no underscores, no spaces.
- **`/` as level separator** — each `/` adds one level of specificity.
- **Max 4 levels deep** — deeper nesting produces keys that are hard to scan
  and query; restructure the namespace instead.
- **Stable identifier** — do not rename a `topic_key` after first use.
  Engram's upsert relies on the exact key; renaming breaks revision history and
  leaves orphaned observations.
- **Same `topic_key` + same `project` = upsert** — engram increments
  `revision_count`. The previous revision is preserved in history.
- **Different `topic_key` = new observation** — use a new phase suffix when
  moving from one artifact type to the next (e.g., `/spec` → `/design`).

### Namespace Registry

| Namespace | Purpose | Documented in |
|-----------|---------|---------------|
| `sdd/<change>/<phase>` | SDD workflow artifacts | `workflows/sdd.md` |
| `superpowers/<feature>/<phase>` | Superpowers skill outputs | `workflows/superpowers.md` |
| `architecture/<area>` | Standalone architectural decisions | `workflows/ad-hoc.md` |
| `decision/<short-id>` | Standalone tactical decisions | `workflows/ad-hoc.md` |
| `bugfix/<short-id>` | Bug fixes | `workflows/ad-hoc.md` |
| `pattern/<name>` | Project conventions | `workflows/ad-hoc.md` |
| `discovery/<short-id>` | Non-obvious findings | `workflows/ad-hoc.md` |
| `config/<area>` | Configuration changes | `workflows/ad-hoc.md` |
| `preference/<area>` | User preferences | `workflows/ad-hoc.md` |

### Good vs Bad Examples

**Good:**
```
sdd/auth-refactor/spec
superpowers/payment-flow/plan
architecture/auth-model
bugfix/n-plus-one-userlist
pattern/e2e-test-suffix
decision/cookie-vs-localstorage
```

**Bad — and why:**
```
sdd-auth-spec            # no namespace separation — looks like a flat tag
Sdd/Auth/Spec            # uppercase — breaks exact-match queries
sdd/very/deeply/nested/thing/here  # too deep — max 4 levels
auth                     # too generic, no namespace — collides with everything
the-thing-i-just-fixed   # not stable, not searchable, will be renamed
```

### Revisions vs New Observations

| Case | Action |
|------|--------|
| Same artifact evolving (more detail, correction) | Same `topic_key` → upsert, `revision_count++` |
| Moving from one phase to next (spec → design) | New `topic_key` with new phase suffix |
| Completely unrelated artifact | New `topic_key` with appropriate namespace |

### Timeline Queries

`mem_search` with an exact `topic_key` returns all revisions of that
observation, oldest to newest. This is how engram-ui renders the evolution of
an artifact over time. Consistent `topic_key` naming is what makes timeline
views possible.

---

## 8. Multi-Repo Handling

### Default Behavior

Engram resolves the current project automatically from the working directory.
Resolution priority (as implemented in `engram/internal/project/detect.go`):

1. `.engram/config.json` → `project_name` field **(highest priority)**
2. Git remote URL (extracts repo name)
3. Git root basename
4. Directory basename (fallback)

For single-repo work, the default is fine. For multi-repo products, all repos
must write to the same engram project.

### Multi-Repo Setup

Create `.engram/config.json` in each participating repo:

```json
{"project_name": "myapp"}
```

All repos with this config write to engram project `myapp`, enabling unified
queries across the product.

### Signal Detection

An agent should suspect multi-repo when **two or more** of these signals are
present:

- Sibling repos in the parent directory share a prefix (e.g., `myapp-frontend`,
  `myapp-backend`, `myapp-infra`).
- Git remotes point to the same org with related names.
- `README`, `package.json`, or `go.mod` references related repos explicitly.
- Cross-references in source code or docs (imports, links, `depends_on`).

When two or more signals match and no `.engram/config.json` exists, the agent
**must ask the user** before creating any config. Do not auto-create.

### Sample Prompt to User

> I see `myapp-frontend`, `myapp-backend`, and `myapp-infra` — these look like
> one product. How should memories be saved?
>
> a) **Unified** (`myapp`) — I'll create `.engram/config.json` in each repo so
>    all memories land in the same project.
> b) **Repo-local** (default) — each repo keeps its own engram project.
> c) **Don't ask again** — keep current behavior and record this preference.

Record the user's choice immediately as a `preference` observation with
`topic_key=preference/multi-repo-setup`.

### Topic_key Dimensions in Multi-Repo

When unified under one `project_name`, the leading topic_key slot becomes the
repo identifier when scope matters:

```
frontend/auth/spec          # auth spec scoped to frontend repo
backend/auth/spec           # auth spec scoped to backend repo
architecture/auth-model     # cross-cutting arch decision — no repo prefix, uses ad-hoc namespace
```

Workflow stacks below repo:
```
frontend/sdd/auth-refactor/spec
```

### Cross-Repo Reads

```
mem_search(query="auth", project="myapp")
```

Returns observations from all repos in the product. To scope to one repo, add
a `topic_key_prefix`:

```
mem_search(query="auth", project="myapp", topic_key_prefix="frontend/")
```

### Pitfalls

- **Renaming `project_name` mid-stream** orphans all existing observations.
  The new name starts a fresh project; old memories are unreachable under the
  new name.
- **Mixed config** — some repos with `.engram/config.json`, others without —
  creates ambiguous queries. Either all repos in a product use the config or
  none of them do.
- **Generic `project_name`** (e.g., `team`, `work`, `misc`) defeats filtering.
  Use a specific product name.

---

## 9. Lifecycle Operations

### Read Patterns

| Tool | When to use | Notes |
|------|-------------|-------|
| `mem_context` | Start of any session | Cheap, always the first call when picking up work |
| `mem_search` | Find by keywords, type, or namespace | Filter by `type`, `project`, `scope`, `topic_key_prefix`. Results are **truncated** — follow up with `mem_get_observation`. |
| `mem_get_observation` | Full content of one observation | Required after a search hit — search results are previews only |
| `mem_search` with exact `topic_key` | Timeline of an artifact | Returns all revisions oldest → newest |
| `mem_current_project` | Verify cwd → project mapping | Use when multi-repo setup is uncertain |

**Query shape**: combine `query` (keywords) + `type` + `topic_key_prefix`.
Bare keyword search in a busy project is noisy. At minimum, add a `type` or
`topic_key_prefix` to narrow results.

### Session Lifecycle

1. **Start**: call `mem_context(project=<auto>)` — reads recent session
   history without searching.
2. **During**: call `mem_save` proactively after decisions, bugfixes,
   discoveries, and conventions. Do not wait to be asked.
3. **End** (MANDATORY — before saying "done", "listo", or equivalent):
   call `mem_session_summary`.

Skipping the session summary means the next session starts blind. This is not
optional.

### `mem_session_summary` Content Shape

```markdown
## Goal
[What this session was about — one sentence]

## Discoveries
- [Non-obvious technical findings]

## Accomplished
- [Completed items with key details]

## Next Steps
- [What remains for the next session]

## Relevant Files
- path/to/file — [what it does or what changed]
```

### Conflict Resolution (`mem_judge`)

After every `mem_save`, check the response envelope:

- If `judgment_required: true` → iterate `candidates[]`
- For each candidate, call `mem_judge(judgment_id=<candidate.judgment_id>, ...)`
- Use **each candidate's own `judgment_id`**, not the top-level one

**Heuristic — ask the user when:**
- `confidence < 0.7`, OR
- `relation` is `supersedes` or `conflicts_with` AND `type` is `architecture`,
  `policy`, or `decision`

**Resolve silently when:**
- `confidence ≥ 0.7` AND `relation` is `related`, `compatible`, `scoped`, or
  `not_conflict`

### Updates Decision Table

| Situation | Action |
|-----------|--------|
| Same artifact evolved (more detail, correction) | Same `topic_key`, new `mem_save` → upsert, `revision_count++` |
| Typo or factual error in an existing observation | `mem_update(id, …)` |
| New artifact shape for same feature (spec → design) | New `topic_key` with new phase suffix |
| Observation is a duplicate or wrong | `mem_delete(id)` |

### Compaction Recovery (MANDATORY)

If you see a compaction message or "FIRST ACTION REQUIRED":

1. **Immediately** call `mem_session_summary` with the compacted content —
   this persists pre-compaction work before it is lost.
2. Call `mem_context` to recover context from previous sessions.
3. Only then continue working.

Skipping step 1 means everything done before compaction is permanently lost
from persistent memory.

---

## 10. Workflow Cookbook — SDD (`workflows/sdd.md` outline)

SDD (Spec-Driven Development) is Gentle.AI's agent-teams-lite skill set. Each
phase produces a typed observation with a predictable `topic_key`.

### Phase → Type → topic_key Mapping

| SDD phase | type | topic_key |
|-----------|------|-----------|
| `sdd-explore` | `exploration` | `sdd/<change>/explore` |
| `sdd-propose` | `proposal` | `sdd/<change>/proposal` |
| `sdd-spec` | `spec` | `sdd/<change>/spec` |
| `sdd-design` | `design` | `sdd/<change>/design` |
| `sdd-tasks` | `tasks` | `sdd/<change>/tasks` |
| `sdd-apply` progress | `report` | `sdd/<change>/apply-progress` |
| `sdd-verify` | `report` | `sdd/<change>/verify-report` |
| `sdd-archive` | `report` | `sdd/<change>/archive-report` |
| `sdd-init` project context | `config` | `sdd-init/<project>` |

`<change>` = kebab-case change identifier (e.g., `auth-refactor`). Use the
same identifier throughout all phases of the same change.

### Content Shape Suggestions per Phase

| Phase | Suggested headings |
|-------|--------------------|
| `exploration` | `## Investigated` / `## Approaches` / `## Recommendation` |
| `proposal` | `## Intent` / `## Scope` / `## Approach` |
| `spec` | `## Requirements` / `## Scenarios` |
| `design` | `## Decisions` / `## Approach` / `## Tradeoffs` |
| `tasks` | Markdown checklist (`- [ ] item`) |
| `report` (verify) | `## CRITICAL` / `## WARNING` / `## SUGGESTION` |
| `report` (archive) | `## Closed` / `## Open` / `## Outcome` |

### Useful Read Patterns

```
# Resume a change — get all phases at once
mem_search(topic_key_prefix="sdd/auth-refactor/")

# Timeline of one phase (all revisions)
mem_search(topic_key="sdd/auth-refactor/spec")

# All proposals in this project
mem_search(type="proposal", project="myapp")
```

### Multi-Repo SDD

For SDD work scoped to one repo:
```
frontend/sdd/auth-refactor/spec
backend/sdd/auth-refactor/spec
```

For cross-cutting SDD work:
```
sdd/auth-refactor/architecture   # no repo prefix
```

---

## 11. Workflow Cookbook — Superpowers (`workflows/superpowers.md` outline)

Superpowers is a set of Gentle.AI skills for brainstorming and planning. Each
skill produces typed observations.

### Skill → Type → topic_key Mapping

| Superpowers skill | type | topic_key |
|-------------------|------|-----------|
| `brainstorming` design doc | `design` | `superpowers/<feature>/design` |
| `writing-plans` implementation plan | `plan` | `superpowers/<feature>/plan` |

`<feature>` = stable kebab-case identifier for the feature. Use a meaningful
name, not a date (dates go in the filename when writing to disk).

### Dual Storage Note

Superpowers skills also write design docs to disk
(`docs/superpowers/specs/YYYY-MM-DD-<topic>-design.md`). The engram save
**complements** the disk file — it does not replace it. The disk file serves
git history and team collaboration; the engram observation serves cross-session
searchability and timeline views in engram-ui.

### Content Shape Suggestions

| Skill | Suggested headings |
|-------|--------------------|
| `brainstorming` (design) | `## Architecture` / `## Components` / `## Data flow` / `## Error handling` / `## Testing` |
| `writing-plans` (plan) | `## Steps` / `## Order` / `## Validation` |

### Linking Design and Plan

Both artifacts share the `<feature>` slot in their `topic_key`. To retrieve
both together:

```
mem_search(topic_key_prefix="superpowers/payment-flow/")
```

Returns both `superpowers/payment-flow/design` and
`superpowers/payment-flow/plan`.

### Multi-Repo Superpowers

For work scoped to one repo:
```
frontend/superpowers/payment-flow/design
```

For cross-cutting output:
```
superpowers/payment-flow/design   # no repo prefix
```

---

## 12. Workflow Cookbook — Ad-hoc (`workflows/ad-hoc.md` outline)

Ad-hoc saves are not tied to any workflow. They cover the majority of everyday
saves: decisions made during implementation, bugs fixed, patterns established,
discoveries about the codebase, config changes, and user preferences.

### Type → topic_key Pattern

| Type | topic_key pattern | Example |
|------|-------------------|---------|
| `decision` | `decision/<short-id>` | `decision/cookie-vs-localstorage` |
| `architecture` | `architecture/<area>` | `architecture/auth-model` |
| `bugfix` | `bugfix/<short-id>` | `bugfix/n-plus-one-userlist` |
| `pattern` | `pattern/<name>` | `pattern/e2e-test-suffix` |
| `discovery` | `discovery/<short-id>` | `discovery/fts5-strips-digits` |
| `config` | `config/<area>` | `config/test-db-sqlite` |
| `preference` | `preference/<area>` | `preference/commit-style` |

### short-id Rules

- Kebab-case, 3–5 words max.
- No dates — engram tracks `created_at` automatically.
- Stable — do not rename after first use.

**Good**: `n-plus-one-userlist`, `cookie-vs-localstorage`, `fts5-strips-digits`
**Bad**: `bug-2024-05-14`, `the-thing`, `fix` (too vague or date-stamped)

### Ad-hoc vs Workflow Output

If a save is the output of an SDD or Superpowers phase, use the workflow
namespace (`sdd/`, `superpowers/`). If it is a standalone save made during
implementation without a governing workflow, use the ad-hoc namespaces above.
Do not mix namespaces for the same artifact.

### Multi-Repo Ad-hoc

For saves scoped to one repo:
```
frontend/bugfix/n-plus-one
```

For cross-cutting saves (e.g., an architecture decision affecting all repos):
```
architecture/auth-model   # no repo prefix
```

---

## 13. Compatibility

This skill is designed against the Agent Skills spec. It is self-contained and
portable — it requires no specific tool, no proprietary extension, and no
changes to any existing system.

**Claude Code** loads skills from:
- `~/.claude/skills/` (global — applies to all projects)
- `.claude/skills/` (project-local — applies to this project only)

**opencode** loads skills from its configured skills directory (per the Agent
Skills spec).

**Installation pattern**: keep the canonical copy at
`skills/engram-conventions/` and symlink (or copy) the directory into each
tool's skills location. A single source of truth avoids drift between tool
copies.

**Version compatibility**: the skill relies only on standard MCP tool names
(`mem_save`, `mem_search`, etc.) and the Agent Skills spec `SKILL.md` loading
convention. No version-specific feature flags required.

---

## 14. Open Questions / Future Work

Items deliberately excluded from V1, noted for later iteration:

- **Automated validation hook** — a post-save check that verifies `type` is
  in the canonical list and `topic_key` matches the namespace pattern. Would
  run as a tool wrapper or MCP middleware, not in the skill itself.
- **`mem_save_prompt` integration** — when to capture the user prompt alongside
  a save (useful for surfacing the originating intent in engram-ui). The
  current skill does not prescribe this.
- **Custom workflow registration** — a mechanism for a new plugin or skill to
  add its own namespace and content shape to the workflow cookbook without
  editing this skill's source.
- **engram-ui-side UI changes** — filter panels, grouping by `type`, timeline
  views, and namespace-aware search. These are the primary consumers of these
  conventions but are tracked as a separate scope in engram-ui.
