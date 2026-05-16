---
name: debug
description: >
  Use this BEFORE proposing or applying any fix to a bug, test failure,
  unexpected behavior, performance regression, build failure, or production
  incident. The skill enforces root-cause-first investigation, bifurcates
  into fix-mode (apply a TDD fix) or document-mode (record the finding
  without coding), and saves the investigation as a structured observation
  in engram so future debugging sessions can find prior work.
triggers:
  - debug
  - mr fix-it
  - depurar
  - debuggear
  - fixealo
  - arreglame esto
  - hay un error en
  - hay un bug en
---

# debug

Investigate bugs systematically. Find the root cause before touching a
fix. Two exits: apply a TDD fix, or document the finding for someone else
to act on.

## The Iron Law

```
NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST
```

If you have not completed Phase 1, you cannot propose fixes. Symptom
fixes are failure. **Violating the letter of this process is violating
the spirit of debugging.**

## When to use

ANY technical issue:
- Test failures
- Bugs in production or local
- Unexpected behavior
- Performance regressions
- Build failures
- Integration issues

**Use this ESPECIALLY when:**
- Under time pressure (emergencies make guessing tempting)
- "Just one quick fix" seems obvious
- You have already tried multiple fixes
- Previous fix did not work
- You do not fully understand the issue

## Pre-flight: search engram first (MANDATORY)

**BEFORE Phase 1, you MUST run a search for prior work on this issue:**

```
mem_search({
  "query": "<keywords from the bug report or error message>",
  "type": "bugfix",
  "project": "{project}"
})
```

Why: an identical or similar bug may have been investigated or fixed
before. The prior observation tells you what was tried, what worked, and
what was learned. Skipping this step wastes 5-30 minutes of repeat work
per session over time.

If results look relevant, follow up with `mem_get_observation(id)` to
read full content (search results are previews, not full bodies).

If nothing relevant comes back, proceed to Phase 1.

## The four phases

You MUST complete each phase before proceeding to the next.

### Phase 1 — Root cause investigation

**BEFORE attempting ANY fix:**

1. **Read errors carefully** — full stack traces, all warnings, every line.
   Note file paths, line numbers, error codes. Errors often contain the
   exact answer.
2. **Reproduce consistently** — can you trigger it reliably? What are the
   exact steps? Does it happen every time? If not reproducible → gather
   more data, do NOT guess.
3. **Check recent changes** — `git log`, `git diff`, recent dependency
   updates, config changes, environment differences.
4. **Gather evidence at component boundaries** — when the system has
   multiple layers (CI → build → signing, API → service → DB), instrument
   each boundary BEFORE proposing fixes. Log what enters and what exits
   each component. Run once to find WHERE it breaks, then investigate
   that specific component.
5. **Trace data flow backward** — where does the bad value originate?
   What called this with the bad value? Keep tracing up until you find
   the source. Fix at source, not at symptom.

**Save the Phase 1 outcome to engram BEFORE moving on.** See "Engram saves"
below — investigation observation with what you found.

### Phase 2 — Pattern analysis

**Find the pattern before fixing:**

1. **Find working examples** — locate similar working code in the same
   codebase. What works that resembles what is broken?
2. **Compare against references** — if implementing a known pattern, read
   the reference implementation COMPLETELY. Do not skim.
3. **Identify differences** — list every difference between working and
   broken, however small. Do NOT assume "that can't matter."
4. **Understand dependencies** — what other components, settings, env,
   assumptions does the broken code rely on?

### Phase 3 — Hypothesis and testing

**Scientific method:**

1. **Form a single hypothesis** — "I think X is the root cause because Y."
   Be specific. Write it down (in the investigation observation).
2. **Test minimally** — make the SMALLEST possible change to test the
   hypothesis. One variable at a time. Do NOT fix multiple things at
   once.
3. **Verify before continuing** — did it work? Yes → Phase 4. No → form
   a NEW hypothesis. Do NOT pile more fixes on top.
4. **When you do not know** — say "I don't understand X." Do not pretend.
   Ask the user. Research more.

### Phase 4 — Implementation (fix mode only)

**If this debugging session ends in "document only" mode, skip Phase 4.**

Fix the root cause, not the symptom.

