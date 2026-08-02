# Changelog

## Unreleased

后续变更写在这里，发布时移动到对应版本。

## [0.1.5] - 2026-08-02

### Added

- Token 级 JSON Schema 推断，减少中间反序列化树。
- Context 取消 API：`ParseWithContext`、`ParseNDJSONParallelContext`。
- `MaxTotalBytes`、`MaxSamples`、`MaxSchemaNodes` 限制。
- 数组采样：`SampleArrayItems`。

### Changed

- 批量 Schema 使用树形合并。
- 资源限制在解析阶段执行。
- 支持超时、主动取消和 Worker 快速退出。

## [0.1.4] - 2026-08-02

### Added

- `MaxStringBytes` 和 `MaxNumberBytes` 输入限制。
- `schema.MergeAll` 平衡树批量合并。
- 并行 NDJSON 按输入序号确定性合并。

### Fixed

- 降低普通 JSON Reader 解析时的额外整包内存拷贝。
- 修复并行 NDJSON Worker 在生产者或消费者异常时的清理流程。
- 升级 CodeQL Action 并固定 gosec Action 版本。

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
