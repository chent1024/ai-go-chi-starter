# AI Go Chi Starter

一个面向 Codex 的最小 Go 服务模板仓库。

## 模板硬约束

- handler 不写业务逻辑，只负责 HTTP 协议转换
- service 不依赖 `chi`、`net/http` 或 HTTP DTO
- repository 不返回 transport DTO
- 所有 env 读取只允许放在 `internal/config/config.go`
- 改 API 路由、字段或错误语义时，必须同步 `docs/api.md`、`openapi/openapi.yaml` 和测试
- 改配置时，必须同步 `.env.example`、`deploy/.env.runtime.example`、`deploy/.env.dev.example`、`docs/config.md`
- 改 runtime wiring、并发、存储、契约或 migration 时，必须跑 `make verify-strict`

Codex 扩展这个模板前，建议先读 [docs/codex-guide.md](docs/codex-guide.md) 和 [docs/recipes/add-domain.md](docs/recipes/add-domain.md)。

## 特性

- 基于 `chi` 的 HTTP 路由，内置 request ID、trace、recover、access log middleware
- 提供 body limit、安全响应头、`/version`、`/metrics` 等产品级 HTTP 基线
- 所有环境变量统一在 `internal/config` 中读取和校验
- 基于 `slog` 的运行时日志，支持标准输出、文件输出、基础脱敏、request/trace 上下文和 outbound logging
- 明确区分 `transport`、`service`、`infra`，避免 handler/service/repository 职责混乱
- 提供 PostgreSQL repository 示例和 SQL migration
- 提供 API、worker、migrate 三个基础进程入口

## 目录结构

```text
cmd/                    进程入口
internal/config/        配置加载与校验
internal/runtime/       日志、trace、outbound logging
internal/transport/     HTTP 协议层
internal/service/       业务逻辑层
internal/infra/         具体基础设施适配层
db/migrations/          前向 SQL migration
deploy/                 本地运行示例
docs/                   架构与配置文档
openapi/                API 契约
```

## 快速开始

1. 将 `.env.example` 复制为 `.env`。
2. 使用以下命令启动 PostgreSQL：

```bash
make dev-up
```

3. 执行 migration：

```bash
make migrate
```

4. 启动 API：

```bash
make run-api
```

5. 启动 worker：

```bash
make run-worker
```

## 常用入口

- `make build`
- `make build-api`
- `make build-worker`
- `make build-migrate`
- `make release-build`
- `make run-api`
- `make run-worker`
- `make migrate`
- `make migrate-version`
- `make dev-up`
- `make dev-down`
- `make dev-logs`
- `make dev-ps`
- `make test-integration`
- `make smoke`

## 验证

- `make verify`
- `make verify-strict`
- `make test-integration`
- `make smoke`

## Release 构建

- 本地开发构建默认输出到 `bin/`
- 统一发布构建使用 `make release-build`
- 发布产物固定输出到 `dist/<VERSION>/<TARGET_OS>-<TARGET_ARCH>/`
- `/version` 的 `version`、`commit`、`build_time` 都来自 `VERSION`、`COMMIT`、`BUILD_TIME` 这组 Make 变量

示例：

```bash
make release-build VERSION=v0.1.0 TARGET_OS=linux TARGET_ARCH=amd64
```

产物目录示例：

```text
dist/v0.1.0/linux-amd64/api
dist/v0.1.0/linux-amd64/worker
dist/v0.1.0/linux-amd64/migrate
```

补充说明：

- `go run ./cmd/api` 或普通开发场景下，`/version` 通常显示 `dev/unknown/unknown`
- 通过 `make build*` 或 `make release-build*` 构建时，会统一注入真实 build 信息

## 默认管理端点

- `GET /healthz`
- `GET /readyz`
- `GET /version`
- `GET /metrics`
