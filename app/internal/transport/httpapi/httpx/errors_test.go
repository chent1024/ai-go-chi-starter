package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-go-chi-starter/internal/service/example"
	"ai-go-chi-starter/internal/service/shared"
)

func TestWriteDomainError(t *testing.T) {
	rec := httptest.NewRecorder()
	err := shared.ErrNotFound(example.CodeNotFound, "example not found", shared.WithRetryable(true))

	WriteDomainError(rec, "req_01", err)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	var env Envelope
	if decodeErr := json.Unmarshal(rec.Body.Bytes(), &env); decodeErr != nil {
		t.Fatalf("Unmarshal() error = %v", decodeErr)
	}
	if env.Code != example.CodeNotFound || env.RequestID != "req_01" || env.Retryable == nil || !*env.Retryable {
		t.Fatalf("unexpected envelope = %+v", env)
	}
}

func TestWriteDomainErrorIncludesDetails(t *testing.T) {
	rec := httptest.NewRecorder()
	err := shared.ErrInvalidArgument("name is required", shared.WithFieldErrors(shared.RequiredField("name")))

	WriteDomainError(rec, "req_03", err)

	var env Envelope
	if decodeErr := json.Unmarshal(rec.Body.Bytes(), &env); decodeErr != nil {
		t.Fatalf("Unmarshal() error = %v", decodeErr)
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

func TestWriteDomainErrorFallsBackForPlainError(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteDomainError(rec, "req_02", errors.New("plain"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestWriteRequestDomainErrorLogsStructuredFailure(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	req := httptest.NewRequest(http.MethodGet, "/v1/examples/exm_01", nil)
	req.Header.Set(RequestIDHeader, "req_99")
	req = req.WithContext(WithLogger(req.Context(), logger.With("request_id", "req_99")))

	rec := httptest.NewRecorder()
	WriteRequestDomainError(rec, req, example.ErrNotFound())

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	output := logs.String()
	for _, want := range []string{`"kind":"error"`, `"error_code":"EXAMPLE_NOT_FOUND"`, `"request_id":"req_99"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("log output missing %q: %s", want, output)
		}
	}
}

func TestWriteRequestDomainErrorFallbackLogsOriginalError(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	req := httptest.NewRequest(http.MethodGet, "/v1/examples/exm_01", nil)
	req.Header.Set(RequestIDHeader, "req_100")
	req = req.WithContext(WithLogger(req.Context(), logger.With("request_id", "req_100")))

	rec := httptest.NewRecorder()
	WriteRequestDomainError(rec, req, errors.New("database socket closed"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	output := logs.String()
	for _, want := range []string{`"error_code":"INTERNAL"`, `"err":"database socket closed"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("log output missing %q: %s", want, output)
		}
	}
}
