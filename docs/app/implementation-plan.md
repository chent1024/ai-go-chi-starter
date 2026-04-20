# AI Go Chi Starter 实施计划

## 目标

这个仓库的目标是成为一个对 Codex 友好的 Go 服务 starter template。
它需要给新 Go 服务提供一套干净、明确、有取舍的起始骨架，让 Codex 可以直接在稳定结构上继续实现，而不是每次从零重搭基础设施，或者把项目逐步写成随意拼接的目录结构。

这个模板应当保持小而明确，并以可落地为导向。重点优化这些方面：

- 文件放置位置可预测
- 依赖方向清晰
- 基础设施尽量少但足够扎实
- 使用 chi 做 HTTP 路由
- 使用结构化日志
- 配置集中读取
- 提供基础 PostgreSQL wiring
- 用一个小型 demo domain 展示推荐编码方式

这个模板**不应该**演变成通用应用框架，也不应该变成无关复用代码的堆积区。

## 设计参考源

这个 starter 可以**参考**以下服务骨架：

- `/Users/xihe0000/workspace/code/app-storage-transfer/app`

但**不能**整体照搬那个服务。当前 upload 服务里有大量领域专属代码，不应该泄漏到这个 starter：

- upload 生命周期与状态流转
- object store 抽象
- provider 适配器
- Minimax 集成
- staging 管理
- queue / poller / purger / expirer 治理逻辑
- upload 专属 API surface 和领域模型

这里只应吸收可复用的骨架、分层思路和横切基础设施。

## 仓库定位

这个仓库应实现为一个**starter template repository**，而不是共享 Go library，也不是 upload 服务的直接克隆。

建议达成的结果：

- 一套干净的仓库骨架
- 一个 demo HTTP 资源
- 一个 demo worker skeleton
- 一个基于 PostgreSQL 的 demo repository
- 一套规则 / 校验配置
- 一份简短的 AGENTS 契约，告诉 Codex 代码应该放在哪里

## 架构原则

这个 starter 必须强制保持这些边界：

1. `app/cmd/*` 只负责进程装配。
2. `transport` 只处理 HTTP 协议相关问题。
3. `service` 承载业务规则。
4. `infra` 承载具体适配器、repository 和 client。
5. `config` 是唯一允许读取环境变量的地方。
6. `runtime` 承载日志、trace context、outbound logging 等横切基础设施。
7. Handler 里不能写业务逻辑。
8. Service 不能依赖 chi 或 HTTP 类型。
9. Repository 不能返回 HTTP DTO。
10. 优先使用明确的小抽象，而不是框架式 magic。

## V1 非目标

第一版**不要**包含这些内容：

- object storage
- 面向 AI vendor 的 provider/client 抽象
- 文件 staging 或 upload orchestration
- 分布式任务队列
- 复杂重试治理
- 高级鉴权流程
- 多租户抽象
- 超出基础 OpenAPI stub 的代码生成流水线
- 泛化的领域工具包

## 推荐仓库结构

starter 目标结构如下：

```text
ai-go-chi-starter/
├── AGENTS.md
├── Makefile
├── Makefile.rules
├── .gitignore
├── README.md
├── .orch/rules/
├── docs/
│   ├── implementation-plan.md
│   ├── architecture.md
│   ├── config.md
│   └── api.md
├── deploy/
│   ├── .env.runtime.example
│   ├── .env.dev.example
│   ├── docker-compose.dev.yaml
│   └── README.md
├── app/
│   ├── go.mod
│   ├── cmd/
│   │   ├── api/
│   │   │   ├── main.go
│   │   │   └── app.go
│   │   ├── worker/
│   │   │   ├── main.go
│   │   │   └── app.go
│   │   └── migrate/
│   │       └── main.go
│   ├── db/
│   │   └── migrations/
│   │       └── 001_init.sql
│   ├── openapi/
│   │   └── openapi.yaml
│   └── internal/
│       ├── config/
│       │   ├── config.go
│       │   └── config_test.go
│       ├── runtime/
│       │   ├── logging.go
│       │   ├── log_file.go
│       │   ├── log_fields.go
│       │   ├── trace.go
│       │   ├── log_context.go
│       │   └── outbound.go
│       ├── transport/
│       │   └── httpapi/
│       │       ├── router.go
│       │       ├── middleware/
│       │       │   ├── request_id.go
│       │       │   ├── trace.go
│       │       │   ├── access_log.go
│       │       │   └── recover.go
│       │       ├── httpx/
│       │       │   ├── envelope.go
│       │       │   ├── errors.go
│       │       │   ├── request_context.go
│       │       │   └── response_recorder.go
│       │       └── v1/
│       │           └── example_handler.go
│       ├── service/
│       │   ├── shared/
│       │   │   ├── error.go
│       │   │   ├── trace.go
│       │   │   └── ids.go
│       │   └── example/
│       │       ├── model.go
│       │       ├── service.go
│       │       └── repository.go
│       └── infra/
│           ├── store/
│           │   └── postgres/
│           │       ├── db.go
│           │       └── example_repository.go
│           └── client/
│               └── httpclient.go
└── .runtime/
```

