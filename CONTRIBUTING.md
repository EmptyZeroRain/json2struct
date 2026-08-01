# Contributing

感谢贡献代码。请在提交 Pull Request 前：

1. 为新功能或 bug 修复补充测试。
2. 运行 `gofmt -w .`。
3. 运行 `go test ./...`、`go test -race ./...` 和 `go vet ./...`。
4. 保持项目仅依赖 Go 标准库。
5. PR 描述中说明变更内容、兼容性影响和验证命令。

请不要提交密钥、真实业务数据或 IDE 配置文件。
