package runtime

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

var (
	bearerTokenPattern    = regexp.MustCompile(`(?i)bearer\s+[a-z0-9\-._~+/]+=*`)
	keyValueSecretPattern = regexp.MustCompile(
		`(?i)(authorization|cookie|set-cookie|x-api-key|api_key|access_token|refresh_token|token)\s*([=:])\s*([^\s,;]+)`,
	)
)

func redactAttr(attr slog.Attr) slog.Attr {
	if isSensitiveAttrKey(attr.Key) {
		return slog.String(attr.Key, "[REDACTED]")
	}

	switch attr.Value.Kind() {
	case slog.KindString:
		return slog.String(attr.Key, RedactText(attr.Value.String()))
	case slog.KindAny:
		if err, ok := attr.Value.Any().(error); ok && err != nil {
			return slog.String(attr.Key, RedactText(err.Error()))
		}
		if value, ok := attr.Value.Any().(fmt.Stringer); ok && value != nil {
			return slog.String(attr.Key, RedactText(value.String()))
		}
	}
	return attr
}

func RedactText(value string) string {
	if value == "" {
		return value
	}
	value = bearerTokenPattern.ReplaceAllString(value, "Bearer [REDACTED]")
	return keyValueSecretPattern.ReplaceAllString(value, `$1$2[REDACTED]`)
}

func isSensitiveAttrKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "authorization", "cookie", "set-cookie", "x-api-key", "api_key", "access_token", "refresh_token", "token":
		return true
	default:
		return false
	}
}
