package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ai-go-chi-starter/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, errors.New("APP_DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	configurePool(db, cfg)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

func configurePool(db *sql.DB, cfg config.DatabaseConfig) {
	if db == nil {
		return
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
}

type ReadyChecker struct {
	DB *sql.DB
}

func (r ReadyChecker) Ready(ctx context.Context) error {
	if r.DB == nil {
		return errors.New("database is not configured")
	}
	return r.DB.PingContext(ctx)
}
