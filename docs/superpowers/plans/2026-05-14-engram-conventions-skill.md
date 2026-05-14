# Engram Conventions Skill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the `engram-conventions` skill (SKILL.md + 5 reference subfiles + 3 workflow cookbook files) at `skills/engram-conventions/` per the approved design spec.

**Architecture:** 9 markdown files organized as one entry point (SKILL.md) plus reference subfiles loaded on-demand by agents. Follows the Agent Skills spec for portability across Claude Code, opencode, and other compatible tools. Zero changes to engram source — purely agent-side guidance.

**Tech Stack:** Markdown, YAML frontmatter (Agent Skills spec). No code, no tests, no build step.

---

## File Map

| File | Responsibility |
|------|---------------|
| `skills/engram-conventions/SKILL.md` | Entry point — frontmatter triggers + 5-rule TL;DR + decision table + compatibility note |
| `skills/engram-conventions/types.md` | 14 canonical observation types with use-when rules, title/content shapes, example calls, decision-aid tree |
| `skills/engram-conventions/topic-keys.md` | Namespace convention, rules, registry table, good/bad examples, revisions vs new obs |
| `skills/engram-conventions/multi-repo.md` | Default behavior, `.engram/config.json` setup, signal detection, ask-user prompt, topic_key dimensions, cross-repo reads, pitfalls |
| `skills/engram-conventions/lifecycle.md` | Read patterns, session lifecycle, `mem_session_summary` shape, conflict resolution, updates decision table, compaction recovery |
| `skills/engram-conventions/workflows/sdd.md` | SDD phase → type → topic_key mapping, content shapes, read patterns, multi-repo SDD |
| `skills/engram-conventions/workflows/superpowers.md` | Superpowers skill output mapping, dual storage note, content shapes, linking, multi-repo |
| `skills/engram-conventions/workflows/ad-hoc.md` | Ad-hoc type → topic_key patterns, short-id rules, ad-hoc vs workflow, multi-repo ad-hoc |

---

## Task 1: Create directory structure and write SKILL.md

**Files:**
- Create: `skills/engram-conventions/SKILL.md`
- Create: `skills/engram-conventions/workflows/` (directory — created implicitly by later tasks)

- [ ] **Step 1: Verify the file does not exist yet**

  Run:
  ```
  Glob pattern: skills/engram-conventions/SKILL.md
  ```
  Expected: no files found.

- [ ] **Step 2: Create directory structure**

  ```bash
  mkdir -p skills/engram-conventions/workflows
  ```
  Expected: directories created, no error.

- [ ] **Step 3: Write `skills/engram-conventions/SKILL.md`**

  Write this exact content:

  ````markdown
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
  ````

- [ ] **Step 4: Read back and confirm key sections present**

  Read `skills/engram-conventions/SKILL.md`. Confirm all of the following are present:
  - YAML frontmatter with `name: engram-conventions` and `triggers:` list
  - `## Quick Rules` section with 5 numbered rules
  - `## Decision Table` with all 9 rows (types.md, topic-keys.md, multi-repo.md, lifecycle.md ×3, workflows/sdd.md, workflows/superpowers.md, workflows/ad-hoc.md)
  - `## Compatibility` section mentioning Claude Code and opencode

- [ ] **Step 5: Commit**

  ```bash
  git add skills/engram-conventions/SKILL.md
  git commit -m "feat(engram-conventions): add SKILL.md entry point with frontmatter and decision table"
  ```

---

## Task 2: Write types.md

**Files:**
- Create: `skills/engram-conventions/types.md`

- [ ] **Step 1: Verify the file does not exist yet**

  Run:
  ```
  Glob pattern: skills/engram-conventions/types.md
  ```
  Expected: no files found.

