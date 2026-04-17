package runtime

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"ai-go-chi-starter/internal/config"
)

func TestRedactTextMasksBearerAndTokenValues(t *testing.T) {
	redacted := RedactText("Authorization=Bearer abc.def token=secret refresh_token:xyz")
	if strings.Contains(redacted, "abc.def") || strings.Contains(redacted, "secret") || strings.Contains(redacted, "xyz") {
		t.Fatalf("redacted text still contains secret: %s", redacted)
	}
	if !strings.Contains(redacted, "[REDACTED]") {
		t.Fatalf("redacted text = %s", redacted)
	}
}

func TestNewLoggerRedactsSensitiveAttrs(t *testing.T) {
	var logs bytes.Buffer
	logger, closer := NewLogger(config.LoggingConfig{
		Level:           "info",
		Format:          "json",
		Output:          "stdout",
		RetentionDays:   1,
		CleanupInterval: 1,
		Timezone:        "UTC",
	}, "api", &logs)
	defer closer.Close()

	logger.Error("request failed", "authorization", "Bearer top-secret", "err", errors.New("token=secret"))

	output := logs.String()
	if strings.Contains(output, "top-secret") || strings.Contains(output, "secret") {
		t.Fatalf("output still contains secret: %s", output)
	}
}
