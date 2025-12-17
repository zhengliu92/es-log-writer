# es-log-writer

一个用于将日志同步写入 Elasticsearch 的 Golang 工具库。

## 功能特性

- ✅ **核心库不依赖 go-zero**，可独立使用
- ✅ 提供 go-zero `logx.Writer` 适配器（可选）
- ✅ 支持批量写入，提高性能
- ✅ 自动按日期创建索引（格式：`{prefix}-{YYYY.MM.DD}`）
- ✅ 支持缓冲区刷新机制，可配置刷新间隔和缓冲区大小
- ✅ 支持 Elasticsearch 认证（用户名密码或 API Key）
- ✅ 支持 trace/span 字段
- ✅ 线程安全，支持并发写入
- ✅ 异步写入，不阻塞业务代码
- ✅ 优雅关闭，确保所有日志都被写入
- ✅ 提供 `MultiWriter`，支持同时输出到多个目标（控制台 + Elasticsearch）
- ✅ 提供 `ConsoleWriter`，支持控制台输出

## 安装

```bash
go get github.com/zhengliu92/es-log-writer
```

## 使用方式

### 方式一：独立使用（不依赖 go-zero）

```go
package main

import (
    "time"
    writer "github.com/zhengliu92/es-log-writer"
)

func main() {
    config := &writer.Config{
        Addresses:     []string{"http://localhost:9200"},
        IndexPrefix:   "app-logs",
        BufferSize:    100,
        FlushInterval: 5 * time.Second,
    }
    
    w, err := writer.NewElasticsearchWriter(config)
    if err != nil {
        panic(err)
    }
    defer w.Close()
    
    // 直接使用
    w.Info("用户登录成功")
    w.Error("数据库连接失败")
    w.Info("请求处理完成",
        writer.Field("duration", 50*time.Millisecond),
        writer.Field("trace", "abc123"),
        writer.Field("user_id", 12345),
    )
}
```

### 方式二：配合 go-zero 使用（需要导入适配器）

```go
package main

import (
    "time"
    "github.com/zeromicro/go-zero/core/logx"
    writer "github.com/zhengliu92/es-log-writer"
    logxadapter "github.com/zhengliu92/es-log-writer/logx"  // 适配器
)

func main() {
    config := &writer.Config{
        Addresses:     []string{"http://localhost:9200"},
        IndexPrefix:   "go-zero-logs",
        BufferSize:    100,
        FlushInterval: 5 * time.Second,
    }
    
    // 创建 logx 适配器
    adapter, err := logxadapter.NewAdapter(config)
    if err != nil {
        panic(err)
    }
    defer adapter.Close()
    
    // 设置 logx 使用 Elasticsearch Writer
    logx.SetWriter(adapter)
    
    // 正常使用 logx
    logx.Info("日志内容")
    logx.Infow("带字段的日志",
        logx.Field("trace", "abc123"),
        logx.Field("duration", 20*time.Millisecond),
    )
}
```

### 方式三：同时输出到控制台和 Elasticsearch（推荐用于开发环境）

使用 `MultiWriter` 可以同时将日志输出到控制台和 Elasticsearch，方便开发调试：

```go
package main

import (
    "time"
    "github.com/zeromicro/go-zero/core/logx"
    writer "github.com/zhengliu92/es-log-writer"
    logxadapter "github.com/zhengliu92/es-log-writer/logx"
)

func main() {
    config := &writer.Config{
        Addresses:     []string{"http://localhost:9200"},
        IndexPrefix:   "go-zero-logs",
        BufferSize:    100,
        FlushInterval: 5 * time.Second,
    }
    
    // 创建 Elasticsearch 适配器
    adapter, err := logxadapter.NewAdapter(config)
    if err != nil {
        panic(err)
    }
    defer adapter.Close()
    
    // 创建控制台 Writer
    consoleWriter := logxadapter.NewConsoleWriter()
    
    // 创建多路复用 Writer，同时写入控制台和 Elasticsearch
    multiWriter := logxadapter.NewMultiWriter(consoleWriter, adapter)
    
    // 设置 logx 使用多路复用 Writer
    logx.SetWriter(multiWriter)
    
    // 日志会同时输出到控制台和 Elasticsearch
    logx.Info("这条日志会同时出现在控制台和 ES")
    logx.Infow("带字段的日志",
        logx.Field("trace", "abc123"),
        logx.Field("duration", 20*time.Millisecond),
    )
}
```

## 包结构

