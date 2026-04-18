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

### Added (LLMS-015)
- Per-IP fixed-window rate limiting on the public API (`internal/api/RateLimiter`)
- `WithRateLimiter` functional option on `api.New()`; default 60 req/min configurable via `API_RATE_LIMIT` env var
- Standard `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` response headers on every request
- `Retry-After` header and `429 Too Many Requests` response when limit is exceeded
- Client IP extracted from `X-Forwarded-For` (nginx first-entry) with fallback to TCP remote address

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

### Added (LLMS-014 + LLMS-018)
- `internal/store/influx/history_reader.go` — `HistoryReader` interface +
  `influxHistoryReader` implementation; queries InfluxDB 3 via
  `POST /api/v3/query_sql` using `date_bin()` for time-bucketed aggregation;
  sanitises `provider_id` to prevent SQL injection; 5 unit tests
- `internal/api/history.go` — `GET /v1/providers/{id}/history?window=24h|7d|30d`
  handler; `WithHistoryReader` functional option keeps `api.New(store)` backward-
  compatible; returns 503 when reader not configured; 4 handler tests
- `internal/api/server.go` — added optional `history HistoryReader` field;
  `New()` now accepts variadic `func(*Server)` options; new route registered
- `cmd/api/main.go` — wires `influx.NewHistoryReader` via `WithHistoryReader`;
  three new required env vars: `INFLUX_HOST`, `INFLUX_TOKEN`, `INFLUX_DATABASE`
- `web/lib/api.ts` — `HistoryBucket` type; `getProviderHistory(id, window)` with
  300s revalidate (history changes slowly)
- `web/components/UptimeSparkline.tsx` — 30-bar daily uptime chart; CSS-only,
  no animation library; bar color maps uptime: ≥99% → `--signal-ok`,
  ≥95% → `--signal-warn`, else → `--signal-down`; empty slots shown as
  `--ink-600` stubs; summary line shows mean uptime %
- `web/app/providers/[id]/page.tsx` — fetches 30d history alongside provider
  data (best-effort: sparkline hidden on fetch failure); renders `UptimeSparkline`
  in an "Uptime History" section above the model list

### Added (LLMS-013)
- `web/app/incidents/page.tsx` — incidents list (all statuses, limit 50); ISR 30s;
  graceful API-error fallback; empty-state message
- `web/app/incidents/[slug]/page.tsx` — permanent incident detail page; ISR 60s;
  `generateMetadata` with `{provider} incident on {date}: {title}` title format;
  JSON-LD `Event` schema (`safeJsonLd` escapes `<` → `\u003c` to prevent
  `</script>` injection from API data); breadcrumb navigation; timeline, description,
  affected models/regions sections; `notFound()` on 404
- `web/components/IncidentCard.tsx` — added optional `href` prop; when set the card
  renders as a `<Link>` with hover border transition; `formatDate` exported for reuse
- `web/lib/api.ts` — replaced `IncidentSummary` with fuller `IncidentDetail` (matches
  Go `incidentResponse` exactly); added `getIncident(slug)` function; `listIncidents`
  now accepts `status` and `limit` params
- `web/app/providers/[id]/page.tsx` — incident cards now link to `/incidents/{slug}`

### Added (LLMS-012)
- `web/app/providers/[id]/page.tsx` — provider detail page; server component,
  ISR 60s; `generateMetadata` produces `Is {Provider} API down?` title + data-driven
  description; calls `notFound()` on `ApiNotFoundError` (HTTP 404), re-throws
  other errors so Next.js retries ISR
- `web/components/IncidentCard.tsx` — card showing severity, title, and UTC
  started_at timestamp for active incidents
- `web/components/ModelList.tsx` — table of active monitored models with
  model ID (monospace) and type
- `web/lib/api.ts` — added `ApiNotFoundError` class, `ProviderDetail`,
  `IncidentRef`, `ModelSummary` types; `getProvider(id)` function; `apiFetch`
  now throws `ApiNotFoundError` on HTTP 404
- `web/components/ProviderTable.tsx` — provider name cells now link to
  `/providers/{id}`

### Added (LLMS-011)
- `web/app/globals.css` — replaced scaffold with brand system CSS variables
  (`--canvas-*`, `--ink-*`, `--signal-*`, `--viz-*`) from BRAND_SYSTEM.md §4;
  dark observatory theme, no light-mode media query
