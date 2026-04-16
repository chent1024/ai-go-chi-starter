# `.orch/rules`

这个目录保存当前仓库的本地规则运行态，不是模板目录。

主要生效文件：

- `manifest.json`：定义 target 结构与受管安装布局
- `<target>/rules.env`：定义 target 级受管默认值
- `<target>/local.env`：定义仓库私有覆盖项
- `global.env`：定义仓库级校验默认值
- `scripts/*`：`make verify` / `make verify-strict` 实际执行的校验脚本
- `lint/.golangci.base.yml`：Go lint 配置

这些文件通常由 toolkit 安装脚本生成或刷新；优先使用
`rules install ... --yes`。只有在迁移旧版 `AGENTS.md` 或需要显式覆盖受管 base 文件时才使用 `--force --yes`。
根级 `rules-kit/` 不再是运行态必需目录。

日常优先修改 `manifest.json` 或 `<target>/local.env`；只有在明确调整受管默认值时才直接修改 `<target>/rules.env`。不要随意手改其他生成文件。
