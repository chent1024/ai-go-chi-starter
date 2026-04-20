package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Options struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func Open(ctx context.Context, options Options) (*sql.DB, error) {
	if strings.TrimSpace(options.URL) == "" {
		return nil, errors.New("APP_DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", options.URL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	configurePool(db, options)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

func configurePool(db *sql.DB, options Options) {
	if db == nil {
		return
	}
	db.SetMaxOpenConns(options.MaxOpenConns)
	db.SetMaxIdleConns(options.MaxIdleConns)
	db.SetConnMaxLifetime(options.ConnMaxLifetime)
	db.SetConnMaxIdleTime(options.ConnMaxIdleTime)
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