- [ ] **Step 2: Write `skills/engram-conventions/types.md`**

  Write this exact content:

  ````markdown
  # Observation Types

  14 canonical types for engram observations. Always pick from this list.

  **Fallback rule**: if no type fits cleanly, use `discovery` and make the
  mismatch explicit in the title (e.g., `"Discovery: [process note] team deploys
  Fridays only"`). The list is guidance, not an enforced enum — engram remains
  flexible.

  ---

  ## Taxonomy Overview

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

  ---

  ## Per-Type Reference

  ### `exploration`

  **Use when**: investigating multiple approaches before committing to one.
  Typically the output of a brainstorming or sdd-explore phase.

  **Don't use when**: you've already chosen an approach and are documenting the
  decision — use `decision` or `architecture` instead.

  **Title shape**: `"Compared X vs Y for Z"` or `"Explored options for Z"`

  **Content shape**:
  ```
  ## Investigated
  [What was explored]

  ## Approaches
  [Each approach with pros/cons]

  ## Recommendation
  [Preferred approach and why]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Compared SSE vs WebSocket for live observation updates",
    type="exploration",
    topic_key="sdd/live-updates/explore",
    project="engram-ui",
    content="## Investigated\nReal-time delivery of new observations to engram-ui.\n\n## Approaches\n- SSE: simpler, HTTP/1.1 compatible, server-push only\n- WebSocket: bidirectional, more complex, overkill here\n\n## Recommendation\nSSE — sufficient for read-only push, less infrastructure."
  )
  ```

  ---

  ### `proposal`

  **Use when**: defining intent and scope at the start of a change, before specs
  are written. The output of `sdd-propose`.

  **Don't use when**: the proposal is already being implemented — use `spec`,
  `design`, or `tasks` for subsequent phases.

  **Title shape**: `"Proposal: <change name> with <key approach>"`

  **Content shape**:
  ```
  ## Intent
  [What problem this solves and why now]

  ## Scope
  [What is in scope / out of scope]

  ## Approach
  [High-level technical direction]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Proposal: auth refactor with JWT rotation",
    type="proposal",
    topic_key="sdd/auth-refactor/proposal",
    project="myapp",
    content="## Intent\nCurrent session tokens never expire. Need short-lived JWTs with refresh rotation.\n\n## Scope\nIn: token issuance, refresh endpoint, middleware. Out: OAuth providers (V2).\n\n## Approach\nReplace session table with signed JWTs. Refresh tokens stored httpOnly cookies."
  )
  ```

  ---

  ### `spec`

  **Use when**: documenting requirements and acceptance scenarios for a change.
  The output of `sdd-spec`.

  **Don't use when**: documenting technical implementation choices — use `design`
  for that.

  **Title shape**: `"Spec: <change name> requirements"`

  **Content shape**:
  ```
  ## Requirements
  - [Functional requirement 1]
  - [Functional requirement 2]

  ## Scenarios
  - Given X, when Y, then Z
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Spec: auth refactor requirements",
    type="spec",
    topic_key="sdd/auth-refactor/spec",
    project="myapp",
    content="## Requirements\n- Tokens expire after 15 minutes\n- Refresh tokens valid 7 days\n- Refresh rotates on use\n\n## Scenarios\n- Given expired token, when client sends request, then 401 returned\n- Given valid refresh token, when client calls /refresh, then new token pair issued"
  )
  ```

  ---

  ### `design`

  **Use when**: documenting technical approach and per-feature architecture
  decisions. The output of `sdd-design` or `superpowers:brainstorming`.

  **Don't use when**: the decision is system-wide and involves explicit tradeoffs
  across multiple domains — use `architecture` for that.

  **Title shape**: `"Design: <feature> <aspect>"`

  **Content shape**:
  ```
  ## Decisions
  [Key technical choices made]

  ## Approach
  [How the feature will be built]

  ## Tradeoffs
  [What was traded off and why]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Design: token rotation strategy",
    type="design",
    topic_key="sdd/auth-refactor/design",
    project="myapp",
    content="## Decisions\n- Stateless JWTs signed with HMAC-SHA256\n- Refresh tokens stored in DB with revocation list\n\n## Approach\nAccess token: 15-min signed JWT. Refresh token: opaque 32-byte random, stored hashed.\n\n## Tradeoffs\nStateless access tokens cannot be revoked early — acceptable for 15-min TTL."
  )
  ```

  ---

  ### `plan`

  **Use when**: recording an implementation roadmap — steps, order, dependencies.
  The output of `superpowers:writing-plans`.

  **Don't use when**: recording the task checklist itself — use `tasks` for that.

  **Title shape**: `"Plan: <change name> implementation"`

  **Content shape**:
  ```
  ## Steps
  [Ordered implementation steps]

  ## Order
  [Dependency rationale if non-obvious]

  ## Validation
  [How to verify each step succeeded]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Plan: auth refactor implementation",
    type="plan",
    topic_key="superpowers/auth-refactor/plan",
    project="myapp",
    content="## Steps\n1. Add JWT library\n2. Implement token issuance endpoint\n3. Implement refresh endpoint\n4. Update middleware\n5. Write integration tests\n\n## Order\nMiddleware last — depends on issuance being complete.\n\n## Validation\nIntegration test suite passes. Existing session tests updated."
  )
  ```

  ---

  ### `tasks`

  **Use when**: breaking a change into a numbered checklist of work items.
  The output of `sdd-tasks`.

  **Don't use when**: recording progress or completion status — use `report`
  for that.

  **Title shape**: `"Tasks: <change name> (<N> items)"`

  **Content shape**: Markdown checklist (`- [ ] item`), one item per line.

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Tasks: auth refactor (8 items)",
    type="tasks",
    topic_key="sdd/auth-refactor/tasks",
    project="myapp",
    content="- [ ] Add `golang-jwt/jwt` dependency\n- [ ] Create `internal/auth/token.go` with Issue() and Verify()\n- [ ] Add POST /api/refresh endpoint\n- [ ] Update auth middleware to parse JWT\n- [ ] Add refresh token DB table and migration\n- [ ] Write unit tests for token.go\n- [ ] Write integration test for full auth flow\n- [ ] Update README auth section"
  )
  ```

  ---

  ### `report`

  **Use when**: recording the output of a verify or archive phase, or any
  status snapshot (apply progress, partial completion, CI result).

  **Don't use when**: recording task breakdowns — use `tasks` for that.

  **Title shape**: `"Verify report: <change> — <N> CRITICAL"` or
  `"Apply progress: <change> — <N>/<total> done"` or
  `"Archive: <change> — closed"`

  **Content shape (verify)**:
  ```
  ## CRITICAL
  - [Blocking issues that must be fixed]

  ## WARNING
  - [Non-blocking issues worth addressing]

  ## SUGGESTION
  - [Optional improvements]
  ```

  **Content shape (archive)**:
  ```
  ## Closed
  [What was completed]

  ## Open
  [What was deferred or left out of scope]

  ## Outcome
  [Net result — what the system can do now that it couldn't before]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Verify report: auth refactor — 0 CRITICAL",
    type="report",
    topic_key="sdd/auth-refactor/verify-report",
    project="myapp",
    content="## CRITICAL\n(none)\n\n## WARNING\n- Refresh token cleanup job not implemented (deferred to V2)\n\n## SUGGESTION\n- Consider adding token blacklist for admin-forced logout"
  )
  ```

  ---

  ### `decision`

  **Use when**: recording a tactical, local choice — storage format, library
  selection, naming convention, workflow preference. Not system-wide.

  **Don't use when**: the choice affects the whole system architecture or
  multiple domains — use `architecture` for that.

  **Title shape**: `"Decision: <option A> over <option B>"` or
  `"Decision: use <X> for <purpose>"`

  **Content shape**:
  ```
  ## What
  [The decision made]

  ## Why
  [Motivation — what drove this choice]

  ## Where
  [Files or paths affected]

  ## Learned
  [Gotchas or edge cases, if any]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Decision: httpOnly cookies over localStorage for refresh tokens",
    type="decision",
    topic_key="decision/cookie-vs-localstorage",
    project="myapp",
    content="## What\nRefresh tokens stored in httpOnly cookies, not localStorage.\n\n## Why\nXSS cannot read httpOnly cookies. localStorage is accessible to JS.\n\n## Where\nPOST /api/refresh sets Set-Cookie header. Frontend never reads the token.\n\n## Learned\nSameSite=Strict breaks cross-origin flows — use SameSite=Lax for redirect-based OAuth."
  )
  ```

  ---

  ### `architecture`

  **Use when**: recording a system-wide architectural choice with explicit
  tradeoffs — hexagonal layout, domain boundaries, persistence strategy.
  Multiple domains or subsystems affected.

  **Don't use when**: the choice is scoped to one feature or tactical — use
  `decision` for that.

  **Title shape**: `"Architecture: <area> <approach>"`

  **Content shape**:
  ```
  ## What
  [The architectural decision]

  ## Why
  [What drove this — forces, constraints, alternatives rejected]

  ## Where
  [Domains, packages, or system boundaries affected]

  ## Tradeoffs
  [What is gained and what is given up]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Architecture: hexagonal layout for auth domain",
    type="architecture",
    topic_key="architecture/auth-model",
    project="myapp",
    content="## What\nAuth domain follows hexagonal architecture with explicit ports and adapters.\n\n## Why\nAllows swapping HTTP handler or DB adapter without touching business logic. Testable in isolation.\n\n## Where\ninternal/auth/ — core/, ports/, adapters/http/, adapters/db/\n\n## Tradeoffs\nMore files and interfaces up front. Worth it for testability and replaceability."
  )
  ```

  ---

  ### `bugfix`

  **Use when**: a bug has been identified, root-caused, and fixed. Captures the
  finding so future sessions don't re-investigate the same issue.

  **Don't use when**: recording a to-do for a bug not yet fixed — use `tasks`
  or `discovery` instead.

  **Title shape**: `"Fixed <what was wrong>"` (verb past tense, no "bug" prefix)

  **Content shape**:
  ```
  ## What
  [One-sentence description of what was fixed]

  ## Why
  [Root cause — what actually caused the bug]

  ## Where
  [Files or paths modified]

  ## Learned
  [Edge case or gotcha that is now known]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Fixed N+1 query in UserList",
    type="bugfix",
    topic_key="bugfix/n-plus-one-userlist",
    project="myapp",
    content="## What\nUserList was issuing one SQL query per row to fetch the associated org.\n\n## Why\nGORM's lazy-loading triggered on `user.Org` access inside the loop.\n\n## Where\ninternal/user/list.go — added Preload(\"Org\") to the base query.\n\n## Learned\nGORM does not preload by default. Any association access in a loop is an N+1 unless Preload() is explicit."
  )
  ```

  ---

  ### `pattern`

  **Use when**: a recurring convention has been established — file naming,
  test structure, code organization, team agreement on style.

  **Don't use when**: the observation is a one-off decision — use `decision`
  for that.

  **Title shape**: `"Pattern: <convention name>"` or
  `"Pattern: <what the convention is>"`

  **Content shape**:
  ```
  ## What
  [The convention]

  ## Why
  [Why this pattern was established]

  ## Where
  [Where it applies]

  ## Learned
  [How to spot violations or edge cases, if any]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Pattern: prefix integration tests with _e2e",
    type="pattern",
    topic_key="pattern/e2e-test-suffix",
    project="myapp",
    content="## What\nAll integration tests that hit the real DB or network are prefixed with `_e2e` in the filename.\n\n## Why\nAllows `go test ./... -run '^[^_]'` to skip integration tests in CI fast lane.\n\n## Where\nAll test files in tests/integration/\n\n## Learned\nGo build tags are an alternative but require flag discipline. Prefix is simpler and visible."
  )
  ```

  ---

  ### `discovery`

  **Use when**: finding non-obvious behavior, surprising interactions, or hidden
  constraints in the codebase, libraries, or infrastructure.

  **Don't use when**: the finding is a known bug you've already fixed — use
  `bugfix` for that.

  **Title shape**: `"Discovery: <what was found>"` (noun phrase, short)

  **Content shape**:
  ```
  ## What
  [What was discovered]

  ## Why it matters
  [Impact on the codebase or workflow]

  ## Where
  [Where it shows up]

  ## Learned
  [What to do differently, or what to watch for]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Discovery: FTS5 strips digits, breaks search for version strings",
    type="discovery",
    topic_key="discovery/fts5-strips-digits",
    project="engram-ui",
    content="## What\nSQLite FTS5 tokenizer strips standalone digit sequences by default.\n\n## Why it matters\nSearching for 'v1' or '404' returns no results even when present in content.\n\n## Where\nengram/internal/store/fts.go — tokenizer config\n\n## Learned\nUse `unicode61 tokenchars` or `ascii` tokenizer to preserve digits. Requires FTS index rebuild."
  )
  ```

  ---

  ### `config`

  **Use when**: recording a configuration or environment change — DB connection,
  test setup, feature flags, tooling settings.

  **Don't use when**: the change is a team-level architectural decision —
  use `architecture` for that.

  **Title shape**: `"Config: <what changed>"` (noun phrase, imperative or
  descriptive)

  **Content shape**:
  ```
  ## What
  [The configuration change]

  ## Why
  [Why this change was made]

  ## Where
  [Files, env vars, or config keys changed]

  ## Learned
  [Side effects or gotchas, if any]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Config: switched test DB to SQLite in-memory",
    type="config",
    topic_key="config/test-db-sqlite",
    project="myapp",
    content="## What\nTest suite now uses `sqlite3 :memory:` instead of a local Postgres instance.\n\n## Why\nCI setup was flaky due to Postgres port conflicts. In-memory SQLite is zero-setup.\n\n## Where\nconfig/test.yaml — DB_URL changed. Makefile target `test` no longer needs `docker compose up`.\n\n## Learned\nSQLite lacks some Postgres features (e.g., RETURNING on UPDATE). Two tests were rewritten."
  )
  ```

  ---

  ### `preference`

  **Use when**: recording a user constraint or preference that the agent should
  respect in future sessions.

  **Don't use when**: the observation is a technical decision — use `decision`
  or `architecture` for that.

  **Title shape**: `"Preference: <what the user prefers>"` (present tense)

  **Content shape**:
  ```
  ## What
  [The preference]

  ## Why
  [Context or constraint behind it, if known]

  ## Where
  [Where it applies]
  ```

  **Example `mem_save` call**:
  ```
  mem_save(
    title="Preference: always use httpOnly cookies, never localStorage",
    type="preference",
    topic_key="preference/cookie-style",
    project="myapp",
    content="## What\nUser requires all tokens to be stored in httpOnly cookies.\n\n## Why\nSecurity policy — XSS protection is non-negotiable.\n\n## Where\nAny feature that stores tokens or session identifiers."
  )
  ```

  ---

  ## Decision-Aid Tree

  When unsure which type to use, follow this tree:

  ```
  Is this the OUTPUT of a workflow phase (SDD or Superpowers)?
    → Yes: use the workflow mapping in workflows/sdd.md or workflows/superpowers.md
    → No: continue

  Is this something you FOUND (not decided, not fixed)?
    → Yes, it's a bug you fixed: bugfix
    → Yes, it's surprising behavior or hidden constraint: discovery
    → No: continue

  Is this a CHOICE you made?
    → Yes, affects the whole system or multiple domains with tradeoffs: architecture
    → Yes, scoped to one feature or tactical: decision
    → Yes, a recurring convention for the project: pattern
    → No: continue

  Is this about ENVIRONMENT or SETUP?
    → Yes, configuration or tooling: config
    → Yes, something the user explicitly prefers or requires: preference
    → No: continue

  Still unsure?
    → Use `discovery` and note the mismatch in the title.
  ```
  ````

- [ ] **Step 3: Read back and confirm key sections present**

  Read `skills/engram-conventions/types.md`. Confirm all of the following are present:
  - Taxonomy overview table with all 14 types
  - A `## Per-Type Reference` section with 14 named subsections (`exploration`, `proposal`, `spec`, `design`, `plan`, `tasks`, `report`, `decision`, `architecture`, `bugfix`, `pattern`, `discovery`, `config`, `preference`)
  - Each subsection has Use when / Don't use when / Title shape / Content shape / Example mem_save call
  - `## Decision-Aid Tree` section at the end

