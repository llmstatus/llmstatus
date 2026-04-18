# Database migrations

All schema changes for the relational store (PostgreSQL) live in this
directory as numbered SQL files.

## Convention

- Files are named `NNNN_short_snake_case_description.sql` (up migration)
  and `NNNN_short_snake_case_description.down.sql` (reversal).
- `NNNN` is a zero-padded 4-digit sequential integer, starting at `0001`.
- Every up migration must have a matching down migration. A down that
  cannot fully reverse (for example, because the up migration destroys
  data) must still be present and must be explicit about what it does or
  does not recover.
- Migrations are applied in lexicographical order by the `migrate`
  runner (`cmd/migrate`).

## Rules

- **Never `DROP TABLE` or perform other destructive changes** without a
  backup and explicit operator approval. See `CONTRIBUTING.md`.
- **Never edit a migration that has already been applied in any
  environment**; create a new one instead.
- **Never commit runtime data changes.** Runtime configuration
  (provider channels, model mappings, quotas) is managed separately and
  must not be altered by migrations.
- Migrations run inside a single transaction unless they use operations
  that Postgres cannot perform transactionally (for example,
  `CREATE INDEX CONCURRENTLY`). Document the reason in a comment at the
  top of any non-transactional migration.

## Time-series data

Time-series data (probe samples and meter samples) lives in InfluxDB,
not Postgres. InfluxDB schema is managed from code at startup, not
through SQL migrations.
