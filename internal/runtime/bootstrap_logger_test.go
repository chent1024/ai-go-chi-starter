package runtime

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestNewBootstrapLoggerIncludesServiceAndRedactsSensitiveValues(t *testing.T) {
	var logs bytes.Buffer
	logger := NewBootstrapLogger("api", &logs)

	logger.Error("bootstrap failed", "kind", "fatal", "err", errors.New("token=secret"))

	output := logs.String()
	for _, want := range []string{`"service":"api"`, `"component":"bootstrap"`, `"kind":"fatal"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("bootstrap log missing %q: %s", want, output)
		}
	}
	if strings.Contains(output, "secret") {
		t.Fatalf("bootstrap log still contains secret: %s", output)
	}
}
