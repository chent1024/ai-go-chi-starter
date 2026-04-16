# Configuration

All environment variables are loaded only in `internal/config/config.go`.

## Runtime Keys

| Key | Default | Notes |
| --- | --- | --- |
| `APP_ENV` | `development` | Application environment label. |
| `APP_DATABASE_URL` | empty | Required by `cmd/api` and `cmd/migrate`. |
| `APP_DATABASE_MAX_OPEN_CONNS` | `25` | Database pool max open connections. |
| `APP_DATABASE_MAX_IDLE_CONNS` | `25` | Database pool max idle connections. |
| `APP_DATABASE_CONN_MAX_LIFETIME` | `30m` | Database connection max lifetime. |
| `APP_DATABASE_CONN_MAX_IDLE_TIME` | `15m` | Database connection max idle time. |
| `APP_API_LISTEN_ADDR` | `:8080` | API bind address. |
| `APP_API_SHUTDOWN_TIMEOUT` | `10s` | API graceful shutdown timeout. |
| `APP_API_READ_TIMEOUT` | `15s` | HTTP server read timeout. |
| `APP_API_WRITE_TIMEOUT` | `30s` | HTTP server write timeout. |
| `APP_API_IDLE_TIMEOUT` | `60s` | HTTP keep-alive idle timeout. |
| `APP_API_REQUEST_TIMEOUT` | `30s` | Per-request context timeout middleware. |
| `APP_API_MAX_HEADER_BYTES` | `1048576` | HTTP max request header size. |
| `APP_WORKER_ENABLED` | `true` | Enables the worker loop. |
| `APP_WORKER_POLL_INTERVAL` | `5s` | Worker ticker interval. |
| `APP_WORKER_SHUTDOWN_TIMEOUT` | `10s` | Worker graceful shutdown timeout. |
| `APP_LOG_LEVEL` | `info` | One of `debug`, `info`, `warn`, `error`. |
| `APP_LOG_FORMAT` | `text` | One of `text`, `json`. |
| `APP_LOG_OUTPUT` | `stdout` | One of `stdout`, `file`, `both`. |
| `APP_LOG_DIR` | `./.runtime/logs` | File log output directory. |
| `APP_LOG_RETENTION_DAYS` | `7` | Daily log retention window. |
| `APP_LOG_CLEANUP_INTERVAL` | `1h` | Cleanup ticker interval for file logs. |
| `APP_LOG_ACCESS_ENABLED` | `true` | Enables HTTP access logs. |
| `APP_LOG_SOURCE_ENABLED` | `false` | Enables slog source info. |
| `APP_LOG_OUTBOUND_ENABLED` | `true` | Enables outbound request logging. |
| `APP_LOG_OUTBOUND_LEVEL` | `debug` | Success log level for outbound calls. |
| `APP_OUTBOUND_TIMEOUT` | `30s` | Whole outbound HTTP request timeout. |
| `APP_OUTBOUND_MAX_IDLE_CONNS` | `100` | Global outbound keep-alive idle connection cap. |
| `APP_OUTBOUND_MAX_IDLE_CONNS_PER_HOST` | `10` | Per-host outbound keep-alive idle connection cap. |
| `APP_OUTBOUND_IDLE_CONN_TIMEOUT` | `90s` | Outbound idle keep-alive connection timeout. |
| `APP_OUTBOUND_TLS_HANDSHAKE_TIMEOUT` | `10s` | Outbound TLS handshake timeout. |
| `APP_OUTBOUND_RESPONSE_HEADER_TIMEOUT` | `15s` | Wait timeout for outbound response headers. |
| `APP_OUTBOUND_EXPECT_CONTINUE_TIMEOUT` | `1s` | Wait timeout for `100-continue` responses. |
| `APP_TIMEZONE` | `UTC` | Timezone for log timestamps and file rotation. |

## Docker Dev Keys

These keys are loaded into config so the examples stay aligned with the repository contract, but they only affect local Compose usage.

| Key | Default | Notes |
| --- | --- | --- |
| `DOCKER_POSTGRES_HOST_PORT` | `5432` | Host port for local PostgreSQL. |
| `DOCKER_POSTGRES_DB` | `app` | Local PostgreSQL database name. |
| `DOCKER_POSTGRES_USER` | `app` | Local PostgreSQL user. |
| `DOCKER_POSTGRES_PASSWORD` | `app` | Local PostgreSQL password. |

## Validation Rules

- shutdown and poll durations must be positive
- API timeouts and max header bytes must be positive
- database pool sizes must be valid and idle must not exceed open
- database pool lifetimes must be positive
- log enums must be valid
- log retention and cleanup interval must be positive
- outbound timeouts and idle connection caps must be positive
- per-host outbound idle connections must not exceed the global outbound idle limit
- timezone must resolve to a valid Go location
