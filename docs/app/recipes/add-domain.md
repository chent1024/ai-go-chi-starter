# Recipe：新增一个 Domain

这个模板推荐用“复制现有模式”的方式扩展新资源，而不是先抽象一层框架。

以下用 `<domain>` 表示新资源名。

## 需要新增或修改的文件

1. `app/internal/service/<domain>/model.go`
2. `app/internal/service/<domain>/repository.go`
3. `app/internal/service/<domain>/service.go`
4. 可选：`app/internal/service/<domain>/error_codes.go`
5. `app/db/queries/<domain>.sql`
6. `app/sqlc.yaml` 通常不需要改，除非引入新的生成分组
7. `app/internal/infra/store/postgres/sqlc/*` 通过 `make sqlc-generate` 生成
8. `app/internal/infra/store/postgres/<domain>_repository.go`
9. `app/internal/transport/httpapi/v1/<domain>_handler.go`
10. `app/internal/transport/httpapi/router.go`
11. `app/db/migrations/<nnn>_<domain>.sql`
12. `docs/app/api.md`
13. `app/openapi/openapi.yaml`
14. 对应的 service / handler / repository tests

## 分层要求

- handler 只做 request decode、response encode、错误写回
- service 负责业务规则、输入校验、领域错误
- repository interface 放在 `app/internal/service/<domain>`
- SQL 查询定义放在 `app/db/queries/*.sql`
- `sqlc` 生成代码放在 `app/internal/infra/store/postgres/sqlc`
- postgres 实现放在 `app/internal/infra/store/postgres`，但这里只保留薄包装，不直接写 SQL 执行
- transport DTO 不能穿透到 service / repository

## 推荐顺序

1. 先定义 `model.go`
2. 定义 `repository.go` 接口
3. 实现 `service.go`
4. 编写 `app/db/queries/<domain>.sql`
5. 执行 `make sqlc-generate`
6. 实现 postgres repository 薄包装
7. 实现 handler
8. 在 router 注册路由
9. 加 migration
10. 同步 API 文档和 OpenAPI
11. 补测试

## 配置和契约同步

- 如果新增 env，必须只在 `app/internal/config/config.go` 读取
- 如果新增路由、字段、错误码，必须同步文档和 OpenAPI
- 如果改了 `app/db/queries/*.sql` 或 `app/sqlc.yaml`，必须重新执行 `make sqlc-generate`
- 如果改了 migration、runtime wiring、存储契约，必须跑 `make verify-strict`

## 最终验证

- 每次改完查询定义后先跑：`make sqlc-generate`
- 普通 domain 改动：`make verify`
- 含存储、migration、runtime 或 API 契约变更：`make verify-strict`
