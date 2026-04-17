package shared

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
)

const traceparentVersion = "00"

type Trace struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
}

type traceContextKey struct{}

func NewRootTrace() Trace {
	return Trace{
		TraceID: randomHex(16),
		SpanID:  randomHex(8),
	}
}

func NewChildTrace(parent Trace) Trace {
	traceID := parent.TraceID
	if !validHex(traceID, 32) {
		traceID = randomHex(16)
	}
	return Trace{
		TraceID:      traceID,
		SpanID:       randomHex(8),
		ParentSpanID: parent.SpanID,
	}
}

func ParseTraceparent(value string) (Trace, bool) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 4 {
		return Trace{}, false
	}
	if parts[0] != traceparentVersion {
		return Trace{}, false
	}
	if !validHex(parts[1], 32) || !validHex(parts[2], 16) || !validHex(parts[3], 2) {
		return Trace{}, false
	}
	return Trace{
		TraceID: parts[1],
		SpanID:  parts[2],
	}, true
}

func (t Trace) Valid() bool {
	return validHex(t.TraceID, 32) && validHex(t.SpanID, 16)
}

func (t Trace) Traceparent() string {
	if !t.Valid() {
		return ""
	}
	return strings.Join([]string{traceparentVersion, t.TraceID, t.SpanID, "01"}, "-")
}

func WithTrace(ctx context.Context, trace Trace) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if !trace.Valid() {
		return ctx
	}
	return context.WithValue(ctx, traceContextKey{}, trace)
}

func TraceFromContext(ctx context.Context) (Trace, bool) {
	if ctx == nil {
		return Trace{}, false
	}
	trace, ok := ctx.Value(traceContextKey{}).(Trace)
	if !ok || !trace.Valid() {
		return Trace{}, false
	}
	return trace, true
}

func TraceLogFields(trace Trace) []any {
	if !trace.Valid() {
		return nil
	}
	fields := []any{
		LogFieldTraceID, trace.TraceID,
		LogFieldSpanID, trace.SpanID,
	}
	if trace.ParentSpanID != "" {
		fields = append(fields, LogFieldParentSpanID, trace.ParentSpanID)
	}
	return fields
}

func randomHex(byteLen int) string {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}

func validHex(value string, expectedLen int) bool {
	if len(value) != expectedLen {
		return false
	}
	if value == strings.Repeat("0", expectedLen) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
