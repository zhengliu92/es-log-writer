# es-log-writer

一个用于将日志同步写入 Elasticsearch 的 Golang 工具库。

## 功能特性

- ✅ **核心库不依赖 go-zero**，可独立使用
- ✅ 提供 go-zero `logx.Writer` 适配器（可选）
- ✅ 支持批量写入，提高性能
- ✅ 自动按日期创建索引（格式：`{prefix}-{YYYY.MM.DD}`）
- ✅ 支持缓冲区刷新机制，可配置刷新间隔和缓冲区大小
- ✅ 支持 Elasticsearch 认证（用户名密码或 API Key）
- ✅ 支持 trace/span/duration 字段自动提取和存储
- ✅ 线程安全，支持并发写入
- ✅ 异步写入，不阻塞业务代码
- ✅ 优雅关闭，确保所有日志都被写入
- ✅ 提供 `MultiWriter`，支持同时输出到多个目标（控制台 + Elasticsearch）
- ✅ 提供 `ConsoleWriter`，支持控制台输出（支持彩色输出，error/warn 输出到 stderr）
- ✅ 自动提取调用者信息（文件名和行号）
- ✅ 支持 go-zero 所有日志方法（Info, Error, Debug, Slow, Stat, Stack, Alert, Severe）

## 安装

```bash
go get github.com/zhengliu92/es-log-writer
```

## 使用方式

### 方式一：独立使用（不依赖 go-zero）

#### 1.1 仅使用 Elasticsearch Writer

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

#### 1.2 使用 MultiWriter 同时输出到控制台和 Elasticsearch

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
    
    // 创建 Elasticsearch Writer
    esWriter, err := writer.NewElasticsearchWriter(config)
    if err != nil {
        panic(err)
    }
    defer esWriter.Close()
    
    // 创建控制台 Writer
    consoleWriter := writer.NewConsoleWriter()
    
    // 创建多路复用 Writer，同时输出到控制台和 Elasticsearch
    w := writer.NewMultiWriter(consoleWriter, esWriter)
    defer w.Close()
    
    // 日志会同时输出到控制台和 Elasticsearch
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
    adapter, err := logxadapter.NewEsAdapter(config)
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
    adapter, err := logxadapter.NewEsAdapter(config)
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
├── types.go          # 类型定义和接口（LogField, LogEntry, Config, FieldAccessor, Writer）
├── writer.go         # ElasticsearchWriter 核心实现
├── console.go        # ConsoleWriter 核心实现（不依赖 go-zero）
├── multi.go          # MultiWriter 核心实现（不依赖 go-zero）
├── utils.go          # 工具函数（FormatContent, GetCaller, 字段转换/提取）
└── logx/
    ├── adapter.go    # go-zero logx.Writer 适配器（ES）
    ├── console.go    # 控制台 Writer（logx 适配器版本）
    ├── multi.go      # 多路复用 Writer（logx 适配器版本）
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
| `IndexPrefix` | `string` | 索引名称前缀，最终索引格式为 `{prefix}-{YYYY.MM.DD}` | `"go-zero-logs"` |
| `BufferSize` | `int` | 缓冲区大小，达到此大小后立即批量写入 | `100` |
| `FlushInterval` | `time.Duration` | 刷新间隔，定期刷新缓冲区（即使未达到 BufferSize） | `5 * time.Second` |
| `EnableSSL` | `bool` | 是否启用 SSL（可选） | `false` |
| `SkipSSLVerify` | `bool` | 是否跳过 SSL 验证（可选） | `false` |

### 配置建议

- **BufferSize**: 根据日志量调整，建议 50-500。值越大，批量写入效率越高，但内存占用也越大。
- **FlushInterval**: 建议 3-10 秒。间隔越短，日志实时性越高，但会增加写入频率。
- **IndexPrefix**: 建议使用应用名称，如 `myapp-logs`，便于在 Kibana 中区分不同应用的日志。

## 核心库 API

### Writer 接口

所有 Writer 都实现了 `Writer` 接口：

```go
type Writer interface {
    Info(content any, fields ...LogField)
    Error(content any, fields ...LogField)
    Debug(content any, fields ...LogField)
    Warn(content any, fields ...LogField)
    Log(level string, content any, fields ...LogField)
    Close() error
}
```

### 创建 Writer

```go
// 创建 Elasticsearch 写入器
esWriter, err := writer.NewElasticsearchWriter(config)

// 创建控制台写入器
consoleWriter := writer.NewConsoleWriter()

// 创建多路复用写入器（可组合多个 Writer）
multiWriter := writer.NewMultiWriter(consoleWriter, esWriter, ...)
```

### 写入日志

```go
// 基本日志方法
w.Info(content, fields...)
w.Error(content, fields...)
w.Debug(content, fields...)
w.Warn(content, fields...)

// 通用日志方法（可指定任意级别）
w.Log("custom", content, fields...)
```

### 创建字段

```go
// 创建字段
writer.Field("key", value)

// 特殊字段（会自动提取到对应位置）
writer.Field("trace", "trace-id")      // 提取到 LogEntry.Trace
writer.Field("span", "span-id")        // 提取到 LogEntry.Span
writer.Field("duration", time.Duration) // 提取到 LogEntry.Duration，自动格式化
```

### 其他方法

```go
// 检查 Elasticsearch 连接（仅 ElasticsearchWriter）
err := esWriter.Ping(ctx)

// 关闭 Writer（会刷新所有缓冲的日志）
err := w.Close()
```

## logx 适配器 API

### 创建适配器

```go
// 创建 ES 适配器（实现 logx.Writer 接口）
adapter, err := logxadapter.NewEsAdapter(config)

// 创建控制台 Writer（实现 logx.Writer 接口）
consoleWriter := logxadapter.NewConsoleWriter()

// 创建多路复用 Writer（可组合多个 Writer）
multiWriter := logxadapter.NewMultiWriter(consoleWriter, adapter, ...)
```

