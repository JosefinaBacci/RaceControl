package migrations

import "embed"

// Files holds the SQL migrations applied by cmd/migrate.
//
//go:embed *.sql
var Files embed.FS