- [ ] **Step 4: Commit**

  ```bash
  git add skills/engram-conventions/types.md
  git commit -m "feat(engram-conventions): add types.md with 14 canonical observation types"
  ```

---

## Task 3: Write topic-keys.md

**Files:**
- Create: `skills/engram-conventions/topic-keys.md`

- [ ] **Step 1: Verify the file does not exist yet**

  Run:
  ```
  Glob pattern: skills/engram-conventions/topic-keys.md
  ```
  Expected: no files found.

- [ ] **Step 2: Write `skills/engram-conventions/topic-keys.md`**

  Write this exact content:

  ````markdown
  # topic_key Namespacing Convention

  `topic_key` is the stable identifier that groups related observations and
  enables timeline views in engram-ui. Using a consistent shape makes
  observations queryable, groupable, and revisable.

  ---

  ## Convention

  ```
  <namespace>/<feature>[/<phase>][/<sub>]
  ```

  ---

  ## Rules

  - **Lowercase kebab-case only** — no uppercase, no underscores, no spaces.
  - **`/` as level separator** — each `/` adds one level of specificity.
  - **Max 4 levels deep** — deeper nesting produces keys that are hard to scan
    and query; restructure the namespace instead.
  - **Stable identifier** — do not rename a `topic_key` after first use.
    Engram's upsert relies on the exact key; renaming breaks revision history
    and leaves orphaned observations.
  - **Same `topic_key` + same `project` = upsert** — engram increments
    `revision_count`. The previous revision is preserved in history.
  - **Different `topic_key` = new observation** — use a new phase suffix when
    moving from one artifact type to the next (e.g., `/spec` → `/design`).

  ---

  ## Namespace Registry

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

  For multi-repo products, a repo-prefix slot is prepended. See `multi-repo.md`.

  ---

  ## Good vs Bad Examples

  **Good:**
  ```
  sdd/auth-refactor/spec
  superpowers/payment-flow/plan
  architecture/auth-model
  bugfix/n-plus-one-userlist
  pattern/e2e-test-suffix
  decision/cookie-vs-localstorage
  config/test-db-sqlite
  preference/commit-style
  discovery/fts5-strips-digits
  ```

  **Bad — and why:**
  ```
  sdd-auth-spec              # no namespace separation — looks like a flat tag
  Sdd/Auth/Spec              # uppercase — breaks exact-match queries
  sdd/very/deeply/nested/thing/here  # too deep — max 4 levels
  auth                       # too generic, no namespace — collides across projects
  the-thing-i-just-fixed     # not stable, not searchable, will be renamed
  bug-2024-05-14             # date in key — redundant (engram tracks created_at) and unstable
  ```

  ---

  ## Revisions vs New Observations

  | Case | Action |
  |------|--------|
  | Same artifact evolving (more detail, correction) | Same `topic_key` → upsert, `revision_count++` |
  | Moving from one phase to next (spec → design) | New `topic_key` with new phase suffix |
  | Completely unrelated artifact | New `topic_key` with appropriate namespace |

  **Example — evolving spec (same `topic_key`, upsert):**
  ```
  # First save
  mem_save(topic_key="sdd/auth-refactor/spec", ...)

  # Later: add more scenarios to the same spec
  mem_save(topic_key="sdd/auth-refactor/spec", ...)  # increments revision_count
  ```

  **Example — moving from spec to design (new `topic_key`):**
  ```
  mem_save(topic_key="sdd/auth-refactor/spec", ...)    # spec phase done
  mem_save(topic_key="sdd/auth-refactor/design", ...)  # design phase — new key
  ```

  ---

  ## Timeline Queries

  `mem_search` with an exact `topic_key` returns all revisions of that
  observation, oldest to newest. This is how engram-ui renders the evolution of
  an artifact over time.

  ```
  # All revisions of the spec for auth-refactor
  mem_search(topic_key="sdd/auth-refactor/spec")

  # All artifacts for the auth-refactor change
  mem_search(topic_key_prefix="sdd/auth-refactor/")

  # All proposals in the project
  mem_search(type="proposal", project="myapp")
  ```

  Consistent `topic_key` naming is what makes these timeline views possible.
  Inconsistent keys (mixed case, renamed keys, no namespace) break timeline
  grouping and require manual investigation to reconstruct history.
  ````

- [ ] **Step 3: Read back and confirm key sections present**

  Read `skills/engram-conventions/topic-keys.md`. Confirm all of the following are present:
  - `## Convention` with the pattern `<namespace>/<feature>[/<phase>][/<sub>]`
  - `## Rules` with 6 bullet rules (lowercase, slash separator, max 4 levels, stable, upsert, new key)
  - `## Namespace Registry` table with all 9 namespace entries, each with "Documented in" column
  - `## Good vs Bad Examples` with labeled good/bad sections
  - `## Revisions vs New Observations` table with 3 rows
  - `## Timeline Queries` section with example `mem_search` calls

- [ ] **Step 4: Commit**

  ```bash
  git add skills/engram-conventions/topic-keys.md
  git commit -m "feat(engram-conventions): add topic-keys.md with namespace convention and registry"
  ```

---

## Task 4: Write multi-repo.md

**Files:**
- Create: `skills/engram-conventions/multi-repo.md`

- [ ] **Step 1: Verify the file does not exist yet**

  Run:
  ```
  Glob pattern: skills/engram-conventions/multi-repo.md
  ```
  Expected: no files found.

