//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ai-go-chi-starter/internal/config"
	"ai-go-chi-starter/internal/service/example"
)

func TestExampleRepositoryIntegrationCRUD(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if strings.TrimSpace(cfg.Database.URL) == "" {
		t.Skip("APP_DATABASE_URL is not configured")
	}

	ctx := context.Background()
	db, err := Open(ctx, cfg.Database)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `TRUNCATE TABLE examples`)
		_ = db.Close()
	})

	if err := applyExampleMigration(ctx, db); err != nil {
		t.Fatalf("applyExampleMigration() error = %v", err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE TABLE examples`); err != nil {
		t.Fatalf("truncate examples: %v", err)
	}

	repo := NewExampleRepository(db)
	suffix := time.Now().UnixNano()

	first, err := repo.Create(ctx, example.Example{
		ID:   fmt.Sprintf("exm_it_%d_a", suffix),
		Name: "first",
	})
	if err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	second, err := repo.Create(ctx, example.Example{
		ID:   fmt.Sprintf("exm_it_%d_b", suffix),
		Name: "second",
	})
	if err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}

	got, err := repo.Get(ctx, second.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != second.ID || got.Name != second.Name {
		t.Fatalf("Get() = %+v, want %+v", got, second)
	}

	items, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("List() len = %d, want 2", len(items))
	}
	if items[0].ID != second.ID || items[1].ID != first.ID {
		t.Fatalf("List() order = [%s %s], want [%s %s]", items[0].ID, items[1].ID, second.ID, first.ID)
	}
}

func applyExampleMigration(ctx context.Context, db interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}) error {
	content, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "db", "migrations", "001_init.sql"))
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, string(content))
	return err
}
