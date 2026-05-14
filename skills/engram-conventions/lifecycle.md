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

> **Note on `topic_key_prefix`**: The examples below use `topic_key_prefix` as the canonical query parameter, which engram-ui supports natively against its store. Engram's MCP `mem_search` tool today only exposes `query`, `type`, `project`, `scope`, and `limit` parameters — to achieve prefix queries via MCP, either (a) use `query` with the prefix as a keyword (FTS will match observations whose content or title mentions the prefix), or (b) request observations by exact `topic_key` for revisions, or (c) fetch a broader set and post-filter results by their `topic_key` field. The convention itself remains: stable namespaced keys enable prefix grouping wherever the consumer supports it.

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

## Instructions
[User preferences or constraints discovered — skip if none]

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