- [ ] **Step 2: Write `skills/engram-conventions/multi-repo.md`**

  Write this exact content:

  ````markdown
  # Multi-Repo Handling

  For single-repo work, the engram default (auto-detect from cwd) is fine. For
  multi-repo products, all repos must write to the same engram project so
  observations can be queried together.

  ---

  ## Default Behavior

  Engram resolves the current project automatically from the working directory.
  Resolution priority (as implemented in `engram/internal/project/detect.go`):

  1. `.engram/config.json` → `project_name` field **(highest priority)**
  2. Git remote URL (extracts repo name)
  3. Git root basename
  4. Directory basename (fallback)

  For single-repo work, the default is fine. When working across related repos,
  fall through to the setup below.

  ---

  ## Multi-Repo Setup

  Create `.engram/config.json` in each participating repo:

  ```json
  {"project_name": "myapp"}
  ```

  All repos with this config write to engram project `myapp`, enabling unified
  queries across the product.

  Place this file in the repo root (alongside `.git/`). Commit it so all team
  members and CI use the same project name.

  ---

  ## Priority

  `.engram/config.json` overrides git remote and directory detection. Once
  created, engram always uses `project_name` from the config — regardless of
  the git remote or directory name.

  To verify the current project is resolved correctly:
  ```
  mem_current_project()
  ```
  Returns the resolved project name. Call this whenever multi-repo setup is
  uncertain.

  ---

  ## Signal Detection

  An agent should suspect multi-repo when **two or more** of these signals are
  present:

  - Sibling repos in the parent directory share a prefix (e.g., `myapp-frontend`,
    `myapp-backend`, `myapp-infra`).
  - Git remotes point to the same org with related names.
  - `README`, `package.json`, or `go.mod` references related repos explicitly.
  - Cross-references in source code or docs (imports, links, `depends_on`).

  When two or more signals match and no `.engram/config.json` exists, the agent
  **must ask the user** before creating any config. Do not auto-create.

  ---

  ## Ask-User Flow

  When signals detected and no config exists, prompt the user:

  > I see `myapp-frontend`, `myapp-backend`, and `myapp-infra` — these look like
  > one product. How should memories be saved?
  >
  > a) **Unified** (`myapp`) — I'll create `.engram/config.json` in each repo so
  >    all memories land in the same project.
  > b) **Repo-local** (default) — each repo keeps its own engram project.
  > c) **Don't ask again** — keep current behavior and record this preference.

  Record the user's choice immediately as a `preference` observation:

  ```
  mem_save(
    title="Preference: multi-repo memory setup for myapp",
    type="preference",
    topic_key="preference/multi-repo-setup",
    project="myapp",
    content="## What\nUser chose unified memory under project name 'myapp'.\n\n## Why\nAll three repos (frontend, backend, infra) are part of the same product.\n\n## Where\n.engram/config.json created in each repo root."
  )
  ```

  ---

  ## topic_key Dimensions in Multi-Repo

  When unified under one `project_name`, the leading `topic_key` slot becomes
  the repo identifier when scope matters:

  ```
  frontend/auth/spec          # auth spec scoped to frontend repo
  backend/auth/spec           # auth spec scoped to backend repo
  architecture/auth-model     # cross-cutting arch decision — no repo prefix
  ```

  Workflow namespaces stack below the repo prefix:
  ```
  frontend/sdd/auth-refactor/spec
  backend/sdd/auth-refactor/spec
  ```

  Cross-cutting work (decisions or discoveries that affect all repos) omits the
  repo prefix and uses the ad-hoc namespaces directly:
  ```
  architecture/auth-model
  decision/cookie-vs-localstorage
  ```

  ---

  ## Cross-Repo Reads

  Query all repos in the product:
  ```
  mem_search(query="auth", project="myapp")
  ```

  Scope to one repo with `topic_key_prefix`:
  ```
  mem_search(query="auth", project="myapp", topic_key_prefix="frontend/")
  ```

  Scope to one workflow and repo:
  ```
  mem_search(project="myapp", topic_key_prefix="frontend/sdd/auth-refactor/")
  ```

  ---

  ## Pitfalls

  - **Renaming `project_name` mid-stream** orphans all existing observations.
    The new name starts a fresh project; old memories are unreachable under the
    new name. Set the name once and never change it.

  - **Mixed config** — some repos with `.engram/config.json`, others without —
    creates ambiguous queries. Either all repos in a product use the config or
    none of them do.

  - **Generic `project_name`** (e.g., `team`, `work`, `misc`) defeats filtering.
    Use a specific, meaningful product name.

  - **Auto-creating config without asking** — do not auto-create
    `.engram/config.json`. Always confirm with the user first (see ask-user flow
    above). The config affects ALL future saves from that repo.
  ````

- [ ] **Step 3: Read back and confirm key sections present**

  Read `skills/engram-conventions/multi-repo.md`. Confirm all of the following are present:
  - `## Default Behavior` section with the 4-level priority list
  - `## Multi-Repo Setup` section with the `.engram/config.json` JSON example
  - `## Priority` section with the `mem_current_project()` call
  - `## Signal Detection` section with 4 bullet signals and the "two or more" rule
  - `## Ask-User Flow` section with the verbatim user prompt and `mem_save` call for recording the preference
  - `## topic_key Dimensions in Multi-Repo` section with examples showing repo prefix
  - `## Cross-Repo Reads` section with 3 `mem_search` examples
  - `## Pitfalls` section with 4 named pitfalls

- [ ] **Step 4: Commit**

  ```bash
  git add skills/engram-conventions/multi-repo.md
  git commit -m "feat(engram-conventions): add multi-repo.md with config setup and signal detection"
  ```

---

## Task 5: Write lifecycle.md

**Files:**
- Create: `skills/engram-conventions/lifecycle.md`

- [ ] **Step 1: Verify the file does not exist yet**

  Run:
  ```
  Glob pattern: skills/engram-conventions/lifecycle.md
  ```
  Expected: no files found.

- [ ] **Step 2: Write `skills/engram-conventions/lifecycle.md`**

  Write this exact content:

  ````markdown
  # Observation Lifecycle

  Full lifecycle operations: reads, session management, conflict resolution,
  and updates.

  ---

  ## Read Patterns

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

  **Examples:**
  ```
  # Start of session — always first
  mem_context(project="myapp")

  # Find by type and prefix (narrow, low noise)
  mem_search(type="spec", topic_key_prefix="sdd/auth-refactor/", project="myapp")

  # Timeline of one artifact
  mem_search(topic_key="sdd/auth-refactor/spec", project="myapp")

  # Get full content (required after search)
  mem_get_observation(id=<id from search result>)

  # Verify project mapping
  mem_current_project()
  ```

  ---

  ## Session Lifecycle

  ### 1. Session Start

  Call `mem_context(project=<auto>)` at the start of any session. This reads
  recent session history without searching — cheap, always the first call when
  picking up work in a project.

  ### 2. During Session

  Call `mem_save` proactively after each of these triggers — do not wait to be
  asked:

  - Architecture or design decision made
  - Team convention documented or established
  - Workflow change agreed upon
  - Tool or library choice made with tradeoffs
  - Bug fix completed (include root cause)
  - Feature implemented with non-obvious approach
  - Non-obvious discovery about the codebase
  - Gotcha, edge case, or unexpected behavior found
  - Pattern established (naming, structure, convention)
  - User preference or constraint learned
  - Configuration change or environment setup done

  ### 3. Session End (MANDATORY)

  Before saying "done", "listo", or any equivalent, call `mem_session_summary`.

  Skipping the session summary means the next session starts blind. This is not
  optional.

  ---

  ## `mem_session_summary` Content Shape

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

  **Example call:**
  ```
  mem_session_summary(
    content="## Goal\nImplemented JWT token rotation for auth refactor.\n\n## Discoveries\n- GORM lazy-loads associations; Preload() required to avoid N+1\n- SameSite=Strict breaks OAuth redirect flow\n\n## Accomplished\n- Token issuance endpoint complete\n- Refresh endpoint complete\n- Integration tests passing (8/8)\n\n## Next Steps\n- Implement token cleanup job\n- Update OpenAPI spec\n\n## Relevant Files\n- internal/auth/token.go — JWT issue and verify\n- internal/auth/refresh.go — refresh endpoint handler",
    project="myapp"
  )
  ```

  ---

  ## Conflict Resolution (`mem_judge`)

  After every `mem_save`, check the response envelope for `judgment_required`.

  ### When `judgment_required: true`

  Iterate `candidates[]`. For each candidate, call `mem_judge` once using **that
  candidate's own `judgment_id`** — not the top-level one.

  ```
  # For each candidate in candidates[]:
  mem_judge(
    judgment_id=<candidate.judgment_id>,
    relation=<your assessed relation>,
    confidence=<0.0–1.0>
  )
  ```

  ### Heuristic — ask the user when:

  - `confidence < 0.7`, OR
  - `relation` is `supersedes` or `conflicts_with` AND `type` is `architecture`,
    `policy`, or `decision`

  Ask conversationally in your next reply — never via a blocking prompt.
  Example: *"I saved the auth decision, but it looks like it might supersede an
  earlier architecture observation about session tokens. Want me to mark it as
  superseding the old one?"*

  ### Resolve silently when:

  - `confidence ≥ 0.7` AND `relation` is `related`, `compatible`, `scoped`, or
    `not_conflict`

  ---

  ## Updates Decision Table

  | Situation | Action |
  |-----------|--------|
  | Same artifact evolved (more detail, correction) | Same `topic_key`, new `mem_save` → upsert, `revision_count++` |
  | Typo or factual error in an existing observation | `mem_update(id=<id>, …)` |
  | New artifact shape for same feature (spec → design) | New `topic_key` with new phase suffix |
  | Observation is a duplicate or wrong | `mem_delete(id=<id>)` |

  **Example — correcting a typo:**
  ```
  # Find the observation to fix
  mem_search(topic_key="sdd/auth-refactor/spec", project="myapp")
  # Get the full content and ID
  mem_get_observation(id=<id>)
  # Update with the corrected content
  mem_update(id=<id>, title="Spec: auth refactor requirements", content="<corrected content>")
  ```

  **Example — upsert (same topic evolving):**
  ```
  # Add a new requirement to an existing spec — same topic_key
  mem_save(
    topic_key="sdd/auth-refactor/spec",
    type="spec",
    title="Spec: auth refactor requirements",
    project="myapp",
    content="<updated content with new requirements added>"
  )
  # revision_count is now incremented; previous version preserved in history
  ```

  ---

  ## Compaction Recovery (MANDATORY)

  If you see a compaction message or the text "FIRST ACTION REQUIRED":

  1. **Immediately** call `mem_session_summary` with the compacted content —
     this persists pre-compaction work before it is lost.
  2. Call `mem_context` to recover context from previous sessions.
  3. Only then continue working.

  **Why step 1 is non-negotiable**: the compacted content represents everything
  done in this session before the context window was trimmed. Without saving it
  explicitly, those observations are permanently lost from persistent memory.

  **Example recovery:**
  ```
  # Step 1 — save pre-compaction work
  mem_session_summary(
    content="<paste or reconstruct the compacted summary here>",
    project="myapp"
  )

  # Step 2 — recover broader context
  mem_context(project="myapp")

  # Step 3 — continue where you left off
  ```
  ````

