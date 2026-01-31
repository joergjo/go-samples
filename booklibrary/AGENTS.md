# Project overview
- Go port of Addy Osmani's Book Library sample; exposes a REST API on 0.0.0.0:8000 under `/api/books`.
- Uses Go `log/slog`, exports Prometheus metrics, and supports MongoDB or DocumentDB backends via Docker Compose profiles.
- Task runner (`task`) drives formatting, builds, tests, and container workflows; API-only (no client bundled).

# Build and test commands
- `task` / `task default`: go fmt + go mod tidy, build, then run tests.
- `task go:build`: compile binary with version metadata embedded via `-ldflags`.
- `task go:run`: run the API locally with debug logging; reads overrides from `.env` if present.
- `task go:test`: run `go test -v -count 2 -shuffle on ./...`.
- `task go:tidy`: run `go fmt ./...` and `go mod tidy -v`.
- `task docker:build` / `task docker:up` / `task docker:down`: build and orchestrate app + MongoDB containers; `task docdb:up` / `down` for DocumentDB local, `task mongo:up` / `down` / `sh` for MongoDB-only workflows.

# Code style guidelines
- Use iditomatic Go.
- Format and tidy with `task go:tidy` before committing; keep `go.mod`/`go.sum` clean.
- Prefer the shared logger setup in `internal/log` and structured `slog` fields for observability.
- Keep handlers in `internal/webapi` small; push business logic into `internal/model` and `internal/mongo` services.

# Testing instructions
- Primary test entrypoints: `task go:test` or `go test -v -count 2 -shuffle on ./...`.
- Tests rely on bundled fixtures (e.g., `testdata/books.json`) and should not require running databases by default.

# Security considerations
- Keep secrets and credentials out of the repo; use `.env` or environment variables for local overrides and Docker Compose settings.
- API currently exposes unauthenticated endpoints and metrics.
- Review container images and pinned tags; update base images and dependencies regularly.
