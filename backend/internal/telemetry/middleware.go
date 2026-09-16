package telemetry

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterInterceptor) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

// Middleware returns a Chi-compatible HTTP middleware that extracts W3C trace context,
// creates a server span, instruments HTTP metrics/attributes, and injects context into outgoing headers.
func Middleware() func(http.Handler) http.Handler {
	tracer := otel.GetTracerProvider().Tracer("github.com/choresync/backend/http")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Exclude healthcheck and internal metrics endpoints from tracing & metrics to prevent skew
			if path == "/healthz" || path == "/health" || path == "/metrics" || strings.HasPrefix(path, "/healthz/") {
				next.ServeHTTP(w, r)
				return
			}

			// 1. Extract W3C distributed trace context from incoming request headers
			propagator := otel.GetTextMapPropagator()
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// 2. Start server span
			spanName := fmt.Sprintf("%s %s", r.Method, path)
			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPMethodKey.String(r.Method),
					semconv.HTTPTargetKey.String(r.URL.RequestURI()),
					semconv.HTTPRouteKey.String(path),
					semconv.HTTPUserAgentKey.String(r.UserAgent()),
					semconv.NetHostNameKey.String(r.Host),
				),
			)
			defer span.End()

			// 3. Inject trace context into response headers for downstream correlation
			propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			// 4. Wrap response writer to capture HTTP status code
			wrapped := &responseWriterInterceptor{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			start := time.Now()

			// 5. Execute downstream handlers with trace context in request
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			duration := time.Since(start).Seconds()

			// 6. Record status code and error flags on span
			span.SetAttributes(
				semconv.HTTPStatusCodeKey.Int(wrapped.statusCode),
			)

			if wrapped.statusCode >= 500 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d: %s", wrapped.statusCode, http.StatusText(wrapped.statusCode)))
			} else if wrapped.statusCode >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d Client Error", wrapped.statusCode))
			} else {
				span.SetStatus(codes.Ok, "OK")
			}

			// 7. Record OpenTelemetry HTTP metrics (Counter & Latency Histogram)
			RecordHTTPRequest(ctx, r.Method, path, wrapped.statusCode, duration)
		})
	}
}
