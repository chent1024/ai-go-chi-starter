# Deploy Notes

## Local Development

1. Copy `deploy/.env.dev.example` to a local env file if needed.
2. Start PostgreSQL:

```bash
docker compose -f deploy/docker-compose.dev.yaml --env-file deploy/.env.dev.example up -d
```

3. Run migrations:

```bash
go run ./cmd/migrate
```

4. Start API and worker in separate shells:

```bash
go run ./cmd/api
go run ./cmd/worker
```

## Runtime Env

`deploy/.env.runtime.example` contains only the application runtime keys.
`deploy/.env.dev.example` adds the Docker-only PostgreSQL keys used by Compose.
