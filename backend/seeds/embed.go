package seeds

import "embed"

// Files holds the idempotent reference-data SQL applied by cmd/seed. Unlike
// migrations, seeds are not versioned in schema_migrations and can be re-run.
//
//go:embed *.sql
var Files embed.FS
