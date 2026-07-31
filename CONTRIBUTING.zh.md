# 为 mihomo-exporter 贡献

[English](CONTRIBUTING.md)

感谢你的贡献。欢迎提交缺陷报告、文档修正、测试和聚焦的改进。

## 开始之前

- 创建新 issue 或 pull request 前，请先搜索现有内容。
- 大型功能或行为变更请先创建 issue，确认范围后再实施。
- 保持变更聚焦。不要在同一个 pull request 中混入无关重构、依赖更新和功能变更。

## 开发环境

需要 Go 1.26、Make 和 Git。Node.js 为可选项，仅用于安装 agent skills。

```bash
make setup
make test
```

所有开发命令均由 Make target 提供。使用 `make help` 查看完整列表，不要用临时命令替代仓库检查。

## 编写变更

- 添加依赖前优先使用标准库。新的库或部署工具应先在 issue 中讨论。
- `cmd/mihomo-exporter` 仅负责装配和退出处理；具体行为应放在职责明确的 `internal/` 包中。
- 显式传递依赖，避免全局可变状态。
- 保持 Prometheus label 有界。不要添加按连接划分的主机或目标 label。
- `Collect()` 中不得执行网络或文件系统 I/O；它只能从最新的缓存快照输出指标。
- 不要提交凭据、`.env` 文件、生成的二进制文件或真实的 Mihomo API secret。

## 测试与验证

行为变更应添加或更新测试。适用时使用具名的表驱动用例；测试中不要使用 `time.Sleep`。

创建 pull request 前，请运行：

```bash
make check
make vuln
```

`make check` 会执行格式检查、模块整洁度检查、lint、启用 race detector 的测试和构建。不要为了通过检查而禁用 linter 或削弱断言。

## 提交与 Pull Request

使用带有匹配 emoji 前缀的 Conventional Commits。例如：

```text
✨ feat(collector): expose a new bounded metric
🐛 fix(mihomo): handle an API timeout
📚 docs: clarify Docker Compose deployment
```

提交信息应使用祈使语气、小写且不带句号。每个提交应尽可能可以独立构建。

Pull request 应说明问题、行为变更、已执行的验证，以及对配置、指标或部署的影响。用户可见行为变更时，请同步更新英文和中文文档。

## 行为准则

请在 issue、pull request 和 review 中保持尊重、建设性和体谅。如遇不可接受的行为，请私下联系仓库维护者。
