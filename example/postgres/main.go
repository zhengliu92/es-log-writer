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
	// 配置 PostgreSQL Writer
	// 替换为您的 DSN
	dsn := "postgres://postgres:password@localhost:5432/logs?sslmode=disable"
	
	config := &writer.PostgresConfig{
		DSN:           dsn,
		TableName:     "app_logs",
		BufferSize:    50,
		FlushInterval: 3 * time.Second,
	}

	// 创建 logx 适配器
	adapter, err := logxadapter.NewPostgresAdapter(config)
	if err != nil {
		fmt.Printf("创建 PostgreSQL 适配器失败: %v\n", err)
		return
	}
	defer adapter.Close()

	// 检查 PostgreSQL 连接是否正常
	ctx := context.Background()
	if err := adapter.Ping(ctx); err != nil {
		fmt.Printf("PostgreSQL 连接检查失败: %v\n", err)
		// 在实际应用中，如果连接失败可能需要处理
	} else {
		fmt.Println("PostgreSQL 连接检查成功")
	}

	consoleWriter := logxadapter.NewConsoleWriter()

	// 创建多路复用 Writer，同时写入控制台和 PostgreSQL
	multiWriter := logxadapter.NewMultiWriter(consoleWriter, adapter)

	// 设置 logx 使用多路复用 Writer
	logx.SetWriter(multiWriter)

	// 使用 logx 打印日志
	logx.Info("这是一条写入 PostgreSQL 的 info 日志")
	logx.Infow("这是一条带字段的日志", logx.Field("user_id", 1001), logx.Field("action", "login"))
	logx.Error("这是一条 error 日志")
	logx.Slow("这是一条 slow 日志")

	// 模拟持续写入日志
	fmt.Println("开始持续写入日志...")
	for i := 0; i < 5; i++ {
		logx.Infof("PostgreSQL 日志测试编号: %d", i)
		time.Sleep(500 * time.Millisecond)
	}

	// 等待日志刷新
	fmt.Println("等待日志刷新到数据库...")
	time.Sleep(5 * time.Second)
	fmt.Println("测试完成")
}
