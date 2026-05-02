# Changelog

All notable changes to llmstatus.io are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Every pull request that changes user-visible behavior, data semantics, or
public APIs must add an entry under `## [Unreleased]`.

## [Unreleased]

### Fixed
- Add Redis 7 service to docker-compose; wire api healthcheck dependency — `REDIS_URL` was set in the Ansible env template but no Redis container existed, causing OTP rate limiting to silently fail on first use

### Changed (UI overhaul)
- Lighten canvas color tokens: base `#0D1117`, raised `#161D28`, overlay `#1C2638` — increases card-to-background contrast and makes the dark theme feel less cave-dark
- Raise ink-400 (`#7B8899`) and ink-500 (`#4D5B6C`) for readable secondary text; ink-600 borders now `#2D3A4A`
- Eliminate all sub-12px text across the UI: replace `text-[10px]`/`text-[11px]` with `text-xs` (12px) across all components and pages
- Lighten hero subtitle from `ink-400` to `ink-300` for better contrast
- Scale up hero h1 to `text-4xl sm:text-5xl` for stronger visual impact
- Fix undefined `--ink-700` CSS variable reference in `UserReportHistogram` (no-data bar was transparent); use `--ink-600`
- Bring histogram axis labels up from `text-[9px]` to `text-xs`

### Changed (LLMS-070 cont.)
- Eliminate `dupl` violations: extract `openAICompatProvider` + `newOpenAICompatProvider` (compat_provider.go) reducing 13 identical provider files from ~67 to ~28 lines; extract `runLightProbe` (probe_exec.go) to deduplicate the `ProbeLightInference` skeleton in Anthropic, Cohere, Gemini, Mistral; convert DeepSeek to pure-compat adapter
- Fix goimports separator in `internal/api/reports.go`

### Changed (LLMS-070)
- Refactor 7 production functions exceeding cyclomatic complexity 10: `handleCreateSubscription`, `handleUpdateSubscription`, `handleOAuthUpsert`, `AllModelSparklines`, `ProviderRegionStats`, `cmd/migrate main`, `cmd/api main`; extract helpers `fetchOwnedSub`, `mergeSubFields`, `resolveEmailPrefs`, `classifySubCreateError`, `decodeOAuthBody`, `indexSparklinesFromRows`, `foldRegionRows`, `newMigrator`, `closeMigrator`, `runDown`, `runForce`, `buildAuthConfig`, `runServer`
- All extracted helpers are ≤ 7 complexity; no behaviour changes

### Fixed
- Add missing `internal/api/websocket.go` (WebSocket Hub/Client types) omitted from prior commits, causing all PR CI runs to fail with undefined symbols
- Fix `otprl.RedisLimiter.Allow`: use `SetArgs` method instead of incorrect `Set` call for NX+TTL semantics
- Fix `FixtureProvider` missing `ProbeScope: "global"` field, causing all integration tests to fail with `providers_probe_scope_check` constraint violation
- Fix all `gofmt`/`goimports` formatting issues across 20+ Go files
- Add `statusOngoing`, `statusOperational`, `statusDown`, `statusDegraded`, `severityMajor` constants in `providers.go` to satisfy `goconst`
- Fix `errcheck` violations in `websocket.go`/`websocket_test.go`: handle `conn.Close` and deadline error returns
- Fix `errcheck` violations in `auth/oauth.go`, `email/client.go`, `notifier/webhook.go`: `defer resp.Body.Close()` annotated
- Add `//nolint:gosec` for G706 log-injection false positives (structured logging, not string formatting) and G115 int32 conversion
- Add `//nolint:unparam` for test helper functions where fixed-value params are intentional
- Remove always-nil `error` return from `handleSubscription` in `realtime.go`
- Add package-level and exported-symbol doc comments to satisfy `revive` across 8 packages
- Fix TypeScript `no-explicit-any` errors in `lib/types.ts`, `lib/optimistic-updates.ts`, `lib/performance.ts`, `lib/swr-config.tsx`, `__tests__/lib/websocket.test.ts`, `tests/e2e/real-time-updates.spec.ts`
- Fix `jest.setup.js` to use ES module import instead of `require()`
- Fix `react/no-unescaped-entities` in `app/methodology/page.tsx`
- Add `eslint-disable` for intentional `setState` in `ProbeTimestamp.tsx` effect

