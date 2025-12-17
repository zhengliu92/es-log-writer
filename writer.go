package writer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticsearchWriter 核心写入器（不依赖 go-zero）
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
	flushChan  chan struct{}
}

// NewElasticsearchWriter 创建一个新的 Elasticsearch Writer
func NewElasticsearchWriter(config *Config) (*ElasticsearchWriter, error) {
	if config == nil {
		config = DefaultConfig()
	}

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

	esConfig := elasticsearch.Config{
		Addresses: config.Addresses,
	}

	if config.APIKey != "" {
		esConfig.APIKey = config.APIKey
	} else if config.Username != "" && config.Password != "" {
		esConfig.Username = config.Username
		esConfig.Password = config.Password
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	w := &ElasticsearchWriter{
		client:     client,
		config:     config,
		buffer:     make([]LogEntry, 0, config.BufferSize),
		bufferSize: config.BufferSize,
		indexName:  config.IndexPrefix,
		ctx:        ctx,
		cancel:     cancel,
		flushChan:  make(chan struct{}, 1),
	}

	w.wg.Add(1)
	go w.flushLoop()

	return w, nil
}

// Log 写入日志（核心方法）
func (w *ElasticsearchWriter) Log(level string, content any, fields ...LogField) {
	trace, span, duration := extractFields(fields)
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Content:   FormatContent(content),
		Caller:    GetCaller(2),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertFields(fields),
	}
	w.AddEntry(entry)
}

// Info 写入 info 级别日志
func (w *ElasticsearchWriter) Info(content any, fields ...LogField) {
	w.Log("info", content, fields...)
}

// Error 写入 error 级别日志
func (w *ElasticsearchWriter) Error(content any, fields ...LogField) {
	w.Log("error", content, fields...)
}

// Debug 写入 debug 级别日志
func (w *ElasticsearchWriter) Debug(content any, fields ...LogField) {
	w.Log("debug", content, fields...)
}

// Warn 写入 warn 级别日志
func (w *ElasticsearchWriter) Warn(content any, fields ...LogField) {
	w.Log("warn", content, fields...)
}

// Ping 检查 Elasticsearch 连接是否正常
func (w *ElasticsearchWriter) Ping(ctx context.Context) error {
	res, err := w.client.Info()
	if err != nil {
		return fmt.Errorf("failed to ping elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch ping failed: %s", res.String())
	}

	return nil
}

// Close 关闭写入器
func (w *ElasticsearchWriter) Close() error {
	w.cancel()
	w.wg.Wait()
	return w.flush()
}

// AddEntry 添加日志条目到缓冲区（导出供适配器使用）
func (w *ElasticsearchWriter) AddEntry(entry LogEntry) {
	w.bufferMu.Lock()
	w.buffer = append(w.buffer, entry)
	shouldFlush := len(w.buffer) >= w.bufferSize
	w.bufferMu.Unlock()

	if shouldFlush {
		select {
		case w.flushChan <- struct{}{}:
		default:
		}
	}
}

// flush 刷新缓冲区
func (w *ElasticsearchWriter) flush() error {
	w.bufferMu.Lock()
	if len(w.buffer) == 0 {
		w.bufferMu.Unlock()
		return nil
	}

	entries := make([]LogEntry, len(w.buffer))
	copy(entries, w.buffer)
	w.buffer = w.buffer[:0]
	w.bufferMu.Unlock()

	if len(entries) == 0 {
		return nil
	}

	indexName := w.getIndexName()

	var buf bytes.Buffer
	for _, entry := range entries {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
			},
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			continue
		}
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		docJSON, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	if buf.Len() == 0 {
		return nil
	}

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

// getIndexName 获取索引名称（按日期）
func (w *ElasticsearchWriter) getIndexName() string {
	today := time.Now().Format("2006.01.02")
	return fmt.Sprintf("%s-%s", w.indexName, today)
}

// flushLoop 刷新循环（后台 goroutine）
func (w *ElasticsearchWriter) flushLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.flush()
			return
		case <-ticker.C:
			w.flush()
		case <-w.flushChan:
			w.flush()
		}
	}
}
