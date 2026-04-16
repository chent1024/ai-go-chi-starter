package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	body := ioNopCloser{Reader: strings.NewReader(`{"name":"demo","extra":true}`)}
	var payload struct {
		Name string `json:"name"`
	}

	err := DecodeJSON(body, &payload)
	if err == nil {
		t.Fatal("DecodeJSON() error = nil, want unknown field error")
	}
}

func TestWriteEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteEnvelope(rec, http.StatusCreated, "req_01", map[string]string{"id": "exm_01"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d", rec.Code)
	}
	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if env.Code != "OK" || env.RequestID != "req_01" {
		t.Fatalf("unexpected envelope = %+v", env)
	}
}

type ioNopCloser struct {
	*strings.Reader
}

func (ioNopCloser) Close() error { return nil }
