package postgres

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"ai-go-chi-starter/internal/service/example"
	"ai-go-chi-starter/internal/service/shared"
)

func TestExampleRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO examples (id, name) VALUES ($1, $2)
		 RETURNING id, name, created_at, updated_at`,
	)).
		WithArgs("exm_01", "demo").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow("exm_01", "demo", now, now))

	item, err := NewExampleRepository(db).Create(context.Background(), example.Example{ID: "exm_01", Name: "demo"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.ID != "exm_01" {
		t.Fatalf("Create() id = %q", item.ID)
	}
}

func TestExampleRepositoryGetNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, name, created_at, updated_at
		 FROM examples
		 WHERE id = $1`,
	)).
		WithArgs("exm_missing").
		WillReturnError(sql.ErrNoRows)

	_, err = NewExampleRepository(db).Get(context.Background(), "exm_missing")
	if shared.Code(err) != "EXAMPLE_NOT_FOUND" || shared.HTTPStatus(err) != http.StatusNotFound {
		t.Fatalf("unexpected error = %v", err)
	}
}

func TestExampleRepositoryList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, name, created_at, updated_at
		 FROM examples
		 ORDER BY created_at DESC, id DESC`,
	)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow("exm_02", "two", now, now).
			AddRow("exm_01", "one", now, now))

	items, err := NewExampleRepository(db).List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("List() len = %d", len(items))
	}
}

func TestExampleRepositoryGetWrapsUnexpectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	wantErr := errors.New("boom")
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, name, created_at, updated_at
		 FROM examples
		 WHERE id = $1`,
	)).
		WithArgs("exm_01").
		WillReturnError(wantErr)

	_, err = NewExampleRepository(db).Get(context.Background(), "exm_01")
	if err == nil || shared.Code(err) != "" {
		t.Fatalf("unexpected error = %v", err)
	}
}
