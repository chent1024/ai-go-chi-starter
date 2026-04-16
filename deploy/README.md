# 部署说明

## 本地开发

1. 如有需要，先将 `deploy/.env.dev.example` 复制为本地 env 文件。
2. 启动 PostgreSQL：

```bash
make dev-up
```

3. 执行 migration：

```bash
make migrate
```

4. 在两个终端里分别启动 API 和 worker：

```bash
make run-api
make run-worker
```

## 常用 Make 入口

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

## 运行时环境变量

`deploy/.env.runtime.example` 只包含应用运行时所需的配置项。  
`deploy/.env.dev.example` 在此基础上额外加入了 Compose 使用的 Docker 本地 PostgreSQL 配置项。