## V1 必须具备的能力

第一版可用版本至少应包含以下内容：

### 1. API 进程骨架

实现：

- `app/cmd/api/main.go`
- `app/cmd/api/app.go`

能力要求：

- 加载配置
- 构造 logger
- 构造 HTTP server
- graceful shutdown
- 启动失败时返回非零退出码

### 2. Worker 进程骨架

实现：

- `app/cmd/worker/main.go`
- `app/cmd/worker/app.go`

能力要求：

- 加载配置
- 构造 logger
- 启动 ticker loop
- 提供一个 `JobHandler` 接口示例
- graceful shutdown

这个 worker 应刻意保持小而通用，不应把 upload 服务那套治理逻辑带进来。

### 3. Migration 进程

实现：

- `app/cmd/migrate/main.go`

能力要求：

- 连接 PostgreSQL
- 执行 `app/db/migrations` 下的 SQL migration

### 4. 配置系统

实现：

- `app/internal/config/config.go`
- `app/internal/config/config_test.go`

要求：

- 只通过环境变量配置
- 集中解析
- 明确默认值
- 启动阶段完成校验
- `app/internal/config` 之外不允许读取环境变量

### 5. 日志与 Runtime 横切层

实现：

- `app/internal/runtime/logging.go`
- `app/internal/runtime/log_file.go`
- `app/internal/runtime/log_fields.go`
- `app/internal/runtime/log_context.go`
- `app/internal/runtime/trace.go`
- `app/internal/runtime/outbound.go`

要求：

- `slog`
- `text/json` 格式
- `stdout/file/both` 输出
- 按日期轮转日志文件
- 基础保留与清理能力
- 支持 trace 感知的 logger 增强
- outbound request logging helper

### 6. 基于 chi 的 HTTP 层

实现：

- `app/internal/transport/httpapi/router.go`
- `app/internal/transport/httpapi/middleware/*`
- `app/internal/transport/httpapi/httpx/*`

要求：

- 使用 `github.com/go-chi/chi/v5`
- middleware 顺序：
  1. recover
  2. request id
  3. trace
  4. access log
- 暴露：
  - `GET /healthz`
  - `GET /readyz`
  - `POST /v1/examples`
  - `GET /v1/examples/{id}`
  - `GET /v1/examples`

### 7. Demo 领域

实现一个最小领域 `example`。

目的：

- 展示推荐的 handler / service / repository 分层方式
- 为测试和文档提供一条真实可执行的链路
- 避免仓库只剩一个空的基础设施骨架

### 8. PostgreSQL Repository

实现：

- `app/internal/infra/store/postgres/db.go`
- `app/internal/infra/store/postgres/example_repository.go`

要求：

- 使用 PostgreSQL 和 `database/sql`
- 提供 example 的 create / get / list 持久化实现
- 不返回 HTTP DTO

### 9. 规则与开发约束

starter 需要包含：

- `.orch/rules/`
- `Makefile.rules`
- `AGENTS.md`

这些约束需要保证：

- 配置与文档保持同步
- 目录职责边界清晰
- `config` 之外禁止读取 env
- 路由变化时同步更新 HTTP contract

## Demo 领域规格

starter 使用一个名为 `example` 的 demo 资源。

### 领域模型

```go
type Example struct {
    ID        string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### HTTP 面

- `POST /v1/examples`
- `GET /v1/examples/{id}`
- `GET /v1/examples`

### Service 接口

- `Create(ctx, input)`
- `Get(ctx, id)`
- `List(ctx)`

### Repository 接口

- `Create(ctx, Example) error`
- `Get(ctx, id) (Example, error)`
- `List(ctx) ([]Example, error)`

这个领域应刻意保持简单直接，重点是展示结构，而不是展示业务复杂度。

## V1 环境变量

建议的第一版运行时配置项：

```env
APP_ENV=development

APP_DATABASE_URL=postgres://user:pass@127.0.0.1:5432/ai_go_chi_starter?sslmode=disable

APP_API_LISTEN_ADDR=:8080
APP_API_SHUTDOWN_TIMEOUT=10s

APP_WORKER_ENABLED=true
APP_WORKER_POLL_INTERVAL=5s
APP_WORKER_SHUTDOWN_TIMEOUT=10s