- [ ] **Step 3: Read back and confirm key sections present**

  Read `skills/engram-conventions/lifecycle.md`. Confirm all of the following are present:
  - `## Read Patterns` table with 5 rows and example queries block
  - `## Session Lifecycle` section with 3 numbered sub-stages (Start, During, End)
  - `## \`mem_session_summary\` Content Shape` section with the content template and example call
  - `## Conflict Resolution` section with the candidate iteration pattern, ask-user heuristic, and silent-resolve heuristic
  - `## Updates Decision Table` with 4 rows and 2 example code blocks
  - `## Compaction Recovery (MANDATORY)` section with 3-step protocol and example

- [ ] **Step 4: Commit**

  ```bash
  git add skills/engram-conventions/lifecycle.md
  git commit -m "feat(engram-conventions): add lifecycle.md with reads, session management, and conflict resolution"
  ```

---

## Task 6: Write workflows/sdd.md

**Files:**
- Create: `skills/engram-conventions/workflows/sdd.md`

- [ ] **Step 1: Verify the file does not exist yet**

  Run:
  ```
  Glob pattern: skills/engram-conventions/workflows/sdd.md
  ```
  Expected: no files found.

- [ ] **Step 2: Write `skills/engram-conventions/workflows/sdd.md`**

  Write this exact content:

  ````markdown
  # SDD Workflow Cookbook

  SDD (Spec-Driven Development) is Gentle.AI's agent-teams-lite skill set. Each
  phase produces a typed observation with a predictable `topic_key`. Use the
  mapping below to ensure consistent, queryable artifacts across all SDD changes.

  ---

  ## Phase → Type → topic_key Mapping

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
  **same identifier** throughout all phases of the same change — the entire
  history of a change is queryable via `topic_key_prefix="sdd/<change>/"`.

  ---

  ## Content Shapes per Phase

  | Phase | Suggested headings |
  |-------|--------------------|
  | `exploration` | `## Investigated` / `## Approaches` / `## Recommendation` |
  | `proposal` | `## Intent` / `## Scope` / `## Approach` |
  | `spec` | `## Requirements` / `## Scenarios` |
  | `design` | `## Decisions` / `## Approach` / `## Tradeoffs` |
  | `tasks` | Markdown checklist (`- [ ] item`) |
  | `report` (apply-progress) | `## Completed` / `## In Progress` / `## Remaining` |
  | `report` (verify) | `## CRITICAL` / `## WARNING` / `## SUGGESTION` |
  | `report` (archive) | `## Closed` / `## Open` / `## Outcome` |

  ---

  ## Example Saves per Phase

  ### sdd-explore → `exploration`
  ```
  mem_save(
    title="Explored SSE vs WebSocket for live observation updates",
    type="exploration",
    topic_key="sdd/live-updates/explore",
    project="engram-ui",
    content="## Investigated\nReal-time delivery of new observations to engram-ui dashboard.\n\n## Approaches\n- SSE: simpler, HTTP/1.1 compatible, server-push only\n- WebSocket: bidirectional, more complex, overkill here\n\n## Recommendation\nSSE — sufficient for read-only push, no bidirectional need."
  )
  ```

  ### sdd-propose → `proposal`
  ```
  mem_save(
    title="Proposal: auth refactor with JWT rotation",
    type="proposal",
    topic_key="sdd/auth-refactor/proposal",
    project="myapp",
    content="## Intent\nCurrent session tokens never expire. Need short-lived JWTs with refresh rotation.\n\n## Scope\nIn: token issuance, refresh endpoint, middleware. Out: OAuth providers (V2).\n\n## Approach\nReplace session table with signed JWTs. Refresh tokens stored httpOnly cookies."
  )
  ```

  ### sdd-spec → `spec`
  ```
  mem_save(
    title="Spec: auth refactor requirements",
    type="spec",
    topic_key="sdd/auth-refactor/spec",
    project="myapp",
    content="## Requirements\n- Tokens expire after 15 minutes\n- Refresh tokens valid 7 days\n- Refresh rotates on use\n\n## Scenarios\n- Given expired token, when client sends request, then 401 returned\n- Given valid refresh token, when client calls /refresh, then new token pair issued"
  )
  ```

  ### sdd-design → `design`
  ```
  mem_save(
    title="Design: token rotation strategy",
    type="design",
    topic_key="sdd/auth-refactor/design",
    project="myapp",
    content="## Decisions\n- Stateless JWTs signed HMAC-SHA256\n- Refresh tokens stored in DB with revocation list\n\n## Approach\nAccess token 15-min signed JWT. Refresh token opaque 32-byte random, stored hashed.\n\n## Tradeoffs\nStateless access tokens cannot be revoked early — acceptable for 15-min TTL."
  )
  ```

  ### sdd-tasks → `tasks`
  ```
  mem_save(
    title="Tasks: auth refactor (8 items)",
    type="tasks",
    topic_key="sdd/auth-refactor/tasks",
    project="myapp",
    content="- [ ] Add golang-jwt/jwt dependency\n- [ ] Create internal/auth/token.go\n- [ ] Add POST /api/refresh endpoint\n- [ ] Update auth middleware\n- [ ] Add refresh token DB table\n- [ ] Write unit tests for token.go\n- [ ] Write integration test for full auth flow\n- [ ] Update README"
  )
  ```

  ### sdd-apply progress → `report`
  ```
  mem_save(
    title="Apply progress: auth refactor — 5/8 done",
    type="report",
    topic_key="sdd/auth-refactor/apply-progress",
    project="myapp",
    content="## Completed\n- Added golang-jwt/jwt\n- Created token.go\n- POST /api/refresh endpoint\n- Auth middleware updated\n- Refresh token DB table\n\n## In Progress\n- Unit tests (partial)\n\n## Remaining\n- Integration test\n- README update"
  )
  ```

  ### sdd-verify → `report`
  ```
  mem_save(
    title="Verify report: auth refactor — 0 CRITICAL",
    type="report",
    topic_key="sdd/auth-refactor/verify-report",
    project="myapp",
    content="## CRITICAL\n(none)\n\n## WARNING\n- Token cleanup job not implemented (deferred)\n\n## SUGGESTION\n- Consider admin-forced logout via blacklist"
  )
  ```

  ### sdd-archive → `report`
  ```
  mem_save(
    title="Archive: auth refactor — closed",
    type="report",
    topic_key="sdd/auth-refactor/archive-report",
    project="myapp",
    content="## Closed\nJWT rotation fully implemented. 8/8 tasks complete.\n\n## Open\nToken cleanup job deferred to V2.\n\n## Outcome\nSession tokens now expire after 15 minutes. Refresh rotation prevents token theft."
  )
  ```

  ### sdd-init project context → `config`
  ```
  mem_save(
    title="Config: sdd-init context for myapp",
    type="config",
    topic_key="sdd-init/myapp",
    project="myapp",
    content="## Stack\nGo 1.22, Chi router, GORM, SQLite\n\n## Test command\ngo test ./... -v\n\n## Strict TDD\ntrue\n\n## Conventions\n- Integration tests prefixed _e2e\n- Conventional commits"
  )
  ```

  ---

  ## Useful Read Patterns

  ```
  # Resume a change — get all phases at once
  mem_search(topic_key_prefix="sdd/auth-refactor/", project="myapp")

  # Timeline of one phase (all revisions)
  mem_search(topic_key="sdd/auth-refactor/spec", project="myapp")

  # All proposals in this project
  mem_search(type="proposal", project="myapp")

  # All reports (verify + archive + apply progress)
  mem_search(type="report", topic_key_prefix="sdd/", project="myapp")
  ```

  After each search hit, call `mem_get_observation(id=<id>)` to get the full
  content — search results are truncated previews.

  ---

  ## Multi-Repo SDD

  For SDD work scoped to one repo, prepend the repo identifier:
  ```
  frontend/sdd/auth-refactor/spec
  backend/sdd/auth-refactor/spec
  ```

  For cross-cutting SDD work (affects all repos), omit the repo prefix:
  ```
  sdd/auth-refactor/architecture   # no repo prefix — cross-cutting
  ```

  Querying across repos:
  ```
  # All frontend SDD artifacts for auth-refactor
  mem_search(topic_key_prefix="frontend/sdd/auth-refactor/", project="myapp")

  # All SDD artifacts for auth-refactor across all repos
  mem_search(topic_key_prefix="sdd/auth-refactor/", project="myapp")
  # Note: also catches cross-cutting artifacts (no repo prefix)
  ```
  ````

