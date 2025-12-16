package logx

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

// MultiWriter 多路复用 Writer，可以同时写入多个 Writer
type MultiWriter struct {
	writers []logx.Writer
}

// NewMultiWriter 创建一个多路复用 Writer
func NewMultiWriter(writers ...logx.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

// Alert 实现 logx.Writer 接口
func (m *MultiWriter) Alert(v any) {
	for _, w := range m.writers {
		w.Alert(v)
	}
}

// Debug 实现 logx.Writer 接口
func (m *MultiWriter) Debug(v any, fields ...logx.LogField) {
	for _, w := range m.writers {
		w.Debug(v, fields...)
	}
}

// Error 实现 logx.Writer 接口
func (m *MultiWriter) Error(v any, fields ...logx.LogField) {
	for _, w := range m.writers {
		w.Error(v, fields...)
	}
}

// Info 实现 logx.Writer 接口
func (m *MultiWriter) Info(v any, fields ...logx.LogField) {
	for _, w := range m.writers {
		w.Info(v, fields...)
	}
}

// Severe 实现 logx.Writer 接口
func (m *MultiWriter) Severe(v any) {
	for _, w := range m.writers {
		w.Severe(v)
	}
}

// Slow 实现 logx.Writer 接口
func (m *MultiWriter) Slow(v any, fields ...logx.LogField) {
	for _, w := range m.writers {
		w.Slow(v, fields...)
	}
}

// Stack 实现 logx.Writer 接口
func (m *MultiWriter) Stack(v any) {
	for _, w := range m.writers {
		w.Stack(v)
	}
}

// Stat 实现 logx.Writer 接口
func (m *MultiWriter) Stat(v any, fields ...logx.LogField) {
	for _, w := range m.writers {
		w.Stat(v, fields...)
	}
}

// Close 关闭所有 Writer
func (m *MultiWriter) Close() error {
	var errs []error
	for _, w := range m.writers {
		if closer, ok := w.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing writers: %v", errs)
	}
	return nil
}

