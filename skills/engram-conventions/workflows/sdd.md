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
