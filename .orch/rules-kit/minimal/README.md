# Minimal Constraint Bundle

这一套是最小可复制版，目标是让一个新仓库用最少文件先建立可执行约束。

## 包含内容

- `rules-kit/doctor.sh`
- `rules-kit/fixup.sh`
- `Makefile.snippet`
- `repo/rules.env`
- `scripts/verify-config-docs.sh`
- `scripts/verify-generated.sh`
- `scripts/verify-secrets.sh`
- `ci/verify.yml`
- `review/CODEOWNERS.template`

## 适用场景

- 新仓库还没有完整分层规则
- 先要把配置漂移、generated 文件误改、secret 泄露挡住
- 只想接入一条主验证流水线，不想一次性上太多脚本

## 接入步骤

1. 复制整个 `minimal/` 目录到目标仓库。
2. 如果只是仓库私有路径或 env 覆盖，优先写到 `repo/local.env`；只有在明确调整受管默认值时才直接修改 `repo/rules.env`。
3. 把 `Makefile.snippet` 合并到目标仓库 `Makefile`。
4. 如果你的平台支持仓库级软规则，可在平台侧直接配置对应文本。
5. 按 CI 平台改写 `ci/verify.yml`。
6. 按团队职责改写 `review/CODEOWNERS.template`。

也可以直接安装：

```bash
bash rules-kit/install-into-repo.sh /path/to/repo minimal
```

安装后建议检查一次：

```bash
cd /path/to/repo
bash rules-kit/doctor.sh .
```

如果仓库里还没有 `Makefile`，可以再执行：

```bash
bash rules-kit/fixup.sh .
```

如果这是一个全新 Go 仓库，建议直接改用 `go-starter` profile 配合：

```bash
bash rules-kit/init-go-service.sh /path/to/new-repo github.com/acme/new-repo
```

如果是全新 TypeScript 仓库，对应使用：

```bash
bash rules-kit/init-ts-service.sh /path/to/new-repo @acme/new-repo
```

## 故意没放进去的内容

- `verify-arch.sh`
- `verify-contract-sync.sh`
- `verify-migrations.sh`
- lint 模板

这些规则更依赖仓库具体结构，建议在最小版稳定后再加。
