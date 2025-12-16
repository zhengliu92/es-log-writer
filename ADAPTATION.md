# 日志格式适配说明

## 你提供的日志格式

```
2025-12-16T15:58:24.202Z         info   [HTTP]  200  -  GET  /api/v1/metrics/range?window=3m - 192.168.31.1:64741 - Apifox/1.0.0 (https://apifox.com)     duration=20.6ms      caller=handler/loghandler.go:167    trace=5a98a59d88786b63d4605481b542dd83   span=4df29a5b1c46695d
```

## 适配情况

✅ **完全适配**

### 工作原理

1. **直接实现 logx.Writer 接口**
   - 我们的实现直接实现了 `logx.Writer` 接口的所有方法（Info, Error, Debug, Slow, Stat 等）
   - go-zero 会直接调用这些方法，传入结构化的数据，而不是文本格式

2. **字段提取**
   - 当使用 `logx.Infow()` 等方法时，可以通过 `logx.Field()` 传入字段
   - 我们的实现会自动提取以下特殊字段：
     - `trace` → 存储到 `LogEntry.Trace`
     - `span` → 存储到 `LogEntry.Span`
     - `duration` → 存储到 `LogEntry.Duration`
     - `caller` → 自动从调用栈获取，或从字段中提取
   - 其他字段存储在 `LogEntry.Fields` 中

3. **使用示例**

```go
// 方式1: 使用 logx.Infow() 传入字段
logx.Infow(
    "[HTTP] 200 - GET /api/v1/metrics/range?window=3m",
    logx.Field("duration", 20*time.Millisecond),
    logx.Field("trace", "5a98a59d88786b63d4605481b542dd83"),
    logx.Field("span", "4df29a5b1c46695d"),
    logx.Field("status", 200),
    logx.Field("method", "GET"),
)

// 方式2: 如果 go-zero 中间件已经添加了这些字段
// 那么它们会自动被提取和存储
```

## 存储到 Elasticsearch 的格式

日志会被存储为以下 JSON 格式：

```json
{
  "@timestamp": "2025-12-16T15:58:24.202Z",
  "level": "info",
  "content": "[HTTP] 200 - GET /api/v1/metrics/range?window=3m",
  "caller": "handler/loghandler.go:167",
  "duration": "20.6ms",
  "trace": "5a98a59d88786b63d4605481b542dd83",
  "span": "4df29a5b1c46695d",
  "fields": {
    "status": 200,
    "method": "GET",
    "path": "/api/v1/metrics/range",
    "client_ip": "192.168.31.1:64741",
    "user_agent": "Apifox/1.0.0 (https://apifox.com)"
  }
}
```

## 注意事项

1. **文本格式 vs 结构化格式**
   - 你看到的文本格式是 go-zero 输出到控制台或文件的格式
   - 我们的实现接收的是 go-zero 内部的结构化数据
   - 不需要解析文本，因为 go-zero 已经为我们处理好了

2. **字段传递**
   - 确保在使用 `logx.Infow()` 等方法时，通过 `logx.Field()` 传入 trace 和 span
   - 如果使用 go-zero 的中间件（如 HTTP 中间件），这些字段通常会自动添加

3. **测试**
   - 运行 `example/test_trace.go` 可以测试 trace 和 span 字段的写入
   - 在 Kibana 中查看索引 `go-zero-logs-*` 验证数据格式

