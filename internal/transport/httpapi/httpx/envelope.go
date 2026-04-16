package httpx

import (
	"encoding/json"
	"io"
	"net/http"
)

type Envelope struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
	Retryable *bool  `json:"retryable,omitempty"`
}

func DecodeJSON(body io.ReadCloser, out any) error {
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
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

func WriteError(w http.ResponseWriter, status int, requestID, code, message string, retryable bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Code:      code,
		Message:   message,
		RequestID: requestID,
		Data:      nil,
		Retryable: boolPtr(retryable),
	})
}

func boolPtr(value bool) *bool {
	return &value
}
