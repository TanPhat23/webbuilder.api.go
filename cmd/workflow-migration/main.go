// Command workflow-migration reports how legacy workflows could migrate.
// WORKFLOW_MIGRATION_DATABASE_URL must be set explicitly; .env is not loaded.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"my-go-app/internal/workflowmigration"
)

func main() { os.Exit(run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr)) }
func run(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("workflow-migration", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dryRun := flags.Bool("dry-run", false, "read and report proposed changes (default)")
	apply := flags.Bool("apply", false, "apply a clean migration to a local database")
	offline := flags.Bool("offline", false, "confirm realtime and editor processes are stopped and their caches are clear")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || (*dryRun && *apply) || (*apply && !*offline) || (!*apply && *offline) {
		fmt.Fprintln(stderr, "--apply requires --offline; it cannot be combined with --dry-run")
		return 2
	}
	dsn := getenv("WORKFLOW_MIGRATION_DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(stderr, "WORKFLOW_MIGRATION_DATABASE_URL is required; .env is not loaded")
		return 2
	}
	if *apply {
		parsed, err := url.Parse(dsn)
		config, configErr := pgx.ParseConfig(dsn)
		local := func(host string) bool { return host == "localhost" || host == "127.0.0.1" || host == "::1" }
		validHosts := configErr == nil && local(config.Host)
		if validHosts {
			for _, fallback := range config.Fallbacks {
				if !local(fallback.Host) {
					validHosts = false
				}
			}
		}
		if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || !local(parsed.Hostname()) || !validHosts {
			fmt.Fprintln(stderr, "--apply currently supports only an explicitly local PostgreSQL URL")
			return 2
		}
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		fmt.Fprintln(stderr, "could not configure database connection")
		return 1
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var report workflowmigration.Plan
	if *apply {
		report, err = workflowmigration.Apply(ctx, db)
	} else {
		report, err = workflowmigration.DryRun(ctx, db)
	}
	if err != nil {
		fmt.Fprintln(stderr, "migration failed:", err)
		return 1
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(stderr, "could not write report:", err)
		return 1
	}
	if len(report.Issues) > 0 {
		return 3
	}
	return 0
}