1. **Create a failing test** — simplest reproduction possible. Automated
   if a framework exists, one-off script otherwise. MUST exist before
   fixing. (If `test-driven-development` skill is available, use it.)
2. **Implement a single fix** — address the root cause identified in
   Phase 1. ONE change. No "while I'm here" improvements. No bundled
   refactors.
3. **Verify the fix** — failing test now passes? Other tests still pass?
   Issue actually resolved?
4. **If the fix does NOT work** — STOP. Count attempts. If `< 3`, return
   to Phase 1 with the new information. **If `>= 3`, you have an
   architectural problem, not a hypothesis problem — see step 5.**
5. **If 3+ fixes failed: question architecture** — each fix revealing new
   shared state, new coupling, or new symptoms elsewhere is the signal
   that the pattern itself is wrong. STOP. Discuss with the user. Do
   NOT attempt fix #4.

**Save the fix outcome to engram.** See "Engram saves" below — fix
observation with root cause + fix + test.

## The two exits — bifurcation after Phase 1

After Phase 1 completes (root cause found, evidence gathered), decide:

| Mode | When | Subsequent phases |
|------|------|-------------------|
| **Fix mode** | The bug is in code the user owns and wants fixed now. | Continue to Phase 2 → 3 → 4. Apply the fix. Save a separate `fix` observation. |
| **Document mode** | The bug is in code the user does NOT own (external library, OS, another team's service), OR the user wants to record the finding without coding right now. | Stop after Phase 1. The investigation observation is the deliverable. |

You can also enter document mode after Phase 2 or 3 if the investigation
reveals the issue is out of scope or requires external action.

## Engram saves — required shape

### Investigation observation (always saved)

The investigation evolves through phases. Use `mem_save` with the same
`topic_key` to upsert as you learn more. Each upsert increments
`revision_count`.

```
mem_save({
  "title": "Debug: {short description of the issue}",
  "type": "discovery" | "decision",
  "topic_key": "debug/{tema}/investigation",
  "project": "{project}",
  "content": "<see content shape below>"
})
```

`{tema}` is kebab-case, derived from the issue (e.g., `debug/login-redirect-loop/investigation`).

**IMPORTANT — engram upserts in place. There is no historical content
preservation across revisions today.** If you need to keep an earlier
version of the investigation, copy it into the new content before
upserting.

After the upsert, emit the URL:
```
Review: http://localhost:7438/m/{id}
```

### Fix observation (fix mode only)

When Phase 4 completes successfully, save a SEPARATE observation:

```
mem_save({
  "title": "Fixed: {what was fixed}",
  "type": "bugfix",
  "topic_key": "debug/{tema}/fix",
  "project": "{project}",
  "content": "<see fix content shape below>"
})
```

Same `{tema}` as the investigation. Emit the review URL.

### Content shape — investigation

```markdown
## Issue
[1-2 sentences describing the observable symptom]

## Reproduction
[Exact steps or command(s) to trigger it]

## Evidence
[What you observed at each component boundary. Errors, log lines, state.]

## Recent changes
[git log / diff context — what changed that could explain this]

## Pattern analysis
[(Phase 2) Working examples found. Differences between working and broken.]

## Hypothesis
[(Phase 3) "I think X is the root cause because Y."]

## Status
[Open / Hypothesis-confirmed / Document-only / Fix-applied]
```

### Content shape — fix

```markdown
## Root cause
[1-2 sentences — what was actually wrong]

## Fix
[What you changed. Files. Lines. The single change.]

## Test
[The failing test you wrote, and how you verified it passes.]

## Related investigation
[Link to investigation: http://localhost:7438/m/{investigation_id}]

## Learned
[Gotchas, edge cases, anything that would help future debugging.]
```

## Engram conventions — embedded alma

This skill is autocontained — it does NOT require `engram-conventions`
to be installed. If `engram-conventions` is installed, defer to it for
fuller guidance.

### T1 (load-bearing) — search + topic_keys + URL emission + update vs save

- **`mem_search` BEFORE Phase 1**: mandatory. Use `type: bugfix` filter.
- **Topic_keys**: `debug/{tema}/investigation` (evolves via repeated
  `mem_save` to the same key) and `debug/{tema}/fix` (separate, only in
  fix mode).
- **URL emission**: `Review: http://localhost:7438/m/{id}` after each
  significant save. Respect `ENGRAM_UI_URL` if set.
- **`mem_save` upsert behavior**: same `topic_key` overwrites in place
  and increments `revision_count`. Engram does NOT preserve historical
  content snapshots today. If you want to keep prior investigation state
  visible, include it in the new content before upserting.

### T2 (quality) — type taxonomy + conflict handling

- **Type taxonomy**: `bugfix` for the fix observation; `discovery` for
  the investigation observation when the finding is non-obvious;
  `decision` for investigation observations that conclude with a chosen
  path (e.g., "we will not fix this — out of scope").
- **Conflict handling**: after `mem_save`, check the response envelope
  for `judgment_required`. If true, iterate `candidates[]` and call
  `mem_judge` per candidate using THAT candidate's `judgment_id` (not
  the top-level one). Heuristic:
  - confidence ≥ 0.7 AND relation is `related`, `compatible`, `scoped`,
    or `not_conflict` → call `mem_judge` silently.
  - Otherwise → ask the user in your next reply.

