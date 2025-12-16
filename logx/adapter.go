// Package logx 提供 go-zero logx.Writer 接口的适配器
// 使用此包需要依赖 go-zero
package logx

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	writer "github.com/zheng/log-writer"
	"github.com/zeromicro/go-zero/core/logx"
)

// Adapter 将 ElasticsearchWriter 适配到 logx.Writer 接口
type Adapter struct {
	*writer.ElasticsearchWriter
}

// NewAdapter 创建一个适配 logx.Writer 的写入器
func NewAdapter(config *writer.Config) (*Adapter, error) {
	w, err := writer.NewElasticsearchWriter(config)
	if err != nil {
		return nil, err
	}
	return &Adapter{ElasticsearchWriter: w}, nil
}

// 以下方法实现 logx.Writer 接口

func (a *Adapter) Alert(v any) {
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "alert",
		Content:   formatContent(v),
		Caller:    getCaller(1),
	}
	a.AddEntry(entry)
}

func (a *Adapter) Debug(v any, fields ...logx.LogField) {
	trace, span, duration := extractLogxFields(fields...)
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "debug",
		Content:   formatContent(v),
		Caller:    getCaller(1),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
	a.AddEntry(entry)
}

func (a *Adapter) Error(v any, fields ...logx.LogField) {
	trace, span, duration := extractLogxFields(fields...)
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "error",
		Content:   formatContent(v),
		Caller:    getCaller(1),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
	a.AddEntry(entry)
}

func (a *Adapter) Info(v any, fields ...logx.LogField) {
	trace, span, duration := extractLogxFields(fields...)
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "info",
		Content:   formatContent(v),
		Caller:    getCaller(1),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
	a.AddEntry(entry)
}

func (a *Adapter) Severe(v any) {
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "severe",
		Content:   formatContent(v),
		Caller:    getCaller(1),
	}
	a.AddEntry(entry)
}

func (a *Adapter) Slow(v any, fields ...logx.LogField) {
	trace, span, duration := extractLogxFields(fields...)
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "slow",
		Content:   formatContent(v),
		Caller:    getCaller(1),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
	a.AddEntry(entry)
}

func (a *Adapter) Stack(v any) {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stackTrace := string(buf[:n])

	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "stack",
		Content:   formatContent(v),
		Caller:    getCaller(1),
		Fields: map[string]interface{}{
			"stack": stackTrace,
		},
	}
	a.AddEntry(entry)
}

func (a *Adapter) Stat(v any, fields ...logx.LogField) {
	trace, span, duration := extractLogxFields(fields...)
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "stat",
		Content:   formatContent(v),
		Caller:    getCaller(1),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
	a.AddEntry(entry)
}

func (a *Adapter) Close() error {
	return a.ElasticsearchWriter.Close()
}

// 辅助函数

func formatContent(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case error:
		return val.Error()
	default:
		return fmt.Sprintf("%v", val)
	}
}

func convertLogxFields(fields ...logx.LogField) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	result := make(map[string]interface{})
	for _, field := range fields {
		result[field.Key] = field.Value
	}
	return result
}

func extractLogxFields(fields ...logx.LogField) (trace, span, duration string) {
	for _, field := range fields {
		switch field.Key {
		case "trace":
			trace = fmt.Sprintf("%v", field.Value)
		case "span":
			span = fmt.Sprintf("%v", field.Value)
		case "duration":
			if dur, ok := field.Value.(time.Duration); ok {
				duration = dur.String()
			} else {
				duration = fmt.Sprintf("%v", field.Value)
			}
		}
	}
	return
}

func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return ""
	}
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}
	return fmt.Sprintf("%s:%d", file, line)
}

