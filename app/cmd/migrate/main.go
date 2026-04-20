package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strconv"
	"strings"

	migrations "ai-go-chi-starter/db"
	"ai-go-chi-starter/internal/config"
	"ai-go-chi-starter/internal/infra/store/postgres"
	rtlog "ai-go-chi-starter/internal/runtime/logging"
)

func main() {
	var action string

	flag.StringVar(&action, "action", "up", "migration action: up, version")
	flag.Parse()

	logger := rtlog.NewBootstrapLogger("migrate", os.Stderr)
	cfg, err := config.Load()
	if err != nil {
		logger.Error("migrate bootstrap failed", "kind", "fatal", "stage", "config", "err", err)
		os.Exit(1)
	}

	db, err := postgres.Open(context.Background(), postgres.Options{
		URL:             cfg.Database.URL,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
	})
	if err != nil {
		logger.Error("migrate startup failed", "kind", "fatal", "stage", "database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	runner := Runner{
		db:            db,
		migrationsFS:  migrations.FS,
		migrationsDir: "migrations",
	}
	if err := runner.Run(context.Background(), action); err != nil {
		logger.Error("migrate command failed", "kind", "fatal", "action", action, "err", err)
		os.Exit(1)
	}
}

type Runner struct {
	db            *sql.DB
	migrationsFS  fs.FS
	migrationsDir string
}

const migrationLockKey int64 = 2026041601

func (r Runner) Run(ctx context.Context, action string) error {
	if err := r.acquireLock(ctx); err != nil {
		return err
	}
	defer r.releaseLock(context.Background())

	if err := r.ensureTable(ctx); err != nil {
		return err
	}
	switch action {
	case "up":
		return r.up(ctx)
	case "version":
		version, err := r.version(ctx)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(os.Stdout, "%d\n", version)
		return nil
	default:
		return fmt.Errorf("unsupported action %q", action)
	}
}

func (r Runner) acquireLock(ctx context.Context) error {
	var acquired bool
	if err := r.db.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, migrationLockKey).Scan(&acquired); err != nil {
		return fmt.Errorf("acquire migration advisory lock: %w", err)
	}
	if !acquired {
		return fmt.Errorf("migration advisory lock is already held")
	}
	return nil
}

func (r Runner) releaseLock(ctx context.Context) {
	if _, err := r.db.ExecContext(ctx, `SELECT pg_advisory_unlock($1)`, migrationLockKey); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "release migration advisory lock: %v\n", err)
	}
}

func (r Runner) ensureTable(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`)
	return err
}

func (r Runner) up(ctx context.Context) error {
	names, err := migrationNamesFromFS(r.filesystem(), r.migrationsDir)
	if err != nil {
		return err
	}
	for _, name := range names {
		version, err := parseVersion(name)
		if err != nil {
			return err
		}
		applied, err := r.isApplied(ctx, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := r.applyFile(ctx, name, version); err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) applyFile(ctx context.Context, name string, version int64) error {
	content, err := fs.ReadFile(r.filesystem(), r.migrationPath(name))
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
		version,
		name,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration %s: %w", name, err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stdout, "applied migration %s\n", name)
	return nil
}

func (r Runner) version(ctx context.Context) (int64, error) {
	var version sql.NullInt64
	err := r.db.QueryRowContext(ctx, `SELECT MAX(version) FROM schema_migrations`).Scan(&version)
	if err != nil {
		return 0, err
	}
	if !version.Valid {
		return 0, nil
	}
	return version.Int64, nil
}

func (r Runner) isApplied(ctx context.Context, version int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`,
		version,
	).Scan(&exists)
	return exists, err
}

func migrationNames(migrationsDir string) ([]string, error) {
	return migrationNamesFromFS(os.DirFS("."), migrationsDir)
}

func migrationNamesFromFS(filesystem fs.FS, migrationsDir string) ([]string, error) {
	entries, err := fs.ReadDir(filesystem, migrationsDir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func (r Runner) filesystem() fs.FS {
	if r.migrationsFS != nil {
		return r.migrationsFS
	}
	return os.DirFS(".")
}

func (r Runner) migrationPath(name string) string {
	if strings.TrimSpace(r.migrationsDir) == "" {
		return name
	}
	return strings.TrimPrefix(r.migrationsDir+"/"+name, "/")
}

func parseVersion(name string) (int64, error) {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid migration filename %q", name)
	}
	value, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse migration version %q: %w", name, err)
	}
	return value, nil
}
