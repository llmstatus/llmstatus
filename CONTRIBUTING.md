# Contributing to llmstatus.io

Thank you for considering a contribution. llmstatus.io is a public good,
and we welcome improvements from the community.

Before you start, please read this document completely.

---

## Code of Conduct

Participation is governed by our [Code of Conduct](CODE_OF_CONDUCT.md).
By participating, you agree to uphold it.

## How decisions are made

See [GOVERNANCE.md](GOVERNANCE.md) for the approval process, maintainer
responsibilities, and release cadence.

---

## What We Accept

- **New provider adapters** — if you have a paid account with a provider we
  do not yet monitor.
- **Bug fixes** — in probe logic, error classification, data processing.
- **Methodology improvements** — proposals to make measurement more accurate.
- **Documentation** — clarifications, typo fixes.
- **Performance improvements** — without sacrificing correctness.
- **Test coverage** — we can always use more tests.
- **Translations** — of i18n files under `web/i18n/` (the English source
  documentation stays English-only).

## What We Decline (usually)

- **New features we have not discussed** — please open an issue first.
- **Architectural changes** — same; discuss first.
- **Dependency additions** — we keep the dependency tree small deliberately.
- **UI redesigns** — the visual system is deliberate and specified in
  `docs/brand/BRAND_SYSTEM.md`.

---

## Before You Submit a Pull Request

1. **Open an issue first** for anything non-trivial. Use the `LLMS-` prefix
   in the issue title. A brief discussion saves both of us time.

2. **Read these documents**:
   - [`METHODOLOGY.md`](METHODOLOGY.md) — the measurement contract we make
     with users.
   - [`GOVERNANCE.md`](GOVERNANCE.md) — who approves what.
   - [`docs/brand/BRAND_SYSTEM.md`](docs/brand/BRAND_SYSTEM.md) — if your
     contribution touches anything visual or user-facing copy.

3. **Run the full local checks** before pushing:

   ```bash
   # Go backend
   go vet ./...
   golangci-lint run
   go test ./... -race -short

   # Next.js frontend
   cd web
   npm run lint
   npm run typecheck
   npm run build
   ```

4. **Write tests** for new code. Coverage thresholds are enforced in CI
   (see `.github/workflows/ci.yml`).

---

## Adding a New Provider Adapter

This is the most common contribution. Follow this process:

1. **Read the provider's API documentation end to end** before writing code.
2. **Use a paid account**, not a free trial. Free-tier probes produce
   unreliable data; we will not merge adapters tested only on free tiers.
3. **Prototype first**: write a throwaway script under `scripts/` that
   makes one real call and dumps the response shape.
4. **Create the adapter** at `internal/probes/adapters/{provider_id}.go`,
   implementing the `Provider` interface defined in
   `internal/probes/adapters/base.go`.
5. **Classify errors via the taxonomy** in [`METHODOLOGY.md`](METHODOLOGY.md)
   §5.4. Do not invent new error types without discussion.
6. **Capture fixtures**: 5–10 real responses (success + at least 3 error
   types) under `internal/probes/adapters/{provider_id}/testdata/`.
7. **Write table-driven tests** against those fixtures.
8. **Register the adapter** in `internal/probes/adapters/registry.go`.
9. **Add a database migration** under `store/migrations/` for the new
   `providers` and `models` rows.
10. **Manually verify** with `go run ./scripts/test_provider {provider_id}`
    against a real key.
11. **Document quirks** in [`docs/known-quirks.md`](docs/known-quirks.md) —
    every provider has some.
12. **Open a pull request** using the provider-addition template.

Essential rules for adapters:

1. **Never commit API keys.** Use `.env` and document required variables
   in `.env.example`.
2. **Handle timeouts, 4xx, 5xx, and provider-specific error codes.**
3. **Never store raw provider response bodies** in the database. Only
   metadata needed to classify the response.

---

## Code Style

### Go

- Go 1.26+ (see `go.mod`).
- Run `gofmt` and `golangci-lint run` before every commit.
- Absolute imports via the module path; no relative imports.
- Public functions require doc comments.
- No function longer than 50 lines; no file longer than 500 lines;
  cyclomatic complexity ≤ 10 (enforced by `gocyclo` in CI).
- Use `context.Context` as the first parameter of all I/O functions.

### TypeScript / Next.js

- Next.js 16+ App Router, TypeScript 5.x.
- Styling is Tailwind with the CSS variables defined in
  `docs/brand/BRAND_SYSTEM.md` §4.
- Server components by default; client components only when interactivity
  requires them.
- No external UI libraries. We write our own primitives.
- No animation libraries in the first major version — CSS only.

### Database

- Schema changes go through numbered SQL files in `store/migrations/`.
- Never `DROP TABLE` or perform destructive changes without explicit
  operator approval.
- Time-series data (probe samples) lives in InfluxDB; relational state
  (providers, incidents, subscriptions) lives in Postgres.

---

## Commit Messages

- Short, imperative mood: `add moonshot adapter` — not `added` or `adds`.
- One logical change per commit.
- Reference the issue: `fix LLMS-42: handle anthropic 529`.
- Keep the subject under 60 characters.
- Do **not** include `Co-Authored-By:` trailers (including for AI-assistant
  commits) unless a human co-author actually co-wrote the change.

---

## Pull Request Process

1. Fork and create a branch from `main`:
   ```
   git checkout -b feature/add-moonshot-adapter
   ```
2. Make your changes.
3. Run the full local checks (above).
4. Update [`CHANGELOG.md`](CHANGELOG.md) under `## [Unreleased]`.
5. Open a pull request using the template. Reference the `LLMS-` issue.
6. A maintainer will review. We aim to respond within 7 days.
7. Be prepared to iterate — reviewer requests are normal.

---

## Reporting Bugs

Open an issue at https://github.com/llmstatus/llmstatus/issues using the bug
template.

For **security vulnerabilities**, do **not** open a public issue. See
[SECURITY.md](SECURITY.md).

## Reporting Incorrect Data

If you believe llmstatus.io is reporting incorrect data about a provider,
email `methodology@llmstatus.io` with:

- The provider in question.
- The specific metric or claim you dispute.
- Your alternative measurement or evidence.
- How we can reproduce your finding.

We will respond within 48 hours and publish a correction if we are wrong.
Data correctness is why this project exists.

---

## License

By contributing, you agree that your contributions will be licensed under
the [Apache License 2.0](LICENSE). Methodology and data files are
additionally made available under [CC BY 4.0](LICENSE-DATA).

---

## Questions?

- Technical discussion: open a GitHub issue or discussion.
- General inquiry: `contact@llmstatus.io`
- Methodology: `methodology@llmstatus.io`
- Security: `security@llmstatus.io`
- Conduct: `conduct@llmstatus.io`
