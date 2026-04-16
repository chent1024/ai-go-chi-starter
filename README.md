# AI Go Chi Starter

一个面向 Codex 的最小 Go 服务模板仓库。

## 特性

- 基于 `chi` 的 HTTP 路由，内置 request ID、trace、recover、access log middleware
- 所有环境变量统一在 `internal/config` 中读取和校验
- 基于 `slog` 的运行时日志，支持标准输出、文件输出和 outbound logging
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
- `make run-api`
- `make run-worker`
- `make migrate`
- `make migrate-version`
- `make dev-up`
- `make dev-down`
- `make dev-logs`
- `make dev-ps`

## 验证

- `make verify`
- `make verify-strict`
