package example

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"ai-go-chi-starter/internal/service/shared"
)

func TestServiceCreate(t *testing.T) {
	repo := &stubRepository{
		createFn: func(ctx context.Context, item Example) (Example, error) {
			_ = ctx
			now := time.Now().UTC()
			item.CreatedAt = now
			item.UpdatedAt = now
			return item, nil
		},
	}

	item, err := NewService(repo).Create(context.Background(), CreateInput{Name: " example "})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.ID == "" {
		t.Fatal("Create() id = empty")
	}
	if item.Name != "example" {
		t.Fatalf("Create() name = %q", item.Name)
	}
}

func TestServiceCreateRejectsEmptyName(t *testing.T) {
	_, err := NewService(&stubRepository{}).Create(context.Background(), CreateInput{Name: "  "})
	if shared.Code(err) != shared.CodeInvalidArgument || shared.HTTPStatus(err) != http.StatusBadRequest {
		t.Fatalf("unexpected error = %v", err)
	}
	details, ok := shared.Details(err).(shared.ValidationDetails)
	if !ok || len(details.FieldErrors) != 1 {
		t.Fatalf("unexpected details = %#v", shared.Details(err))
	}
	if details.FieldErrors[0].Field != "name" {
		t.Fatalf("field error = %+v", details.FieldErrors[0])
	}
}

func TestServiceGet(t *testing.T) {
	repo := &stubRepository{
		getFn: func(ctx context.Context, id string) (Example, error) {
			_ = ctx
			return Example{ID: id, Name: "found"}, nil
		},
	}

	item, err := NewService(repo).Get(context.Background(), "exm_123")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if item.ID != "exm_123" {
		t.Fatalf("Get() id = %q", item.ID)
	}
}

func TestServiceList(t *testing.T) {
	repo := &stubRepository{
		listFn: func(ctx context.Context) ([]Example, error) {
			_ = ctx
			return []Example{{ID: "exm_1"}, {ID: "exm_2"}}, nil
		},
	}

	items, err := NewService(repo).List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("List() len = %d", len(items))
	}
}

func TestServiceReturnsRepositoryError(t *testing.T) {
	wantErr := ErrNotFound()
	repo := &stubRepository{
		getFn: func(ctx context.Context, id string) (Example, error) {
			_ = ctx
			_ = id
			return Example{}, wantErr
		},
	}

	_, err := NewService(repo).Get(context.Background(), "exm_missing")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Get() error = %v, want %v", err, wantErr)
	}
}

type stubRepository struct {
	createFn func(context.Context, Example) (Example, error)
	getFn    func(context.Context, string) (Example, error)
	listFn   func(context.Context) ([]Example, error)
}

func (s *stubRepository) Create(ctx context.Context, item Example) (Example, error) {
	if s.createFn == nil {
		return item, nil
	}
	return s.createFn(ctx, item)
}

func (s *stubRepository) Get(ctx context.Context, id string) (Example, error) {
	if s.getFn == nil {
		return Example{}, nil
	}
	return s.getFn(ctx, id)
}

func (s *stubRepository) List(ctx context.Context) ([]Example, error) {
	if s.listFn == nil {
		return nil, nil
	}
	return s.listFn(ctx)
}
