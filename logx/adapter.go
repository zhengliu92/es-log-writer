// Package logx 提供 go-zero logx.Writer 接口的适配器
// 使用此包需要依赖 go-zero
package logx

import (
	"context"
	"runtime"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	writer "github.com/zhengliu92/es-log-writer"
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

// Alert 实现 logx.Writer 接口
func (a *Adapter) Alert(v any) {
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "alert",
		Content:   writer.FormatContent(v),
		Caller:    writer.GetCaller(1),
	}
	a.AddEntry(entry)
}

// Debug 实现 logx.Writer 接口
func (a *Adapter) Debug(v any, fields ...logx.LogField) {
	trace, span, duration := extractLogxFields(fields...)
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "debug",
		Content:   writer.FormatContent(v),
		Caller:    writer.GetCaller(1),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
	a.AddEntry(entry)
}

// Error 实现 logx.Writer 接口
func (a *Adapter) Error(v any, fields ...logx.LogField) {
	trace, span, duration := extractLogxFields(fields...)
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "error",
		Content:   writer.FormatContent(v),
		Caller:    writer.GetCaller(1),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
	a.AddEntry(entry)
}

// Info 实现 logx.Writer 接口
func (a *Adapter) Info(v any, fields ...logx.LogField) {
	trace, span, duration := extractLogxFields(fields...)
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "info",
		Content:   writer.FormatContent(v),
		Caller:    writer.GetCaller(1),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
	a.AddEntry(entry)
}

// Severe 实现 logx.Writer 接口
func (a *Adapter) Severe(v any) {
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "severe",
		Content:   writer.FormatContent(v),
		Caller:    writer.GetCaller(1),
	}
	a.AddEntry(entry)
}

// Slow 实现 logx.Writer 接口
func (a *Adapter) Slow(v any, fields ...logx.LogField) {
	trace, span, duration := extractLogxFields(fields...)
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "slow",
		Content:   writer.FormatContent(v),
		Caller:    writer.GetCaller(1),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
	a.AddEntry(entry)
}

// Stack 实现 logx.Writer 接口
func (a *Adapter) Stack(v any) {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stackTrace := string(buf[:n])

	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "stack",
		Content:   writer.FormatContent(v),
		Caller:    writer.GetCaller(1),
		Fields: map[string]interface{}{
			"stack": stackTrace,
		},
	}
	a.AddEntry(entry)
}

// Stat 实现 logx.Writer 接口
func (a *Adapter) Stat(v any, fields ...logx.LogField) {
	trace, span, duration := extractLogxFields(fields...)
	entry := writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "stat",
		Content:   writer.FormatContent(v),
		Caller:    writer.GetCaller(1),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
	a.AddEntry(entry)
}

// Close 关闭适配器
func (a *Adapter) Close() error {
	return a.ElasticsearchWriter.Close()
}

// Ping 检查 Elasticsearch 连接是否正常
func (a *Adapter) Ping(ctx context.Context) error {
	return a.ElasticsearchWriter.Ping(ctx)
}
