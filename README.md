# engram-ui

Web viewer for [engram](https://github.com/Gentleman-Programming/engram) — a persistent memory system for AI coding agents.

Reads engram's data through its public HTTP REST API. Decoupled from engram's internal schema.

## Installation

### Homebrew (macOS / Linux)

```bash
brew install gentleman-programming/tap/engram-ui
```

### Go install

```bash
go install github.com/Gentleman-Programming/engram-ui/cmd/engram-ui@latest
```

### Direct download

Download the binary for your platform from the [Releases page](https://github.com/Gentleman-Programming/engram-ui/releases):

| Platform | Archive |
|----------|---------|
| macOS amd64 | `engram-ui_{version}_darwin_amd64.tar.gz` |
| macOS arm64 (Apple Silicon) | `engram-ui_{version}_darwin_arm64.tar.gz` |
| Linux amd64 | `engram-ui_{version}_linux_amd64.tar.gz` |
| Linux arm64 | `engram-ui_{version}_linux_arm64.tar.gz` |
| Windows amd64 | `engram-ui_{version}_windows_amd64.zip` |
| Windows arm64 | `engram-ui_{version}_windows_arm64.zip` |

Extract and place the binary somewhere on your `PATH`.

### Platform notes

**macOS — Gatekeeper warning**: the first time you run engram-ui, macOS may block it with an "unidentified developer" warning. Right-click the binary → Open → Open anyway. Alternatively, use `go install` or Homebrew which handles signing automatically.

**Windows — AV false positives**: some antivirus tools flag unsigned Go binaries. If blocked, use `go install` (builds locally so the binary is trusted). The source is fully open.

**Linux — non-systemd distros**: `engram-ui setup os-autostart` writes a systemd user unit. The `systemctl --user enable` invocation will fail on Alpine, Void, or other non-systemd distros, but the unit file is still written (failure is non-fatal). Enable using your init system's mechanism manually.

---

## Usage

```
engram-ui [serve] [flags]    start the web UI (default)
engram-ui setup <target>     install skills or configure OS autostart
engram-ui version            print version and exit
engram-ui help               print help
```

### serve flags

```
--engram=<url>     engram REST API base URL (default: http://localhost:7437)
--listen=<addr>    address engram-ui listens on   (default: :7438)
--no-spawn         fail instead of auto-spawning 'engram serve'
```

### setup subcommands

| Target | What it does | Destination |
|--------|-------------|-------------|
| `claude-code` | Copies the engram-conventions skill into Claude Code's skills directory | `~/.claude/skills/engram-conventions/` |
| `opencode` | Copies the engram-conventions skill into OpenCode's config directory | `~/.config/opencode/skills/engram-conventions/` |
| `os-autostart` | Registers engram-ui as an OS-level autostart entry | Windows: `%APPDATA%\...\Startup\engram-ui.bat`<br>macOS: `~/Library/LaunchAgents/com.gentleman-programming.engram-ui.plist`<br>Linux: `~/.config/systemd/user/engram-ui.service` |
| `remove-autostart` | Removes the OS autostart entry | Same paths as above |

All `setup` subcommands are idempotent — safe to run more than once.

---

## GitHub Actions secrets

The release workflow (`.github/workflows/release.yml`) requires two secrets configured on the repository:

- `GITHUB_TOKEN` — automatically provided by GitHub Actions (no setup needed).
- `HOMEBREW_TAP_TOKEN` — a Personal Access Token with `repo` scope on the `Gentleman-Programming/homebrew-tap` repository. Create at GitHub → Settings → Developer settings → Personal access tokens.

---

## Requirements (development)

- Go 1.25+
- [`engram`](https://github.com/Gentleman-Programming/engram) on `PATH`
- [`templ`](https://templ.guide) for template generation:

  ```bash
  go install github.com/a-h/templ/cmd/templ@latest
  ```

## Run from source

```bash
templ generate         # compile .templ → _templ.go
go run ./cmd/engram-ui
```

Open http://localhost:7438

## How it talks to engram

On startup, `engram-ui serve`:

1. Pings the `--listen` address `/healthz` — exits cleanly if already running.
2. Pings `--engram` health endpoint.
3. If engram is reachable → connects.
4. If not → spawns `engram serve` as a child process and waits for health.
5. On shutdown, terminates the spawned process (only if we spawned it).

## Layout

```
cmd/engram-ui/           # binary entrypoint (thin — delegates to internal/cli)
internal/
  cli/                   # Dispatch + subcommand handlers
  installer/             # skill copy + OS autostart
    skills/
      engram-conventions/ # skill payload (embedded into binary at build time)
  client/                # HTTP client → engram REST
  server/                # chi router + handlers
  views/                 # templ templates
```

> **Note for contributors**: the engram-conventions skill source lives at
> `internal/installer/skills/engram-conventions/` (moved from the repo root in v3).
> Edit skill content there — it is embedded into the binary at build time via `//go:embed`.

## License

MIT
