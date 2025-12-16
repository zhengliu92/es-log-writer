package logwriter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/zeromicro/go-zero/core/logx"
)

// ElasticsearchWriter 实现了 logx.Writer 接口，将日志写入 Elasticsearch
type ElasticsearchWriter struct {
	client     *elasticsearch.Client
	config     *Config
	buffer     []LogEntry
	bufferMu   sync.Mutex
	bufferSize int
	indexName  string
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	flushChan  chan struct{} // 用于触发刷新
}

// LogEntry 表示一条日志条目
type LogEntry struct {
	Timestamp string                 `json:"@timestamp"`
	Level     string                 `json:"level"`
	Content   string                 `json:"content"`
	Caller    string                 `json:"caller,omitempty"`
	Duration  string                 `json:"duration,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Config Elasticsearch Writer 配置
type Config struct {
	// Elasticsearch 地址列表
	Addresses []string `json:"addresses"`
	// 用户名（可选）
	Username string `json:"username,omitempty"`
	// 密码（可选）
	Password string `json:"password,omitempty"`
	// API Key（可选，优先级高于用户名密码）
	APIKey string `json:"api_key,omitempty"`
	// 索引名称前缀，实际索引名称为 {IndexPrefix}-{YYYY.MM.DD}
	IndexPrefix string `json:"index_prefix"`
	// 缓冲区大小，达到此大小后批量写入
	BufferSize int `json:"buffer_size"`
	// 刷新间隔，定期刷新缓冲区
	FlushInterval time.Duration `json:"flush_interval"`
	// 是否启用 SSL
	EnableSSL bool `json:"enable_ssl,omitempty"`
	// SSL 跳过验证（仅用于开发环境）
	SkipSSLVerify bool `json:"skip_ssl_verify,omitempty"`
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

// NewElasticsearchWriter 创建一个新的 Elasticsearch Writer
func NewElasticsearchWriter(config *Config) (*ElasticsearchWriter, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 设置默认值
	if len(config.Addresses) == 0 {
		config.Addresses = []string{"http://localhost:9200"}
	}
	if config.IndexPrefix == "" {
		config.IndexPrefix = "go-zero-logs"
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 100
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 5 * time.Second
	}

	// 构建 Elasticsearch 客户端配置
	esConfig := elasticsearch.Config{
		Addresses: config.Addresses,
	}

	// 设置认证
	if config.APIKey != "" {
		esConfig.APIKey = config.APIKey
	} else if config.Username != "" && config.Password != "" {
		esConfig.Username = config.Username
		esConfig.Password = config.Password
	}

	// 创建 Elasticsearch 客户端
	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	writer := &ElasticsearchWriter{
		client:     client,
		config:     config,
		buffer:     make([]LogEntry, 0, config.BufferSize),
		bufferSize: config.BufferSize,
		indexName:  config.IndexPrefix,
		ctx:        ctx,
		cancel:     cancel,
		flushChan:  make(chan struct{}, 1), // 带缓冲的 channel，避免阻塞
	}

	// 启动后台刷新协程
	writer.wg.Add(1)
	go writer.flushLoop()

	return writer, nil
}

// addEntry 添加日志条目到缓冲区
func (w *ElasticsearchWriter) addEntry(entry LogEntry) {
	w.bufferMu.Lock()
	w.buffer = append(w.buffer, entry)
	shouldFlush := len(w.buffer) >= w.bufferSize
	w.bufferMu.Unlock()

	// 如果缓冲区满了，触发刷新（非阻塞）
	if shouldFlush {
		select {
		case w.flushChan <- struct{}{}:
		default:
			// channel 已满，说明已经有 flush 请求在等待，跳过
		}
	}
}

// formatContent 格式化日志内容
func (w *ElasticsearchWriter) formatContent(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case error:
		return val.Error()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatFields 格式化日志字段
func (w *ElasticsearchWriter) formatFields(fields ...logx.LogField) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	result := make(map[string]interface{})
	for _, field := range fields {
		result[field.Key] = field.Value
	}
	return result
}

// getCaller 获取调用者信息
func (w *ElasticsearchWriter) getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return ""
	}
	// 只保留文件名和行号
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// Alert 实现 logx.Writer 接口
func (w *ElasticsearchWriter) Alert(v any) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "alert",
		Content:   w.formatContent(v),
		Caller:    w.getCaller(1),
	}
	w.addEntry(entry)
}

// Debug 实现 logx.Writer 接口
func (w *ElasticsearchWriter) Debug(v any, fields ...logx.LogField) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "debug",
		Content:   w.formatContent(v),
		Caller:    w.getCaller(1),
		Fields:    w.formatFields(fields...),
	}
	w.addEntry(entry)
}

// Error 实现 logx.Writer 接口
func (w *ElasticsearchWriter) Error(v any, fields ...logx.LogField) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "error",
		Content:   w.formatContent(v),
		Caller:    w.getCaller(1),
		Fields:    w.formatFields(fields...),
	}
	w.addEntry(entry)
}

// Info 实现 logx.Writer 接口
func (w *ElasticsearchWriter) Info(v any, fields ...logx.LogField) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "info",
		Content:   w.formatContent(v),
		Caller:    w.getCaller(1),
		Fields:    w.formatFields(fields...),
	}
	w.addEntry(entry)
}

// Severe 实现 logx.Writer 接口
func (w *ElasticsearchWriter) Severe(v any) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "severe",
		Content:   w.formatContent(v),
		Caller:    w.getCaller(1),
	}
	w.addEntry(entry)
}

// Slow 实现 logx.Writer 接口
func (w *ElasticsearchWriter) Slow(v any, fields ...logx.LogField) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "slow",
		Content:   w.formatContent(v),
		Caller:    w.getCaller(1),
		Fields:    w.formatFields(fields...),
	}
	// 尝试从 fields 中提取 duration
	for _, field := range fields {
		if field.Key == "duration" {
			if dur, ok := field.Value.(time.Duration); ok {
				entry.Duration = dur.String()
			} else {
				entry.Duration = fmt.Sprintf("%v", field.Value)
			}
			break
		}
	}
	w.addEntry(entry)
}

// Stack 实现 logx.Writer 接口
func (w *ElasticsearchWriter) Stack(v any) {
	// 获取堆栈信息
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stackTrace := string(buf[:n])

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "stack",
		Content:   w.formatContent(v),
		Caller:    w.getCaller(1),
		Fields: map[string]interface{}{
			"stack": stackTrace,
		},
	}
	w.addEntry(entry)
}

// Stat 实现 logx.Writer 接口
func (w *ElasticsearchWriter) Stat(v any, fields ...logx.LogField) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "stat",
		Content:   w.formatContent(v),
		Caller:    w.getCaller(1),
		Fields:    w.formatFields(fields...),
	}
	w.addEntry(entry)
}

// Close 实现 logx.Writer 接口
func (w *ElasticsearchWriter) Close() error {
	w.cancel()
	w.wg.Wait()
	return w.flush()
}

// flush 刷新缓冲区，将日志批量写入 Elasticsearch
func (w *ElasticsearchWriter) flush() error {
	w.bufferMu.Lock()
	if len(w.buffer) == 0 {
		w.bufferMu.Unlock()
		return nil
	}

	// 复制缓冲区内容
	entries := make([]LogEntry, len(w.buffer))
	copy(entries, w.buffer)
	w.buffer = w.buffer[:0]
	w.bufferMu.Unlock()

	if len(entries) == 0 {
		return nil
	}

	// 生成索引名称（按日期）
	indexName := w.getIndexName()

	// 构建批量请求
	var buf bytes.Buffer
	for _, entry := range entries {
		// 构建 bulk API 的元数据行
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
			},
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			// 如果序列化失败，跳过这条日志
			continue
		}
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		// 构建文档内容
		docJSON, err := json.Marshal(entry)
		if err != nil {
			// 如果序列化失败，跳过这条日志
			continue
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	// 如果没有有效数据，直接返回
	if buf.Len() == 0 {
		return nil
	}

	// 执行批量写入
	req := esapi.BulkRequest{
		Body:    bytes.NewReader(buf.Bytes()),
		Refresh: "false",
	}

	res, err := req.Do(context.Background(), w.client)
	if err != nil {
		return fmt.Errorf("failed to execute bulk request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch error: %s", res.String())
	}

	return nil
}

// getIndexName 获取当前日期的索引名称
func (w *ElasticsearchWriter) getIndexName() string {
	today := time.Now().Format("2006.01.02")
	return fmt.Sprintf("%s-%s", w.indexName, today)
}

// flushLoop 定期刷新缓冲区
func (w *ElasticsearchWriter) flushLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			// 退出前刷新剩余日志
			w.flush()
			return
		case <-ticker.C:
			// 定期刷新
			w.flush()
		case <-w.flushChan:
			// 缓冲区满时触发的刷新
			w.flush()
		}
	}
}
