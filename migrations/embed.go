// Package migrations holds the versioned SQL migrations for the webbuilder
// Postgres schema. The Go API owns the schema; GORM is used for queries only
// (no AutoMigrate). Add a new file with `make migrate-create name=<name>`.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS

// PrismaHandoverMigration is the last migration Prisma applied before the
// schema moved to goose. `migrate baseline` requires it in _prisma_migrations.
const PrismaHandoverMigration = "20261111120000_add_element_order_key"

// BaselineVersion is the goose version of 00001_baseline.sql.
const BaselineVersion int64 = 1