- `web/app/layout.tsx` — updated with site-wide metadata title template
  (`%s — LLM Status`) and description
- `web/lib/api.ts` — typed server-side API client; `listProviders()` and
  `listIncidents()` fetch from `API_URL` env var with Next.js `next.revalidate`
  cache; graceful null return on network failure
- `web/components/StatusPill.tsx` — operational/degraded/down pill using brand
  signal colors; server component
- `web/components/ProviderTable.tsx` — provider list table with alternating row
  backgrounds; server component
- `web/app/page.tsx` — homepage server component; fetches live provider status,
  renders summary banner and `ProviderTable`; `revalidate = 30`; degrades
  gracefully when the API is unreachable
- `next build` produces a static pre-render with 30 s ISR, zero TypeScript errors

### Added (LLMS-009)
- `internal/detector/` — event-detection subsystem with three layers:
  - `reader.go` — `ProbeReader` interface + `InfluxReader` that queries
    InfluxDB 3 via `POST /api/v3/query_sql`; no gRPC dependency
  - `rules.go` — `EvaluateRules` applies Rule 6.1 (≥50% error rate in 5m →
    `provider_down`, critical) and Rule 6.2 (>5% in 10m → `elevated_errors`,
    major); Rule 6.1 suppresses Rule 6.2 for the same provider;
    minimum 3 probes required before any rule fires
  - `runner.go` — `Runner` fetches stats, evaluates rules, deduplicates via
    `GetOngoingByProviderAndRule`, creates incidents, and auto-resolves
    stale auto-detected incidents; manual incidents are never auto-resolved
- `cmd/detector/main.go` — production binary with graceful shutdown via
  `signal.NotifyContext`; config via `DATABASE_URL`, `INFLUX_HOST`,
  `INFLUX_TOKEN`, `INFLUX_DATABASE`, `DETECTOR_INTERVAL` (default 60s)
- Incident slug format: `YYYY-MM-DD-{provider_id}-{rule-with-dashes}`
  (e.g. `2026-04-18-openai-provider-down`)
- 17 unit tests: 8 rule-evaluation cases, 3 InfluxReader cases (httptest),
  6 runner cases with `fakeReader` + `fakeIncidentStore`

### Added (LLMS-010)
- `internal/api/` — public read API using Go 1.22+ `net/http.ServeMux`
  with method+path patterns (no external router dependency)
- Endpoints: `GET /v1/providers`, `GET /v1/providers/{id}`,
  `GET /v1/incidents`, `GET /v1/incidents/{id}`, `GET /healthz`
- Provider status derived from the incidents table: critical ongoing →
  `down`, major/minor ongoing → `degraded`, no ongoing → `operational`
- `GET /v1/incidents/{id}` accepts both UUID and slug
- Response envelope: `{"data": ..., "meta": {"generated_at": "...", "cache_ttl_s": 30}}`
- `coalesceSlice` ensures all array fields serialize as `[]`, never JSON `null`
- `Store` interface (subset of `pgstore.Querier`) keeps handlers decoupled
  from the DB layer and fully testable with a fake store
- `cmd/api/main.go` — server with graceful 5s shutdown; config via
  `DATABASE_URL`, `API_ADDR`
- 14 unit tests covering list, filter, get-by-uuid/slug, not-found, 405

### Added (LLMS-004)
- `internal/store/influx/writer.go` — `Writer` interface + `lineWriter` implementation
  using the InfluxDB v2 line-protocol HTTP write endpoint (compatible with InfluxDB 3);
  no extra dependencies — uses `net/http` only
- `internal/ingest/handler.go` — `Handler` serving `POST /v1/probe`; validates
  `provider_id`, `model`, `probe_type`, `region_id`, `started_at`; writes via
  `influx.Writer`; returns 204/400/405/500
- `cmd/ingest/main.go` — HTTP server with graceful shutdown; config from
  `INFLUX_HOST`, `INFLUX_TOKEN`, `INFLUX_DATABASE`, `INGEST_ADDR`; exposes
  `/healthz` for load-balancer health checks
- `internal/probes/httpsink.go` — `HTTPSink` implements `ResultSink` by POSTing
  JSON to the ingest URL; `cmd/prober` uses it when `INGEST_URL` is set, falls
  back to `LogSink`
- 13 new unit tests across `internal/store/influx`, `internal/ingest`, and
  `internal/probes` (HTTPSink)

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
