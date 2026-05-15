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
