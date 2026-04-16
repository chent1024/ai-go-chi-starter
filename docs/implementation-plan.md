# AI Go Chi Starter Implementation Plan

## Purpose

This repository is intended to become a Codex-friendly Go service starter template.
It should provide a clean, opinionated baseline for new Go services so Codex can start
from a stable structure instead of rebuilding infrastructure from scratch or drifting
into ad hoc project layouts.

The template should be small, explicit, and production-oriented. It should optimize for:

- predictable file placement
- clear dependency direction
- minimal but solid infrastructure
- chi-based HTTP routing
- structured logging
- centralized configuration
- basic PostgreSQL wiring
- a small demo domain that shows the intended coding style

This template should **not** become a generic application framework or a dumping ground
for unrelated reusable code.

## Source of Truth for Design

This starter should be **inspired by** the service skeleton in:

- `/Users/xihe0000/workspace/code/app-storage-transfer/app`

But it should **not** copy that service wholesale. The current upload service contains
domain-specific code that must not leak into this starter:

- upload lifecycle and status transitions
- object store abstractions
- provider adapters
- Minimax integration
- staging management
- queue/poller/purger/expirer governance
- upload-specific API surface and domain models

Only the reusable skeleton, layering ideas, and horizontal infrastructure should be
adapted into this repository.

## Target Positioning

This repository should be implemented as a **starter template repository**, not as a
shared Go library and not as a direct clone of the upload service.

Recommended outcome:

- a clean repository scaffold
- one demo HTTP resource
- one demo worker skeleton
- one demo PostgreSQL-backed repository
- one rule/verification setup
- one short AGENTS contract that tells Codex where code belongs

## Architectural Principles

The starter must enforce these boundaries:

1. `cmd/*` only wires processes together.
2. `transport` handles HTTP protocol concerns only.
3. `service` contains business rules.
4. `infra` contains concrete adapters, repositories, and clients.
5. `config` is the only place where environment variables are read.
6. `runtime` contains cross-cutting infrastructure such as logging, tracing context, and outbound logging.
7. Handlers must not contain business logic.
8. Services must not depend on chi or HTTP types.
9. Repositories must not return HTTP DTOs.
10. The starter must prefer explicit small abstractions over framework-like magic.

## Non-Goals for V1

Do **not** include these in the first version:

- object storage
- provider/client abstractions for AI vendors
- file staging or upload orchestration
- distributed job queues
- complex retry governance
- advanced authentication flows
- tenancy abstractions
- code generation pipelines beyond a basic OpenAPI stub
- generalized domain toolkit packages

## Recommended Repository Layout

The starter should be built toward this structure:

```text
ai-go-chi-starter/
├── AGENTS.md
├── Makefile
├── Makefile.rules
├── .gitignore
├── go.mod
├── README.md
├── deploy/
│   ├── .env.runtime.example
│   ├── .env.dev.example
│   ├── docker-compose.dev.yaml
│   └── README.md
├── .orch/rules/
├── cmd/
│   ├── api/
│   │   ├── main.go
│   │   └── app.go
│   ├── worker/
│   │   ├── main.go
│   │   └── app.go
│   └── migrate/
│       └── main.go
├── db/
│   └── migrations/
│       └── 001_init.sql
├── docs/
│   ├── implementation-plan.md
│   ├── architecture.md
│   ├── config.md
│   └── api.md
├── openapi/
│   └── openapi.yaml
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   ├── runtime/
│   │   ├── logging.go
│   │   ├── log_file.go
│   │   ├── log_fields.go
│   │   ├── trace.go
│   │   ├── log_context.go
│   │   └── outbound.go
│   ├── transport/
│   │   └── httpapi/
│   │       ├── router.go
│   │       ├── middleware/
│   │       │   ├── request_id.go
│   │       │   ├── trace.go
│   │       │   ├── access_log.go
│   │       │   └── recover.go
│   │       ├── httpx/
│   │       │   ├── envelope.go
│   │       │   ├── errors.go
│   │       │   ├── request_context.go
│   │       │   └── response_recorder.go
│   │       └── v1/
│   │           └── example_handler.go
│   ├── service/
│   │   ├── shared/
│   │   │   ├── error.go
│   │   │   ├── trace.go
│   │   │   └── ids.go
│   │   └── example/
│   │       ├── model.go
│   │       ├── service.go
│   │       └── repository.go
│   └── infra/
│       ├── store/
│       │   └── postgres/
│       │       ├── db.go
│       │       └── example_repository.go
│       └── client/
│           └── httpclient.go
└── .runtime/
```

