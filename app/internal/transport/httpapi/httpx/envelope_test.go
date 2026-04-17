package httpx

import (
	"encoding/json"
	"errors"
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

func TestDecodeJSONRejectsMultipleDocuments(t *testing.T) {
	body := ioNopCloser{Reader: strings.NewReader(`{"name":"demo"}{"name":"extra"}`)}
	var payload struct {
		Name string `json:"name"`
	}

	err := DecodeJSON(body, &payload)
	if !errors.Is(err, ErrMultipleJSONDocuments) {
		t.Fatalf("DecodeJSON() error = %v, want ErrMultipleJSONDocuments", err)
	}
}

func TestHasJSONContentType(t *testing.T) {
	testCases := []struct {
		name        string
		contentType string
		want        bool
	}{
		{name: "json", contentType: "application/json", want: true},
		{name: "json with charset", contentType: "application/json; charset=utf-8", want: true},
		{name: "missing", contentType: "", want: false},
		{name: "invalid", contentType: `application/json; charset="utf-8`, want: false},
		{name: "text plain", contentType: "text/plain", want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/examples", nil)
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			if got := HasJSONContentType(req); got != tc.want {
				t.Fatalf("HasJSONContentType() = %v, want %v", got, tc.want)
			}
		})
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
