# Repository Guidelines

## Project Structure & Module Organization
- Go-first layout. Place executables under `cmd/<app>/` and shared packages in `internal/` or `pkg/`. Keep feature-local helpers inside the feature folder rather than global utils.
- Keep tests alongside code as `_test.go` files. Example: `internal/foo/foo.go` with `internal/foo/foo_test.go`.
- Store lightweight docs (like this file) in the repo root. Add specs/design notes under `docs/` if they grow beyond a single page.

## Build, Test, and Development Commands
- `go test ./...` — run the full test suite; fails on lint-level compile errors too.
- `go run ./cmd/<app>` — run a command locally; prefer `go run ./cmd/server` for services.
- `go fmt ./...` — format code before sending a PR.
- `go vet ./...` — static checks for common mistakes; run before reviews.
- Use modules pinned in `go.mod`; prefer standard library and only de facto dependencies per `SPEC.md`.

## Coding Style & Naming Conventions
- Follow default `gofmt` output (tabs for indentation). Avoid custom formatting.
- Exported identifiers should be clear and Go-idiomatic (`NewThing`, `Do`); unexported helpers should stay scoped and avoid abbreviations unless common (`ctx`, `err`).
- Keep functions small; return errors with context (`fmt.Errorf("load config: %w", err)`).
- Prefer composition over inheritance; pass interfaces at call sites when mocking is needed in tests.

## Testing Guidelines
- Use Go’s standard testing package; keep table-driven tests for pure functions.
- Name tests `Test<Thing>` and example tests `Example<Thing>` for documentation-style coverage.
- Aim for covering edge cases where inputs are empty, nil, or malformed. Include concurrency tests when goroutines are used.
- For services, add short integration tests that can run with in-memory fakes; avoid external dependencies unless guarded by build tags.

## Commit & Pull Request Guidelines
- Commits: write imperative, concise messages (e.g., `add user lookup`, `fix retry backoff`). Group related changes; avoid mixing refactors and behavior changes without notes.
- PRs: include a short description of behavior changes, testing performed (`go test ./...`), and any follow-up tasks. Link issues when available.
- Screenshots or logs are welcome when updating observable behavior or diagnostics.

## Security & Configuration Tips
- Keep secrets out of the repo; load via environment variables and document defaults in README/ENV sample files.
- Validate all external inputs; wrap network and filesystem calls with clear timeouts and context cancellation.
- If introducing dependencies, prefer vetted libraries and justify them briefly in the PR description.
