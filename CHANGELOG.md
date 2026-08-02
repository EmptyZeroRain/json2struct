# Changelog

## Unreleased

后续变更写在这里，发布时移动到对应版本。

## [0.1.3] - 2026-08-02

### Fixed

- 修复并行 NDJSON 解析在 Scanner 读取失败时 Worker 清理不完整的问题。
- 固定 gosec Action 版本，并明确忽略 `AddFile` 的受控文件路径检查告警。

## [0.1.2] - 2026-08-02

### Added

- `MaxNodes` 总节点资源限制。
- 非递归 JSON 资源校验，降低深层输入栈风险。
- 有界并行 NDJSON 解析：`ParseNDJSONParallel`、`AddNDJSONParallel`。
- 可使用 `errors.Is` 判断的解析限制错误。
- CodeQL 和 gosec GitHub Actions 安全扫描。

### Fixed

- Merge 模式不再保留每个输入 Schema，降低长期批量处理的内存增长。
- 并行 NDJSON 解析发生错误时及时停止 Worker，避免 Goroutine 泄漏。

## [0.1.1] - 2026-08-02

### Added

- 批量并行 JSON 推断：`InferBatch`、`AddBatchWithOptions`。
- NDJSON 流式解析和输入资源限制。
- Schema 增量合并：`schema.MergeInto`。
- 并发安全的 Generator、Fuzz、Race 和 Benchmark 测试。

### Fixed

- 修复对象、数组和标量类型冲突时的 Schema 误判。
- 修复特殊 JSON Key 导致生成 Go Tag 非法的问题。
- 修复 Windows CI 的 gofmt 检查。

### Changed

- 模块兼容 Go 1.22 及以上版本。
- 项目保持零第三方依赖。
