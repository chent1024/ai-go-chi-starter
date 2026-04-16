# Constraint Starter Kit

这一套是完整骨架，适合新项目一开始就建立“软规则 + 硬校验 + CI + review 边界”。

## 目录说明

- `Makefile.template`
  - 如何接入统一的 verify 目标
- `bootstrap.sh`
  - 从当前仓库的模板导出一份独立骨架目录
- `rules.manifest`
  - 导出的文件清单

## 建议使用方式

1. 在当前仓库根目录运行 `bash rules-kit/install.sh /tmp/constraint-starter starter`。
2. 把生成出来的目录复制到目标仓库根目录。
3. 将 `Makefile.template` 中的约束入口合并到目标仓库 `Makefile`。
4. 如果平台支持仓库级软规则，可在平台侧直接配置对应文本。
5. 调整 target 结构时修改 `.orch/rules/manifest.json`；仓库私有路径或 env 覆盖优先写 `.orch/rules/<target>/local.env`；只有在明确调整受管默认值时才直接修改 `.orch/rules/<target>/rules.env`。
6. 再按需要逐步启用 `verify-arch`、`verify-contract-sync`、`verify-migrations`。

保留 `bootstrap.sh` 只是为了兼容前一版用法；新的统一入口是 `rules-kit/install.sh`。

如果你希望直接安装到目标仓库，可以改用：

```bash
bash rules-kit/install-into-repo.sh /path/to/repo starter
```

安装完成后建议再执行：

```bash
cd /path/to/repo
bash rules-kit/doctor.sh . --strict
```

如果想把最常见的接线动作自动补齐，可以先运行：

```bash
bash rules-kit/fixup.sh .
```

如果目标目录还没有任何 Go 服务骨架，直接使用：

```bash
bash rules-kit/init-go-service.sh /path/to/new-repo github.com/acme/new-repo
bash rules-kit/init-ts-service.sh /path/to/new-repo @acme/new-repo
```

## 最适合的项目阶段

- 已经有明确配置入口
- 已经有接口契约目录或 API schema
- 有 CI 和受保护分支
- 希望把约束从 agent 文本转成脚本和 review 规则
