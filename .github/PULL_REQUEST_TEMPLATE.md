## Summary

<!-- What does this PR do and why? One or two sentences. -->

## Related issue

Closes LLMS-XXX

## Type of change

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] New provider adapter
- [ ] Breaking change (fix or feature that would cause existing behavior to change)
- [ ] Documentation only
- [ ] Methodology change (requires operator approval — see `GOVERNANCE.md`)

## Checklist

- [ ] I read `CONTRIBUTING.md`
- [ ] Commits are in imperative mood, no `Co-Authored-By` trailers
- [ ] `go test ./... -race` passes locally
- [ ] `golangci-lint run` passes locally
- [ ] `npm -C web run lint` and `npm -C web run build` pass locally
- [ ] New or changed behavior has tests
- [ ] Coverage for changed code ≥ the module's minimum (see `CONTRIBUTING.md`)
- [ ] `CHANGELOG.md` updated under `## [Unreleased]`
- [ ] No new `TODO` / `FIXME` in code (open a follow-up LLMS- issue instead)
- [ ] No API keys, secrets, or raw provider response bodies committed

## For provider-adapter PRs

- [ ] Uses a paid API key (documented in `costs.md`, not committed)
- [ ] Handles timeout, 4xx, 5xx, and provider-specific error codes
- [ ] Fixtures captured under `internal/probes/adapters/{provider}/testdata/`
- [ ] Quirks documented in `docs/known-quirks.md`
- [ ] Registered in `internal/probes/adapters/registry.go`

## Methodology impact

<!-- If this PR changes how any metric is computed, how an incident is
     detected, or how data is published, describe the change and link to
     the methodology PR. Otherwise write "None". -->

None

## Screenshots / terminal output

<!-- Include screenshots for UI changes, or curl output for API changes. -->
