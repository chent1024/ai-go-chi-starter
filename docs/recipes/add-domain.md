# Recipe：新增一个 Domain

这个模板推荐用“复制现有模式”的方式扩展新资源，而不是先抽象一层框架。

以下用 `<domain>` 表示新资源名。

## 需要新增或修改的文件

1. `internal/service/<domain>/model.go`
2. `internal/service/<domain>/repository.go`
3. `internal/service/<domain>/service.go`
4. 可选：`internal/service/<domain>/error_codes.go`
5. `internal/infra/store/postgres/<domain>_repository.go`
6. `internal/transport/httpapi/v1/<domain>_handler.go`
7. `internal/transport/httpapi/router.go`
8. `db/migrations/<nnn>_<domain>.sql`
9. `docs/api.md`
10. `openapi/openapi.yaml`
11. 对应的 service / handler / repository tests

## 分层要求

- handler 只做 request decode、response encode、错误写回
- service 负责业务规则、输入校验、领域错误
- repository interface 放在 `internal/service/<domain>`
- postgres 实现放在 `internal/infra/store/postgres`
- transport DTO 不能穿透到 service / repository

## 推荐顺序

1. 先定义 `model.go`
2. 定义 `repository.go` 接口
3. 实现 `service.go`
4. 实现 postgres repository
5. 实现 handler
6. 在 router 注册路由
7. 加 migration
8. 同步 API 文档和 OpenAPI
9. 补测试

## 配置和契约同步

- 如果新增 env，必须只在 `internal/config/config.go` 读取
- 如果新增路由、字段、错误码，必须同步文档和 OpenAPI
- 如果改了 migration、runtime wiring、存储契约，必须跑 `make verify-strict`

## 最终验证

- 普通 domain 改动：`make verify`
- 含存储、migration、runtime 或 API 契约变更：`make verify-strict`
