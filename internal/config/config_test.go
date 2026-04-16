package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	unsetEnv(
		t,
		"APP_ENV",
		"APP_DATABASE_URL",
		"APP_API_LISTEN_ADDR",
		"APP_API_SHUTDOWN_TIMEOUT",
		"APP_WORKER_ENABLED",
		"APP_WORKER_POLL_INTERVAL",
		"APP_WORKER_SHUTDOWN_TIMEOUT",
		"APP_LOG_LEVEL",
		"APP_LOG_FORMAT",
		"APP_LOG_OUTPUT",
		"APP_LOG_DIR",
		"APP_LOG_RETENTION_DAYS",
		"APP_LOG_CLEANUP_INTERVAL",
		"APP_LOG_ACCESS_ENABLED",
		"APP_LOG_SOURCE_ENABLED",
		"APP_LOG_OUTBOUND_ENABLED",
		"APP_LOG_OUTBOUND_LEVEL",
		"APP_TIMEZONE",
		"DOCKER_POSTGRES_HOST_PORT",
		"DOCKER_POSTGRES_DB",
		"DOCKER_POSTGRES_USER",
		"DOCKER_POSTGRES_PASSWORD",
	)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q", cfg.AppEnv)
	}
	if cfg.API.ListenAddr != ":8080" || cfg.API.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected API defaults: %+v", cfg.API)
	}
	if !cfg.Worker.Enabled || cfg.Worker.PollInterval != 5*time.Second {
		t.Fatalf("unexpected worker defaults: %+v", cfg.Worker)
	}
	if cfg.Logging.Level != "info" || cfg.Logging.Format != "text" || cfg.Logging.Output != "stdout" {
		t.Fatalf("unexpected logging defaults: %+v", cfg.Logging)
	}
	if cfg.Logging.RetentionDays != 7 || cfg.Logging.CleanupInterval != time.Hour {
		t.Fatalf("unexpected log retention defaults: %+v", cfg.Logging)
	}
	if !cfg.Logging.AccessEnabled || cfg.Logging.SourceEnabled || !cfg.Logging.OutboundEnabled {
		t.Fatalf("unexpected logging toggles: %+v", cfg.Logging)
	}
	if cfg.Logging.OutboundLevel != "debug" {
		t.Fatalf("unexpected outbound level: %+v", cfg.Logging)
	}
	if cfg.Logging.Timezone != "UTC" || cfg.Logging.Location == nil {
		t.Fatalf("unexpected timezone defaults: %+v", cfg.Logging)
	}
	if cfg.Docker.PostgresDB != "app" || cfg.Docker.PostgresUser != "app" {
		t.Fatalf("unexpected docker defaults: %+v", cfg.Docker)
	}
}

func TestLoadAppliesOverrides(t *testing.T) {
	setEnv(t, "APP_ENV", "test")
	setEnv(t, "APP_DATABASE_URL", "postgres://user:pass@127.0.0.1:5432/app?sslmode=disable")
	setEnv(t, "APP_API_LISTEN_ADDR", ":18080")
	setEnv(t, "APP_API_SHUTDOWN_TIMEOUT", "15s")
	setEnv(t, "APP_WORKER_ENABLED", "false")
	setEnv(t, "APP_WORKER_POLL_INTERVAL", "11s")
	setEnv(t, "APP_WORKER_SHUTDOWN_TIMEOUT", "25s")
	setEnv(t, "APP_LOG_LEVEL", "debug")
	setEnv(t, "APP_LOG_FORMAT", "json")
	setEnv(t, "APP_LOG_OUTPUT", "both")
	setEnv(t, "APP_LOG_DIR", "/tmp/app-logs")
	setEnv(t, "APP_LOG_RETENTION_DAYS", "14")
	setEnv(t, "APP_LOG_CLEANUP_INTERVAL", "30m")
	setEnv(t, "APP_LOG_ACCESS_ENABLED", "false")
	setEnv(t, "APP_LOG_SOURCE_ENABLED", "true")
	setEnv(t, "APP_LOG_OUTBOUND_ENABLED", "false")
	setEnv(t, "APP_LOG_OUTBOUND_LEVEL", "warn")
	setEnv(t, "APP_TIMEZONE", "Asia/Shanghai")
	setEnv(t, "DOCKER_POSTGRES_HOST_PORT", "15432")
	setEnv(t, "DOCKER_POSTGRES_DB", "starter")
	setEnv(t, "DOCKER_POSTGRES_USER", "starter_user")
	setEnv(t, "DOCKER_POSTGRES_PASSWORD", "starter_pass")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "test" || cfg.Database.URL == "" {
		t.Fatalf("unexpected app config: %+v", cfg)
	}
	if cfg.API.ListenAddr != ":18080" || cfg.API.ShutdownTimeout != 15*time.Second {
		t.Fatalf("unexpected API config: %+v", cfg.API)
	}
	if cfg.Worker.Enabled || cfg.Worker.PollInterval != 11*time.Second {
		t.Fatalf("unexpected worker config: %+v", cfg.Worker)
	}
	if cfg.Logging.Level != "debug" || cfg.Logging.Format != "json" || cfg.Logging.Output != "both" {
		t.Fatalf("unexpected logging config: %+v", cfg.Logging)
	}
	if cfg.Logging.Dir != "/tmp/app-logs" || cfg.Logging.RetentionDays != 14 {
		t.Fatalf("unexpected log output config: %+v", cfg.Logging)
	}
	if cfg.Logging.AccessEnabled || !cfg.Logging.SourceEnabled || cfg.Logging.OutboundEnabled {
		t.Fatalf("unexpected logging toggles: %+v", cfg.Logging)
	}
	if cfg.Logging.OutboundLevel != "warn" || cfg.Logging.Location == nil {
		t.Fatalf("unexpected outbound config: %+v", cfg.Logging)
	}
	if cfg.Docker.PostgresHostPort != "15432" || cfg.Docker.PostgresPassword != "starter_pass" {
		t.Fatalf("unexpected docker config: %+v", cfg.Docker)
	}
}

func TestLoadReturnsParseErrors(t *testing.T) {
	setEnv(t, "APP_API_SHUTDOWN_TIMEOUT", "bad")
	setEnv(t, "APP_WORKER_ENABLED", "bad")
	setEnv(t, "APP_LOG_RETENTION_DAYS", "bad")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
}

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("Setenv(%q) error = %v", key, err)
	}
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%q) error = %v", key, err)
		}
	}
}