### Added (LLMS-045)
- Fixed host port assignments for all dev services (avoid conflicts on this host): db→15432, influx→18086, ingest→18080, api→18081, Next.js dev→13000
- `web/.env.local.example` added — copy to `web/.env.local` before running `npm run dev`
- Docker `web` service moved to `web-docker` profile; default dev workflow uses `npm run dev` outside Docker
- `/about` page — who we are, why we built it, independence statement, contact
- `/methodology` page — probe types, 7 global locations, all 4 detection rules with triggers, data retention, known limitations
- Footer now links to /about and /methodology on every page

### Added (LLMS-044)
- Homepage OG image (`/opengraph-image`) — edge-rendered brand card: site name, headline, tagline, and "20+ AI providers tracked" badge; no live data fetch (avoids stale-status problem on social shares)

### Added (LLMS-043)
- Incident detail pages now have edge-rendered OG images (1200×630): incident title, severity color, provider ID, and status pill on dark background; graceful fallback when API unavailable

### Added (LLMS-042)
- Provider table columns are now sortable: click Name, Uptime 24h, P95, or Status to cycle asc → desc → default
- Default sort: worst status first (down → degraded → operational), then alphabetical within each group
- Active sort column shows ▲/▼ indicator; unsorted columns show a muted ↕

### Added (LLMS-041)
- `/incidents` page now has filter chips: status (All / Ongoing / Resolved) and a provider dropdown; filters combine with AND logic
- Result count shown; empty-state message when no incidents match
- Fetch limit raised from 50 → 200 so client-side filtering has full dataset; server page remains static with 30s revalidate

### Added (LLMS-040)
- Homepage now shows a "Recent Incidents" section (up to 5, ongoing/monitoring first) above the provider table; hidden when no incidents exist
- Incidents are fetched in parallel with providers via `Promise.allSettled` — no added latency, each fails independently
- "All incidents →" link to /incidents included in section header

### Added (LLMS-039)
- `/compare` page — side-by-side provider comparison: current status, 24h uptime %, p95 latency, active incident count, category, region, and 30-day uptime sparklines for both providers
- `CompareSelector` client component — two dropdowns that push URL params on change, so the Server Component page re-fetches data for the selected pair
- "Compare" nav link added to SiteHeader
- `/compare` added to sitemap

### Added (LLMS-038)
- Provider detail pages now emit Schema.org `Service` JSON-LD for SEO (includes provider name, service type, status-page URL when present)
- Incident detail pages now emit Schema.org `Event` JSON-LD (startDate, endDate if resolved, eventStatus)
- All JSON-LD uses the React 19 native `<script>` children pattern — React 19 auto-escapes `</script>` sequences, no manual sanitization needed
- `providers/[id]` OG image (`opengraph-image.tsx`) — edge-rendered 1200×630 PNG showing provider name and live status pill on dark background; auto-served by Next.js convention

### Added (LLMS-037)
- `/api` documentation page — fully static, lists all public endpoints (status, providers, history, incidents, badges, RSS) with curl examples, response envelope format, rate limit headers, and error format
- "API" link added to SiteHeader nav
- `/api` added to sitemap

### Added (LLMS-036)
- `/providers` page — dedicated provider list page with filter chips for status (all / operational / degraded / down) and category (all / official / aggregator / CN official); filters combine with AND logic
- `ProvidersClient` client component — owns filter state, renders result count and filtered `ProviderTable`
- SiteHeader "Providers" nav link now points to `/providers` (previously pointed to `/`)
- `/providers` added to sitemap with `changeFrequency: "always"`, priority 0.95

### Added (LLMS-035)
- `GET /badge/{id}.svg?style=detailed` — badge variant that appends live 24h uptime percentage when `LiveStatsReader` is wired (e.g. `operational · 99.9%`); falls back to simple format when live stats are unavailable
- `/badges` page now shows both Simple and Detailed preview images side-by-side, with embed snippets for each style
- 3 new badge tests: `DetailedStyle_WithUptime`, `DetailedStyle_NoLiveStats_FallsBack`, `UnknownStyle_FallsBack`

### Added (LLMS-034)
- RSS discovery: `<link rel="alternate" type="application/rss+xml">` added to root `layout.tsx` (global feed) and to provider detail `generateMetadata` (per-provider feed) — satisfies ROADMAP §8.3 requirement
- `web/app/api/feed/route.ts` — Next.js proxy for global feed (`/api/feed` → Go API `/feed.xml`)
- `web/app/api/feed/[id]/route.ts` — Next.js proxy for per-provider feed (`/api/feed/{id}` → Go API `/v1/providers/{id}/feed.xml`)
- RSS link added to `SiteFooter` and to provider detail page subtitle row
- Provider detail page now shows "RSS" link alongside "Status page ↗"

