package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"ai-go-chi-starter/internal/service/example"
	"ai-go-chi-starter/internal/service/shared"
)

type ExampleRepository struct {
	db *sql.DB
}

func NewExampleRepository(db *sql.DB) *ExampleRepository {
	return &ExampleRepository{db: db}
}

func (r *ExampleRepository) Create(ctx context.Context, item example.Example) (example.Example, error) {
	if r.db == nil {
		return example.Example{}, shared.NewError(shared.CodeInternal, "database is not configured", http.StatusInternalServerError)
	}
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO examples (id, name) VALUES ($1, $2)
		 RETURNING id, name, created_at, updated_at`,
		item.ID,
		item.Name,
	)
	return scanExample(row)
}

func (r *ExampleRepository) Get(ctx context.Context, id string) (example.Example, error) {
	if r.db == nil {
		return example.Example{}, shared.NewError(shared.CodeInternal, "database is not configured", http.StatusInternalServerError)
	}
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, created_at, updated_at
		 FROM examples
		 WHERE id = $1`,
		id,
	)
	item, err := scanExample(row)
	if err == nil {
		return item, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return example.Example{}, shared.NewError("EXAMPLE_NOT_FOUND", "example not found", http.StatusNotFound)
	}
	return example.Example{}, fmt.Errorf("get example: %w", err)
}

func (r *ExampleRepository) List(ctx context.Context) ([]example.Example, error) {
	if r.db == nil {
		return nil, shared.NewError(shared.CodeInternal, "database is not configured", http.StatusInternalServerError)
	}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, name, created_at, updated_at
		 FROM examples
		 ORDER BY created_at DESC, id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list examples: %w", err)
	}
	defer rows.Close()

	items := make([]example.Example, 0)
	for rows.Next() {
		var item example.Example
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan example: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate examples: %w", err)
	}
	return items, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanExample(source scanner) (example.Example, error) {
	var item example.Example
	err := source.Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}
