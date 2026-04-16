# Codex 开发指南

这份模板不是靠 prompt 自觉维持结构，而是靠固定分层、同步清单和仓库校验来约束。

## 先看什么

在这个仓库里实现需求时，优先阅读：

1. `AGENTS.md`
2. `README.md`
3. `docs/architecture.md`
4. 对应 recipe，例如 `docs/recipes/add-domain.md`

## 硬约束

- handler 不写业务逻辑，只做 HTTP 协议转换
- service 不依赖 `chi`、`net/http` 或 HTTP DTO
- repository 不返回 transport DTO
- 所有 env 读取只能放在 `internal/config/config.go`
- 改 API surface 时，必须同步 `docs/api.md`、`openapi/openapi.yaml` 和测试
- 改配置时，必须同步 `.env.example`、`deploy/.env.runtime.example`、`deploy/.env.dev.example`、`docs/config.md`
- 改 runtime wiring、并发、存储、契约或 migration 时，必须跑 `make verify-strict`

## 推荐工作流

1. 先读现有 domain、handler、repository 和 router wiring
2. 尽量复制现有模式，不新造抽象
3. 改代码时同步改测试和文档
4. 普通改动跑 `make verify`
5. 涉及 runtime / storage / migration / contract 的改动跑 `make verify-strict`

## 扩展入口

- 新增 domain：看 `docs/recipes/add-domain.md`
- 新增 HTTP 接口：沿用 `internal/transport/httpapi/v1/example_handler.go`
- 新增 PostgreSQL repository：沿用 `internal/infra/store/postgres/example_repository.go`
- 新增错误码：沿用 `internal/service/<domain>/error_codes.go` 和 `docs/errors.md`

## 不要做的事

- 不要在 `internal/service` 里直接 import `net/http`
- 不要在 handler 里直接访问数据库
- 不要在 `internal/config` 之外直接读取 env
- 不要因为一个新需求把 `example` 域抽成新的大框架
- 不要跳过 `make verify` / `make verify-strict`
