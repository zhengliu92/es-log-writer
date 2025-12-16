# log-writer

一个用于将 go-zero 的 logx 日志同步写入 Elasticsearch 的 Golang 工具库。

## 功能特性

- ✅ 完整实现 `logx.Writer` 接口，无缝集成 go-zero 日志系统
- ✅ 支持所有日志级别：Alert, Debug, Error, Info, Severe, Slow, Stack, Stat
- ✅ 支持批量写入，提高性能
- ✅ 自动按日期创建索引（格式：`{prefix}-{YYYY.MM.DD}`）
- ✅ 支持缓冲区刷新机制，可配置刷新间隔和缓冲区大小
- ✅ 支持 Elasticsearch 认证（用户名密码或 API Key）
- ✅ 自动记录调用者信息（文件名和行号）
- ✅ 支持日志字段（LogField）
- ✅ 线程安全，支持并发写入
- ✅ 优雅关闭，确保所有日志都被写入
- ✅ **异步写入**：日志写入不阻塞业务代码

## 安装

```bash
go get github.com/zheng/log-writer
```

## 环境准备

### 使用 Docker Compose 启动 Elasticsearch 和 Kibana

项目提供了两个 Docker Compose 配置文件：

1. **docker-compose.yml** - 开发环境（无认证）
   ```bash
   docker-compose up -d
   ```

2. **docker-compose.secure.yml** - 生产环境（带认证）
   ```bash
   # 设置密码（可选，默认为 changeme）
   export ELASTIC_PASSWORD=your-secure-password
   docker-compose -f docker-compose.secure.yml up -d
   ```

启动后：
- **Elasticsearch**: http://localhost:9200
- **Kibana**: http://localhost:5601

查看服务状态：
```bash
# 查看所有服务
docker-compose ps

# 查看 Elasticsearch 日志
docker-compose logs -f elasticsearch

# 查看 Kibana 日志
docker-compose logs -f kibana
```

停止服务：
```bash
# 停止服务（保留数据）
docker-compose down

# 停止服务并删除数据卷
docker-compose down -v
```

### 在 Kibana 中查看日志

1. 打开 Kibana: http://localhost:5601
2. 进入 **Stack Management** > **Index Patterns**
3. 创建索引模式：`go-zero-logs-*`
4. 选择时间字段：`@timestamp`
5. 进入 **Discover** 查看日志

### 验证 Elasticsearch 是否运行

```bash
# 无认证版本
curl http://localhost:9200

# 带认证版本
curl -u elastic:changeme http://localhost:9200
```

## 快速开始

### 基本使用

```go
package main

import (
    "time"
    "github.com/zeromicro/go-zero/core/logx"
    logwriter "github.com/zheng/log-writer"
)

func main() {
    // 创建 Elasticsearch Writer
    config := &logwriter.Config{
        Addresses:     []string{"http://localhost:9200"},
        IndexPrefix:   "go-zero-logs",
        BufferSize:    100,
        FlushInterval: 5 * time.Second,
    }
    
    esWriter, err := logwriter.NewElasticsearchWriter(config)
    if err != nil {
        panic(err)
    }
    defer esWriter.Close()
    
    // 设置 logx 使用 Elasticsearch Writer
    logx.SetWriter(esWriter)
    
    // 正常使用 logx
    logx.Info("日志内容")
}
```

### 使用认证（使用 docker-compose.secure.yml 时）

```go
config := &logwriter.Config{
    Addresses:   []string{"http://localhost:9200"},
    IndexPrefix: "go-zero-logs",
    Username:    "elastic",
    Password:    "your-password", // 或使用环境变量 ELASTIC_PASSWORD
}
```

或使用 API Key：

```go
config := &logwriter.Config{
    Addresses:   []string{"https://your-es-cluster:9200"},
    IndexPrefix: "go-zero-logs",
    APIKey:      "your-api-key",
    EnableSSL:   true,
}
```

### 与 go-zero 配置集成

```go
package main

import (
    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/core/logx"
    logwriter "github.com/zheng/log-writer"
)

func main() {
    // 加载 go-zero 配置
    var c logx.LogConf
    conf.MustLoad("config.yaml", &c)
    logx.MustSetup(c)
    
    // 添加 Elasticsearch Writer（同时输出到文件和控制台）
    esConfig := &logwriter.Config{
        Addresses:     []string{"http://localhost:9200"},
        IndexPrefix:   "go-zero-logs",
        BufferSize:    100,
        FlushInterval: 5 * time.Second,
    }
    
    esWriter, err := logwriter.NewElasticsearchWriter(esConfig)
    if err != nil {
        panic(err)
    }
    defer esWriter.Close()
    
    // 添加额外的 Writer（不影响原有的文件输出）
    logx.AddWriter(esWriter)
    
    // 正常使用 logx
    logx.Info("日志内容")
}
```

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
| `EnableSSL` | `bool` | 是否启用 SSL | `false` |
| `SkipSSLVerify` | `bool` | SSL 跳过验证（仅用于开发环境） | `false` |

### 索引命名规则

索引名称格式：`{IndexPrefix}-{YYYY.MM.DD}`

例如：
- 前缀为 `go-zero-logs`
- 2024年1月15日的日志会写入索引：`go-zero-logs-2024.01.15`

## 性能优化

1. **批量写入**：日志会先缓存在内存中，达到 `BufferSize` 或达到 `FlushInterval` 时批量写入 Elasticsearch
2. **异步刷新**：后台协程定期刷新缓冲区，不影响主业务逻辑
3. **可配置缓冲区**：根据日志量调整 `BufferSize` 和 `FlushInterval` 以平衡性能和实时性

## 注意事项

1. **优雅关闭**：程序退出前务必调用 `Close()` 方法，确保所有日志都被写入
2. **索引管理**：建议在 Elasticsearch 中配置索引生命周期策略（ILM）自动管理旧索引
3. **性能影响**：虽然使用了批量写入和异步刷新，但仍可能对性能有轻微影响，建议在生产环境中进行压测
4. **Docker 资源**：Elasticsearch 默认使用 512MB 内存，可根据需要调整 `ES_JAVA_OPTS` 环境变量

## 示例

完整示例请参考 [example/main.go](example/main.go)

## 参考文档

- [go-zero 日志文档](https://go-zero.dev/docs/tutorials/go-zero/log/overview)
- [Elasticsearch Go Client](https://github.com/elastic/go-elasticsearch)
- [Elasticsearch 官方文档](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [Kibana 官方文档](https://www.elastic.co/guide/en/kibana/current/index.html)

## License

MIT
