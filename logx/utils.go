package logx

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// formatContent 格式化内容为字符串
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

// convertLogxFields 将 logx.LogField 切片转换为 map
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

// extractLogxFields 从 logx.LogField 中提取 trace、span、duration 特殊字段
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

// getCaller 获取调用者信息
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