## Required V1 Capabilities

The first usable version should include the following:

### 1. API Process Skeleton

Implement:

- `cmd/api/main.go`
- `cmd/api/app.go`

Capabilities:

- load config
- construct logger
- construct HTTP server
- graceful shutdown
- return non-zero on startup failure

### 2. Worker Process Skeleton

Implement:

- `cmd/worker/main.go`
- `cmd/worker/app.go`

Capabilities:

- load config
- construct logger
- start ticker loop
- demonstrate a `JobHandler` interface
- graceful shutdown

This worker should be intentionally small and generic. It should not bring over the
upload service governance machinery.

### 3. Migration Process

Implement:

- `cmd/migrate/main.go`

Capabilities:

- connect to PostgreSQL
- run SQL migrations from `db/migrations`

### 4. Configuration System

Implement:

- `internal/config/config.go`
- `internal/config/config_test.go`

Requirements:

- environment-only configuration
- centralized parsing
- explicit defaults
- startup validation
- no environment reads outside `internal/config`

### 5. Logging and Runtime Cross-Cutting Layer

Implement:

- `internal/runtime/logging.go`
- `internal/runtime/log_file.go`
- `internal/runtime/log_fields.go`
- `internal/runtime/log_context.go`
- `internal/runtime/trace.go`
- `internal/runtime/outbound.go`

Requirements:

- `slog`
- text/json format
- stdout/file/both output
- daily log file rotation by date
- basic retention cleanup
- trace-aware logger enrichment
- outbound request logging helper

### 6. HTTP Layer with chi

Implement:

- `internal/transport/httpapi/router.go`
- `internal/transport/httpapi/middleware/*`
- `internal/transport/httpapi/httpx/*`

Requirements:

- use `github.com/go-chi/chi/v5`
- middleware order:
  1. recover
  2. request id
  3. trace
  4. access log
- expose:
  - `GET /healthz`
  - `GET /readyz`
  - `POST /v1/examples`
  - `GET /v1/examples/{id}`
  - `GET /v1/examples`

### 7. Demo Domain

Implement one minimal domain named `example`.

Purpose:

- show the intended handler/service/repository split
- provide a real executable path for tests and docs
- avoid shipping an empty infrastructure-only skeleton

### 8. PostgreSQL Repository

Implement:

- `internal/infra/store/postgres/db.go`
- `internal/infra/store/postgres/example_repository.go`

### 9. Rules and Developer Guardrails

The starter should include:

- `.orch/rules/`
- `Makefile.rules`
- `AGENTS.md`

These should enforce:

- config/docs sync discipline
- directory ownership
- no env reads outside config
- HTTP contract updates when routes change

## Demo Domain Specification

The starter should use a single demo resource named `example`.

### Domain model

