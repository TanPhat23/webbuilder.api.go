package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"my-go-app/internal/models"
)

func TestPostgresCustomElementNamesAreUniquePerOwner(t *testing.T) {
	dsn := os.Getenv("UUID_GORM_TEST_DSN")
	if dsn == "" {
		t.Skip("requires dedicated migrated PostgreSQL database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	ctx := context.Background()
	repo := NewCustomElementRepository(db)
	users := [2]string{uuid.NewString(), uuid.NewString()}
	for _, userID := range users {
		if err := db.Exec(`INSERT INTO "User" ("Id", "Email", "CreatedAt", "UpdatedAt") VALUES (?, ?, NOW(), NOW())`, userID, userID+"@example.invalid").Error; err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := db.Exec(`DELETE FROM "User" WHERE "Id" = ?`, userID).Error; err != nil {
				t.Error(err)
			}
		})
	}
	create := func(userID, name string) error {
		_, err := repo.CreateCustomElement(ctx, &models.CustomElement{
			Id: uuid.NewString(), UserId: userID, Name: name, Structure: json.RawMessage(`{}`), IsPublic: true,
		})
		return err
	}
	if err := create(users[0], "Card"); err != nil {
		t.Fatal(err)
	}
	if err := create(users[1], "Card"); err != nil {
		t.Fatalf("another user could not use the same name: %v", err)
	}
	if err := create(users[0], "card"); !errors.Is(err, ErrCustomElementAlreadyExists) {
		t.Errorf("case-insensitive duplicate create: %v", err)
	}
	onlyPublic := true
	items, err := repo.GetCustomElements(ctx, users[0], &onlyPublic)
	if err != nil || len(items) != 1 || items[0].UserId != users[0] {
		t.Fatalf("public filter must stay scoped to the caller: %+v (%v)", items, err)
	}
	if err := repo.DeleteCustomElement(ctx, items[0].Id, users[1]); !errors.Is(err, ErrCustomElementUnauthorized) {
		t.Fatalf("another user deleted the snapshot: %v", err)
	}
	if err := repo.DeleteCustomElement(ctx, items[0].Id, users[0]); err != nil {
		t.Fatal(err)
	}
	if err := create(users[0], "Card"); err != nil {
		t.Fatalf("name could not be reused after deletion: %v", err)
	}
}
