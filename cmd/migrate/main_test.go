package main

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"testing/fstest"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestRunnerRunVersionUsesAdvisoryLock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT pg_try_advisory_lock($1)`)).
		WithArgs(migrationLockKey).
		WillReturnRows(sqlmock.NewRows([]string{"pg_try_advisory_lock"}).AddRow(true))
	mock.ExpectExec(regexp.QuoteMeta(`
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT MAX(version) FROM schema_migrations`)).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(sql.NullInt64{}))
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_unlock($1)`)).
		WithArgs(migrationLockKey).
		WillReturnResult(sqlmock.NewResult(0, 1))

	runner := Runner{db: db, migrationsFS: fstest.MapFS{}, migrationsDir: "migrations"}
	if err := runner.Run(context.Background(), "version"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunnerRunFailsWhenLockAlreadyHeld(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT pg_try_advisory_lock($1)`)).
		WithArgs(migrationLockKey).
		WillReturnRows(sqlmock.NewRows([]string{"pg_try_advisory_lock"}).AddRow(false))

	runner := Runner{db: db}
	if err := runner.Run(context.Background(), "version"); err == nil {
		t.Fatal("Run() error = nil, want advisory lock error")
	}
}

func TestMigrationNamesFromFS(t *testing.T) {
	filesystem := fstest.MapFS{
		"migrations/001_init.sql": {Data: []byte("select 1;")},
		"migrations/010_more.sql": {Data: []byte("select 2;")},
		"migrations/readme.md":    {Data: []byte("ignore")},
	}

	names, err := migrationNamesFromFS(filesystem, "migrations")
	if err != nil {
		t.Fatalf("migrationNamesFromFS() error = %v", err)
	}
	want := []string{"001_init.sql", "010_more.sql"}
	if len(names) != len(want) {
		t.Fatalf("migrationNamesFromFS() len = %d, want %d", len(names), len(want))
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("migrationNamesFromFS()[%d] = %q, want %q", i, names[i], want[i])
		}
	}
}
