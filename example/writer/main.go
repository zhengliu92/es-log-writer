package main

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	writer "github.com/zhengliu92/es-log-writer"
	logxadapter "github.com/zhengliu92/es-log-writer/logx"
)

func main() {
	// 配置 Elasticsearch Writer
	config := &writer.Config{
		Addresses:     []string{"http://localhost:9200"},
		IndexPrefix:   "go-zero-logs",
		BufferSize:    50,
		FlushInterval: 3 * time.Second,
		// 如果需要认证，可以设置以下字段
		Username: "elastic",
		Password: "my_elastic_password",
		// 或者使用 API Key
		// APIKey: "your-api-key",
	}

	// 创建 logx 适配器
	adapter, err := logxadapter.NewAdapter(config)
	if err != nil {
		panic(err)
	}
	defer adapter.Close()

	// 检查 Elasticsearch 连接是否正常
	ctx := context.Background()
	if err := adapter.Ping(ctx); err != nil {
		logx.Errorf("Elasticsearch 连接检查失败: %v", err)
		panic(fmt.Sprintf("无法连接到 Elasticsearch: %v", err))
	}
	logx.Info("Elasticsearch 连接检查成功")

	// 设置 logx 使用 Elasticsearch Writer
	// logx.SetWriter(adapter)

	consoleWriter := logxadapter.NewConsoleWriter()

	// 创建多路复用 Writer，同时写入控制台和 Elasticsearch
	multiWriter := logxadapter.NewMultiWriter(consoleWriter, adapter)

	// 设置 logx 使用多路复用 Writer
	logx.SetWriter(multiWriter)

	// 使用 logx 打印日志
	logx.Info("这是一条 info 日志")
	logx.Infow("这是一条带字段的日志", logx.Field("key1", "value1"), logx.Field("key2", 123))
	logx.Error("这是一条 error 日志")
	logx.Slow("这是一条 slow 日志")

	// 使用 logc（带 context）
	logx.WithContext(ctx).Info("使用 context 的日志")
	logx.WithContext(ctx).Errorf("错误日志: %s", "某个错误")

	// 模拟持续写入日志
	for i := 0; i < 10; i++ {
		logx.Infof("日志编号: %d", i)
		time.Sleep(500 * time.Millisecond)
	}

	// 等待日志刷新
	time.Sleep(5 * time.Second)
}
