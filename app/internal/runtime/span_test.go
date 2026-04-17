package runtime

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"ai-go-chi-starter/internal/service/shared"
)

func TestStartSpanContinuesTraceAndLogsDebugLifecycle(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	parent := shared.NewRootTrace()
	ctx := shared.WithRequestID(shared.WithTrace(context.Background(), parent), "req_01")

	spanCtx, span := StartSpan(ctx, logger, "example.operation", "component", "test")
	trace, ok := shared.TraceFromContext(spanCtx)
	if !ok {
		t.Fatal("trace missing from span context")
	}
	if trace.TraceID != parent.TraceID {
		t.Fatalf("trace id = %q, want %q", trace.TraceID, parent.TraceID)
	}
	if trace.ParentSpanID != parent.SpanID {
		t.Fatalf("parent span id = %q, want %q", trace.ParentSpanID, parent.SpanID)
	}

	span.End(nil, "status", 200)

	output := logs.String()
	for _, want := range []string{
		`"span_name":"example.operation"`,
		`"request_id":"req_01"`,
		`"trace_id":"` + parent.TraceID + `"`,
		`"span_status":"ok"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("log output missing %q: %s", want, output)
		}
	}
}
