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
- `make release-build`
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

补充说明：

- `APP_API_MAX_BODY_BYTES` 用于限制单次 HTTP 请求体大小，默认 `1048576`
- `/metrics` 输出 Prometheus 文本格式指标
- `/version` 输出当前 API build 信息

## 发布构建

- 发布构建统一使用 `make release-build`
- 产物固定输出到 `dist/<VERSION>/<TARGET_OS>-<TARGET_ARCH>/`
- `VERSION`、`COMMIT`、`BUILD_TIME` 会同步注入 `/version`

示例：

```bash
make release-build VERSION=v0.1.0 TARGET_OS=linux TARGET_ARCH=amd64
```

## 本地验证

- `make test-integration`
  使用本地 Docker PostgreSQL 跑 postgres repository integration test，并在测试前后自动拉起/清理本地数据库。
- `make smoke`
  跑一轮最小端到端 smoke test：本地 PostgreSQL、migration、API、worker、healthz/readyz 和 example 的 create/get/list。
