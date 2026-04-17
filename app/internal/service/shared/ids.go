package shared

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

const (
	LogFieldService      = "service"
	LogFieldComponent    = "component"
	LogFieldRequestID    = "request_id"
	LogFieldTraceID      = "trace_id"
	LogFieldSpanID       = "span_id"
	LogFieldParentSpanID = "parent_span_id"
	LogFieldErrorCode    = "error_code"
	LogFieldRetryable    = "retryable"
	LogFieldStatus       = "status"
)

func NewID(prefix string) string {
	if strings.TrimSpace(prefix) == "" {
		prefix = "id"
	}
	return prefix + "_" + randomSuffix(12)
}

func randomSuffix(byteLen int) string {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
