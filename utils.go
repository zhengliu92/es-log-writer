package writer

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// FormatContent 格式化内容为字符串
func FormatContent(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case error:
		return val.Error()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// convertFields 通用的字段转换函数，接受 FieldAccessor 切片
func convertFields[T FieldAccessor](fields []T) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	result := make(map[string]interface{})
	for _, field := range fields {
		result[field.GetKey()] = field.GetValue()
	}
	return result
}

// ConvertFields 导出的通用字段转换函数，接受 FieldAccessor 切片
func ConvertFields(fields []FieldAccessor) map[string]interface{} {
	return convertFields(fields)
}

// extractFields 通用的字段提取函数，接受 FieldAccessor 切片
func extractFields[T FieldAccessor](fields []T) (trace, span, duration string) {
	for _, field := range fields {
		switch field.GetKey() {
		case "trace":
			trace = fmt.Sprintf("%v", field.GetValue())
		case "span":
			span = fmt.Sprintf("%v", field.GetValue())
		case "duration":
			if dur, ok := field.GetValue().(time.Duration); ok {
				duration = dur.String()
			} else {
				duration = fmt.Sprintf("%v", field.GetValue())
			}
		}
	}
	return
}

// ExtractFields 导出的通用字段提取函数，接受 FieldAccessor 切片
func ExtractFields(fields []FieldAccessor) (trace, span, duration string) {
	return extractFields(fields)
}

// getCaller 获取调用者信息
func GetCaller(skip int) string {
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