```go
type Example struct {
    ID        string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### HTTP surface

- `POST /v1/examples`
- `GET /v1/examples/{id}`
- `GET /v1/examples`

### Service interface

- `Create(ctx, input)`
- `Get(ctx, id)`
- `List(ctx)`

### Repository interface

- `Create(ctx, Example) error`
- `Get(ctx, id) (Example, error)`
- `List(ctx) ([]Example, error)`

This domain should remain deliberately boring. The goal is to teach shape, not domain
complexity.

## Environment Variables for V1

Recommended initial runtime keys:

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

Development-only example env should additionally include Docker PostgreSQL variables:

```env
DOCKER_POSTGRES_HOST_PORT=5432
DOCKER_POSTGRES_DB=ai_go_chi_starter
DOCKER_POSTGRES_USER=postgres
DOCKER_POSTGRES_PASSWORD=postgres
```

## Migration Plan from app-storage-transfer

### Safe to adapt directly

These areas are good reference sources:

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

### Must be rewritten, not copied

- `app/internal/transport/httpapi/router.go`
  - must be rebuilt for chi
- all `uploads` domain code
- all object store/provider code
- all queue/governance code
- all upload/minimax/staging behavior

## Implementation Phases

The repository should be implemented in this order.

### Phase 1: Bootstrap

Create:

- `go.mod`
- `.gitignore`
- `README.md`
- directory skeleton
- `Makefile`

Success criteria:

- repository layout exists
- `go test ./...` is runnable, even if minimal

### Phase 2: Config and Runtime

Create:

- `internal/config/*`
- `internal/runtime/*`
- env example files

Success criteria:

- config loads successfully
- logger can write to stdout
- config tests pass

### Phase 3: HTTP Foundation

Create:

- chi router
- middleware
- httpx package
- `healthz` and `readyz`

Success criteria:

- server boots
- health endpoints respond
- access log is emitted

### Phase 4: Demo Domain

Create:

- `service/example/*`
- `transport/httpapi/v1/example_handler.go`

Success criteria:

- example handlers compile
- in-memory or stub tests validate flow before DB wiring

### Phase 5: Database

Create:

- migrations
- postgres connection
- postgres repository
- migrate command

Success criteria:

- migration applies
- create/get/list works through postgres repository

### Phase 6: Worker and Docs

Create:

- worker skeleton
- OpenAPI stub
- docs
- rules and AGENTS contract

Success criteria:

- all baseline commands pass
- repository is understandable without tribal knowledge

## Testing Requirements

The first implementation should include:

### Config tests

- valid default loading
- invalid duration handling
- invalid log format handling

### HTTP utility tests

- request id middleware
- trace middleware
- envelope/error writer
- access log basic behavior

### Service tests

- example create validation
- get not found
- list behavior

### Handler tests

- create success/failure
- get success/404
- list success

### Repository tests

- postgres create
- postgres get
- postgres list

## Rules for Codex During Implementation

Codex should follow these implementation constraints:

1. Do not place business logic in handlers.
2. Do not read env vars outside `internal/config`.
3. Do not introduce framework-like magic or hidden global state.
4. Keep files and functions small.
5. Use explicit interfaces only where there is a real boundary.
6. Prefer the standard library plus chi/slog/pgx over larger frameworks.
7. Keep the demo domain intentionally minimal.
8. Do not introduce upload, provider, objectstore, or AI-specific abstractions.

## Acceptance Criteria

The starter is ready for first use when all of the following are true:

- the repository boots an API process
- the repository boots a worker process
- config is centralized and documented
- HTTP routes use chi
- the demo resource works end-to-end
- OpenAPI stub exists
- docs explain structure and setup
- verification commands run successfully
- Codex can extend the repository without first reorganizing it

## Suggested Codex Execution Prompt

Use this repository to build a minimal chi-based Go starter template.

Implement:

- `cmd/api`, `cmd/worker`, `cmd/migrate`
- centralized config loading in `internal/config`
- runtime logging/trace/outbound helpers in `internal/runtime`
- chi router and middleware in `internal/transport/httpapi`
- JSON envelope/error helpers in `internal/transport/httpapi/httpx`
- one minimal demo resource `example` with create/get/list
- postgres repository and SQL migration
- env examples under `deploy/`
- OpenAPI stub under `openapi/`
- docs under `docs/`
- rules and AGENTS guardrails

Constraints:

- no upload/objectstore/provider/minimax code
- no large framework abstractions
- handler -> service -> repository separation is mandatory
- chi is required
- slog is required
- all env reads must stay inside config

Use the upload service repository at:

- `/Users/xihe0000/workspace/code/app-storage-transfer/app`

as a structural reference only for reusable runtime/config/httpx patterns.
Do not copy its domain-specific upload logic.