```
github.com/zhengliu92/es-log-writer
├── types.go          # 类型定义和接口（LogField, LogEntry, Config, FieldAccessor）
├── writer.go         # ElasticsearchWriter 核心实现
├── utils.go          # 工具函数（FormatContent, GetCaller, 字段转换/提取）
└── logx/
    ├── adapter.go    # go-zero logx.Writer 适配器（ES）
    ├── console.go    # 控制台 Writer
    ├── multi.go      # 多路复用 Writer
    └── utils.go      # logx 字段适配工具函数
```

| 包 | 依赖 | 说明 |
|---|------|------|
| `github.com/zhengliu92/es-log-writer` | 仅 Elasticsearch | 核心库，可独立使用 |
| `github.com/zhengliu92/es-log-writer/logx` | go-zero | logx.Writer 适配器、ConsoleWriter、MultiWriter |

## 配置说明

### Config 结构体

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `Addresses` | `[]string` | Elasticsearch 地址列表 | `["http://localhost:9200"]` |
| `Username` | `string` | 用户名（可选） | `""` |
| `Password` | `string` | 密码（可选） | `""` |
| `APIKey` | `string` | API Key（可选，优先级高于用户名密码） | `""` |
| `IndexPrefix` | `string` | 索引名称前缀 | `"go-zero-logs"` |
| `BufferSize` | `int` | 缓冲区大小，达到此大小后批量写入 | `100` |
| `FlushInterval` | `time.Duration` | 刷新间隔，定期刷新缓冲区 | `5 * time.Second` |

## 核心库 API

```go
// 创建写入器
w, err := writer.NewElasticsearchWriter(config)

// 写入日志
w.Info(content, fields...)
w.Error(content, fields...)
w.Debug(content, fields...)
w.Warn(content, fields...)
w.Log(level, content, fields...)

// 创建字段
writer.Field("key", value)

// 检查连接
w.Ping(ctx)

// 关闭
w.Close()
```

## logx 适配器 API

```go
// 创建 ES 适配器
adapter, err := logxadapter.NewAdapter(config)

// 创建控制台 Writer
consoleWriter := logxadapter.NewConsoleWriter()

// 创建多路复用 Writer（可组合多个 Writer）
multiWriter := logxadapter.NewMultiWriter(consoleWriter, adapter, ...)

// 检查 ES 连接
adapter.Ping(ctx)

// 关闭
adapter.Close()
multiWriter.Close()  // 会关闭所有包含的 Writer
```

## 存储格式

日志在 Elasticsearch 中的存储格式：

```json
{
  "@timestamp": "2025-12-17T10:30:00Z",
  "level": "info",
  "content": "[HTTP] 200 - GET /api/users",
  "caller": "main.go:25",
  "duration": "20ms",
  "trace": "5a98a59d88786b63d4605481b542dd83",
  "span": "4df29a5b1c46695d",
  "fields": {
    "status": 200,
    "method": "GET"
  }
}
```

## Elasticsearch 数据结构定义

在使用本库前，建议在 Elasticsearch 中创建索引模板，以确保正确的字段映射。

### 快速设置（使用索引模板）

```bash
# 使用提供的模板文件
curl -X PUT "localhost:9200/_index_template/logs-template" \
  -H 'Content-Type: application/json' \
  -d @elasticsearch-template.json
```

### 字段映射说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `@timestamp` | `date` | 时间戳（RFC3339 格式） |
| `level` | `keyword` | 日志级别（支持精确匹配和聚合） |
| `content` | `text` + `keyword` | 日志内容（支持全文搜索和精确匹配） |
| `caller` | `keyword` | 调用位置 |
| `duration` | `keyword` | 持续时间 |
| `trace` | `keyword` | 追踪 ID |
| `span` | `keyword` | Span ID |
| `fields` | `object` | 动态字段（用户自定义） |

详细设置说明请参考：
- [ELASTICSEARCH_SETUP.md](ELASTICSEARCH_SETUP.md) - 基础数据结构定义
- [CHINESE_ANALYZER.md](CHINESE_ANALYZER.md) - 中文分词配置（IK Analyzer）

## 示例

- [独立使用示例](example/standalone/main.go) - 不依赖 go-zero
- [go-zero 集成示例](example/writer/main.go) - 配合 logx 使用（仅 ES）
- [Trace 示例](example/trace/main.go) - 使用 MultiWriter 同时输出到控制台和 ES，包含 trace/span 字段

## 参考文档

- [go-zero 日志文档](https://go-zero.dev/docs/tutorials/go-zero/log/overview)
- [Elasticsearch Go Client](https://github.com/elastic/go-elasticsearch)

## License

MIT
