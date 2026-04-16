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
	Outbound OutboundConfig
	Docker   DockerConfig
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type APIConfig struct {
	ListenAddr      string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	RequestTimeout  time.Duration
	MaxHeaderBytes  int
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

type OutboundConfig struct {
	Timeout               time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
	ExpectContinueTimeout time.Duration
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
	apiReadTimeout, err := durationFromEnv("APP_API_READ_TIMEOUT", 15*time.Second)
	parseErrs = appendErr(parseErrs, err)
	apiWriteTimeout, err := durationFromEnv("APP_API_WRITE_TIMEOUT", 30*time.Second)
	parseErrs = appendErr(parseErrs, err)
	apiIdleTimeout, err := durationFromEnv("APP_API_IDLE_TIMEOUT", 60*time.Second)
	parseErrs = appendErr(parseErrs, err)
	apiRequestTimeout, err := durationFromEnv("APP_API_REQUEST_TIMEOUT", 30*time.Second)
	parseErrs = appendErr(parseErrs, err)
	apiMaxHeaderBytes, err := intFromEnv("APP_API_MAX_HEADER_BYTES", 1<<20)
	parseErrs = appendErr(parseErrs, err)
	databaseMaxOpenConns, err := intFromEnv("APP_DATABASE_MAX_OPEN_CONNS", 25)
	parseErrs = appendErr(parseErrs, err)
	databaseMaxIdleConns, err := intFromEnv("APP_DATABASE_MAX_IDLE_CONNS", 25)
	parseErrs = appendErr(parseErrs, err)
	databaseConnMaxLifetime, err := durationFromEnv("APP_DATABASE_CONN_MAX_LIFETIME", 30*time.Minute)
	parseErrs = appendErr(parseErrs, err)
	databaseConnMaxIdleTime, err := durationFromEnv("APP_DATABASE_CONN_MAX_IDLE_TIME", 15*time.Minute)
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
	outboundTimeout, err := durationFromEnv("APP_OUTBOUND_TIMEOUT", 30*time.Second)
	parseErrs = appendErr(parseErrs, err)
	outboundMaxIdleConns, err := intFromEnv("APP_OUTBOUND_MAX_IDLE_CONNS", 100)
	parseErrs = appendErr(parseErrs, err)
	outboundMaxIdleConnsPerHost, err := intFromEnv("APP_OUTBOUND_MAX_IDLE_CONNS_PER_HOST", 10)
	parseErrs = appendErr(parseErrs, err)
	outboundIdleConnTimeout, err := durationFromEnv("APP_OUTBOUND_IDLE_CONN_TIMEOUT", 90*time.Second)
	parseErrs = appendErr(parseErrs, err)
	outboundTLSHandshakeTimeout, err := durationFromEnv("APP_OUTBOUND_TLS_HANDSHAKE_TIMEOUT", 10*time.Second)
	parseErrs = appendErr(parseErrs, err)
	outboundResponseHeaderTimeout, err := durationFromEnv("APP_OUTBOUND_RESPONSE_HEADER_TIMEOUT", 15*time.Second)
	parseErrs = appendErr(parseErrs, err)
	outboundExpectContinueTimeout, err := durationFromEnv("APP_OUTBOUND_EXPECT_CONTINUE_TIMEOUT", time.Second)
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
			URL:             stringFromEnv("APP_DATABASE_URL", ""),
			MaxOpenConns:    databaseMaxOpenConns,
			MaxIdleConns:    databaseMaxIdleConns,
			ConnMaxLifetime: databaseConnMaxLifetime,
			ConnMaxIdleTime: databaseConnMaxIdleTime,
		},
		API: APIConfig{
			ListenAddr:      stringFromEnv("APP_API_LISTEN_ADDR", ":8080"),
			ShutdownTimeout: apiShutdownTimeout,
			ReadTimeout:     apiReadTimeout,
			WriteTimeout:    apiWriteTimeout,
			IdleTimeout:     apiIdleTimeout,
			RequestTimeout:  apiRequestTimeout,
			MaxHeaderBytes:  apiMaxHeaderBytes,
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
		Outbound: OutboundConfig{
			Timeout:               outboundTimeout,
			MaxIdleConns:          outboundMaxIdleConns,
			MaxIdleConnsPerHost:   outboundMaxIdleConnsPerHost,
			IdleConnTimeout:       outboundIdleConnTimeout,
			TLSHandshakeTimeout:   outboundTLSHandshakeTimeout,
			ResponseHeaderTimeout: outboundResponseHeaderTimeout,
			ExpectContinueTimeout: outboundExpectContinueTimeout,
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
	if c.API.ReadTimeout <= 0 {
		errs = append(errs, errors.New("APP_API_READ_TIMEOUT must be positive"))
	}
	if c.API.WriteTimeout <= 0 {
		errs = append(errs, errors.New("APP_API_WRITE_TIMEOUT must be positive"))
	}
	if c.API.IdleTimeout <= 0 {
		errs = append(errs, errors.New("APP_API_IDLE_TIMEOUT must be positive"))
	}
	if c.API.RequestTimeout <= 0 {
		errs = append(errs, errors.New("APP_API_REQUEST_TIMEOUT must be positive"))
	}
	if c.API.MaxHeaderBytes <= 0 {
		errs = append(errs, errors.New("APP_API_MAX_HEADER_BYTES must be positive"))
	}
	if c.Database.MaxOpenConns <= 0 {
		errs = append(errs, errors.New("APP_DATABASE_MAX_OPEN_CONNS must be positive"))
	}
	if c.Database.MaxIdleConns < 0 {
		errs = append(errs, errors.New("APP_DATABASE_MAX_IDLE_CONNS must not be negative"))
	}
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		errs = append(errs, errors.New("APP_DATABASE_MAX_IDLE_CONNS must not exceed APP_DATABASE_MAX_OPEN_CONNS"))
	}
	if c.Database.ConnMaxLifetime <= 0 {
		errs = append(errs, errors.New("APP_DATABASE_CONN_MAX_LIFETIME must be positive"))
	}
	if c.Database.ConnMaxIdleTime <= 0 {
		errs = append(errs, errors.New("APP_DATABASE_CONN_MAX_IDLE_TIME must be positive"))
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
	if c.Outbound.Timeout <= 0 {
		errs = append(errs, errors.New("APP_OUTBOUND_TIMEOUT must be positive"))
	}
	if c.Outbound.MaxIdleConns <= 0 {
		errs = append(errs, errors.New("APP_OUTBOUND_MAX_IDLE_CONNS must be positive"))
	}
	if c.Outbound.MaxIdleConnsPerHost <= 0 {
		errs = append(errs, errors.New("APP_OUTBOUND_MAX_IDLE_CONNS_PER_HOST must be positive"))
	}
	if c.Outbound.MaxIdleConnsPerHost > c.Outbound.MaxIdleConns {
		errs = append(
			errs,
			errors.New("APP_OUTBOUND_MAX_IDLE_CONNS_PER_HOST must not exceed APP_OUTBOUND_MAX_IDLE_CONNS"),
		)
	}
	if c.Outbound.IdleConnTimeout <= 0 {
		errs = append(errs, errors.New("APP_OUTBOUND_IDLE_CONN_TIMEOUT must be positive"))
	}
	if c.Outbound.TLSHandshakeTimeout <= 0 {
		errs = append(errs, errors.New("APP_OUTBOUND_TLS_HANDSHAKE_TIMEOUT must be positive"))
	}
	if c.Outbound.ResponseHeaderTimeout <= 0 {
		errs = append(errs, errors.New("APP_OUTBOUND_RESPONSE_HEADER_TIMEOUT must be positive"))
	}
	if c.Outbound.ExpectContinueTimeout <= 0 {
		errs = append(errs, errors.New("APP_OUTBOUND_EXPECT_CONTINUE_TIMEOUT must be positive"))
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
