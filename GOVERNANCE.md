# Governance

llmstatus.io is a public-good monitoring project. This document describes how
decisions are made.

## Roles

### Maintainers
Merge pull requests, triage issues, cut releases, steward the methodology.

Current maintainers are listed in `.github/CODEOWNERS`.

### Contributors
Anyone who opens an issue, submits a pull request, or contributes documentation.
Regular contributors are invited to become maintainers after sustained,
high-quality participation.

### Operator
The person or entity operating the `llmstatus.io` production deployment and the
associated domains. The operator is responsible for:
- Running probe infrastructure and paying API costs
- Publishing methodology updates that carry legal weight
- Responding to provider concerns

The operator does not have unilateral authority over code merges; code
decisions follow the process below.

## Decision process

### Routine changes (bug fixes, documentation, test coverage, minor refactors)
- Open a pull request.
- One maintainer approval required to merge.
- Author may not self-approve.

### Non-routine changes (new features, architectural shifts, new dependencies, public-facing copy, new providers)
- Open a GitHub Issue with the `LLMS-` prefix first.
- At least two maintainer approvals required to merge.
- A minimum 48-hour comment window on the issue, so contributors in other
  timezones can weigh in.

### Methodology changes
- Any change that affects how a metric is computed, how an incident is
  detected, or how data is published.
- Require a pull request against `METHODOLOGY.md` in addition to code.
- Require explicit operator approval (the project's trust depends on the
  methodology being stable and traceable).
- Must be announced in `CHANGELOG.md` under `### Changed` or `### Deprecated`
  with the reason documented.

### Emergency fixes (security, data correctness, provider-contact incidents)
- Maintainer may merge with a single approval and the `emergency` label.
- Must be followed within 48 hours by a retrospective pull request that
  documents what happened and updates tests to prevent regression.

## Releases

- Semantic Versioning (`MAJOR.MINOR.PATCH`).
- Tags are cut by a maintainer from `main`.
- CI publishes container images on tag creation.
- Every release must have a corresponding `CHANGELOG.md` section.

## Conflict resolution

If maintainers disagree and consensus cannot be reached after reasonable
discussion (typically one week):
1. The operator's vote breaks ties on matters affecting production or
   public-facing data.
2. On matters purely internal to the code (architecture, refactors), the
   matter is deferred and the status quo remains until a proposal can reach
   consensus.

## Removing a maintainer

Maintainers may be removed by unanimous vote of the remaining maintainers,
in cases of prolonged inactivity (≥ 6 months), breach of the Code of Conduct,
or breach of the methodology commitments in `METHODOLOGY.md`.

## Changing this document

Changes to `GOVERNANCE.md` require two maintainer approvals and a 7-day
public comment window on the pull request.
