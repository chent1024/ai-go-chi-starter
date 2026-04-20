package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	dbsqlc "ai-go-chi-starter/internal/infra/store/postgres/sqlc"
	rttrace "ai-go-chi-starter/internal/runtime/tracing"
	"ai-go-chi-starter/internal/service/example"
	"ai-go-chi-starter/internal/service/shared"
	"log/slog"
)

type ExampleRepository struct {
	store  *dbsqlc.Queries
	logger *slog.Logger
}

func NewExampleRepository(db *sql.DB) *ExampleRepository {
	repo := &ExampleRepository{}
	if db != nil {
		repo.store = dbsqlc.New(db)
	}
	return repo
}

func (r *ExampleRepository) WithLogger(logger *slog.Logger) *ExampleRepository {
	r.logger = logger
	return r
}

func (r *ExampleRepository) Create(ctx context.Context, item example.Example) (_ example.Example, err error) {
	if r.store == nil {
		return example.Example{}, shared.ErrInternal("database is not configured")
	}
	spanCtx, span := rttrace.StartSpan(ctx, r.logger, "postgres.example.create")
	defer func() {
		span.End(err, "db.system", "postgres", "db.operation", "insert", "db.table", "examples")
	}()
	row, err := r.store.CreateExample(spanCtx, dbsqlc.CreateExampleParams{
		ID:   item.ID,
		Name: item.Name,
	})
	if err != nil {
		return example.Example{}, fmt.Errorf("create example: %w", err)
	}
	return toExample(row), nil
}

func (r *ExampleRepository) Get(ctx context.Context, id string) (_ example.Example, err error) {
	if r.store == nil {
		return example.Example{}, shared.ErrInternal("database is not configured")
	}
	spanCtx, span := rttrace.StartSpan(ctx, r.logger, "postgres.example.get")
	defer func() {
		span.End(err, "db.system", "postgres", "db.operation", "select_one", "db.table", "examples")
	}()
	row, err := r.store.GetExample(spanCtx, id)
	if err == nil {
		return toExample(row), nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return example.Example{}, example.ErrNotFound()
	}
	return example.Example{}, fmt.Errorf("get example: %w", err)
}

func (r *ExampleRepository) List(ctx context.Context) (_ []example.Example, err error) {
	if r.store == nil {
		return nil, shared.ErrInternal("database is not configured")
	}
	spanCtx, span := rttrace.StartSpan(ctx, r.logger, "postgres.example.list")
	defer func() {
		span.End(err, "db.system", "postgres", "db.operation", "select_many", "db.table", "examples")
	}()
	rows, err := r.store.ListExamples(spanCtx)
	if err != nil {
		return nil, fmt.Errorf("list examples: %w", err)
	}
	items := make([]example.Example, 0, len(rows))
	for _, row := range rows {
		items = append(items, toExample(row))
	}
	return items, nil
}

func toExample(row dbsqlc.Example) example.Example {
	return example.Example{
		ID:        row.ID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
