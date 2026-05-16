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
mem_save({
  "title": "Compared SSE vs WebSocket for live observation updates",
  "type": "exploration",
  "topic_key": "sdd/live-updates/explore",
  "project": "engram-ui",
  "content": "## Investigated\nReal-time delivery of new observations to engram-ui.\n\n## Approaches\n- SSE: simpler, HTTP/1.1 compatible, server-push only\n- WebSocket: bidirectional, more complex, overkill here\n\n## Recommendation\nSSE — sufficient for read-only push, less infrastructure."
})
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
mem_save({
  "title": "Proposal: auth refactor with JWT rotation",
  "type": "proposal",
  "topic_key": "sdd/auth-refactor/proposal",
  "project": "myapp",
  "content": "## Intent\nCurrent session tokens never expire. Need short-lived JWTs with refresh rotation.\n\n## Scope\nIn: token issuance, refresh endpoint, middleware. Out: OAuth providers (V2).\n\n## Approach\nReplace session table with signed JWTs. Refresh tokens stored httpOnly cookies."
})
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
mem_save({
  "title": "Spec: auth refactor requirements",
  "type": "spec",
  "topic_key": "sdd/auth-refactor/spec",
  "project": "myapp",
  "content": "## Requirements\n- Tokens expire after 15 minutes\n- Refresh tokens valid 7 days\n- Refresh rotates on use\n\n## Scenarios\n- Given expired token, when client sends request, then 401 returned\n- Given valid refresh token, when client calls /refresh, then new token pair issued"
})
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
mem_save({
  "title": "Design: token rotation strategy",
  "type": "design",
  "topic_key": "sdd/auth-refactor/design",
  "project": "myapp",
  "content": "## Decisions\n- Stateless JWTs signed with HMAC-SHA256\n- Refresh tokens stored in DB with revocation list\n\n## Approach\nAccess token: 15-min signed JWT. Refresh token: opaque 32-byte random, stored hashed.\n\n## Tradeoffs\nStateless access tokens cannot be revoked early — acceptable for 15-min TTL."
})
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
mem_save({
  "title": "Plan: auth refactor implementation",
  "type": "plan",
  "topic_key": "superpowers/auth-refactor/plan",
  "project": "myapp",
  "content": "## Steps\n1. Add JWT library\n2. Implement token issuance endpoint\n3. Implement refresh endpoint\n4. Update middleware\n5. Write integration tests\n\n## Order\nMiddleware last — depends on issuance being complete.\n\n## Validation\nIntegration test suite passes. Existing session tests updated."
})
```

---

### `tasks`

**Use when**: breaking a change into a checkbox checklist of work items.
The output of `sdd-tasks`.

**Don't use when**: recording progress or completion status — use `report`
for that.

**Title shape**: `"Tasks: <change name> (<N> items)"`

**Content shape**: Markdown checklist (`- [ ] item`), one item per line.

**Example `mem_save` call**:
```
mem_save({
  "title": "Tasks: auth refactor (8 items)",
  "type": "tasks",
  "topic_key": "sdd/auth-refactor/tasks",
  "project": "myapp",
  "content": "- [ ] Add `golang-jwt/jwt` dependency\n- [ ] Create `internal/auth/token.go` with Issue() and Verify()\n- [ ] Add POST /api/refresh endpoint\n- [ ] Update auth middleware to parse JWT\n- [ ] Add refresh token DB table and migration\n- [ ] Write unit tests for token.go\n- [ ] Write integration test for full auth flow\n- [ ] Update README auth section"
})
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
mem_save({
  "title": "Verify report: auth refactor — 0 CRITICAL",
  "type": "report",
  "topic_key": "sdd/auth-refactor/verify-report",
  "project": "myapp",
  "content": "## CRITICAL\n(none)\n\n## WARNING\n- Refresh token cleanup job not implemented (deferred to V2)\n\n## SUGGESTION\n- Consider adding token blacklist for admin-forced logout"
})
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
mem_save({
  "title": "Decision: httpOnly cookies over localStorage for refresh tokens",
  "type": "decision",
  "topic_key": "decision/cookie-vs-localstorage",
  "project": "myapp",
  "content": "## What\nRefresh tokens stored in httpOnly cookies, not localStorage.\n\n## Why\nXSS cannot read httpOnly cookies. localStorage is accessible to JS.\n\n## Where\nPOST /api/refresh sets Set-Cookie header. Frontend never reads the token.\n\n## Learned\nSameSite=Strict breaks cross-origin flows — use SameSite=Lax for redirect-based OAuth."
})
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
mem_save({
  "title": "Architecture: hexagonal layout for auth domain",
  "type": "architecture",
  "topic_key": "architecture/auth-model",
  "project": "myapp",
  "content": "## What\nAuth domain follows hexagonal architecture with explicit ports and adapters.\n\n## Why\nAllows swapping HTTP handler or DB adapter without touching business logic. Testable in isolation.\n\n## Where\ninternal/auth/ — core/, ports/, adapters/http/, adapters/db/\n\n## Tradeoffs\nMore files and interfaces up front. Worth it for testability and replaceability."
})
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
mem_save({
  "title": "Fixed N+1 query in UserList",
  "type": "bugfix",
  "topic_key": "bugfix/n-plus-one-userlist",
  "project": "myapp",
  "content": "## What\nUserList was issuing one SQL query per row to fetch the associated org.\n\n## Why\nGORM's lazy-loading triggered on `user.Org` access inside the loop.\n\n## Where\ninternal/user/list.go — added Preload(\"Org\") to the base query.\n\n## Learned\nGORM does not preload by default. Any association access in a loop is an N+1 unless Preload() is explicit."
})
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
mem_save({
  "title": "Pattern: prefix integration tests with _e2e",
  "type": "pattern",
  "topic_key": "pattern/e2e-test-suffix",
  "project": "myapp",
  "content": "## What\nAll integration tests that hit the real DB or network are prefixed with `_e2e` in the filename.\n\n## Why\nAllows `go test ./... -run '^[^_]'` to skip integration tests in CI fast lane.\n\n## Where\nAll test files in tests/integration/\n\n## Learned\nGo build tags are an alternative but require flag discipline. Prefix is simpler and visible."
})
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
mem_save({
  "title": "Discovery: FTS5 strips digits, breaks search for version strings",
  "type": "discovery",
  "topic_key": "discovery/fts5-strips-digits",
  "project": "engram-ui",
  "content": "## What\nSQLite FTS5 tokenizer strips standalone digit sequences by default.\n\n## Why it matters\nSearching for 'v1' or '404' returns no results even when present in content.\n\n## Where\nengram/internal/store/fts.go — tokenizer config\n\n## Learned\nUse `unicode61 tokenchars` or `ascii` tokenizer to preserve digits. Requires FTS index rebuild."
})
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
mem_save({
  "title": "Config: switched test DB to SQLite in-memory",
  "type": "config",
  "topic_key": "config/test-db-sqlite",
  "project": "myapp",
  "content": "## What\nTest suite now uses `sqlite3 :memory:` instead of a local Postgres instance.\n\n## Why\nCI setup was flaky due to Postgres port conflicts. In-memory SQLite is zero-setup.\n\n## Where\nconfig/test.yaml — DB_URL changed. Makefile target `test` no longer needs `docker compose up`.\n\n## Learned\nSQLite lacks some Postgres features (e.g., RETURNING on UPDATE). Two tests were rewritten."
})
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
mem_save({
  "title": "Preference: always use httpOnly cookies, never localStorage",
  "type": "preference",
  "topic_key": "preference/cookie-style",
  "project": "myapp",
  "content": "## What\nUser requires all tokens to be stored in httpOnly cookies.\n\n## Why\nSecurity policy — XSS protection is non-negotiable.\n\n## Where\nAny feature that stores tokens or session identifiers."
})
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
