package main

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	logwriter "github.com/zheng/log-writer"
)

func main() {
	// 配置 Elasticsearch Writer
	config := &logwriter.Config{
		Addresses:     []string{"http://localhost:9200"},
		IndexPrefix:   "go-zero-logs",
		BufferSize:    50,
		FlushInterval: 3 * time.Second,
	}

	// 创建 Elasticsearch Writer
	esWriter, err := logwriter.NewElasticsearchWriter(config)
	if err != nil {
		panic(err)
	}
	defer esWriter.Close()

	// 设置 logx 使用 Elasticsearch Writer
	logx.SetWriter(esWriter)

	// 模拟你提供的日志格式
	// 原始格式: 2025-12-16T15:58:24.202Z  info  [HTTP]  200  -  GET  /api/v1/metrics/range?window=3m - 192.168.31.1:64741 - Apifox/1.0.0 (https://apifox.com)  duration=20.6ms  caller=handler/loghandler.go:167  trace=5a98a59d88786b63d4605481b542dd83  span=4df29a5b1c46695d

	// 使用 logx 打印带 trace 和 span 的日志
	logx.Infow(
		"[HTTP] 200 - GET /api/v1/metrics/range?window=3m - 192.168.31.1:64741 - Apifox/1.0.0 (https://apifox.com)",
		logx.Field("duration", 20*time.Millisecond),
		logx.Field("caller", "handler/loghandler.go:167"),
		logx.Field("trace", "5a98a59d88786b63d4605481b542dd83"),
		logx.Field("span", "4df29a5b1c46695d"),
		logx.Field("status", 200),
		logx.Field("method", "GET"),
		logx.Field("path", "/api/v1/metrics/range"),
		logx.Field("client_ip", "192.168.31.1:64741"),
		logx.Field("user_agent", "Apifox/1.0.0 (https://apifox.com)"),
	)

	// 使用 context（如果有 trace context）
	ctx := context.Background()
	logx.WithContext(ctx).Infow(
		"Another log with trace",
		logx.Field("trace", "abc123def456"),
		logx.Field("span", "xyz789"),
		logx.Field("duration", 15*time.Millisecond),
	)

	// 等待日志刷新
	time.Sleep(5 * time.Second)
}