### Added (LLMS-033)
- `/badges` page — live badge preview + Markdown / HTML / URL embed snippets for every monitored provider; linked from site nav
- `CopyButton` client component — clipboard copy with 1.8s "Copied" confirmation
- `web/app/api/badge/[id]/route.ts` — Next.js proxy route that forwards badge requests to the Go API (`/badge/{id}.svg`); lets the static badges page render `<img>` tags without exposing the internal `API_URL` to the browser
- `/badges` added to sitemap with `changeFrequency: "monthly"`

### Fixed (LLMS-032)
- `docker-compose.yml` rewritten for Go + Next.js stack; removes all Python/uvicorn references which no longer exist
- `deploy/docker/Dockerfile.golang` — single parameterized multi-stage Go build (`--build-arg CMD=<api|ingest|detector|prober|migrate>`); runtime image is `alpine:3.21` with only `ca-certificates` + `tzdata`
- `deploy/docker/Dockerfile.web` — multi-stage Next.js build using `output: "standalone"`; copies only `.next/standalone` for minimal image
- `web/next.config.ts` — adds `output: "standalone"` required by `Dockerfile.web`
- `.env.example` — documents all required and optional env vars with dev-safe defaults

### Added (LLMS-031)
- `GET /v1/status` — aggregate system status endpoint: worst-case status across all active providers (`"operational"` / `"degraded"` / `"down"`) plus breakdown counts; wrapped in standard envelope
- `Pinger` interface + `WithPinger` functional option on `Server` — when wired, `/healthz` pings the DB and returns 503 if unreachable; without it, behaviour is unchanged (always 200)
- `cmd/api/main.go` wires `WithPinger(pool)` so production health checks reflect actual DB state
- 7 new tests: `GetStatus` (all-operational, one-down, one-degraded, empty) + `Healthz` (no pinger, pinger OK, pinger fail)

### Added (LLMS-030)
- `web/app/error.tsx` — client-component error boundary: branded dark page with amber "Error" label, digest-aware message, and secondary "Try again" button that calls Next.js `reset()`
- `web/app/loading.tsx` — server-component loading skeleton: `animate-pulse` blocks mirroring the hero + table layout; shown during client-side navigation suspense

### Added (LLMS-029)
- `web/app/not-found.tsx` — branded 404 page (amber "404" label, dark canvas, "Back to status" link) replacing Next.js default white page
- `web/app/robots.ts` — serves `/robots.txt`: allow all crawlers, points to sitemap; respects `NEXT_PUBLIC_SITE_URL` env var
- `web/app/sitemap.ts` — serves `/sitemap.xml`: static routes (`/`, `/incidents`) + dynamic routes for every active provider and recent incident; revalidates hourly; falls back gracefully when API is unreachable

### Added (LLMS-028)
- Homepage hero section: tagline + subhead verbatim from BRAND_SYSTEM.md §7.1; optional faint grid background per §6.5
- `layout.tsx` base metadata: corrected site name to `llmstatus.io`, added `openGraph.siteName`, `twitter.card`
- `page.tsx` exports named `metadata` with OG title + description
- `providers/[id]/page.tsx` `generateMetadata` now includes `openGraph` with provider-specific title + description
- `incidents/[slug]/page.tsx` `generateMetadata` now includes `openGraph` with incident title + description
- `incidents/page.tsx` now includes `openGraph` metadata
- `ProviderTable` table headers updated to all-caps 11px with 0.08em tracking per brand spec §6.2

### Fixed (LLMS-027)
- `cmd/api/main.go` now passes `historyReader` to both `WithHistoryReader` and `WithLiveStatsReader`; previously `uptime_24h`/`p95_ms` would never appear in production despite the pipeline being complete
- Added `TestListProviders_WithLiveStats` and `TestListProviders_LiveStatsNil_OmitsFields` to confirm field presence/absence

### Fixed (LLMS-020)
- Corrected middleware stack order to `accessLog → requestID → cors → handler` so that X-Request-ID is present on every response, including CORS preflight (OPTIONS) 204 responses that previously short-circuited before `requestIDMiddleware` ran
- Added `TestCORS_Preflight` assertion: `X-Request-ID` header must be non-empty on preflight responses

### Added (LLMS-026)
- `ProviderLiveStat` type + `LiveStatsReader` interface in `internal/store/influx` — one SQL query returns 24h uptime + p95 latency for all providers
- `influxHistoryReader.postQuery` private helper extracted from `ProviderHistory` — eliminates HTTP boilerplate duplication across both query methods
- `WithLiveStatsReader` functional option on API `Server` — optional; summaries still return without uptime/p95 fields when not wired
- `/v1/providers` response now includes `uptime_24h` (0–1) and `p95_ms` fields when live stats available, omitted otherwise
- `ProviderTable` now shows "Uptime 24h" and "p95" columns per brand spec §6.2; values formatted as `99.9%` / `450ms` / `1.2s`; `—` when data unavailable

