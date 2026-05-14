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