### 设置到 logx

```go
// 设置 logx 使用适配器
logx.SetWriter(adapter)
// 或使用多路复用 Writer
logx.SetWriter(multiWriter)
```

### 支持的 logx 方法

适配器支持 go-zero logx 的所有方法：

```go
// 基本方法
logx.Info(v, fields...)
logx.Error(v, fields...)
logx.Debug(v, fields...)

// go-zero 特有方法
logx.Slow(v, fields...)      // 慢日志
logx.Stat(v, fields...)       // 统计日志
logx.Stack(v)                 // 堆栈日志
logx.Alert(v)                 // 告警日志
logx.Severe(v)                // 严重错误日志

// 带字段的方法
logx.Infow("message", logx.Field("key", value), ...)
logx.Errorw("message", logx.Field("key", value), ...)
```

### 其他方法

```go
// 检查 ES 连接
err := adapter.Ping(ctx)

// 关闭适配器
err := adapter.Close()
multiWriter.Close()  // 会关闭所有包含的 Writer
```

## 存储格式

### 日志条目结构

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
    "method": "GET",
    "user_id": 12345
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 | 来源 |
|------|------|------|------|
| `@timestamp` | `string` | 日志时间戳（RFC3339 格式） | 自动生成 |
| `level` | `string` | 日志级别（info/error/debug/warn/slow/stat/stack/alert/severe） | 方法参数 |
| `content` | `string` | 日志内容 | 方法参数 |
| `caller` | `string` | 调用位置（文件名:行号） | 自动提取 |
| `duration` | `string` | 持续时间（如 "20ms"） | 从字段中提取 |
| `trace` | `string` | 追踪 ID | 从字段中提取 |
| `span` | `string` | Span ID | 从字段中提取 |
| `fields` | `object` | 其他自定义字段 | 从字段中提取（排除 trace/span/duration） |

### 索引命名规则

索引名称格式：`{IndexPrefix}-{YYYY.MM.DD}`

例如：
- 配置 `IndexPrefix: "app-logs"`，日期为 2025-12-17
- 生成的索引名：`app-logs-2025.12.17`

每天自动创建新索引，便于按日期管理和清理日志。

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

## 使用示例

### 示例代码

- [独立使用示例](example/standalone/main.go) - 不依赖 go-zero，使用 MultiWriter 同时输出到控制台和 ES
- [go-zero 集成示例](example/writer/main.go) - 配合 logx 使用（仅 ES）
- [Trace 示例](example/trace/main.go) - 使用 MultiWriter 同时输出到控制台和 ES，包含 trace/span 字段

### 运行示例

```bash
# 确保 Elasticsearch 正在运行（默认 localhost:9200）
# 运行独立使用示例
go run example/standalone/main.go

# 运行 go-zero 集成示例
go run example/writer/main.go

# 运行 Trace 示例
go run example/trace/main.go
```

### 验证日志

在 Kibana 中查看日志：

1. 打开 Kibana（通常为 http://localhost:5601）
2. 进入 "Discover" 页面
3. 选择索引模式 `app-logs-*` 或 `go-zero-logs-*`
4. 查看日志数据

## 注意事项

### 错误处理

- `NewElasticsearchWriter` 会立即尝试连接 Elasticsearch，如果连接失败会返回错误
- 写入日志时如果 Elasticsearch 不可用，错误会被静默处理（不会阻塞业务代码）
- 建议在生产环境中监控 Elasticsearch 连接状态，定期调用 `Ping()` 方法

### 性能优化

- 根据日志量调整 `BufferSize` 和 `FlushInterval`
- 批量写入可以提高性能，但会增加内存占用
- 在高并发场景下，建议使用较大的 `BufferSize`（如 200-500）

### 优雅关闭

- 调用 `Close()` 方法会：
  1. 停止后台刷新 goroutine
  2. 等待所有缓冲的日志写入完成
  3. 关闭 Elasticsearch 连接
- 建议在应用退出时调用 `defer w.Close()` 确保所有日志都被写入

### 字段提取规则

- `trace`、`span`、`duration` 字段会被自动提取到顶层字段
- 其他字段存储在 `fields` 对象中
- `duration` 字段如果是 `time.Duration` 类型，会自动格式化为字符串（如 "20ms"）

## 常见问题

### Q: 如何同时输出到控制台和 Elasticsearch？

A: 使用 `MultiWriter`：

```go
consoleWriter := writer.NewConsoleWriter()
esWriter, _ := writer.NewElasticsearchWriter(config)
multiWriter := writer.NewMultiWriter(consoleWriter, esWriter)
```

### Q: 如何自定义索引名称格式？

A: 当前版本固定为 `{prefix}-{YYYY.MM.DD}` 格式。如需自定义，可以修改 `getIndexName()` 方法或提交 Issue。

### Q: 支持哪些 Elasticsearch 版本？

A: 使用 `github.com/elastic/go-elasticsearch/v8`，支持 Elasticsearch 7.x 和 8.x。

### Q: 如何配置认证？

A: 在 `Config` 中设置 `Username`/`Password` 或 `APIKey`：

```go
config := &writer.Config{
    Addresses: []string{"https://your-es:9200"},
    APIKey:    "your-api-key",  // 或使用 Username/Password
}
```

## 参考文档

- [go-zero 日志文档](https://go-zero.dev/docs/tutorials/go-zero/log/overview)
- [Elasticsearch Go Client](https://github.com/elastic/go-elasticsearch)
- [日志格式适配说明](ADAPTATION.md) - 了解如何适配 go-zero 日志格式

## License

MIT
