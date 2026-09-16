package telemetry

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type queryCtxKey struct{}

// PGXQueryTracer implements pgx.QueryTracer to automatically create child spans for SQL statements
// and record database query latency and error metrics.
type PGXQueryTracer struct {
	tracer trace.Tracer
}

// NewPGXQueryTracer creates a new tracer for PostgreSQL database operations.
func NewPGXQueryTracer() *PGXQueryTracer {
	return &PGXQueryTracer{
		tracer: otel.GetTracerProvider().Tracer("github.com/choresync/backend/db"),
	}
}

// TraceQueryStart starts a client span before executing a query on the PostgreSQL connection.
func (t *PGXQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx = context.WithValue(ctx, queryCtxKey{}, time.Now())
	ctx, _ = t.tracer.Start(ctx, "postgresql.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemPostgreSQL,
			semconv.DBNameKey.String("choresync"),
			semconv.DBStatementKey.String(data.SQL),
			semconv.NetPeerNameKey.String("postgres"),
			semconv.NetPeerPortKey.Int(5432),
		),
	)
	return ctx
}

// TraceQueryEnd ends the client span when the database query completes, recording errors and latency.
func (t *PGXQueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if startTime, ok := ctx.Value(queryCtxKey{}).(time.Time); ok {
		duration := time.Since(startTime).Seconds()
		var err error
		if data.Err != nil && data.Err != pgx.ErrNoRows {
			err = data.Err
		}
		RecordDBQuery(ctx, "postgresql.query", duration, err)
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	defer span.End()

	if data.Err != nil && data.Err != pgx.ErrNoRows {
		span.RecordError(data.Err)
		span.SetStatus(codes.Error, data.Err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}
}
