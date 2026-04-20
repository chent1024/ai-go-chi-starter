package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	rttrace "ai-go-chi-starter/internal/runtime/tracing"
	"ai-go-chi-starter/internal/service/example"
	"ai-go-chi-starter/internal/service/shared"
	"log/slog"
)

type ExampleRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewExampleRepository(db *sql.DB) *ExampleRepository {
	return &ExampleRepository{db: db}
}

func (r *ExampleRepository) WithLogger(logger *slog.Logger) *ExampleRepository {
	r.logger = logger
	return r
}

func (r *ExampleRepository) Create(ctx context.Context, item example.Example) (_ example.Example, err error) {
	if r.db == nil {
		return example.Example{}, shared.ErrInternal("database is not configured")
	}
	spanCtx, span := rttrace.StartSpan(ctx, r.logger, "postgres.example.create")
	defer func() {
		span.End(err, "db.system", "postgres", "db.operation", "insert", "db.table", "examples")
	}()
	row := r.db.QueryRowContext(
		spanCtx,
		`INSERT INTO examples (id, name) VALUES ($1, $2)
		 RETURNING id, name, created_at, updated_at`,
		item.ID,
		item.Name,
	)
	return scanExample(row)
}

func (r *ExampleRepository) Get(ctx context.Context, id string) (_ example.Example, err error) {
	if r.db == nil {
		return example.Example{}, shared.ErrInternal("database is not configured")
	}
	spanCtx, span := rttrace.StartSpan(ctx, r.logger, "postgres.example.get")
	defer func() {
		span.End(err, "db.system", "postgres", "db.operation", "select_one", "db.table", "examples")
	}()
	row := r.db.QueryRowContext(
		spanCtx,
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
		return example.Example{}, example.ErrNotFound()
	}
	return example.Example{}, fmt.Errorf("get example: %w", err)
}

func (r *ExampleRepository) List(ctx context.Context) (_ []example.Example, err error) {
	if r.db == nil {
		return nil, shared.ErrInternal("database is not configured")
	}
	spanCtx, span := rttrace.StartSpan(ctx, r.logger, "postgres.example.list")
	defer func() {
		span.End(err, "db.system", "postgres", "db.operation", "select_many", "db.table", "examples")
	}()
	rows, err := r.db.QueryContext(
		spanCtx,
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
