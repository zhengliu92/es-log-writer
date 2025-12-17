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

// NewEsAdapter 创建一个适配 logx.Writer 的 Elasticsearch 写入器
func NewEsAdapter(config *writer.Config) (*Adapter, error) {
	w, err := writer.NewElasticsearchWriter(config)
	if err != nil {
		return nil, err
	}
	return &Adapter{ElasticsearchWriter: w}, nil
}

// createLogEntry 创建日志条目（辅助函数）
func createLogEntry(level string, content any, callerSkip int, fields ...logx.LogField) writer.LogEntry {
	trace, span, duration := extractLogxFields(fields...)
	return writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Content:   writer.FormatContent(content),
		Caller:    writer.GetCaller(callerSkip),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertLogxFields(fields...),
	}
}

// createSimpleLogEntry 创建简单日志条目（无字段）
func createSimpleLogEntry(level string, content any, callerSkip int) writer.LogEntry {
	return writer.LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Content:   writer.FormatContent(content),
		Caller:    writer.GetCaller(callerSkip),
	}
}

// Alert 实现 logx.Writer 接口
func (a *Adapter) Alert(v any) {
	a.AddEntry(createSimpleLogEntry("alert", v, 1))
}

// Debug 实现 logx.Writer 接口
func (a *Adapter) Debug(v any, fields ...logx.LogField) {
	a.AddEntry(createLogEntry("debug", v, 1, fields...))
}

// Error 实现 logx.Writer 接口
func (a *Adapter) Error(v any, fields ...logx.LogField) {
	a.AddEntry(createLogEntry("error", v, 1, fields...))
}

// Info 实现 logx.Writer 接口
func (a *Adapter) Info(v any, fields ...logx.LogField) {
	a.AddEntry(createLogEntry("info", v, 1, fields...))
}

// Severe 实现 logx.Writer 接口
func (a *Adapter) Severe(v any) {
	a.AddEntry(createSimpleLogEntry("severe", v, 1))
}

// Slow 实现 logx.Writer 接口
func (a *Adapter) Slow(v any, fields ...logx.LogField) {
	a.AddEntry(createLogEntry("slow", v, 1, fields...))
}

// Stack 实现 logx.Writer 接口
func (a *Adapter) Stack(v any) {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stackTrace := string(buf[:n])

	entry := createSimpleLogEntry("stack", v, 1)
	entry.Fields = map[string]interface{}{
		"stack": stackTrace,
	}
	a.AddEntry(entry)
}

// Stat 实现 logx.Writer 接口
func (a *Adapter) Stat(v any, fields ...logx.LogField) {
	a.AddEntry(createLogEntry("stat", v, 1, fields...))
}

// Close 关闭适配器
func (a *Adapter) Close() error {
	return a.ElasticsearchWriter.Close()
}

// Ping 检查 Elasticsearch 连接是否正常
func (a *Adapter) Ping(ctx context.Context) error {
	return a.ElasticsearchWriter.Ping(ctx)
}
