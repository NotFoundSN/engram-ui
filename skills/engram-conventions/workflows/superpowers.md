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

For design or plan work scoped to one repo, use the repo identifier as the leading slot:
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
