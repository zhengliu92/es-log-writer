package logx

import (
	"github.com/zeromicro/go-zero/core/logx"
	writer "github.com/zhengliu92/es-log-writer"
)

// logxFieldAdapter 适配 logx.LogField 到 writer.FieldAccessor 接口
type logxFieldAdapter struct {
	field logx.LogField
}

func (a logxFieldAdapter) GetKey() string {
	return a.field.Key
}

func (a logxFieldAdapter) GetValue() interface{} {
	return a.field.Value
}

// convertLogxFields 将 logx.LogField 切片转换为 map
func convertLogxFields(fields ...logx.LogField) map[string]interface{} {
	adapters := make([]writer.FieldAccessor, len(fields))
	for i, field := range fields {
		adapters[i] = logxFieldAdapter{field: field}
	}
	return writer.ConvertFields(adapters)
}

// extractLogxFields 从 logx.LogField 中提取 trace、span、duration 特殊字段
func extractLogxFields(fields ...logx.LogField) (trace, span, duration string) {
	adapters := make([]writer.FieldAccessor, len(fields))
	for i, field := range fields {
		adapters[i] = logxFieldAdapter{field: field}
	}
	return writer.ExtractFields(adapters)
}
