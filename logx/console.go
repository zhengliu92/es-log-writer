package logx

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	writer "github.com/zhengliu92/es-log-writer"
)

// ConsoleWriter 控制台 Writer，将日志输出到标准输出
type ConsoleWriter struct{}

// NewConsoleWriter 创建一个控制台 Writer
func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

// Alert 实现 logx.Writer 接口
func (c *ConsoleWriter) Alert(v any) {
	c.log("alert", v)
}

// Debug 实现 logx.Writer 接口
func (c *ConsoleWriter) Debug(v any, fields ...logx.LogField) {
	c.log("debug", v, fields...)
}

// Error 实现 logx.Writer 接口
func (c *ConsoleWriter) Error(v any, fields ...logx.LogField) {
	c.log("error", v, fields...)
}

// Info 实现 logx.Writer 接口
func (c *ConsoleWriter) Info(v any, fields ...logx.LogField) {
	c.log("info", v, fields...)
}

// Severe 实现 logx.Writer 接口
func (c *ConsoleWriter) Severe(v any) {
	c.log("severe", v)
}

// Slow 实现 logx.Writer 接口
func (c *ConsoleWriter) Slow(v any, fields ...logx.LogField) {
	c.log("slow", v, fields...)
}

// Stack 实现 logx.Writer 接口
func (c *ConsoleWriter) Stack(v any) {
	c.log("stack", v)
}

// Stat 实现 logx.Writer 接口
func (c *ConsoleWriter) Stat(v any, fields ...logx.LogField) {
	c.log("stat", v, fields...)
}

// Close 实现 Close 接口（控制台 Writer 不需要关闭）
func (c *ConsoleWriter) Close() error {
	return nil
}

// log 辅助方法，格式化并输出日志
func (c *ConsoleWriter) log(level string, v any, fields ...logx.LogField) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	content := writer.FormatContent(v)

	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(level)))
	parts = append(parts, timestamp)
	parts = append(parts, content)

	// 输出其他字段
	for _, field := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", field.Key, field.Value))
	}

	output := strings.Join(parts, " ")
	if level == "alert" || level == "severe" || level == "stack" {
		fmt.Fprintf(os.Stderr, "%s\n", output)
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
}
