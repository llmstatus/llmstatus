# Changelog

All notable changes to llmstatus.io are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Every pull request that changes user-visible behavior, data semantics, or
public APIs must add an entry under `## [Unreleased]`.

## [Unreleased]

### Added
- Initial open-source repository scaffolding
- Apache-2.0 License for source code, CC BY 4.0 for methodology and data
- GitHub Actions CI pipeline (Go + Next.js)
- Issue templates (bug, feature, provider addition)
- Code of Conduct (Contributor Covenant v2.1)
- Governance document
- Go 1.26 module `github.com/llmstatus/llmstatus` with `cmd/` binaries
  for `api`, `ingest`, `detector`, `prober`, `migrate` (LLMS-001)
- `internal/probes` base types (`Provider` interface, `ProbeResult`,
  `ErrorClass` taxonomy) and an adapter registry
- `pkg/testutil` stubs for Postgres, InfluxDB, Redis, upstream mock,
  and fixtures (skipped with `t.Skip` until LLMS-002)
- `store/migrations/` with a placeholder `0001_init.sql` and a README
  documenting the numbering convention
- `deploy/ansible/` scaffolding (`inventories/{dev,staging,prod}`,
  `roles/`, `playbooks/`)
- `web/` — Next.js 16.2 App Router with TypeScript, Tailwind, ESLint
- `web/i18n/` directory reserved as the only path permitted to hold
  non-English content
- `scripts/check-worktrees.sh` implementing CLAUDE.md §5 worktree
  discipline
- `.golangci.yml` linter config and `.editorconfig`

### Changed
- CI Go jobs now scope to `./cmd/... ./internal/... ./pkg/...` so Go
  sources shipped inside `web/node_modules/` (for example by the
  `flatted` JS package) are not linted, vetted, or tested
- The coverage gate no-ops when no test files exist yet, instead of
  failing during the scaffold phase
- `gocyclo` CI step now ignores `_test.go` files; table-driven tests
  legitimately exceed the 10-complexity threshold

### Added (LLMS-002)
- `internal/httpclient/` — shared HTTP client with default 30s timeout,
  `User-Agent: llmstatus.io/<version>`, per-request `X-Request-ID`, and
  context-aware retry for idempotent methods only
- `internal/probes/adapters/openai.go` — first real adapter
  implementing `ProbeLightInference`; other probe methods return
  `ErrNotSupported` pending LLMS-003
- Recorded fixtures at `internal/probes/adapters/openai/testdata/`
  covering success, 401 auth, 429 rate-limit, 500 server error, and a
  malformed (HTML) body
- Live registration behind the `livekeys` build tag: adapter code
  compiles and is testable without a real API key, but the prober only
  fires real calls when built with `-tags livekeys` and
  `LLMS_OPENAI_API_KEY` + `LLMS_REGION_ID` are set
- `docs/known-quirks.md` — first entries for OpenAI (HTTP 200 + error
  envelope, variable 401 codes)

### Added (LLMS-008)
- `internal/probes/adapters/anthropic.go` — second real adapter implementing
  `ProbeLightInference` for `api.anthropic.com/v1/messages`; handles HTTP 529
  (Anthropic-specific "overloaded" status) as `server_5xx`; uses `x-api-key`
  and `anthropic-version: 2023-06-01` headers; registered via `init()`
- 5 fixtures in `internal/probes/adapters/anthropic/testdata/` covering
  success, 401, 429, 500, and 529; 7 test functions including timeout and
  context-cancel cases
- `docs/known-quirks.md` — Anthropic section expanded: `x-api-key` auth,
  required `anthropic-version` header, 529 overloaded status, and
  `input_tokens`/`output_tokens` naming convention

### Added (LLMS-003)
- `internal/probes/runner.go` — `Runner` schedules `ProbeLightInference` for
  every active (provider, model) pair at a configurable interval; bounded
  concurrency via semaphore + `sync.WaitGroup`; per-probe context deadline;
  functional options `WithInterval`, `WithProbeTimeout`, `WithConcurrency`
- `ResultSink` interface decouples the runner from ingest; `LogSink` is the
  default implementation (`slog.Info` per result)
- `cmd/prober/main.go` — production binary: loads active providers and models
  from Postgres via `pgstore`, matches each against the adapter registry,
  then drives a `Runner`; config via `REGION_ID`, `DATABASE_URL`,
  `PROBE_INTERVAL`, `PROBE_TIMEOUT`, `PROBE_CONCURRENCY` env vars
- 5 unit tests with `goleak.VerifyTestMain`, bounded-concurrency tracking via
  `atomic.Int64`, and a fake adapter + collect sink

### Added (LLMS-007)
- `store/migrations/embed.go` — `//go:embed *.sql` bakes all migration
  files into the binary; no runtime path dependency
- `cmd/migrate` — production-ready migration runner using
  `golang-migrate/v4` with `iofs` source and `pgx/v5` driver; commands:
  `up`, `down [N]`, `version`, `status`, `force N`; DB URL from
  `DATABASE_URL` env or `-db` flag; uses `pgx5://` scheme
- Dependencies: `github.com/golang-migrate/migrate/v4 v4.19.1`

### Added (LLMS-006)
- `pkg/testutil/postgres.go` — `NewPostgres(t)` spins up a Postgres 17
  container via testcontainers-go, applies all `store/migrations/*.sql`
  up files in order, and returns a ready `*pgxpool.Pool`; skips when
  `-short` is set; container cleaned up via `t.Cleanup`
- `pkg/testutil/fixtures.go` — `FixtureProvider` and `FixtureModel`
  helpers for integration tests (replaces stale stubs with real schema)
- `tests/integration/store/store_test.go` — 12 integration sub-tests
  covering full CRUD for providers, models, and incidents including
  upsert idempotency, deduplication query, and resolve flow
- `go.uber.org/goleak v1.3.0` — `TestMain` goroutine-leak guard
- `github.com/google/go-cmp` — deep equality in assertions
- `github.com/testcontainers/testcontainers-go v0.42.0` + postgres module
- CI: `integration` job now uses testcontainers (no pre-wired service
  containers); `go vet` extended to `./tests/...`

### Added (LLMS-005)
- `store/migrations/0002_schema.sql` — `providers`, `models`, and
  `incidents` tables in PostgreSQL 17 with CHECK constraints, FK
  references, and targeted indexes
- `store/migrations/0002_schema.down.sql` — teardown for the above
- `internal/store/postgres/queries.sql` — typed SQL queries (sqlc input)
  covering full CRUD for all three tables plus incident deduplication
- `sqlc.yaml` — sqlc v2 config (engine: postgresql, driver: pgx/v5);
  UUID fields map to `github.com/google/uuid.UUID`, JSONB to
  `encoding/json.RawMessage`
- `internal/store/postgres/gen/` — sqlc-generated Go package `pgstore`
  with typed structs, `Querier` interface, and query implementations
- `.golangci.yml` path exclusion for `internal/store/postgres/gen/`
  (generated code is not linted)
- Dependencies: `github.com/jackc/pgx/v5 v5.9.1`,
  `github.com/google/uuid v1.6.0`

### Added (auto-merge helpers)
- `scripts/merge-pr.sh` — approve + merge a PR end-to-end without
  manual intervention
- `scripts/merge-open-deps.sh` — sweep every open dependabot PR
