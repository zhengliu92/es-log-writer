package writer

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ConsoleWriter 控制台 Writer，将日志输出到标准输出（不依赖 go-zero）
type ConsoleWriter struct{}

// NewConsoleWriter 创建一个控制台 Writer
func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

// Log 写入日志（核心方法）
func (c *ConsoleWriter) Log(level string, content any, fields ...LogField) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	contentStr := FormatContent(content)

	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(level)))
	parts = append(parts, timestamp)
	parts = append(parts, contentStr)

	trace, span, duration := extractFields(fields)
	if trace != "" {
		parts = append(parts, fmt.Sprintf("trace=%s", trace))
	}
	if span != "" {
		parts = append(parts, fmt.Sprintf("span=%s", span))
	}
	if duration != "" {
		parts = append(parts, fmt.Sprintf("duration=%s", duration))
	}

	for _, field := range fields {
		if field.Key != "trace" && field.Key != "span" && field.Key != "duration" {
			parts = append(parts, fmt.Sprintf("%s=%v", field.Key, field.Value))
		}
	}

	output := strings.Join(parts, " ")
	if level == "error" || level == "warn" {
		fmt.Fprintf(os.Stderr, "%s\n", output)
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
}

// Info 写入 info 级别日志
func (c *ConsoleWriter) Info(content any, fields ...LogField) {
	c.Log("info", content, fields...)
}

// Error 写入 error 级别日志
func (c *ConsoleWriter) Error(content any, fields ...LogField) {
	c.Log("error", content, fields...)
}

// Debug 写入 debug 级别日志
func (c *ConsoleWriter) Debug(content any, fields ...LogField) {
	c.Log("debug", content, fields...)
}

// Warn 写入 warn 级别日志
func (c *ConsoleWriter) Warn(content any, fields ...LogField) {
	c.Log("warn", content, fields...)
}

// Close 关闭写入器（控制台 Writer 不需要关闭）
func (c *ConsoleWriter) Close() error {
	return nil
}
