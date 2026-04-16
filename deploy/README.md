# Deploy Notes

## Local Development

1. Copy `deploy/.env.dev.example` to a local env file if needed.
2. Start PostgreSQL:

```bash
make dev-up
```

3. Run migrations:

```bash
make migrate
```

4. Start API and worker in separate shells:

```bash
make run-api
make run-worker
```

## Common Make Targets

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

## Runtime Env

`deploy/.env.runtime.example` contains only the application runtime keys.
`deploy/.env.dev.example` adds the Docker-only PostgreSQL keys used by Compose.
