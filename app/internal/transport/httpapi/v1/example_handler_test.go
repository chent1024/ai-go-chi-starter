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
	"ai-go-chi-starter/internal/transport/httpapi/middleware"
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
	req.Header.Set("Content-Type", "application/json")
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
			return example.Example{}, example.ErrNotFound()
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

func TestExampleHandlerCreateWritesValidationDetails(t *testing.T) {
	handler := NewExampleHandler(handlerService{
		createFn: func(ctx context.Context, input example.CreateInput) (example.Example, error) {
			_ = ctx
			_ = input
			return example.Example{}, shared.ErrInvalidArgument(
				"name is required",
				shared.WithFieldErrors(shared.RequiredField("name")),
			)
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/examples", bytes.NewBufferString(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpx.RequestIDHeader, "req_validation")
	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	var env httpx.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	details, ok := env.Details.(map[string]any)
	if !ok {
		t.Fatalf("details type = %T", env.Details)
	}
	fieldErrors, ok := details["field_errors"].([]any)
	if !ok || len(fieldErrors) != 1 {
		t.Fatalf("field_errors = %#v", details["field_errors"])
	}
}

func TestExampleHandlerCreateRejectsLargeBody(t *testing.T) {
	handler := middleware.BodyLimit(8)(http.HandlerFunc(NewExampleHandler(handlerService{
		createFn: func(ctx context.Context, input example.CreateInput) (example.Example, error) {
			t.Fatal("Create() should not be called")
			return example.Example{}, nil
		},
	}).Create))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/examples", bytes.NewBufferString(`{"name":"demo"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpx.RequestIDHeader, "req_large")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d", rec.Code)
	}
	var env httpx.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if env.Code != shared.CodeInvalidArgument {
		t.Fatalf("unexpected envelope = %+v", env)
	}
}

func TestExampleHandlerCreateRejectsEmptyBody(t *testing.T) {
	handler := NewExampleHandler(handlerService{
		createFn: func(ctx context.Context, input example.CreateInput) (example.Example, error) {
			t.Fatal("Create() should not be called")
			return example.Example{}, nil
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/examples", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpx.RequestIDHeader, "req_empty")
	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestExampleHandlerCreateRejectsUnsupportedContentType(t *testing.T) {
	handler := NewExampleHandler(handlerService{
		createFn: func(ctx context.Context, input example.CreateInput) (example.Example, error) {
			t.Fatal("Create() should not be called")
			return example.Example{}, nil
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/examples", bytes.NewBufferString(`{"name":"demo"}`))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set(httpx.RequestIDHeader, "req_content_type")
	handler.Create(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d", rec.Code)
	}
	var env httpx.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if env.Code != shared.CodeInvalidArgument || env.Message != "Content-Type must be application/json" {
		t.Fatalf("unexpected envelope = %+v", env)
	}
}

func TestExampleHandlerCreateRejectsMultipleJSONDocuments(t *testing.T) {
	handler := NewExampleHandler(handlerService{
		createFn: func(ctx context.Context, input example.CreateInput) (example.Example, error) {
			t.Fatal("Create() should not be called")
			return example.Example{}, nil
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/examples", bytes.NewBufferString(`{"name":"demo"}{"name":"extra"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpx.RequestIDHeader, "req_multiple_docs")
	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	var env httpx.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if env.Code != shared.CodeInvalidArgument || env.Message != "request body must contain a single JSON document" {
		t.Fatalf("unexpected envelope = %+v", env)
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