## Red flags — STOP and return to Phase 1

If you catch yourself thinking:
- "Quick fix for now, investigate later"
- "Just try changing X and see if it works"
- "Add multiple changes, run tests"
- "Skip the test, I'll manually verify"
- "It's probably X, let me fix that"
- "I don't fully understand but this might work"
- "Pattern says X but I'll adapt it differently"
- "Here are the main problems: [lists fixes without investigation]"
- Proposing solutions before tracing data flow
- **"One more fix attempt" (when already tried 2+)**
- **Each fix reveals a new problem in a different place**

**ALL of these mean: STOP. Return to Phase 1.**

If 3+ fixes have failed → question the architecture (Phase 4, step 5).

## User signals you are doing it wrong

Watch for these redirections from the user:
- "Is that not happening?" → you assumed without verifying.
- "Will it show us...?" → you should have added evidence gathering.
- "Stop guessing" → you are proposing fixes without understanding.
- "Ultrathink this" → question fundamentals, not just symptoms.
- "We're stuck?" (frustrated) → your approach is not working.

**When you see these:** STOP. Return to Phase 1.

## Common rationalizations

| Excuse | Reality |
|--------|---------|
| "Issue is simple, don't need process" | Simple issues have root causes too. Process is fast for simple bugs. |
| "Emergency, no time for process" | Systematic debugging is FASTER than guess-and-check thrashing. |
| "Just try this first, then investigate" | First fix sets the pattern. Do it right from the start. |
| "I'll write the test after confirming fix works" | Untested fixes don't stick. Test first proves it. |
| "Multiple fixes at once saves time" | Can't isolate what worked. Causes new bugs. |
| "Reference too long, I'll adapt the pattern" | Partial understanding guarantees bugs. Read it completely. |
| "I see the problem, let me fix it" | Seeing symptoms ≠ understanding root cause. |
| "One more fix attempt" (after 2+ failures) | 3+ failures = architectural problem. |
| "Skip the engram search, this is a new bug" | You don't know it's new until you searched. |
| "engram-conventions isn't installed so I can't save" | This skill's alma covers what you need. |

## Quick reference

| Phase | Key activities | Success criteria |
|-------|---------------|------------------|
| **Pre-flight** | `mem_search` with `type: bugfix` | Past work surfaced or confirmed absent |
| **1. Root cause** | Read errors, reproduce, check changes, gather boundary evidence, trace backward | Understand WHAT and WHY |
| **2. Pattern** | Find working examples, compare, list differences | Identify the deltas |
| **3. Hypothesis** | Form one theory, test minimally, verify | Confirmed or new hypothesis |
| **4. Implementation** | (fix mode) Create failing test, single fix, verify | Bug resolved, tests pass |

## Compatibility

Loads via the Agent Skills spec at:
- **Claude Code**: `~/.claude/skills/debug/SKILL.md`
- **OpenCode**: `~/.config/opencode/skills/debug/SKILL.md` (or
  `~/.claude/skills/debug/` via cross-tool path)

Install via `engram-ui` (TUI installer or `engram-ui setup debug`).
