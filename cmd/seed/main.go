// Command seed loads demo marketplace data for local development.
//
// WARNING: it deletes ALL marketplace items, tags and categories first,
// exactly like the Prisma seed it replaces. Never point it at production.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"my-go-app/seed"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/lucsky/cuid"
)

type seedData struct {
	Categories []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"categories"`
	Tags             []string `json:"tags"`
	MarketplaceItems []struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Preview      string   `json:"preview"`
		TemplateType string   `json:"templateType"`
		Featured     bool     `json:"featured"`
		PageCount    int      `json:"pageCount"`
		Tags         []string `json:"tags"`
		Downloads    int      `json:"downloads"`
		Likes        int      `json:"likes"`
		Categories   []string `json:"categories"`
	} `json:"marketplaceItems"`
}

func main() {
	_ = godotenv.Load(".env")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	var data seedData
	if err := json.Unmarshal(seed.MarketplaceJSON, &data); err != nil {
		log.Fatalf("parse seed data: %v", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	if err := run(ctx, tx, &data); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	log.Printf("seeded %d categories, %d tags, %d marketplace items",
		len(data.Categories), len(data.Tags), len(data.MarketplaceItems))
}

func run(ctx context.Context, tx *sql.Tx, data *seedData) error {
	now := time.Now().UTC()

	// Use any existing user as author, else create a demo user.
	var userID, email string
	var first, last sql.NullString
	err := tx.QueryRowContext(ctx,
		`SELECT "Id", "Email", "FirstName", "LastName" FROM "User" WHERE "Email" LIKE '%@%' LIMIT 1`,
	).Scan(&userID, &email, &first, &last)
	if err == sql.ErrNoRows {
		userID, email = "seed-user-id", "demo@marketplace.com"
		first, last = sql.NullString{String: "Demo", Valid: true}, sql.NullString{String: "User", Valid: true}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO "User" ("Id","Email","FirstName","LastName","ImageUrl","CreatedAt","UpdatedAt")
			 VALUES ($1,$2,$3,$4,$5,$6,$6)`,
			userID, email, first.String, last.String,
			"https://api.dicebear.com/7.x/avataaars/svg?seed=Demo", now); err != nil {
			return fmt.Errorf("create demo user: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	authorName := strings.TrimSpace(first.String + " " + last.String)
	if authorName == "" {
		authorName = "Anonymous"
	}

	for _, table := range []string{"MarketplaceItemTag", "MarketplaceItemCategory", "MarketplaceItem", "Tag", "Category"} {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`DELETE FROM %q`, table)); err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}

	for _, c := range data.Categories {
		if _, err := tx.ExecContext(ctx, `INSERT INTO "Category" ("Id","Name") VALUES ($1,$2)`, c.ID, c.Name); err != nil {
			return fmt.Errorf("insert category %s: %w", c.ID, err)
		}
	}
	tagIDs := make(map[string]string, len(data.Tags))
	for _, name := range data.Tags {
		id := cuid.New()
		if _, err := tx.ExecContext(ctx, `INSERT INTO "Tag" ("Id","Name") VALUES ($1,$2)`, id, name); err != nil {
			return fmt.Errorf("insert tag %s: %w", name, err)
		}
		tagIDs[name] = id
	}

	for _, it := range data.MarketplaceItems {
		itemID := cuid.New()
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO "MarketplaceItem"
			   ("Id","Title","Description","Preview","TemplateType","Featured","PageCount",
			    "Downloads","Likes","AuthorId","AuthorName","Verified","CreatedAt","UpdatedAt")
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,true,$12,$12)`,
			itemID, it.Title, it.Description, it.Preview, it.TemplateType, it.Featured, it.PageCount,
			it.Downloads, it.Likes, userID, authorName, now); err != nil {
			return fmt.Errorf("insert item %q: %w", it.Title, err)
		}
		for _, tag := range it.Tags {
			if tagID, ok := tagIDs[tag]; ok {
				if _, err := tx.ExecContext(ctx,
					`INSERT INTO "MarketplaceItemTag" ("ItemId","TagId") VALUES ($1,$2)`, itemID, tagID); err != nil {
					return fmt.Errorf("link tag %s: %w", tag, err)
				}
			}
		}
		for _, cat := range it.Categories {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO "MarketplaceItemCategory" ("ItemId","CategoryId") VALUES ($1,$2)`, itemID, cat); err != nil {
				return fmt.Errorf("link category %s: %w", cat, err)
			}
		}
	}
	return nil
}
