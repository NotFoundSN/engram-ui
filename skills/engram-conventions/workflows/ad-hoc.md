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

For saves scoped to one repo, use the repo identifier as the leading slot:
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
