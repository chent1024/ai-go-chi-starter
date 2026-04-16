package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"ai-go-chi-starter/internal/service/example"
	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

func TestExampleHandlerCreate(t *testing.T) {
	handler := NewExampleHandler(handlerService{
		createFn: func(ctx context.Context, input example.CreateInput) (example.Example, error) {
			_ = ctx
			now := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
			return example.Example{
				ID:        "exm_01",
				Name:      input.Name,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/examples", bytes.NewBufferString(`{"name":"demo"}`))
	req.Header.Set(httpx.RequestIDHeader, "req_01")
	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestExampleHandlerGet(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/v1/examples/{id}", NewExampleHandler(handlerService{
		getFn: func(ctx context.Context, id string) (example.Example, error) {
			_ = ctx
			return example.Example{ID: id, Name: "demo", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
	}).Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/examples/exm_01", nil)
	req.Header.Set(httpx.RequestIDHeader, "req_02")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestExampleHandlerList(t *testing.T) {
	handler := NewExampleHandler(handlerService{
		listFn: func(ctx context.Context) ([]example.Example, error) {
			_ = ctx
			now := time.Now()
			return []example.Example{{ID: "exm_01", Name: "demo", CreatedAt: now, UpdatedAt: now}}, nil
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/examples", nil)
	req.Header.Set(httpx.RequestIDHeader, "req_03")
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var env httpx.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if env.Code != "OK" {
		t.Fatalf("unexpected envelope = %+v", env)
	}
}

func TestExampleHandlerWritesDomainError(t *testing.T) {
	handler := NewExampleHandler(handlerService{
		getFn: func(ctx context.Context, id string) (example.Example, error) {
			_ = ctx
			_ = id
			return example.Example{}, shared.NewError("EXAMPLE_NOT_FOUND", "example not found", http.StatusNotFound)
		},
	})

	router := chi.NewRouter()
	router.Get("/v1/examples/{id}", handler.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/examples/exm_missing", nil)
	req.Header.Set(httpx.RequestIDHeader, "req_04")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

type handlerService struct {
	createFn func(context.Context, example.CreateInput) (example.Example, error)
	getFn    func(context.Context, string) (example.Example, error)
	listFn   func(context.Context) ([]example.Example, error)
}

func (h handlerService) Create(ctx context.Context, input example.CreateInput) (example.Example, error) {
	return h.createFn(ctx, input)
}

func (h handlerService) Get(ctx context.Context, id string) (example.Example, error) {
	return h.getFn(ctx, id)
}

func (h handlerService) List(ctx context.Context) ([]example.Example, error) {
	return h.listFn(ctx)
}
