# json2struct

[![Go Reference](https://pkg.go.dev/badge/github.com/EmptyZeroRain/json2struct.svg)](https://pkg.go.dev/github.com/EmptyZeroRain/json2struct)
[![CI](https://github.com/EmptyZeroRain/json2struct/actions/workflows/ci.yml/badge.svg)](https://github.com/EmptyZeroRain/json2struct/actions/workflows/ci.yml)

一个仅依赖 Go 标准库、兼容 Go 1.22+ 的 JSON/NDJSON 到 Go struct 推断库。

仅使用 Go 标准库，将 JSON/NDJSON 样本推断为可编译的 Go struct。

## 快速开始

假设项目目录下有两个 JSON 文件：

```text
data/user-1.json
data/user-2.json
```

`data/user-1.json`：

```json
{"id":1,"name":"Ada"}
```

`data/user-2.json`：

```json
{"id":2,"email":"ada@example.com"}
```

```go
package main

import (
    "fmt"
    "log"

    json2struct "github.com/EmptyZeroRain/json2struct"
)

func main() {
    g := json2struct.New(json2struct.Options{
        Name: "User",
        Package: "models",
        Merge: true,
        Omitempty: true,
    })

    if err := g.AddFile("data/user-1.json"); err != nil {
        log.Fatal(err)
    }
    if err := g.AddFile("data/user-2.json"); err != nil {
        log.Fatal(err)
    }

    code, err := g.Generate()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(string(code))
}
```

也可以使用 `os.Open` 后传入 `AddReader`：

```go
file, err := os.Open("data/user-1.json")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

if err := g.AddReader(file); err != nil {
    log.Fatal(err)
}
```

`AddReader` 支持任意 JSON Reader，`AddFile` 支持文件，`AddNDJSON` 按行流式读取并合并。`Schema()` 返回可修改的公开 schema，修改后使用 `GenerateFromSchema` 生成代码。运行 `go test ./...` 验证库。

大批量样本可以并行推断后合并：

```go
schema, err := json2struct.InferBatch(samples, json2struct.BatchOptions{Workers: 8})
if err != nil { log.Fatal(err) }
code, err := json2struct.New(json2struct.Options{Name: "Record"}).GenerateFromSchema(schema)
```

生产环境可使用 `parser.Options` 中的 `option.Limits` 限制最大深度、字段数、单行大小和输入大小，避免异常数据消耗过多资源。

本项目只提供 Go Library，不包含 CLI。使用方可直接导入模块并调用上述 API。

## 安装

```bash
go get github.com/EmptyZeroRain/json2struct
```

```go
import json2struct "github.com/EmptyZeroRain/json2struct"
```

## 项目结构

```text
.
├── generator/       Go 源码生成
├── inference/       JSON 类型推断
├── option/          配置项
├── parser/          JSON/NDJSON 解析
├── schema/          Schema 模型与合并
├── json2struct.go   对外主 API
├── LICENSE
└── .github/workflows/ci.yml
```

## 许可证

本项目使用 MIT License，详见 [LICENSE](LICENSE)。

类型推断采用保守策略：JSON 字符串（包括时间、日期、时间戳）和数字统一生成为 `string`，不会擅自转换为 `time.Time`、`int` 或 `float64`，避免 ID 精度损失。无法识别的直接输入值也默认生成为 `string`；多样本之间发生明确类型冲突时才使用 `interface{}`。
