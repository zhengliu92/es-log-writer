package writer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresqlWriter PostgreSQL 写入器
type PostgresqlWriter struct {
	pool       *pgxpool.Pool
	config     *PostgresConfig
	buffer     []LogEntry
	bufferMu   sync.Mutex
	bufferSize int
	tableName  string
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	flushChan  chan struct{}
}

// NewPostgresqlWriter 创建一个新的 PostgreSQL Writer
func NewPostgresqlWriter(config *PostgresConfig) (*PostgresqlWriter, error) {
	if config == nil {
		config = DefaultPostgresConfig()
	}

	if config.DSN == "" {
		return nil, fmt.Errorf("postgres dsn cannot be empty")
	}
	if config.TableName == "" {
		config.TableName = "logs"
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 100
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	w := &PostgresqlWriter{
		pool:       pool,
		config:     config,
		buffer:     make([]LogEntry, 0, config.BufferSize),
		bufferSize: config.BufferSize,
		tableName:  config.TableName,
		ctx:        ctx,
		cancel:     cancel,
		flushChan:  make(chan struct{}, 1),
	}

	// 自动创建表
	if err := w.ensureTable(); err != nil {
		w.Close()
		return nil, err
	}

	w.wg.Add(1)
	go w.flushLoop()

	return w, nil
}

func (w *PostgresqlWriter) ensureTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGSERIAL PRIMARY KEY,
			timestamp TIMESTAMPTZ NOT NULL,
			level VARCHAR(20) NOT NULL,
			content TEXT,
			duration VARCHAR(50),
			trace VARCHAR(100),
			span VARCHAR(100),
			fields JSONB
		);
		CREATE INDEX IF NOT EXISTS idx_%s_timestamp ON %s(timestamp);
		CREATE INDEX IF NOT EXISTS idx_%s_level ON %s(level);
		CREATE INDEX IF NOT EXISTS idx_%s_trace ON %s(trace);
	`, w.tableName, w.tableName, w.tableName, w.tableName, w.tableName, w.tableName, w.tableName)

	_, err := w.pool.Exec(w.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

// log 内部日志方法
func (w *PostgresqlWriter) log(level string, content any, fields ...LogField) {
	trace, span, duration := extractFields(fields)
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Content:   FormatContent(content),
		Duration:  duration,
		Trace:     trace,
		Span:      span,
		Fields:    convertFields(fields),
	}
	w.AddEntry(entry)
}

// Log 写入日志
func (w *PostgresqlWriter) Log(level string, content any, fields ...LogField) {
	w.log(level, content, fields...)
}

// Info 写入 info 级别日志
func (w *PostgresqlWriter) Info(content any, fields ...LogField) {
	w.log("info", content, fields...)
}

// Error 写入 error 级别日志
func (w *PostgresqlWriter) Error(content any, fields ...LogField) {
	w.log("error", content, fields...)
}

// Debug 写入 debug 级别日志
func (w *PostgresqlWriter) Debug(content any, fields ...LogField) {
	w.log("debug", content, fields...)
}

// Warn 写入 warn 级别日志
func (w *PostgresqlWriter) Warn(content any, fields ...LogField) {
	w.log("warn", content, fields...)
}

// Ping 检查连接是否正常
func (w *PostgresqlWriter) Ping(ctx context.Context) error {
	return w.pool.Ping(ctx)
}

// Close 关闭写入器
func (w *PostgresqlWriter) Close() error {
	w.cancel()
	w.wg.Wait()
	w.flush()
	w.pool.Close()
	return nil
}

// AddEntry 添加日志条目到缓冲区
func (w *PostgresqlWriter) AddEntry(entry LogEntry) {
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
func (w *PostgresqlWriter) flush() error {
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

	// 使用 CopyFrom 进行批量插入
	rows := make([][]any, 0, len(entries))
	for _, entry := range entries {
		ts, _ := time.Parse(time.RFC3339, entry.Timestamp)
		fieldsJSON, _ := json.Marshal(entry.Fields)
		rows = append(rows, []any{
			ts,
			entry.Level,
			entry.Content,
			entry.Duration,
			entry.Trace,
			entry.Span,
			fieldsJSON,
		})
	}

	_, err := w.pool.CopyFrom(
		context.Background(),
		pgx.Identifier{w.tableName},
		[]string{"timestamp", "level", "content", "duration", "trace", "span", "fields"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return fmt.Errorf("failed to bulk insert logs to postgres: %w", err)
	}

	return nil
}

// flushLoop 刷新循环
func (w *PostgresqlWriter) flushLoop() {
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
