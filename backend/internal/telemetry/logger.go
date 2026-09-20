package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// LogEntry represents a structured JSON log record correlated with OpenTelemetry traces.
type LogEntry struct {
	Timestamp   string         `json:"time"`
	Level       string         `json:"level"`
	Message     string         `json:"msg"`
	TraceID     string         `json:"trace_id,omitempty"`
	SpanID      string         `json:"span_id,omitempty"`
	Service     string         `json:"service.name"`
	Environment string         `json:"deployment.environment"`
	Version     string         `json:"service.version"`
	Fields      map[string]any `json:"fields,omitempty"`
}

// Logger provides context-aware structured JSON logging.
type Logger struct {
	cfg Config
}

var globalLogger *Logger

func init() {
	cfg := LoadConfigFromEnv()
	globalLogger = &Logger{cfg: cfg}
}

// Info logs an informational message correlated with the active trace context.
func Info(ctx context.Context, msg string, fields ...any) {
	globalLogger.log(ctx, "INFO", msg, fields...)
}

// Warn logs a warning message correlated with the active trace context.
func Warn(ctx context.Context, msg string, fields ...any) {
	globalLogger.log(ctx, "WARN", msg, fields...)
}

// Error logs an error message correlated with the active trace context.
func Error(ctx context.Context, msg string, fields ...any) {
	globalLogger.log(ctx, "ERROR", msg, fields...)
}

func (l *Logger) log(ctx context.Context, level, msg string, fields ...any) {
	entry := LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
		Level:       level,
		Message:     msg,
		Service:     l.cfg.ServiceName,
		Environment: l.cfg.Environment,
		Version:     l.cfg.Version,
	}

	if ctx != nil {
		span := trace.SpanFromContext(ctx)
		if span != nil && span.SpanContext().IsValid() {
			entry.TraceID = span.SpanContext().TraceID().String()
			entry.SpanID = span.SpanContext().SpanID().String()
		}
	}

	if len(fields) > 0 {
		fieldMap := make(map[string]any)
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				key := fmt.Sprintf("%v", fields[i])
				fieldMap[key] = fields[i+1]
			}
		}
		if len(fieldMap) > 0 {
			entry.Fields = fieldMap
		}
	}

	data, err := json.Marshal(entry)
	if err == nil {
		fmt.Fprintln(os.Stdout, string(data))
	} else {
		fmt.Printf("[%s] %s: %s\n", level, time.Now().Format(time.RFC3339), msg)
	}
}
