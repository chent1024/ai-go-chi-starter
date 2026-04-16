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

	"ai-go-chi-starter/internal/service/shared"
)

func TestWriteDomainError(t *testing.T) {
	rec := httptest.NewRecorder()
	err := shared.NewError("EXAMPLE_NOT_FOUND", "example not found", http.StatusNotFound, shared.WithRetryable(true))

	WriteDomainError(rec, "req_01", err)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	var env Envelope
	if decodeErr := json.Unmarshal(rec.Body.Bytes(), &env); decodeErr != nil {
		t.Fatalf("Unmarshal() error = %v", decodeErr)
	}
	if env.Code != "EXAMPLE_NOT_FOUND" || env.RequestID != "req_01" || env.Retryable == nil || !*env.Retryable {
		t.Fatalf("unexpected envelope = %+v", env)
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
	WriteRequestDomainError(rec, req, shared.NewError("EXAMPLE_NOT_FOUND", "example not found", http.StatusNotFound))

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
