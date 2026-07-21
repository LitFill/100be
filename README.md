# 100be

Monorepo for [LitFill](https://github.com/LitFill)'s backend engineering curriculum:
implement the same service contract across N languages.

## Structure

```
go/         — URL shortener HTTP server
frontend/   — Static HTML frontend (Koka → HTML DSL)
flake.nix   — Nix dev shell (Go, Koka, jq, jj)
```

## Current Backends

| Lang | Description |
|------|-------------|
| Go   | URL shortener: REST API, file-persisted map, `POST /shorten`, `GET /:id`, `GET /`, `DELETE /:id` |

## Build & Run

```sh
# start a backend server (e.g. Go)
cd go && go run .   # listens on :8080
```
 
```
# generate frontend
cd frontend && koka -e main.kk  # writes index.html
```

See per-directory READMEs for details.