- [ ] **Step 3: Read back and confirm key sections present**

  Read `skills/engram-conventions/workflows/sdd.md`. Confirm all of the following are present:
  - `## Phase → Type → topic_key Mapping` table with all 9 SDD phases
  - `## Content Shapes per Phase` table with 8 rows (covering `exploration`, `proposal`, `spec`, `design`, `tasks`, `report` apply-progress, `report` verify, `report` archive)
  - `## Example Saves per Phase` section with 9 complete `mem_save` call examples
  - `## Useful Read Patterns` section with 4 `mem_search` examples
  - `## Multi-Repo SDD` section with repo-prefix examples and cross-repo query

- [ ] **Step 4: Commit**

  ```bash
  git add skills/engram-conventions/workflows/sdd.md
  git commit -m "feat(engram-conventions): add workflows/sdd.md with SDD phase mapping and examples"
  ```

---

## Task 7: Write workflows/superpowers.md

**Files:**
- Create: `skills/engram-conventions/workflows/superpowers.md`

- [ ] **Step 1: Verify the file does not exist yet**

  Run:
  ```
  Glob pattern: skills/engram-conventions/workflows/superpowers.md
  ```
  Expected: no files found.

- [ ] **Step 2: Write `skills/engram-conventions/workflows/superpowers.md`**

  Write this exact content:

  ````markdown
  # Superpowers Workflow Cookbook

  Superpowers is a set of Gentle.AI skills for brainstorming and planning. Each
  skill produces typed observations that complement the disk files the skills
  also write.

  ---

  ## Skill → Type → topic_key Mapping

  | Superpowers skill | type | topic_key |
  |-------------------|------|-----------|
  | `brainstorming` design doc | `design` | `superpowers/<feature>/design` |
  | `writing-plans` implementation plan | `plan` | `superpowers/<feature>/plan` |

  `<feature>` = stable kebab-case identifier for the feature being designed or
  planned. Use a **meaningful name**, not a date — dates belong in the disk
  filename, not in the engram key.

  ---

  ## Dual Storage Note

  Superpowers skills write artifacts to disk:
  - Design docs → `docs/superpowers/specs/YYYY-MM-DD-<topic>-design.md`
  - Plans → `docs/superpowers/plans/YYYY-MM-DD-<feature>.md`

  The engram save **complements** the disk file — it does not replace it.

  | Storage | Purpose |
  |---------|---------|
  | Disk file | Git history, team collaboration, diffs, PR review |
  | Engram observation | Cross-session searchability, timeline views in engram-ui |

  Both saves should happen. The disk file is the canonical artifact; engram is
  the searchable index.

  ---

  ## Content Shapes

  | Skill | Suggested headings |
  |-------|--------------------|
  | `brainstorming` (design) | `## Architecture` / `## Components` / `## Data flow` / `## Error handling` / `## Testing` |
  | `writing-plans` (plan) | `## Steps` / `## Order` / `## Validation` |

  Keep the engram content concise — it is a summary for searchability and
  timeline rendering, not a full duplicate of the disk doc. 3–8 sentences or
  bullet points per section is appropriate.

  ---

  ## Example Saves

  ### `brainstorming` → `design`
  ```
  mem_save(
    title="Design: engram-conventions skill architecture",
    type="design",
    topic_key="superpowers/engram-conventions/design",
    project="engram-ui",
    content="## Architecture\n9 markdown files: SKILL.md entry point + 5 reference subfiles + 3 workflow cookbooks.\n\n## Components\nSKILL.md loads always (Agent Skills spec). Subfiles loaded on-demand via decision table.\n\n## Data flow\nAgent reads SKILL.md → matches decision table → reads relevant subfile → shapes save.\n\n## Error handling\nFallback rule: if no type fits, use 'discovery'. Skill guidance, not API enforcement.\n\n## Testing\nNo code — consistency checked manually via cross-file review task."
  )
  ```

  ### `writing-plans` → `plan`
  ```
  mem_save(
    title="Plan: engram-conventions skill implementation",
    type="plan",
    topic_key="superpowers/engram-conventions/plan",
    project="engram-ui",
    content="## Steps\n9 tasks: SKILL.md, types.md, topic-keys.md, multi-repo.md, lifecycle.md, workflows/sdd.md, workflows/superpowers.md, workflows/ad-hoc.md, cross-file consistency check.\n\n## Order\nSKILL.md first (entry point). Subfiles independent of each other. Workflows last (reference subfiles).\n\n## Validation\nRead-back check per task. Cross-file consistency check in Task 9."
  )
  ```

  ---

  ## Linking Design and Plan

  Both artifacts share the `<feature>` slot in their `topic_key`. To retrieve
  both together:

  ```
  mem_search(topic_key_prefix="superpowers/engram-conventions/", project="engram-ui")
  ```

  Returns both `superpowers/engram-conventions/design` and
  `superpowers/engram-conventions/plan`.

  This prefix query also returns any future artifacts saved under the same
  feature identifier (e.g., a retrospective or a follow-up exploration).

  ---

  ## Useful Read Patterns

  ```
  # All Superpowers artifacts for a feature
  mem_search(topic_key_prefix="superpowers/payment-flow/", project="myapp")

  # All plans across the project
  mem_search(type="plan", project="myapp")

  # All designs across the project
  mem_search(type="design", project="myapp")
  ```

  After each search hit, call `mem_get_observation(id=<id>)` to get the full
  content — search results are truncated previews.

  ---

  ## Multi-Repo Superpowers

  For design or plan work scoped to one repo, prepend the repo identifier:
  ```
  frontend/superpowers/payment-flow/design
  backend/superpowers/payment-flow/design
  ```

  For cross-cutting work (affects all repos), omit the repo prefix:
  ```
  superpowers/payment-flow/design   # no repo prefix — cross-cutting
  ```

  Querying all repos:
  ```
  mem_search(topic_key_prefix="superpowers/payment-flow/", project="myapp")
  # Returns both repo-scoped and cross-cutting artifacts for payment-flow
  ```
  ````

- [ ] **Step 3: Read back and confirm key sections present**

  Read `skills/engram-conventions/workflows/superpowers.md`. Confirm all of the following are present:
  - `## Skill → Type → topic_key Mapping` table with both Superpowers skills
  - `## Dual Storage Note` section with the disk file paths and comparison table
  - `## Content Shapes` table for both skills with suggested headings
  - `## Example Saves` section with 2 complete `mem_save` call examples
  - `## Linking Design and Plan` section with `topic_key_prefix` query example
  - `## Useful Read Patterns` section with 3 `mem_search` examples
  - `## Multi-Repo Superpowers` section with repo-prefix examples

- [ ] **Step 4: Commit**

  ```bash
  git add skills/engram-conventions/workflows/superpowers.md
  git commit -m "feat(engram-conventions): add workflows/superpowers.md with brainstorming and writing-plans mapping"
  ```

---

## Task 8: Write workflows/ad-hoc.md

**Files:**
- Create: `skills/engram-conventions/workflows/ad-hoc.md`

- [ ] **Step 1: Verify the file does not exist yet**

  Run:
  ```
  Glob pattern: skills/engram-conventions/workflows/ad-hoc.md
  ```
  Expected: no files found.