APP_LOG_LEVEL=info
APP_LOG_FORMAT=text
APP_LOG_OUTPUT=stdout
APP_LOG_DIR=./.runtime/logs
APP_LOG_RETENTION_DAYS=7
APP_LOG_CLEANUP_INTERVAL=1h
APP_LOG_ACCESS_ENABLED=true
APP_LOG_SOURCE_ENABLED=false
APP_LOG_OUTBOUND_ENABLED=true
APP_LOG_OUTBOUND_LEVEL=debug
APP_TIMEZONE=UTC
```

开发环境示例还应额外包含 Docker PostgreSQL 变量：

```env
DOCKER_POSTGRES_HOST_PORT=5432
DOCKER_POSTGRES_DB=ai_go_chi_starter
DOCKER_POSTGRES_USER=postgres
DOCKER_POSTGRES_PASSWORD=postgres
```

## 从 app-storage-transfer 迁移的参考范围

### 可以直接参考的部分

这些区域可以作为较安全的参考源：

- `app/internal/config/config.go`
- `app/internal/runtime/logging.go`
- `app/internal/runtime/log_file.go`
- `app/internal/runtime/outbound.go`
- `app/internal/transport/httpapi/httpx/envelope.go`
- `app/internal/transport/httpapi/httpx/access_log.go`
- `app/internal/transport/httpapi/httpx/trace.go`
- `app/internal/service/shared/error.go`
- `app/internal/service/shared/trace.go`
- `app/cmd/api/app.go`

### 必须重写，不能照搬

- `app/internal/transport/httpapi/router.go`
  - 必须改成基于 chi 的实现
- 所有 `uploads` 领域代码
- 所有 object store/provider 代码
- 所有 queue/governance 代码
- 所有 upload/minimax/staging 行为

## 实施阶段

仓库应按以下顺序实施。

### 第一阶段：Bootstrap

创建：

- `app/go.mod`
- `.gitignore`
- `README.md`
- 目录骨架
- `Makefile`

完成标准：

- 仓库目录结构已经建立
- `cd app && go test ./...` 已可执行

### 第二阶段：Config 与 Runtime

创建：

- `app/internal/config/*`
- `app/internal/runtime/*`
- env 示例文件

完成标准：

- 配置能成功加载
- logger 能写到 stdout
- config tests 通过

### 第三阶段：HTTP 基础层

创建：

- chi router
- middleware
- httpx 包
- `healthz` 和 `readyz`

完成标准：

- 服务能正常启动
- health 端点能响应
- access log 能正常输出

### 第四阶段：Demo 领域

创建：

- `app/internal/service/example/*`
- `app/internal/transport/httpapi/v1/example_handler.go`

完成标准：

- example handler 可以通过编译
- 在接数据库之前，先用 in-memory 或 stub tests 验证主流程

### 第五阶段：数据库

创建：

- migrations
- postgres 连接
- postgres repository
- migrate command

完成标准：

- migration 能成功执行
- create / get / list 能通过 postgres repository 跑通

### 第六阶段：Worker 与文档

创建：

- worker skeleton
- OpenAPI stub
- docs
- rules 和 AGENTS 契约

完成标准：

- 所有基础命令都能通过
- 不依赖口口相传也能理解仓库结构和约束

## 测试要求

第一版实现至少应包含：

### 配置测试

- 默认值加载正确
- 非法 duration 能被识别
- 非法日志格式能被识别

### HTTP 工具测试

- request id middleware
- trace middleware
- envelope / error writer
- access log 基础行为

### Service 测试

- example create 校验
- get not found
- list 行为

### Handler 测试

- create success / failure
- get success / 404
- list success

### Repository 测试

- postgres create
- postgres get
- postgres list

## Codex 实施时的规则

Codex 应遵守以下实现约束：

1. 不要把业务逻辑放进 handler。
2. 不要在 `app/internal/config` 之外读取 env。
3. 不要引入框架式 magic 或隐藏的全局状态。
4. 保持文件和函数短小。
5. 只有存在真实边界时才引入显式接口。
6. 优先使用标准库加 chi / slog / pgx，不要上更重的框架。
7. demo domain 要刻意保持最小。
8. 不要引入 upload、provider、objectstore 或 AI 专属抽象。

## 验收标准

当以下条件全部满足时，可以认为这个 starter 已可首次投入使用：

- 仓库可以启动 API 进程
- 仓库可以启动 worker 进程
- config 已集中管理并完成文档说明
- HTTP 路由基于 chi
- demo 资源能端到端工作
- 已有 OpenAPI stub
- 文档已经说明结构和启动方式
- 校验命令可以成功执行
- Codex 无需先重构目录就能继续扩展仓库

## 建议给 Codex 的执行提示词

使用这个仓库构建一个最小的、基于 chi 的 Go starter template。

实现：

- `app/cmd/api`, `app/cmd/worker`, `app/cmd/migrate`
- 在 `app/internal/config` 中集中加载配置
- 在 `app/internal/runtime` 中实现 runtime logging / trace / outbound helpers
- 在 `app/internal/transport/httpapi` 中实现 chi router 和 middleware
- 在 `app/internal/transport/httpapi/httpx` 中实现 JSON envelope / error helpers
- 实现一个最小 demo 资源 `example`，支持 create / get / list
- 实现 postgres repository 和 SQL migration
- 在 `deploy/` 下提供 env 示例
- 在 `app/openapi/` 下提供 OpenAPI stub
- 在 `docs/` 下补齐文档
- 补齐 rules 和 AGENTS 约束

约束：

- 不要引入 upload / objectstore / provider / minimax 代码
- 不要引入大框架式抽象
- 必须保持 handler -> service -> repository 的职责分离
- 必须使用 chi
- 必须使用 slog
- 所有 env 读取都必须留在 config 中

可参考以下 upload 服务仓库：

- `/Users/xihe0000/workspace/code/app-storage-transfer/app`

但只允许把它作为 runtime / config / httpx 结构模式的参考。
不要复制其中领域专属的 upload 逻辑。
