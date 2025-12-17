package writer

import "time"

// FieldAccessor 字段访问接口，用于统一处理不同类型的字段
type FieldAccessor interface {
	GetKey() string
	GetValue() interface{}
}

// LogField 日志字段（自定义类型，不依赖 go-zero）
type LogField struct {
	Key   string
	Value any
}

// GetKey 实现 FieldAccessor 接口
func (f LogField) GetKey() string {
	return f.Key
}

// GetValue 实现 FieldAccessor 接口
func (f LogField) GetValue() interface{} {
	return f.Value
}

// Field 创建一个日志字段
func Field(key string, value any) LogField {
	return LogField{Key: key, Value: value}
}

// LogEntry 表示一条日志条目
type LogEntry struct {
	Timestamp string                 `json:"@timestamp"`
	Level     string                 `json:"level"`
	Content   string                 `json:"content"`
	Caller    string                 `json:"caller,omitempty"`
	Duration  string                 `json:"duration,omitempty"`
	Trace     string                 `json:"trace,omitempty"`
	Span      string                 `json:"span,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Config Elasticsearch Writer 配置
type Config struct {
	Addresses     []string      `json:"addresses"`
	Username      string        `json:"username,omitempty"`
	Password      string        `json:"password,omitempty"`
	APIKey        string        `json:"api_key,omitempty"`
	IndexPrefix   string        `json:"index_prefix"`
	BufferSize    int           `json:"buffer_size"`
	FlushInterval time.Duration `json:"flush_interval"`
	EnableSSL     bool          `json:"enable_ssl,omitempty"`
	SkipSSLVerify bool          `json:"skip_ssl_verify,omitempty"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Addresses:     []string{"http://localhost:9200"},
		IndexPrefix:   "go-zero-logs",
		BufferSize:    100,
		FlushInterval: 5 * time.Second,
		EnableSSL:     false,
	}
}