- [ ] **Step 2: Write `skills/engram-conventions/workflows/ad-hoc.md`**

  Write this exact content:

  ````markdown
  # Ad-hoc Saves Cookbook

  Ad-hoc saves are not tied to any workflow. They cover the majority of
  everyday saves: decisions made during implementation, bugs fixed, patterns
  established, discoveries about the codebase, configuration changes, and user
  preferences.

  ---

  ## Type → topic_key Pattern

  | Type | topic_key pattern | Example |
  |------|-------------------|---------|
  | `decision` | `decision/<short-id>` | `decision/cookie-vs-localstorage` |
  | `architecture` | `architecture/<area>` | `architecture/auth-model` |
  | `bugfix` | `bugfix/<short-id>` | `bugfix/n-plus-one-userlist` |
  | `pattern` | `pattern/<name>` | `pattern/e2e-test-suffix` |
  | `discovery` | `discovery/<short-id>` | `discovery/fts5-strips-digits` |
  | `config` | `config/<area>` | `config/test-db-sqlite` |
  | `preference` | `preference/<area>` | `preference/commit-style` |

  ---

  ## short-id Rules

  - **Kebab-case, 3–5 words max** — short enough to type, specific enough to
    be meaningful.
  - **No dates** — engram tracks `created_at` automatically. A date in the
    `topic_key` is redundant and will make the key look stale.
  - **Stable** — do not rename after first use. Renaming orphans the existing
    observation and breaks the upsert chain.

  **Good**: `n-plus-one-userlist`, `cookie-vs-localstorage`, `fts5-strips-digits`

  **Bad**:
  - `bug-2024-05-14` — date in key, will look stale, renamed next year
  - `the-thing` — not descriptive, impossible to search
  - `fix` — too vague, will collide with other fixes
  - `important-decision-made-today` — date implicit, too long

  ---

  ## Example Saves

  ### `decision`
  ```
  mem_save(
    title="Decision: httpOnly cookies over localStorage for refresh tokens",
    type="decision",
    topic_key="decision/cookie-vs-localstorage",
    project="myapp",
    content="## What\nRefresh tokens stored in httpOnly cookies, not localStorage.\n\n## Why\nXSS cannot read httpOnly cookies. localStorage is accessible to injected JS.\n\n## Where\nPOST /api/refresh sets Set-Cookie. Frontend never reads the token directly.\n\n## Learned\nSameSite=Strict breaks cross-origin OAuth redirects — use SameSite=Lax."
  )
  ```

  ### `architecture`
  ```
  mem_save(
    title="Architecture: hexagonal layout for auth domain",
    type="architecture",
    topic_key="architecture/auth-model",
    project="myapp",
    content="## What\nAuth domain follows hexagonal architecture with explicit ports and adapters.\n\n## Why\nAllows swapping HTTP handler or DB adapter without touching business logic.\n\n## Where\ninternal/auth/ — core/, ports/, adapters/http/, adapters/db/\n\n## Tradeoffs\nMore files and interfaces upfront. Worth it for testability and replaceability."
  )
  ```

  ### `bugfix`
  ```
  mem_save(
    title="Fixed N+1 query in UserList",
    type="bugfix",
    topic_key="bugfix/n-plus-one-userlist",
    project="myapp",
    content="## What\nUserList issued one SQL query per row to fetch the associated org.\n\n## Why\nGORM's lazy-loading triggered on user.Org access inside the render loop.\n\n## Where\ninternal/user/list.go — added Preload(\"Org\") to base query.\n\n## Learned\nGORM does not preload by default. Any association access in a loop is an N+1 unless Preload() is explicit."
  )
  ```

  ### `pattern`
  ```
  mem_save(
    title="Pattern: prefix integration tests with _e2e",
    type="pattern",
    topic_key="pattern/e2e-test-suffix",
    project="myapp",
    content="## What\nAll integration tests that hit the real DB or network use _e2e prefix in the filename.\n\n## Why\nAllows `go test ./... -run '^[^_]'` to skip integration tests in CI fast lane.\n\n## Where\nAll test files under tests/integration/\n\n## Learned\nGo build tags are an alternative but require flag discipline. Prefix is simpler and visible in IDEs."
  )
  ```

  ### `discovery`
  ```
  mem_save(
    title="Discovery: FTS5 strips digits, breaks search for version strings",
    type="discovery",
    topic_key="discovery/fts5-strips-digits",
    project="engram-ui",
    content="## What\nSQLite FTS5 tokenizer strips standalone digit sequences by default.\n\n## Why it matters\nSearching for 'v1' or '404' returns no results even when present in content.\n\n## Where\nengram/internal/store/fts.go — tokenizer config.\n\n## Learned\nUse `unicode61 tokenchars` or `ascii` tokenizer to preserve digits. Requires FTS index rebuild after changing tokenizer."
  )
  ```

  ### `config`
  ```
  mem_save(
    title="Config: switched test DB to SQLite in-memory",
    type="config",
    topic_key="config/test-db-sqlite",
    project="myapp",
    content="## What\nTest suite now uses sqlite3 :memory: instead of local Postgres.\n\n## Why\nCI was flaky due to Postgres port conflicts. In-memory SQLite is zero-setup.\n\n## Where\nconfig/test.yaml — DB_URL changed. Makefile test target no longer needs docker compose.\n\n## Learned\nSQLite lacks some Postgres features (RETURNING on UPDATE). Two tests rewritten."
  )
  ```

  ### `preference`
  ```
  mem_save(
    title="Preference: always use httpOnly cookies, never localStorage",
    type="preference",
    topic_key="preference/cookie-style",
    project="myapp",
    content="## What\nUser requires all tokens stored in httpOnly cookies.\n\n## Why\nSecurity policy — XSS protection non-negotiable.\n\n## Where\nAny feature that stores tokens or session identifiers."
  )
  ```

  ---

  ## Ad-hoc vs Workflow Output

  If a save is the output of an SDD or Superpowers phase, use the workflow
  namespace (`sdd/`, `superpowers/`). If it is a standalone save made during
  implementation without a governing workflow, use the ad-hoc namespaces above.

  **Do not mix namespaces for the same artifact.** An architecture decision
  reached during an SDD design phase belongs in `sdd/<change>/design`, not in
  `architecture/<area>`. After the change is archived, a cross-cutting permanent
  record may be saved separately under `architecture/<area>`.

  | Save context | Namespace |
  |--------------|-----------|
  | Output of `sdd-design` phase | `sdd/<change>/design` |
  | Output of `brainstorming` skill | `superpowers/<feature>/design` |
  | Standalone decision during implementation | `decision/<short-id>` |
  | Permanent cross-cutting architecture record | `architecture/<area>` |

  ---

  ## Multi-Repo Ad-hoc

  For saves scoped to one repo, prepend the repo identifier:
  ```
  frontend/bugfix/n-plus-one
  backend/discovery/fts5-strips-digits
  ```

  For cross-cutting saves (e.g., an architecture decision affecting all repos),
  omit the repo prefix:
  ```
  architecture/auth-model         # no repo prefix — affects all repos
  decision/cookie-vs-localstorage # no repo prefix — product-wide decision
  ```

  Querying ad-hoc artifacts for a specific type across all repos:
  ```
  mem_search(type="bugfix", project="myapp")
  mem_search(type="architecture", project="myapp")
  ```
  ````

- [ ] **Step 3: Read back and confirm key sections present**

  Read `skills/engram-conventions/workflows/ad-hoc.md`. Confirm all of the following are present:
  - `## Type → topic_key Pattern` table with all 7 ad-hoc types
  - `## short-id Rules` section with good/bad examples
  - `## Example Saves` section with 7 complete `mem_save` examples (one per type: decision, architecture, bugfix, pattern, discovery, config, preference)
  - `## Ad-hoc vs Workflow Output` section with the decision table (4 rows)
  - `## Multi-Repo Ad-hoc` section with repo-prefix examples and query examples

- [ ] **Step 4: Commit**

  ```bash
  git add skills/engram-conventions/workflows/ad-hoc.md
  git commit -m "feat(engram-conventions): add workflows/ad-hoc.md with type-to-topic-key patterns and examples"
  ```

---

## Task 9: Cross-file consistency check and final commit

**Files:**
- Read: all 9 files (no writes unless fixes needed)
- Modify: any file with consistency issues found

- [ ] **Step 1: Check — every namespace in topic-keys.md registry appears in a workflow file**

  Open `skills/engram-conventions/topic-keys.md`. For each row in the Namespace Registry table, verify the "Documented in" file exists and covers that namespace:

  | Namespace | Expected in |
  |-----------|-------------|
  | `sdd/<change>/<phase>` | `workflows/sdd.md` — Phase mapping table |
  | `superpowers/<feature>/<phase>` | `workflows/superpowers.md` — Skill mapping table |
  | `architecture/<area>` | `workflows/ad-hoc.md` — Type mapping table |
  | `decision/<short-id>` | `workflows/ad-hoc.md` — Type mapping table |
  | `bugfix/<short-id>` | `workflows/ad-hoc.md` — Type mapping table |
  | `pattern/<name>` | `workflows/ad-hoc.md` — Type mapping table |
  | `discovery/<short-id>` | `workflows/ad-hoc.md` — Type mapping table |
  | `config/<area>` | `workflows/ad-hoc.md` — Type mapping table |
  | `preference/<area>` | `workflows/ad-hoc.md` — Type mapping table |

  If any namespace is not documented in the expected file, add the missing entry to the workflow file.

