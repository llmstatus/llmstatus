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
