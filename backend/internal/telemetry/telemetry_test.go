package telemetry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/choresync/backend/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestInitTracer_And_Config(t *testing.T) {
	os.Setenv("OTEL_SERVICE_NAME", "test-backend")
	os.Setenv("DEPLOYMENT_ENVIRONMENT", "testing")
	os.Setenv("SERVICE_VERSION", "v1.0.0-test")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	defer func() {
		os.Unsetenv("OTEL_SERVICE_NAME")
		os.Unsetenv("DEPLOYMENT_ENVIRONMENT")
		os.Unsetenv("SERVICE_VERSION")
		os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}()

	cfg := telemetry.LoadConfigFromEnv()
	if cfg.ServiceName != "test-backend" {
		t.Errorf("expected ServiceName test-backend, got %s", cfg.ServiceName)
	}
	if cfg.Environment != "testing" {
		t.Errorf("expected Environment testing, got %s", cfg.Environment)
	}
	if cfg.Version != "v1.0.0-test" {
		t.Errorf("expected Version v1.0.0-test, got %s", cfg.Version)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	shutdown, err := telemetry.InitTracer(ctx)
	if err != nil {
		t.Fatalf("InitTracer failed: %v", err)
	}
	if shutdown == nil {
		t.Fatalf("expected non-nil shutdown func")
	}

	_ = shutdown(context.Background())
}

func TestInitMeter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	shutdown, err := telemetry.InitMeter(ctx)
	if err != nil {
		t.Fatalf("InitMeter failed: %v", err)
	}
	if shutdown == nil {
		t.Fatalf("expected non-nil shutdown func")
	}

	// Verify recording functions execute safely without panic
	telemetry.RecordHTTPRequest(ctx, "GET", "/api/v1/households", 200, 0.015)
	telemetry.RecordDBQuery(ctx, "postgresql.query", 0.005, nil)
	telemetry.RecordChoreCompleted(ctx, "h-roommates")

	_ = shutdown(context.Background())
}

func TestHTTPMiddleware_TracePropagation(t *testing.T) {
	// Set up in-memory exporter for testing
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	mw := telemetry.Middleware()

	var handlerExecuted bool
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerExecuted = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	req := httptest.NewRequest("GET", "/api/v1/households", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerExecuted {
		t.Fatalf("handler was not executed")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}

	// Verify traceparent header was injected into response
	traceparent := rec.Header().Get("traceparent")
	if traceparent == "" {
		t.Errorf("expected traceparent header in response, got empty")
	}

	// Verify span recorded
	spans := sr.Ended()
	if len(spans) == 0 {
		t.Fatalf("expected at least 1 ended span, got %d", len(spans))
	}
	span := spans[0]
	if span.Name() != "GET /api/v1/households" {
		t.Errorf("expected span name 'GET /api/v1/households', got '%s'", span.Name())
	}
}

func TestHTTPMiddleware_ExcludesHealthcheck(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)

	mw := telemetry.Middleware()

	var handlerExecuted bool
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerExecuted = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerExecuted {
		t.Fatalf("healthcheck handler was not executed")
	}

	spans := sr.Ended()
	if len(spans) != 0 {
		t.Errorf("expected 0 spans recorded for /healthz, got %d", len(spans))
	}
}

func TestPGXQueryTracer(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)

	tracer := telemetry.NewPGXQueryTracer()
	if tracer == nil {
		t.Fatalf("expected non-nil PGXQueryTracer")
	}
}