- [ ] **Step 2: Check — every type in types.md appears in at least one workflow example**

  The 14 types and where they should appear:

  | Type | Expected coverage |
  |------|-------------------|
  | `exploration` | `workflows/sdd.md` — `sdd-explore` row |
  | `proposal` | `workflows/sdd.md` — `sdd-propose` row |
  | `spec` | `workflows/sdd.md` — `sdd-spec` row |
  | `design` | `workflows/sdd.md` and `workflows/superpowers.md` |
  | `plan` | `workflows/superpowers.md` — `writing-plans` row |
  | `tasks` | `workflows/sdd.md` — `sdd-tasks` row |
  | `report` | `workflows/sdd.md` — apply-progress, verify, archive rows |
  | `decision` | `workflows/ad-hoc.md` — type mapping table |
  | `architecture` | `workflows/ad-hoc.md` — type mapping table |
  | `bugfix` | `workflows/ad-hoc.md` — type mapping table |
  | `pattern` | `workflows/ad-hoc.md` — type mapping table |
  | `discovery` | `workflows/ad-hoc.md` — type mapping table |
  | `config` | `workflows/sdd.md` (sdd-init) and `workflows/ad-hoc.md` |
  | `preference` | `workflows/ad-hoc.md` — type mapping table |

  If any type has no example in any workflow file, add a brief example to the most relevant workflow file.

- [ ] **Step 3: Check — SKILL.md decision table references all subfiles**

  Open `skills/engram-conventions/SKILL.md`. Verify the decision table contains rows pointing to:
  - `types.md`
  - `topic-keys.md`
  - `multi-repo.md`
  - `lifecycle.md` (at least 3 rows: reads, conflicts, updates)
  - `workflows/sdd.md`
  - `workflows/superpowers.md`
  - `workflows/ad-hoc.md`

  If any file is missing from the decision table, add the row.

- [ ] **Step 4: Check — type names are consistent across all files**

  Verify the following type spellings are identical (lowercase, no hyphens) in every file where they appear:

  `exploration`, `proposal`, `spec`, `design`, `plan`, `tasks`, `report`,
  `decision`, `architecture`, `bugfix`, `pattern`, `discovery`, `config`,
  `preference`

  Search for variant spellings: `"archival"`, `"arch"`, `"bug-fix"`,
  `"bug_fix"`, `"note"`, `"architectural-decision"`. If any are found, fix
  them to match the canonical names above.

- [ ] **Step 5: Check — topic_key patterns are consistent**

  Verify these exact patterns appear consistently across files:

  | Pattern | Where it appears |
  |---------|-----------------|
  | `sdd/<change>/explore` | `types.md` exploration example, `workflows/sdd.md` mapping |
  | `sdd/<change>/proposal` | `types.md` proposal example, `workflows/sdd.md` mapping |
  | `sdd/<change>/spec` | `types.md` spec example, `workflows/sdd.md` mapping |
  | `sdd/<change>/design` | `types.md` design example, `workflows/sdd.md` mapping |
  | `sdd/<change>/tasks` | `types.md` tasks example, `workflows/sdd.md` mapping |
  | `sdd/<change>/apply-progress` | `workflows/sdd.md` mapping |
  | `sdd/<change>/verify-report` | `types.md` report example, `workflows/sdd.md` mapping |
  | `sdd/<change>/archive-report` | `workflows/sdd.md` mapping |
  | `sdd-init/<project>` | `workflows/sdd.md` mapping |
  | `superpowers/<feature>/design` | `workflows/superpowers.md` mapping |
  | `superpowers/<feature>/plan` | `workflows/superpowers.md` mapping |

  If any pattern is inconsistent (e.g., `sdd/<change>/archive` vs
  `sdd/<change>/archive-report`), fix the inconsistency to match the spec.

- [ ] **Step 6: Commit (only if changes were made in steps 1–5)**

  If no changes: skip this step.

  If changes were made:
  ```bash
  git add skills/engram-conventions/
  git commit -m "fix(engram-conventions): cross-file consistency fixes"
  ```

- [ ] **Step 7: Final commit tagging the skill as complete**

  ```bash
  git add skills/engram-conventions/
  git commit -m "feat(engram-conventions): complete engram-conventions skill (9 files)"
  ```

  (If step 6 already committed all fixes, run `git status` first to confirm
  there are no remaining unstaged changes before running this commit.)

---

## Self-Review Notes

The following checks were applied before finalizing this plan:

### 1. Spec Coverage

| Spec section | Task that implements it |
|--------------|------------------------|
| §1 Overview | Task 1 (SKILL.md body/preamble) |
| §2 Goals — 14 types | Task 2 (types.md) |
| §2 Goals — topic_key namespace | Task 3 (topic-keys.md) |
| §2 Goals — multi-repo | Task 4 (multi-repo.md) |
| §2 Goals — full lifecycle | Task 5 (lifecycle.md) |
| §2 Goals — workflow cookbooks | Tasks 6, 7, 8 |
| §2 Goals — portability | Task 1 (SKILL.md Compatibility section) |
| §2 Goals — non-invasive | Task 1 (SKILL.md preamble) |
| §3 Non-goals | Not a deliverable — no task needed |
| §4 Architecture / file layout | Tasks 1–8 create exactly the 9 files |
| §5 SKILL.md — frontmatter | Task 1 step 3 |
| §5 SKILL.md — 5-rule TL;DR | Task 1 step 3 |
| §5 SKILL.md — decision table | Task 1 step 3 |
| §5 SKILL.md — compatibility note | Task 1 step 3 |
| §6 Types — taxonomy table | Task 2 taxonomy overview |
| §6 Types — per-type detail | Task 2 per-type reference sections |
| §6 Types — decision-aid tree | Task 2 decision-aid tree |
| §6 Types — decision vs architecture | Task 2 per-type sections |
| §7 topic_key — convention | Task 3 convention section |
| §7 topic_key — rules | Task 3 rules section |
| §7 topic_key — namespace registry | Task 3 namespace registry table |
| §7 topic_key — good/bad examples | Task 3 good vs bad section |
| §7 topic_key — revisions vs new | Task 3 revisions section |
| §7 topic_key — timeline queries | Task 3 timeline queries section |
| §8 Multi-repo — default behavior | Task 4 default behavior section |
| §8 Multi-repo — setup | Task 4 setup section |
| §8 Multi-repo — signal detection | Task 4 signal detection section |
| §8 Multi-repo — ask-user prompt | Task 4 ask-user flow section |
| §8 Multi-repo — topic_key dimensions | Task 4 topic_key dimensions section |
| §8 Multi-repo — cross-repo reads | Task 4 cross-repo reads section |
| §8 Multi-repo — pitfalls | Task 4 pitfalls section |
| §9 Lifecycle — read patterns | Task 5 read patterns section |
| §9 Lifecycle — session lifecycle | Task 5 session lifecycle section |
| §9 Lifecycle — mem_session_summary shape | Task 5 content shape section |
| §9 Lifecycle — mem_judge conflict | Task 5 conflict resolution section |
| §9 Lifecycle — updates table | Task 5 updates decision table |
| §9 Lifecycle — compaction recovery | Task 5 compaction recovery section |
| §10 SDD cookbook — phase mapping | Task 6 phase mapping table |
| §10 SDD cookbook — content shapes | Task 6 content shapes table |
| §10 SDD cookbook — read patterns | Task 6 read patterns section |
| §10 SDD cookbook — multi-repo SDD | Task 6 multi-repo section |
| §11 Superpowers cookbook — mapping | Task 7 skill mapping table |
| §11 Superpowers cookbook — dual storage | Task 7 dual storage note |
| §11 Superpowers cookbook — content shapes | Task 7 content shapes table |
| §11 Superpowers cookbook — linking | Task 7 linking section |
| §11 Superpowers cookbook — multi-repo | Task 7 multi-repo section |
| §12 Ad-hoc cookbook — type→topic_key | Task 8 type mapping table |
| §12 Ad-hoc cookbook — short-id rules | Task 8 short-id rules section |
| §12 Ad-hoc cookbook — ad-hoc vs workflow | Task 8 ad-hoc vs workflow section |
| §12 Ad-hoc cookbook — multi-repo | Task 8 multi-repo section |
| §13 Compatibility | Task 1 compatibility section |
| §14 Open questions | Not implemented (deliberately excluded from V1) |

No gaps found.

### 2. Placeholder Scan

All task steps contain explicit, complete content. No "TBD", "TODO", "XXX",
"see above", "similar to task N", or "appropriate error handling" strings are
present in any step.

### 3. Type Consistency

All 14 type names are spelled identically (lowercase, no hyphens) across all
tasks: `exploration`, `proposal`, `spec`, `design`, `plan`, `tasks`, `report`,
`decision`, `architecture`, `bugfix`, `pattern`, `discovery`, `config`,
`preference`.

Namespace patterns verified consistent:
- `sdd/<change>/apply-progress` (not `apply_progress` or `apply-prog`)
- `sdd/<change>/verify-report` (not `verify_report` or `verify`)
- `sdd/<change>/archive-report` (not `archive` alone)
- `sdd-init/<project>` (not `sdd/init/<project>`)
- `superpowers/<feature>/design` (not `superpowers/<feature>-design`)
