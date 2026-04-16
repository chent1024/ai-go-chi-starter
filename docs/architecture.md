# Architecture

## Layers

- `cmd/*`: process entrypoints only
- `internal/transport/httpapi`: HTTP protocol and middleware only
- `internal/service`: business rules and domain services
- `internal/infra`: concrete adapters such as PostgreSQL and outbound clients
- `internal/runtime`: cross-cutting logging, trace, and outbound logging
- `internal/config`: the only env-loading package
- outbound HTTP clients share one transport profile: timeout, keep-alive pool, trace propagation, and outbound logging are configured centrally

## Process Wiring

### API

`cmd/api` loads config, creates the logger, opens PostgreSQL with an explicit pool configuration, wires the example service and handler, then starts the chi server with graceful shutdown, configured HTTP timeouts, and an explicit drain state that rejects new application requests during shutdown.

### Worker

`cmd/worker` loads config, creates the logger, and runs a minimal ticker loop through a `JobHandler` interface. Shutdown logs include drain start, inflight job count, and drain completion.

### Migrate

`cmd/migrate` loads config, opens PostgreSQL, acquires a PostgreSQL advisory lock, and applies forward-only SQL files from `db/migrations`.

## HTTP Cross-Cutting

- request ID middleware guarantees `X-Request-Id`
- trace middleware guarantees `Traceparent`
- request timeout middleware enforces a per-request context deadline
- drain middleware rejects new application requests once graceful shutdown starts
- access log middleware records route, latency, bytes, and status
- recover middleware converts panics into the standard error envelope

## Domain Example

The `example` resource demonstrates the intended shape:

- transport decodes and encodes DTOs
- service validates input and generates IDs
- repository handles SQL and returns domain models

This keeps handler, service, and repository responsibilities stable and easy for Codex to extend.
