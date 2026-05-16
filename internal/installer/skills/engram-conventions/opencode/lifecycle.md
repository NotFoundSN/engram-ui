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
| `mem_search` with exact `topic_key` | Current state of an artifact | Returns the current row for that `topic_key`. Engram upserts in place — historical content is NOT preserved, only `revision_count` is tracked. |
| `mem_current_project` | Verify cwd → project mapping | Use when multi-repo setup is uncertain |

> **Note on `topic_key_prefix`**: `topic_key_prefix` is a documented
> **convention** in this skill — it is NOT a parameter exposed by the
> engram MCP `mem_search` tool today. MCP `mem_search` accepts `query`,
> `type`, `project`, `scope`, `topic_key`, and `limit`.
>
> To achieve prefix-style queries via MCP, use one of these workarounds:
> (a) request by exact `topic_key` (returns the current row for that
> artifact); (b) filter by `type` plus a `query` keyword and post-filter
> results in your client by `topic_key.startswith("<prefix>")`;
> (c) use `query` with the prefix string as a keyword — FTS will match
> observations whose content or title mentions the prefix.
>
> Engram-side native `topic_key_prefix` support on `mem_search` is tracked
> as the v1 W-1 follow-up. When it ships, the workarounds below collapse
> to a single direct call.

**Query shape**: combine `query` (keywords) + `type`. Bare keyword search in
a busy project is noisy. At minimum, add a `type` to narrow results. Use
`topic_key` for exact artifact lookup. For prefix-style scoping, see the note
above and post-filter by `topic_key.startswith("<prefix>")`.

**Examples:**
```
# Start of session — always first
mem_context({"project": "myapp"})

# Find by type (narrow, low noise)
mem_search({"type": "spec", "query": "auth-refactor", "project": "myapp"})
# Then in your code: filter results where topic_key.startswith("sdd/auth-refactor/")

# Current state of one artifact (engram upserts in place — no historical content)
mem_search({"topic_key": "sdd/auth-refactor/spec", "project": "myapp"})

# Get full content (required after search)
mem_get_observation({"id": "<id from search result>"})

# Verify project mapping
mem_current_project({})
```

---

## Surfacing memories to the user via URL

After saving an observation, the agent MAY emit a shareable URL so the user
can open it in engram-ui for review.

### Format

```
Review: http://localhost:7438/m/{id}
```

`{id}` is the `id` field returned in the `mem_save` response envelope.
`7438` is engram-ui's default port.

### Custom installs — `ENGRAM_UI_URL`

If the user runs engram-ui on a non-default host or port, the convention is
to read `ENGRAM_UI_URL` from the environment and substitute it for the
`http://localhost:7438` prefix. Example: `ENGRAM_UI_URL=http://10.0.0.5:9000`
→ emit `Review: http://10.0.0.5:9000/m/{id}`.

### Fallback — engram-ui not installed

If engram-ui is not installed or not running, the emitted URL is
informational: clicking it fails locally, but the underlying engram save
succeeded. The skill stays useful — emitting the URL is never required for
correctness.

### Presentation

A single line in the agent's response, immediately after the save
confirmation. Example:

> Saved `sdd/auth-refactor/spec` to engram (id 124). Review: http://localhost:7438/m/124

**Situations where surfacing the URL is appropriate:**

- The agent just generated a spec, design, proposal, plan, or report and the
  user is expected to review or approve it before the next step.
- The user is in interactive review mode — a back-and-forth conversation
  where every phase pauses for approval.
- The saved content is substantial (more than ~500 characters) and reading
  it in engram-ui is friendlier than scrolling back through chat.
- The agent suspects the user wants to validate its reasoning before the
  next action (e.g., before launching the next SDD phase).

**Situations where surfacing the URL is noise:**

- A chain of automated saves where a single URL at the end of the chain is
  enough.
- Trivial saves (one-line bugfix note, short pattern note, single sentence
  preference).
- The agent is running in autonomous mode without user checkpoints
  (background work, batch operations).
- Internal bookkeeping saves (apply-progress checkpoints, session summaries
  written implicitly at session end).

Final judgement belongs to the agent reading the live conversation. The skill
provides guidance — never a rigid trigger. The capability stays useful
regardless of which situations the agent chooses.

---

## When to use engram vs a standalone .md file

Workflow artifacts and human-review documents are easier to find, link, and
revise when they live in engram instead of as files on disk. This is a
preference, not an absolute rule — some artifacts genuinely belong on disk.

**PREFER engram for** (no shadow `docs/specs/*.md` or `docs/designs/*.md`):

- Specs, designs, proposals, plans, exploration write-ups
- Verify reports, archive reports, apply-progress snapshots
- Decision and architecture notes generated during a workflow
- Anything intended for human review that the agent just produced

Save these via `mem_save` with the appropriate `type` and `topic_key`. The
agent may surface a `Review: ...` URL (see the section above) so the user
can open it in engram-ui.

**Exceptions — keep these on disk:**

- `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `LICENSE`
- Source-code docstrings and inline comments
- Project-level long-lived public docs (`docs/getting-started.md`,
  user-facing guides, public-API references)
- Architecture decision records (ADRs) intended for permanent public
  consumption — debatable; the agent MAY save these to engram too if it
  wants searchability

The strength of this guidance is **PREFER**, not **NEVER**. If the agent
needs to produce a `.md` file for a legitimate reason (publication,
external sharing, the user explicitly asked for a file), that remains
valid.

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
mem_session_summary({
  "content": "## Goal\nImplemented JWT token rotation for auth refactor.\n\n## Discoveries\n- GORM lazy-loads associations; Preload() required to avoid N+1\n- SameSite=Strict breaks OAuth redirect flow\n\n## Accomplished\n- Token issuance endpoint complete\n- Refresh endpoint complete\n- Integration tests passing (8/8)\n\n## Next Steps\n- Implement token cleanup job\n- Update OpenAPI spec\n\n## Relevant Files\n- internal/auth/token.go — JWT issue and verify\n- internal/auth/refresh.go — refresh endpoint handler",
  "project": "myapp"
})
```

---

## Conflict Resolution (`mem_judge`)

After every `mem_save`, check the response envelope for `judgment_required`.

### When `judgment_required: true`

Iterate `candidates[]`. For each candidate, call `mem_judge` once using **that
candidate's own `judgment_id`** — not the top-level one.

```
# For each candidate in candidates[]:
mem_judge({
  "judgment_id": "<candidate.judgment_id>",
  "relation": "<your assessed relation>",
  "confidence": "<0.0–1.0>"
})
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
mem_search({"topic_key": "sdd/auth-refactor/spec", "project": "myapp"})
# Get the full content and ID
mem_get_observation({"id": "<id>"})
# Update with the corrected content
mem_update({"id": "<id>", "title": "Spec: auth refactor requirements", "content": "<corrected content>"})
```

**Example — upsert (same topic evolving):**
```
# Add a new requirement to an existing spec — same topic_key
mem_save({
  "topic_key": "sdd/auth-refactor/spec",
  "type": "spec",
  "title": "Spec: auth refactor requirements",
  "project": "myapp",
  "content": "<updated content with new requirements added>"
})
# revision_count is incremented. The previous content is OVERWRITTEN in place — engram does not preserve historical content snapshots today.
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
mem_session_summary({
  "content": "<paste or reconstruct the compacted summary here>",
  "project": "myapp"
})

# Step 2 — recover broader context
mem_context({"project": "myapp"})

# Step 3 — continue where you left off
```
