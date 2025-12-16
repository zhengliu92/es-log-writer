// 独立使用示例（不依赖 go-zero）
package main

import (
	"time"

	writer "github.com/zheng/es-log-writer"
)

func main() {
	// 配置 Elasticsearch Writer
	config := &writer.Config{
		Addresses:     []string{"http://localhost:9200"},
		IndexPrefix:   "app-logs",
		BufferSize:    50,
		FlushInterval: 3 * time.Second,
	}

	// 创建 Elasticsearch Writer（不依赖 go-zero）
	w, err := writer.NewElasticsearchWriter(config)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	// 直接使用 writer 打印日志
	w.Info("这是一条 info 日志")
	w.Error("这是一条 error 日志")
	w.Debug("这是一条 debug 日志")
	w.Warn("这是一条 warn 日志")

	// 带字段的日志
	w.Info("用户登录",
		writer.Field("user_id", 12345),
		writer.Field("ip", "192.168.1.100"),
		writer.Field("trace", "abc123"),
		writer.Field("duration", 50*time.Millisecond),
	)

	// 模拟持续写入日志
	for i := 0; i < 10; i++ {
		w.Log("info", "日志编号", writer.Field("index", i))
		time.Sleep(500 * time.Millisecond)
	}

	// 等待日志刷新
	time.Sleep(5 * time.Second)
}
