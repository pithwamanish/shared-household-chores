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

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "colon separator with base64 padding",
			input:    "Authorization: Basic MTAyNzU1MDpnbGMteHl6MTIz==",
			expected: map[string]string{"Authorization": "Basic MTAyNzU1MDpnbGMteHl6MTIz=="},
		},
		{
			name:     "equals separator with base64 padding",
			input:    "Authorization=Basic MTAyNzU1MDpnbGMteHl6MTIz==",
			expected: map[string]string{"Authorization": "Basic MTAyNzU1MDpnbGMteHl6MTIz=="},
		},
		{
			name:     "bare basic auth token without header key",
			input:    "Basic MTAyNzU1MDpnbGMteHl6MTIz==",
			expected: map[string]string{"Authorization": "Basic MTAyNzU1MDpnbGMteHl6MTIz=="},
		},
		{
			name:     "bare bearer token without header key",
			input:    "Bearer eyJhbGciOiJIUzI1NiJ9.test==",
			expected: map[string]string{"Authorization": "Bearer eyJhbGciOiJIUzI1NiJ9.test=="},
		},
		{
			name:     "wrapped in outer quotes",
			input:    `"Authorization=Basic dGVzdA=="`,
			expected: map[string]string{"Authorization": "Basic dGVzdA=="},
		},
		{
			name:     "multiple comma separated headers",
			input:    "Authorization: Basic dGVzdA==, X-Custom-Header: custom-val",
			expected: map[string]string{"Authorization": "Basic dGVzdA==", "X-Custom-Header": "custom-val"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := telemetry.ParseHeaders(tc.input)
			if len(got) != len(tc.expected) {
				t.Fatalf("expected %d headers, got %d: %v", len(tc.expected), len(got), got)
			}
			for k, v := range tc.expected {
				if got[k] != v {
					t.Errorf("header %s: expected %q, got %q", k, v, got[k])
				}
			}
		})
	}
}

func TestNormalizeOTLPPath(t *testing.T) {
	tests := []struct {
		name           string
		rawPath        string
		defaultSubpath string
		isGrafana      bool
		expected       string
	}{
		{
			name:           "Grafana bare domain",
			rawPath:        "",
			defaultSubpath: "/v1/traces",
			isGrafana:      true,
			expected:       "/otlp/v1/traces",
		},
		{
			name:           "Grafana with /otlp",
			rawPath:        "/otlp",
			defaultSubpath: "/v1/traces",
			isGrafana:      true,
			expected:       "/otlp/v1/traces",
		},
		{
			name:           "Grafana with /otlp/ and metrics",
			rawPath:        "/otlp/",
			defaultSubpath: "/v1/metrics",
			isGrafana:      true,
			expected:       "/otlp/v1/metrics",
		},
		{
			name:           "Grafana already with /otlp/v1/traces requested for metrics",
			rawPath:        "/otlp/v1/traces",
			defaultSubpath: "/v1/metrics",
			isGrafana:      true,
			expected:       "/otlp/v1/metrics",
		},
		{
			name:           "Local collector empty path",
			rawPath:        "",
			defaultSubpath: "/v1/traces",
			isGrafana:      false,
			expected:       "/v1/traces",
		},
		{
			name:           "Local collector /v1/traces",
			rawPath:        "/v1/traces",
			defaultSubpath: "/v1/traces",
			isGrafana:      false,
			expected:       "/v1/traces",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := telemetry.NormalizeOTLPPath(tc.rawPath, tc.defaultSubpath, tc.isGrafana)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

