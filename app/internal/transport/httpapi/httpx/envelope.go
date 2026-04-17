package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
)

type Envelope struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
	Details   any    `json:"details,omitempty"`
	Retryable *bool  `json:"retryable,omitempty"`
}

var (
	ErrUnsupportedContentType = errors.New("unsupported content type")
	ErrMultipleJSONDocuments  = errors.New("request body must contain a single JSON document")
)

func DecodeJSON(body io.ReadCloser, out any) error {
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); errors.Is(err, io.EOF) {
		return nil
	} else if err == nil {
		return ErrMultipleJSONDocuments
	} else {
		return err
	}
}

func DecodeRequestJSON(req *http.Request, out any) error {
	if !HasJSONContentType(req) {
		return ErrUnsupportedContentType
	}
	return DecodeJSON(req.Body, out)
}

func HasJSONContentType(req *http.Request) bool {
	contentType := strings.TrimSpace(req.Header.Get("Content-Type"))
	if contentType == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == "application/json"
}

func WriteEnvelope(w http.ResponseWriter, status int, requestID string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Code:      "OK",
		Message:   "",
		RequestID: requestID,
		Data:      data,
	})
}

func WriteError(w http.ResponseWriter, status int, requestID, code, message string, retryable bool, details ...any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Code:      code,
		Message:   message,
		RequestID: requestID,
		Data:      nil,
		Details:   firstDetail(details),
		Retryable: boolPtr(retryable),
	})
}

func boolPtr(value bool) *bool {
	return &value
}

func firstDetail(details []any) any {
	if len(details) == 0 {
		return nil
	}
	return details[0]
}
