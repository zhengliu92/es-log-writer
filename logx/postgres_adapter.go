// Package logx 提供 go-zero logx.Writer 接口的适配器
package logx

import (
	"context"
	"runtime"

	"github.com/zeromicro/go-zero/core/logx"
	writer "github.com/zhengliu92/es-log-writer"
)

// PostgresAdapter 将 PostgresqlWriter 适配到 logx.Writer 接口
type PostgresAdapter struct {
	*writer.PostgresqlWriter
}

// NewPostgresAdapter 创建一个适配 logx.Writer 的 PostgreSQL 写入器
func NewPostgresAdapter(config *writer.PostgresConfig) (*PostgresAdapter, error) {
	w, err := writer.NewPostgresqlWriter(config)
	if err != nil {
		return nil, err
	}
	return &PostgresAdapter{PostgresqlWriter: w}, nil
}

// Alert 实现 logx.Writer 接口
func (a *PostgresAdapter) Alert(v any) {
	a.AddEntry(createSimpleLogEntry("alert", v))
}

// Debug 实现 logx.Writer 接口
func (a *PostgresAdapter) Debug(v any, fields ...logx.LogField) {
	a.AddEntry(createLogEntry("debug", v, fields...))
}

// Error 实现 logx.Writer 接口
func (a *PostgresAdapter) Error(v any, fields ...logx.LogField) {
	a.AddEntry(createLogEntry("error", v, fields...))
}

// Info 实现 logx.Writer 接口
func (a *PostgresAdapter) Info(v any, fields ...logx.LogField) {
	a.AddEntry(createLogEntry("info", v, fields...))
}

// Severe 实现 logx.Writer 接口
func (a *PostgresAdapter) Severe(v any) {
	a.AddEntry(createSimpleLogEntry("severe", v))
}

// Slow 实现 logx.Writer 接口
func (a *PostgresAdapter) Slow(v any, fields ...logx.LogField) {
	a.AddEntry(createLogEntry("slow", v, fields...))
}

// Stack 实现 logx.Writer 接口
func (a *PostgresAdapter) Stack(v any) {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stackTrace := string(buf[:n])

	entry := createSimpleLogEntry("stack", v)
	entry.Fields = map[string]interface{}{
		"stack": stackTrace,
	}
	a.AddEntry(entry)
}

// Stat 实现 logx.Writer 接口
func (a *PostgresAdapter) Stat(v any, fields ...logx.LogField) {
	a.AddEntry(createLogEntry("stat", v, fields...))
}

// Close 关闭适配器
func (a *PostgresAdapter) Close() error {
	return a.PostgresqlWriter.Close()
}

// Ping 检查 PostgreSQL 连接是否正常
func (a *PostgresAdapter) Ping(ctx context.Context) error {
	return a.PostgresqlWriter.Ping(ctx)
}
