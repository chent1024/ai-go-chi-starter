# Architecture

## Layers

- `cmd/*`: process entrypoints only
- `internal/transport/httpapi`: HTTP protocol and middleware only
- `internal/service`: business rules and domain services
- `internal/infra`: concrete adapters such as PostgreSQL and outbound clients
- `internal/runtime`: cross-cutting logging, trace, and outbound logging
- `internal/config`: the only env-loading package

## Process Wiring

### API

`cmd/api` loads config, creates the logger, opens PostgreSQL, wires the example service and handler, then starts the chi server with graceful shutdown.

### Worker

`cmd/worker` loads config, creates the logger, and runs a minimal ticker loop through a `JobHandler` interface. It intentionally does not include queue governance.

### Migrate

`cmd/migrate` loads config, opens PostgreSQL, and applies forward-only SQL files from `db/migrations`.

## HTTP Cross-Cutting

- request ID middleware guarantees `X-Request-Id`
- trace middleware guarantees `Traceparent`
- access log middleware records route, latency, bytes, and status
- recover middleware converts panics into the standard error envelope

## Domain Example

The `example` resource demonstrates the intended shape:

- transport decodes and encodes DTOs
- service validates input and generates IDs
- repository handles SQL and returns domain models

This keeps handler, service, and repository responsibilities stable and easy for Codex to extend.
