# Codex 开发指南

这份模板不是靠 prompt 自觉维持结构，而是靠固定分层、同步清单和仓库校验来约束。

## 先看什么

在这个仓库里实现需求时，优先阅读：

1. `AGENTS.md`
2. `README.md`
3. `docs/app/architecture.md`
4. 对应 recipe，例如 `docs/app/recipes/add-domain.md`

## 硬约束

- handler 不写业务逻辑，只做 HTTP 协议转换
- service 不依赖 `chi`、`net/http` 或 HTTP DTO
- service 不依赖 `internal/config`、`internal/runtime/*` 或具体 `internal/infra/*` 实现
- repository 不返回 transport DTO
- repository / store 包不 import `internal/transport/*`
- 所有 env 读取只能放在 `app/internal/config/config.go`
- `app/internal/runtime` 根目录不放 Go 文件；只允许使用子包，例如 `internal/runtime/logging`、`internal/runtime/tracing`
- 不允许 import `ai-go-chi-starter/internal/runtime` 根包
- `cmd/*` 只负责装配；ticker loop、drain 状态机、长生命周期 worker 逻辑必须下沉到 `internal/*`
- `app/internal/transport/httpapi/drain`、`app/internal/transport/httpapi/metrics` 属于 HTTP 专属能力，不回流到 runtime
- 不要随意新建 `app/internal/*`、`app/internal/runtime/*` 或 `app/internal/transport/httpapi/*` 的新目录；新增这类稳定子域前，必须先补 `docs/app/architecture.md`、本文件，以及对应 repo rules / arch rules
- 生产代码里的 `context.Background()` / `context.TODO()` 只允许留在顶层入口和 runtime 内部；service / infra / transport / worker 必须继续传递调用方 context
- PostgreSQL 访问必须使用 context 版本 API；出站 HTTP 必须使用 `http.NewRequestWithContext`
- PostgreSQL repository 默认通过 `app/db/queries/*.sql` + `sqlc` 生成查询；非生成的 `internal/infra/store/postgres` 代码只做领域映射、trace 和错误转换，不直接执行 SQL
- starter 默认不要把“无上限全表列表查询”作为示例；列表 SQL 至少要带固定上限，后续真实业务再升级为分页
- 只有 observability 子包可以定义 `LogField*` 常量；通用日志字段放 `app/internal/runtime/logging`，span 私有字段留在 `app/internal/runtime/tracing`
- 改 API surface 时，必须同步 `docs/app/api.md`、`app/openapi/openapi.yaml` 和测试
- 改配置时，必须同步 `deploy/.env.runtime.example`、`deploy/.env.dev.example`、`docs/app/config.md`
- 改 runtime wiring、并发、存储、契约或 migration 时，必须跑 `make verify-strict`
- 改性能相关逻辑时，必须额外跑 `make verify-perf`

## 推荐工作流

1. 先读现有 domain、handler、repository 和 router wiring
2. 尽量复制现有模式，不新造抽象
3. 改代码时同步改测试和文档
4. 普通改动跑 `make verify`
5. 涉及 runtime / storage / migration / contract 的改动跑 `make verify-strict`
6. 涉及性能路径或 benchmark 相关改动时，额外跑 `make verify-perf`

## 扩展入口

- 新增 domain：看 `docs/app/recipes/add-domain.md`
- 新增 HTTP 接口：沿用 `app/internal/transport/httpapi/v1/example_handler.go`
- 新增 PostgreSQL repository：先在 `app/db/queries/*.sql` 定义查询并执行 `make sqlc-generate`，再沿用 `app/internal/infra/store/postgres/example_repository.go` 的薄包装模式接入领域层
- 新增错误码：沿用 `app/internal/service/<domain>/error_codes.go` 和 `docs/app/errors.md`
- 新增横切日志能力：放 `app/internal/runtime/logging`
- 新增 trace / request context / span 能力：放 `app/internal/runtime/tracing`
- 新增 HTTP 专属状态或指标：放 `app/internal/transport/httpapi/drain` 或 `.../metrics`
- 新增业务 domain：放 `app/internal/service/<domain>`，这是少数允许按 recipe 扩展的目录；不要为单个需求新建新的 `app/internal/<top-level>` 目录
- 新增 domain 后，默认要同时补齐 `model.go`、`repository.go`、`service.go`、`service_test.go`、postgres `<domain>_repository(.go/_test.go)`、`v1/<domain>_handler(.go/_test.go)`

## 不要做的事

- 不要在 `app/internal/service` 里直接 import `net/http`
- 不要在 `app/internal/service` 里 import `internal/config`、`internal/runtime/*` 或具体 `internal/infra/*`
- 不要在 `app/internal/service` 里 import `log/slog`
- 不要在 handler 里直接访问数据库
- 不要在 repository / store 里 import `internal/transport/*`
- 不要在非生成的 `internal/infra/store/postgres` 代码里直接调用 `QueryContext`、`QueryRowContext`、`ExecContext` 或手写执行 SQL
- 不要在 `app/internal/config` 之外直接读取 env
- 不要新建 `app/internal/runtime/*.go` 根级文件
- 不要 import `ai-go-chi-starter/internal/runtime` 根包
- 不要把 `time.NewTicker`、worker loop 或 drain 状态机直接写进 `cmd/*`
- 不要在生产代码里直接写 `context.Background()` / `context.TODO()`，也不要用 `http.NewRequest`
- 不要在 `runtime/logging` 之外直接 `slog.New(...)`、`slog.NewJSONHandler(...)`、`slog.NewTextHandler(...)`
- 不要在 `runtime/logging`、`runtime/tracing` 之外重新声明 `LogField*` 常量
- 不要为了单个新需求直接新建 `app/internal/foo`、`app/internal/runtime/bar` 或 `app/internal/transport/httpapi/baz`
- 不要因为一个新需求把 `example` 域抽成新的大框架
- 不要跳过 `make verify` / `make verify-strict` / `make verify-perf`
