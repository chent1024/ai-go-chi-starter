# 部署说明

## 本地开发

推荐顺序：

1. 参考 `deploy/.env.dev.example` 准备本地开发环境变量。
2. 启动 PostgreSQL：

```bash
make dev-up
```

3. 执行 migration：

```bash
make migrate
```

4. 启动 API：

```bash
make run-api
```

5. 如需验证 worker 行为，再启动 worker：

```bash
make run-worker
```

开发建议：

- 只调 HTTP API 时，通常只需要 `make dev-up`、`make migrate`、`make run-api`
- 只有联调后台任务或 drain 行为时，再额外启动 `make run-worker`
- 修改配置后，优先对照 `docs/app/config.md` 检查可选值和风险项

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

## 联调与排障

推荐命令：

- `make dev-logs`
  持续观察本地 PostgreSQL 容器日志
- `make dev-ps`
  查看本地 Compose 服务状态
- `make migrate-version`
  快速确认当前 migration 版本
- `make smoke`
  跑一轮最小端到端链路，适合改完路由、migration、repository 之后做回归
- `make test-integration`
  只验证 PostgreSQL repository integration 路径

常见场景：

- 宿主机 `5432` 已占用时，只改 `deploy/.env.dev.example` 里的 `DOCKER_POSTGRES_HOST_PORT`
- API 能启动但数据库连不上时，优先核对 `APP_DATABASE_URL` 是否与 `deploy/.env.dev.example` 里的 Docker 端口一致
- 改了 schema 或 repository 行为后，先跑 `make migrate`，再跑 `make test-integration` 或 `make smoke`

## 运行时环境变量

`deploy/.env.runtime.example` 只包含应用运行时所需的配置项。  
`deploy/.env.dev.example` 在此基础上额外加入了 Compose 使用的 Docker 本地 PostgreSQL 配置项。

补充说明：

- `APP_API_MAX_BODY_BYTES` 用于限制单次 HTTP 请求体大小，默认 `1048576`
- `/metrics` 输出 Prometheus 文本格式指标
- `/version` 输出当前 API build 信息

## 发布构建

推荐做法：

- 发布构建统一使用 `make release-build`
- 产物固定输出到 `app/dist/<VERSION>/<TARGET_OS>-<TARGET_ARCH>/`
- `VERSION`、`COMMIT`、`BUILD_TIME` 会同步注入 `/version`

常见用法：

- 本地快速验证二进制：`make build`
- 需要指定版本号和目标平台时：`make release-build VERSION=v0.1.0 TARGET_OS=linux TARGET_ARCH=amd64`
- CI/CD 中建议显式传入 `VERSION`、`COMMIT`、`BUILD_TIME`，避免使用默认值

示例：

```bash
make release-build VERSION=v0.1.0 TARGET_OS=linux TARGET_ARCH=amd64
```

## 本地验证

- `make test-integration`
  使用本地 Docker PostgreSQL 跑 postgres repository integration test，并在测试前后自动拉起/清理本地数据库。
- `make smoke`
  跑一轮最小端到端 smoke test：本地 PostgreSQL、migration、API、worker、healthz/readyz 和 example 的 create/get/list。

建议：

- 只改 repository / migration 时，优先跑 `make test-integration`
- 改了 HTTP 路由、运行时 wiring、worker 协作或端到端链路时，再跑 `make smoke`