### Fixed (LLMS-025)
- `internal/detector` coverage raised from 79.7% → 89.3% (floor: 85%)
- Added tests: `Run` context-cancel, `runOnce` 5m read error, `ensureIncident` store error + create error, `resolveStale` list error + resolve error, `incidentTitle` all rule branches
- `fakeIncidentStore` extended with `getErr`/`createErr`/`listErr`/`resolveErr` fields for targeted error injection

### Added (LLMS-024)
- `SiteHeader` server component: logo link + nav (`Providers` / `Incidents`) with active-page highlighting via `NavLink` client component (`usePathname`)
- `SiteFooter` server component: standard footer copy, rendered once in root `layout.tsx`
- `NavLink` client component: highlights the active route using `usePathname()`

### Changed (LLMS-024)
- Root `layout.tsx` now renders `SiteHeader` and `SiteFooter` — all 4 pages stripped of duplicated header/footer markup
- `ProviderTable` rows now have `hover:bg-[var(--canvas-overlay)]` per brand spec §6.2

### Added (LLMS-023)
- `HistoryBucket.P95Ms` field — InfluxDB SQL now selects `approx_percentile_cont(0.95)` for successful probes alongside existing uptime data
- `LatencyBar` component — 30-day per-day p95 bar chart, color-coded by threshold (≤500ms green, ≤2000ms amber, >2000ms red), gray stubs for days with no probe data; shows median p95 summary
- Provider detail page: `LatencyBar` section rendered below `UptimeSparkline` when latency data is available

### Added (LLMS-022)
- `ProbeTimestamp` client component — relative "X ago" display, auto-refreshes every 10 s (brand spec §6.4); used on `IncidentCard` and incident detail page header
- `IncidentCard` now shows `ProbeTimestamp` instead of static `formatDate` for started time

### Changed (LLMS-022)
- `StatusPill` updated to match brand spec §6.1: removed background pill (`rounded-full` + bg color), now dot-only with all-caps 11px text at 0.05em tracking

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

### Added (LLMS-020)
- Production middleware stack: **CORS** (`Access-Control-Allow-Origin: *`, OPTIONS preflight → 204), **request ID** (`X-Request-ID` — propagates incoming or generates UUID v4), and **structured access logging** (slog, skips `/healthz`)
- Middleware applied to all routes including rate-limited responses; ordering: `accessLog → cors → requestID → [rateLimiter] → mux`

### Added (LLMS-019)
- Detection rule 6.3 — **latency degradation** (`latency_degradation`, severity `minor`): fires when p95 `duration_ms` over the last 5 min exceeds 3× the p95 over the past 24 h
- Detection rule 6.4 — **regional outage** (`regional_outage`, severity `minor`): fires when one `region_id` has >50% error rate over 5 min while the provider is not globally down
- `LatencyStats` and `RegionalStats` types added to the detector package
- `ProbeReader` interface extended: `LatencyByProvider`, `RegionalErrorRateByProvider`
- InfluxDB 3 queries for both new methods use `approx_percentile_cont` and `region_id` grouping
- `Runner.runOnce` now evaluates all four rules; latency/regional fetch failures are non-fatal (logged as `WARN`)
- **Known limitation**: rule 6.3 baseline is the last 24 h (V1 simplification); METHODOLOGY.md specifies same-hour 7-day median (tracked in REVIEW_QUEUE.md)

### Added (LLMS-017)
- `GET /feed.xml` — global RSS 2.0 feed of all incidents (last 50, all providers)
- `GET /v1/providers/{id}/feed.xml` — per-provider RSS 2.0 feed; returns HTTP 404 for unknown provider IDs
- Feed items include provider name, severity, status, start time, and a permalink GUID
- `Cache-Control: max-age=60`; `X-Forwarded-Proto` honoured for absolute link URLs

### Added (LLMS-016)
- `GET /badge/{provider_id}.svg` — shields.io-compatible flat SVG badge showing provider status
- Colors: operational → green (`#4CAF50`), degraded → amber (`#FF9800`), down → red (`#F44336`), unknown → gray (`#9E9E9E`)
- Unknown provider IDs return a gray "unknown" badge (not a JSON error) for badge-consumer friendliness
- XSS-safe: provider names are HTML-escaped before embedding in SVG

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
