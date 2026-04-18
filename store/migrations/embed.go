// Package migrations embeds the numbered SQL migration files so that the
// migrate binary carries them without a runtime file-path dependency.
package migrations

import "embed"

// FS holds all *.sql files from this directory, including both up and
// down migrations. golang-migrate's iofs source filters by convention
// (N_name.sql for up, N_name.down.sql for down).
//
//go:embed *.sql
var FS embed.FS
