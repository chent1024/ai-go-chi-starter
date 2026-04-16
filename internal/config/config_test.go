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
		"APP_DATABASE_MAX_OPEN_CONNS",
		"APP_DATABASE_MAX_IDLE_CONNS",
		"APP_DATABASE_CONN_MAX_LIFETIME",
		"APP_DATABASE_CONN_MAX_IDLE_TIME",
		"APP_API_LISTEN_ADDR",
		"APP_API_SHUTDOWN_TIMEOUT",
		"APP_API_READ_TIMEOUT",
		"APP_API_WRITE_TIMEOUT",
		"APP_API_IDLE_TIMEOUT",
		"APP_API_REQUEST_TIMEOUT",
		"APP_API_MAX_HEADER_BYTES",
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
		"APP_OUTBOUND_TIMEOUT",
		"APP_OUTBOUND_MAX_IDLE_CONNS",
		"APP_OUTBOUND_MAX_IDLE_CONNS_PER_HOST",
		"APP_OUTBOUND_IDLE_CONN_TIMEOUT",
		"APP_OUTBOUND_TLS_HANDSHAKE_TIMEOUT",
		"APP_OUTBOUND_RESPONSE_HEADER_TIMEOUT",
		"APP_OUTBOUND_EXPECT_CONTINUE_TIMEOUT",
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
	if cfg.Database.MaxOpenConns != 25 || cfg.Database.MaxIdleConns != 25 {
		t.Fatalf("unexpected database defaults: %+v", cfg.Database)
	}
	if cfg.Database.ConnMaxLifetime != 30*time.Minute || cfg.Database.ConnMaxIdleTime != 15*time.Minute {
		t.Fatalf("unexpected database duration defaults: %+v", cfg.Database)
	}
	if cfg.API.ListenAddr != ":8080" || cfg.API.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected API defaults: %+v", cfg.API)
	}
	if cfg.API.ReadTimeout != 15*time.Second ||
		cfg.API.WriteTimeout != 30*time.Second ||
		cfg.API.IdleTimeout != 60*time.Second ||
		cfg.API.RequestTimeout != 30*time.Second ||
		cfg.API.MaxHeaderBytes != 1<<20 {
		t.Fatalf("unexpected API timeout defaults: %+v", cfg.API)
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
	if cfg.Outbound.Timeout != 30*time.Second ||
		cfg.Outbound.MaxIdleConns != 100 ||
		cfg.Outbound.MaxIdleConnsPerHost != 10 ||
		cfg.Outbound.IdleConnTimeout != 90*time.Second ||
		cfg.Outbound.TLSHandshakeTimeout != 10*time.Second ||
		cfg.Outbound.ResponseHeaderTimeout != 15*time.Second ||
		cfg.Outbound.ExpectContinueTimeout != time.Second {
		t.Fatalf("unexpected outbound defaults: %+v", cfg.Outbound)
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
	setEnv(t, "APP_DATABASE_MAX_OPEN_CONNS", "40")
	setEnv(t, "APP_DATABASE_MAX_IDLE_CONNS", "10")
	setEnv(t, "APP_DATABASE_CONN_MAX_LIFETIME", "45m")
	setEnv(t, "APP_DATABASE_CONN_MAX_IDLE_TIME", "20m")
	setEnv(t, "APP_API_LISTEN_ADDR", ":18080")
	setEnv(t, "APP_API_SHUTDOWN_TIMEOUT", "15s")
	setEnv(t, "APP_API_READ_TIMEOUT", "12s")
	setEnv(t, "APP_API_WRITE_TIMEOUT", "40s")
	setEnv(t, "APP_API_IDLE_TIMEOUT", "75s")
	setEnv(t, "APP_API_REQUEST_TIMEOUT", "9s")
	setEnv(t, "APP_API_MAX_HEADER_BYTES", "2097152")
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
	setEnv(t, "APP_OUTBOUND_TIMEOUT", "45s")
	setEnv(t, "APP_OUTBOUND_MAX_IDLE_CONNS", "120")
	setEnv(t, "APP_OUTBOUND_MAX_IDLE_CONNS_PER_HOST", "24")
	setEnv(t, "APP_OUTBOUND_IDLE_CONN_TIMEOUT", "95s")
	setEnv(t, "APP_OUTBOUND_TLS_HANDSHAKE_TIMEOUT", "7s")
	setEnv(t, "APP_OUTBOUND_RESPONSE_HEADER_TIMEOUT", "18s")
	setEnv(t, "APP_OUTBOUND_EXPECT_CONTINUE_TIMEOUT", "2s")
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
	if cfg.Database.MaxOpenConns != 40 || cfg.Database.MaxIdleConns != 10 {
		t.Fatalf("unexpected database config: %+v", cfg.Database)
	}
	if cfg.Database.ConnMaxLifetime != 45*time.Minute || cfg.Database.ConnMaxIdleTime != 20*time.Minute {
		t.Fatalf("unexpected database duration config: %+v", cfg.Database)
	}
	if cfg.API.ListenAddr != ":18080" || cfg.API.ShutdownTimeout != 15*time.Second {
		t.Fatalf("unexpected API config: %+v", cfg.API)
	}
	if cfg.API.ReadTimeout != 12*time.Second ||
		cfg.API.WriteTimeout != 40*time.Second ||
		cfg.API.IdleTimeout != 75*time.Second ||
		cfg.API.RequestTimeout != 9*time.Second ||
		cfg.API.MaxHeaderBytes != 2097152 {
		t.Fatalf("unexpected API timeout config: %+v", cfg.API)
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
	if cfg.Outbound.Timeout != 45*time.Second ||
		cfg.Outbound.MaxIdleConns != 120 ||
		cfg.Outbound.MaxIdleConnsPerHost != 24 ||
		cfg.Outbound.IdleConnTimeout != 95*time.Second ||
		cfg.Outbound.TLSHandshakeTimeout != 7*time.Second ||
		cfg.Outbound.ResponseHeaderTimeout != 18*time.Second ||
		cfg.Outbound.ExpectContinueTimeout != 2*time.Second {
		t.Fatalf("unexpected outbound transport config: %+v", cfg.Outbound)
	}
	if cfg.Docker.PostgresHostPort != "15432" || cfg.Docker.PostgresPassword != "starter_pass" {
		t.Fatalf("unexpected docker config: %+v", cfg.Docker)
	}
}

func TestLoadReturnsParseErrors(t *testing.T) {
	setEnv(t, "APP_API_SHUTDOWN_TIMEOUT", "bad")
	setEnv(t, "APP_DATABASE_MAX_OPEN_CONNS", "bad")
	setEnv(t, "APP_WORKER_ENABLED", "bad")
	setEnv(t, "APP_LOG_RETENTION_DAYS", "bad")
	setEnv(t, "APP_OUTBOUND_TIMEOUT", "bad")

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
