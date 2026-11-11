// Command migrate applies the versioned SQL migrations in ./migrations.
//
//	go run ./cmd/migrate up          apply all pending migrations
//	go run ./cmd/migrate status      list applied / pending migrations
//	go run ./cmd/migrate down        roll back the latest migration
//	go run ./cmd/migrate baseline    adopt a database Prisma already migrated
//
// DATABASE_URL is read from the environment (or .env).
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"my-go-app/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down|status|baseline>")
		os.Exit(2)
	}
	_ = godotenv.Load(".env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	switch cmd := os.Args[1]; cmd {
	case "up":
		err = goose.UpContext(ctx, db, ".")
	case "down":
		err = goose.DownContext(ctx, db, ".")
	case "status":
		err = goose.StatusContext(ctx, db, ".")
	case "baseline":
		err = baseline(ctx, db)
	default:
		err = fmt.Errorf("unknown command %q", cmd)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// baseline marks 00001_baseline.sql as applied on a database whose schema
// was created by Prisma, so `up` only runs migrations written after the
// handover. It refuses unless Prisma's history shows the handover migration
// finished, which guards against adopting a database that is behind.
func baseline(ctx context.Context, db *sql.DB) error {
	var finished bool
	err := db.QueryRowContext(ctx,
		`SELECT finished_at IS NOT NULL AND rolled_back_at IS NULL
		   FROM "_prisma_migrations" WHERE migration_name = $1
		   ORDER BY started_at DESC LIMIT 1`,
		migrations.PrismaHandoverMigration,
	).Scan(&finished)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !finished) {
		return fmt.Errorf("prisma migration %s is not applied; run `prisma migrate deploy` from webbuilderv2 at its last Prisma commit first",
			migrations.PrismaHandoverMigration)
	}
	if err != nil {
		return fmt.Errorf("read _prisma_migrations (is this a Prisma-managed database?): %w", err)
	}

	current, err := goose.EnsureDBVersionContext(ctx, db)
	if err != nil {
		return err
	}
	if current >= migrations.BaselineVersion {
		log.Printf("already baselined (goose version %d)", current)
		return nil
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, true)`,
		migrations.BaselineVersion,
	); err != nil {
		return fmt.Errorf("record baseline: %w", err)
	}
	log.Printf("baselined at version %d; run `migrate up` for later migrations", migrations.BaselineVersion)
	return nil
}
