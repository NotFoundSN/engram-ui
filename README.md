# engram-ui

Web viewer for [engram](https://github.com/Gentleman-Programming/engram) — a persistent memory system for AI coding agents.

Reads engram's data through its public HTTP REST API. Decoupled from engram's internal schema.

## Status

Early scaffold. Boots, auto-spawns `engram serve` if needed, renders stats + recent observations.

## Requirements

- Go 1.25+
- [`engram`](https://github.com/Gentleman-Programming/engram) installed and on `PATH`
- [`templ`](https://templ.guide) for template generation:

  ```bash
  go install github.com/a-h/templ/cmd/templ@latest
  ```

## Run

```bash
templ generate         # compile .templ → _templ.go
go run ./cmd/engram-ui
```

Open http://localhost:7438

## How it talks to engram

On startup, `engram-ui`:

1. Pings `http://localhost:7437/health`
2. If engram is reachable → connects
3. If not → spawns `engram serve` as a child process and waits for health
4. On shutdown, terminates the spawned process (only if we spawned it)

Override with flags:

```bash
go run ./cmd/engram-ui \
  --engram=http://localhost:7437 \
  --listen=:7438 \
  --no-spawn        # fail instead of auto-spawning
```

## Layout

```
cmd/engram-ui/        # binary entrypoint
internal/
  client/             # HTTP client → engram REST
  server/             # chi router + handlers
  views/              # templ templates
```

## License

MIT
