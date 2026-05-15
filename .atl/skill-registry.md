# Skill Registry — engram-ui

Last updated: 2026-05-14
Scan source: project `skills/` + user `~/.claude/skills/` (referenced via available-skills system list)

## Project Skills (loaded from `skills/`)

| Name | Path | Triggers |
|------|------|----------|
| engram-conventions | `skills/engram-conventions/SKILL.md` | `mem_save`, `mem_search`, `mem_context`, `mem_session_summary`, `mem_judge`, `mem_update`, "save to engram", "engram memory", "observation save", "topic_key" |

## User-Level Skills (relevant subset)

### Code & Testing
| Name | Triggers |
|------|----------|
| go-testing | Go tests, Bubbletea TUI testing, `teatest`, adding test coverage |
| skill-creator | Creating new AI agent skills following Agent Skills spec |
| skill-registry | "update skills", "skill registry", after installing/removing skills |
| simplify | Review changed code for reuse, quality, and efficiency |

### Review & PRs
| Name | Triggers |
|------|----------|
| review | Review a pull request |
| security-review | Complete security review of pending changes on current branch |
| judgment-day | "judgment day", "review adversarial", "dual review", "doble review", "juzgar" |
| branch-pr | Creating a PR, opening a PR, preparing changes for review |
| issue-creation | Creating a GitHub issue, reporting a bug, requesting a feature |
| commit-commands:commit | Create a git commit |
| commit-commands:commit-push-pr | Commit, push, and open a PR |
| commit-commands:clean_gone | Clean up `[gone]` branches and worktrees |

### SDD Pipeline
| Name | Role |
|------|------|
| sdd-init | Initialize SDD context |
| sdd-explore | Investigate ideas, compare approaches |
| sdd-propose | Create change proposal (intent, scope, approach) |
| sdd-spec | Write specs with requirements + Given/When/Then scenarios |
| sdd-design | Technical design with architecture decisions |
| sdd-tasks | Break down into implementation task checklist |
| sdd-apply | Implement tasks following specs |
| sdd-verify | Validate implementation against specs |
| sdd-archive | Sync specs and close the change |
| sdd-onboard | Guided full-cycle walkthrough |
| sdd-new / sdd-ff / sdd-continue | Meta orchestrator commands |

### Mode & Output
| Name | Triggers |
|------|----------|
| caveman:caveman | "caveman mode", "talk like caveman", "less tokens" |
| caveman:caveman-commit | "/commit", "generate commit", staging changes |
| caveman:caveman-review | "/review", "code review", "review the diff" |
| engram:memory | Always active — persistent memory protocol |

## Compact Rules (auto-injected into sub-agent prompts)

### engram-conventions
- Always namespace `topic_key` as `<namespace>/<feature>[/<phase>]`. Use leading repo-slot for multi-repo (e.g., `engram-ui/ui-refinement/scope-v1`).
- Stick to 14 canonical types when possible: `decision`, `architecture`, `bugfix`, `pattern`, `discovery`, `config`, `preference`, `feature`, `incident`, `note`, `session_summary`, `prompt`, `tool_use`, `command`.
- Same evolving topic → reuse `topic_key` (upsert). New topic → distinct key.
- Session close → MUST call `mem_session_summary` before declaring done.

### go-testing (for Go code in this project)
- Test runner: `go test ./...`
- Coverage: `go test -cover ./...`
- Use `net/http/httptest` for handler/router tests.
- Do not test generated `*_templ.go` directly — test via rendered HTML output.
- Follow Go table-driven test pattern where it adds value.

### Conventional commits (engram-ui repo convention)
- Format: `type(scope): subject` — types: `feat`, `fix`, `refactor`, `chore`, `docs`, `style`, `test`.
- No `Co-Authored-By` lines.
- Subject ≤ 72 chars.

## Project Conventions

- **Stack**: Go 1.25 + templ + chi + Tailwind (CDN) + goldmark/bluemonday for markdown rendering.
- **Layout**: `cmd/engram-ui/` (binary), `internal/{client,server,views,render}/`.
- **Views**: `*.templ` + generated `*_templ.go` (via `templ generate`).
- **Routing**: chi with middleware (RequestID, Recoverer, Logger).
- **Backend dependency**: engram HTTP API at `localhost:7437`. Health endpoint `/health`, observations at `/observations/recent` + `/observations/{id}`, search at `/search?q=`.
- **Make targets**: `generate`, `build`, `run`, `dev` (templ watch + proxy), `tidy`, `clean`.
