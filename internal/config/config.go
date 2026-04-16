package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv   string
	Database DatabaseConfig
	API      APIConfig
	Worker   WorkerConfig
	Logging  LoggingConfig
	Docker   DockerConfig
}

type DatabaseConfig struct {
	URL string
}

type APIConfig struct {
	ListenAddr      string
	ShutdownTimeout time.Duration
}

type WorkerConfig struct {
	Enabled         bool
	PollInterval    time.Duration
	ShutdownTimeout time.Duration
}

type LoggingConfig struct {
	Level           string
	Format          string
	AccessEnabled   bool
	SourceEnabled   bool
	OutboundEnabled bool
	OutboundLevel   string
	Output          string
	Dir             string
	RetentionDays   int
	CleanupInterval time.Duration
	Timezone        string
	Location        *time.Location
}

type DockerConfig struct {
	PostgresHostPort string
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string
}

func Load() (Config, error) {
	var parseErrs []error

	apiShutdownTimeout, err := durationFromEnv("APP_API_SHUTDOWN_TIMEOUT", 10*time.Second)
	parseErrs = appendErr(parseErrs, err)
	workerEnabled, err := boolFromEnv("APP_WORKER_ENABLED", true)
	parseErrs = appendErr(parseErrs, err)
	workerPollInterval, err := durationFromEnv("APP_WORKER_POLL_INTERVAL", 5*time.Second)
	parseErrs = appendErr(parseErrs, err)
	workerShutdownTimeout, err := durationFromEnv("APP_WORKER_SHUTDOWN_TIMEOUT", 10*time.Second)
	parseErrs = appendErr(parseErrs, err)
	logAccessEnabled, err := boolFromEnv("APP_LOG_ACCESS_ENABLED", true)
	parseErrs = appendErr(parseErrs, err)
	logSourceEnabled, err := boolFromEnv("APP_LOG_SOURCE_ENABLED", false)
	parseErrs = appendErr(parseErrs, err)
	logOutboundEnabled, err := boolFromEnv("APP_LOG_OUTBOUND_ENABLED", true)
	parseErrs = appendErr(parseErrs, err)
	logLevel, err := parseLevelValue("APP_LOG_LEVEL", stringFromEnv("APP_LOG_LEVEL", "info"))
	parseErrs = appendErr(parseErrs, err)
	logFormat, err := parseFormatValue("APP_LOG_FORMAT", stringFromEnv("APP_LOG_FORMAT", "text"))
	parseErrs = appendErr(parseErrs, err)
	logOutput, err := parseOutputValue("APP_LOG_OUTPUT", stringFromEnv("APP_LOG_OUTPUT", "stdout"))
	parseErrs = appendErr(parseErrs, err)
	logOutboundLevel, err := parseLevelValue("APP_LOG_OUTBOUND_LEVEL", stringFromEnv("APP_LOG_OUTBOUND_LEVEL", "debug"))
	parseErrs = appendErr(parseErrs, err)
	logRetentionDays, err := intFromEnv("APP_LOG_RETENTION_DAYS", 7)
	parseErrs = appendErr(parseErrs, err)
	logCleanupInterval, err := durationFromEnv("APP_LOG_CLEANUP_INTERVAL", time.Hour)
	parseErrs = appendErr(parseErrs, err)
	timezone := stringFromEnv("APP_TIMEZONE", time.UTC.String())
	location, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil {
		parseErrs = appendErr(parseErrs, fmt.Errorf("APP_TIMEZONE: %w", err))
	}

	if err := errors.Join(parseErrs...); err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv: stringFromEnv("APP_ENV", "development"),
		Database: DatabaseConfig{
			URL: stringFromEnv("APP_DATABASE_URL", ""),
		},
		API: APIConfig{
			ListenAddr:      stringFromEnv("APP_API_LISTEN_ADDR", ":8080"),
			ShutdownTimeout: apiShutdownTimeout,
		},
		Worker: WorkerConfig{
			Enabled:         workerEnabled,
			PollInterval:    workerPollInterval,
			ShutdownTimeout: workerShutdownTimeout,
		},
		Logging: LoggingConfig{
			Level:           logLevel,
			Format:          logFormat,
			AccessEnabled:   logAccessEnabled,
			SourceEnabled:   logSourceEnabled,
			OutboundEnabled: logOutboundEnabled,
			OutboundLevel:   logOutboundLevel,
			Output:          logOutput,
			Dir:             stringFromEnv("APP_LOG_DIR", "./.runtime/logs"),
			RetentionDays:   logRetentionDays,
			CleanupInterval: logCleanupInterval,
			Timezone:        timezone,
			Location:        location,
		},
		Docker: DockerConfig{
			PostgresHostPort: stringFromEnv("DOCKER_POSTGRES_HOST_PORT", "5432"),
			PostgresDB:       stringFromEnv("DOCKER_POSTGRES_DB", "app"),
			PostgresUser:     stringFromEnv("DOCKER_POSTGRES_USER", "app"),
			PostgresPassword: stringFromEnv("DOCKER_POSTGRES_PASSWORD", "app"),
		},
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error

	if strings.TrimSpace(c.AppEnv) == "" {
		errs = append(errs, errors.New("APP_ENV must not be empty"))
	}
	if strings.TrimSpace(c.API.ListenAddr) == "" {
		errs = append(errs, errors.New("APP_API_LISTEN_ADDR must not be empty"))
	}
	if c.API.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("APP_API_SHUTDOWN_TIMEOUT must be positive"))
	}
	if c.Worker.PollInterval <= 0 {
		errs = append(errs, errors.New("APP_WORKER_POLL_INTERVAL must be positive"))
	}
	if c.Worker.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("APP_WORKER_SHUTDOWN_TIMEOUT must be positive"))
	}
	if !isOneOf(c.Logging.Level, "debug", "info", "warn", "error") {
		errs = append(errs, fmt.Errorf("APP_LOG_LEVEL must be one of debug, info, warn, error"))
	}
	if !isOneOf(c.Logging.Format, "text", "json") {
		errs = append(errs, fmt.Errorf("APP_LOG_FORMAT must be one of text, json"))
	}
	if !isOneOf(c.Logging.Output, "stdout", "file", "both") {
		errs = append(errs, fmt.Errorf("APP_LOG_OUTPUT must be one of stdout, file, both"))
	}
	if !isOneOf(c.Logging.OutboundLevel, "debug", "info", "warn", "error") {
		errs = append(errs, fmt.Errorf("APP_LOG_OUTBOUND_LEVEL must be one of debug, info, warn, error"))
	}
	if c.Logging.RetentionDays <= 0 {
		errs = append(errs, errors.New("APP_LOG_RETENTION_DAYS must be positive"))
	}
	if c.Logging.CleanupInterval <= 0 {
		errs = append(errs, errors.New("APP_LOG_CLEANUP_INTERVAL must be positive"))
	}
	if strings.TrimSpace(c.Logging.Timezone) == "" || c.Logging.Location == nil {
		errs = append(errs, errors.New("APP_TIMEZONE must resolve to a valid location"))
	}

	return errors.Join(errs...)
}

func stringFromEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func boolFromEnv(key string, fallback bool) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("%s: %w", key, err)
	}
	return parsed, nil
}

func intFromEnv(key string, fallback int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return parsed, nil
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return parsed, nil
}

func appendErr(errs []error, err error) []error {
	if err == nil {
		return errs
	}
	return append(errs, err)
}

func normalizeLevel(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeFormat(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeOutput(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isOneOf(value string, expected ...string) bool {
	for _, item := range expected {
		if value == item {
			return true
		}
	}
	return false
}

func parseLevelValue(key, value string) (string, error) {
	normalized := normalizeLevel(value)
	if !isOneOf(normalized, "debug", "info", "warn", "error") {
		return "", fmt.Errorf("%s: invalid value %q", key, value)
	}
	return normalized, nil
}

func parseFormatValue(key, value string) (string, error) {
	normalized := normalizeFormat(value)
	if !isOneOf(normalized, "text", "json") {
		return "", fmt.Errorf("%s: invalid value %q", key, value)
	}
	return normalized, nil
}

func parseOutputValue(key, value string) (string, error) {
	normalized := normalizeOutput(value)
	if !isOneOf(normalized, "stdout", "file", "both") {
		return "", fmt.Errorf("%s: invalid value %q", key, value)
	}
	return normalized, nil
}
