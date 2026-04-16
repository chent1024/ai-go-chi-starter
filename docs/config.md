# 配置说明

所有环境变量都只允许在 `internal/config/config.go` 中读取。

## 运行时配置项

| Key | 默认值 | 说明 |
| --- | --- | --- |
| `APP_ENV` | `development` | 应用运行环境标识。 |
| `APP_DATABASE_URL` | 空 | `cmd/api` 和 `cmd/migrate` 必需。 |
| `APP_DATABASE_MAX_OPEN_CONNS` | `25` | 数据库连接池最大打开连接数。 |
| `APP_DATABASE_MAX_IDLE_CONNS` | `25` | 数据库连接池最大空闲连接数。 |
| `APP_DATABASE_CONN_MAX_LIFETIME` | `30m` | 数据库连接最大生命周期。 |
| `APP_DATABASE_CONN_MAX_IDLE_TIME` | `15m` | 数据库连接最大空闲时间。 |
| `APP_API_LISTEN_ADDR` | `:8080` | API 监听地址。 |
| `APP_API_SHUTDOWN_TIMEOUT` | `10s` | API graceful shutdown 超时。 |
| `APP_API_READ_TIMEOUT` | `15s` | HTTP server 读取超时。 |
| `APP_API_WRITE_TIMEOUT` | `30s` | HTTP server 写入超时。 |
| `APP_API_IDLE_TIMEOUT` | `60s` | HTTP keep-alive 空闲超时。 |
| `APP_API_REQUEST_TIMEOUT` | `30s` | 每个请求的 context timeout middleware。 |
| `APP_API_MAX_HEADER_BYTES` | `1048576` | HTTP 请求头最大大小。 |
| `APP_WORKER_ENABLED` | `true` | 是否启用 worker loop。 |
| `APP_WORKER_POLL_INTERVAL` | `5s` | Worker ticker 间隔。 |
| `APP_WORKER_SHUTDOWN_TIMEOUT` | `10s` | Worker graceful shutdown 超时。 |
| `APP_LOG_LEVEL` | `info` | 可选值：`debug`、`info`、`warn`、`error`。 |
| `APP_LOG_FORMAT` | `text` | 可选值：`text`、`json`。 |
| `APP_LOG_OUTPUT` | `stdout` | 可选值：`stdout`、`file`、`both`。 |
| `APP_LOG_DIR` | `./.runtime/logs` | 文件日志输出目录。 |
| `APP_LOG_RETENTION_DAYS` | `7` | 按天保留日志的窗口。 |
| `APP_LOG_CLEANUP_INTERVAL` | `1h` | 文件日志清理 ticker 间隔。 |
| `APP_LOG_ACCESS_ENABLED` | `true` | 是否开启 HTTP access log。 |
| `APP_LOG_SOURCE_ENABLED` | `false` | 是否开启 slog source 信息。 |
| `APP_LOG_OUTBOUND_ENABLED` | `true` | 是否开启 outbound request logging。 |
| `APP_LOG_OUTBOUND_LEVEL` | `debug` | outbound 成功请求的日志级别。 |
| `APP_OUTBOUND_TIMEOUT` | `30s` | 单次 outbound HTTP 请求的整体超时。 |
| `APP_OUTBOUND_MAX_IDLE_CONNS` | `100` | outbound keep-alive 全局最大空闲连接数。 |
| `APP_OUTBOUND_MAX_IDLE_CONNS_PER_HOST` | `10` | outbound keep-alive 每个主机的最大空闲连接数。 |
| `APP_OUTBOUND_IDLE_CONN_TIMEOUT` | `90s` | outbound keep-alive 空闲连接超时。 |
| `APP_OUTBOUND_TLS_HANDSHAKE_TIMEOUT` | `10s` | outbound TLS 握手超时。 |
| `APP_OUTBOUND_RESPONSE_HEADER_TIMEOUT` | `15s` | 等待 outbound 响应头的超时。 |
| `APP_OUTBOUND_EXPECT_CONTINUE_TIMEOUT` | `1s` | 等待 `100-continue` 响应的超时。 |
| `APP_TIMEZONE` | `UTC` | 日志时间戳和日志轮转使用的时区。 |

## Docker 本地开发配置项

这些 key 也会被加载进 config，目的是让示例文件和仓库契约保持一致；它们只影响本地 Compose 使用。

| Key | 默认值 | 说明 |
| --- | --- | --- |
| `DOCKER_POSTGRES_HOST_PORT` | `5432` | 本地 PostgreSQL 映射到宿主机的端口。 |
| `DOCKER_POSTGRES_DB` | `app` | 本地 PostgreSQL 数据库名。 |
| `DOCKER_POSTGRES_USER` | `app` | 本地 PostgreSQL 用户名。 |
| `DOCKER_POSTGRES_PASSWORD` | `app` | 本地 PostgreSQL 密码。 |

## 校验规则

- shutdown 和 poll duration 必须为正数
- API timeout 和 max header bytes 必须为正数
- 数据库连接池参数必须合法，且 idle 不能大于 open
- 数据库连接生命周期参数必须为正数
- 日志枚举值必须合法
- 日志保留天数和清理间隔必须为正数
- outbound timeout 和 idle 连接上限必须为正数
- outbound 每主机空闲连接上限不能大于全局空闲连接上限
- `APP_TIMEZONE` 必须能解析成有效的 Go location
