# Go — URL Shortener

HTTP server that maps long URLs to short codes and persists mappings to disk.

## Quickstart

```sh
go run .          # listens on :8080
# or
go build -o be100 && ./be100
```

## API

| Method | Path       | Body                       | Response             |
|--------|------------|----------------------------|----------------------|
| GET    | /          | —                          | `[{short, long}]`    |
| POST   | /shorten   | `original-url=<encoded>`   | `{short, long}`      |
| GET    | /:id       | —                          | `{short, long}` / 404|
| DELETE | /:id       | —                          | 204 / 404            |

All endpoints set CORS headers (`Access-Control-Allow-Origin: *`).

## Architecture

- **Short code**: first 5 base-52 chars of the original URL after a deterministic
  mapping — no external dependencies.
- **Storage**: `SafeUrlMap` — `sync.RWMutex`-guarded `map[string]string`,
  persisted to `urlmap.txt` (tab-delimited `short => long`).
- **Persistence**: flushed on every `Set` / `Delete` via atomic write.
- **Data file**: loaded on startup from the working directory; created empty
  if absent.

## Test

```sh
go test -v .
```

Covers: short-code generation, `SafeUrlMap` CRUD, persistence round-trip, HTTP
handler integration (unit + `httptest`).

## Binary

Built as `be100` (see `go.mod`: `module be100`).
