package telemetry

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Config holds the standard telemetry resource attributes and collector configuration.
type Config struct {
	ServiceName string
	Environment string
	Version     string
	Endpoint    string
	Headers     map[string]string
}

// ParseHeaders parses comma-separated headers in either key=value or key: value format,
// safely handling base64 padding, quotes, and space-prefixed basic/bearer tokens.
func ParseHeaders(rawHeaders string) map[string]string {
	headers := make(map[string]string)
	rawHeaders = strings.Trim(strings.TrimSpace(rawHeaders), "\"'")
	if rawHeaders == "" {
		return headers
	}

	for _, part := range strings.Split(rawHeaders, ",") {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "\"'")
		if part == "" {
			continue
		}

		var key, val string
		idxColon := strings.Index(part, ":")
		idxEq := strings.Index(part, "=")

		// A valid HTTP header name cannot contain whitespace
		validKey := func(s string) bool {
			return len(s) > 0 && !strings.ContainsAny(s, " \t\r\n")
		}

		if idxColon != -1 && validKey(strings.TrimSpace(part[:idxColon])) && (idxEq == -1 || idxColon < idxEq) {
			key = strings.TrimSpace(part[:idxColon])
			val = strings.TrimSpace(part[idxColon+1:])
		} else if idxEq != -1 && validKey(strings.TrimSpace(part[:idxEq])) {
			key = strings.TrimSpace(part[:idxEq])
			val = strings.TrimSpace(part[idxEq+1:])
		} else if strings.HasPrefix(strings.ToLower(part), "basic ") || strings.HasPrefix(strings.ToLower(part), "bearer ") {
			key = "Authorization"
			val = part
		}

		val = strings.Trim(val, "\"'")
		if key != "" && val != "" {
			headers[key] = val
		}
	}

	return headers
}

// NormalizeOTLPPath ensures the URL path is appropriately formatted for OTLP HTTP endpoints.
// For Grafana Cloud endpoints, it ensures the required /otlp prefix is present.
func NormalizeOTLPPath(rawPath, defaultSubpath string, isGrafana bool) string {
	path := strings.TrimSuffix(rawPath, "/")
	if isGrafana && !strings.HasPrefix(path, "/otlp") {
		path = "/otlp" + path
	}
	if strings.HasSuffix(path, "/v1/traces") {
		path = strings.TrimSuffix(path, "/v1/traces")
	} else if strings.HasSuffix(path, "/v1/metrics") {
		path = strings.TrimSuffix(path, "/v1/metrics")
	}
	return strings.TrimSuffix(path, "/") + defaultSubpath
}

// LoadConfigFromEnv reads OpenTelemetry settings from environment variables with safe defaults.
func LoadConfigFromEnv() Config {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "api-backend"
	}

	env := os.Getenv("DEPLOYMENT_ENVIRONMENT")
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env == "" {
		env = "development"
	}

	ver := os.Getenv("SERVICE_VERSION")
	if ver == "" {
		ver = os.Getenv("APP_VERSION")
	}
	if ver == "" {
		ver = "dev-latest"
	}

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://otel-collector:4318"
	}

	rawHeaders := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	if rawHeaders == "" {
		rawHeaders = os.Getenv("OTEL_EXPORTER_OTLP_TRACES_HEADERS")
	}
	headers := ParseHeaders(rawHeaders)

	return Config{
		ServiceName: serviceName,
		Environment: env,
		Version:     ver,
		Endpoint:    endpoint,
		Headers:     headers,
	}
}

func buildResource(ctx context.Context, cfg Config) *resource.Resource {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
			semconv.ServiceVersionKey.String(cfg.Version),
		),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		res = resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
			semconv.ServiceVersionKey.String(cfg.Version),
		)
	}
	return res
}

// InitTracer initializes the OpenTelemetry TracerProvider, resource attributes, and OTLP exporter.
func InitTracer(ctx context.Context) (func(context.Context) error, error) {
	cfg := LoadConfigFromEnv()
	res := buildResource(ctx, cfg)

	// Register global error handler to surface OTel export failures to logs
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Printf("[OpenTelemetry Error] %v", err)
	}))

	// Configure OTLP HTTP trace exporter options
	opts := []otlptracehttp.Option{}
	parsedURL, err := url.Parse(cfg.Endpoint)
	if err == nil && parsedURL.Host != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(parsedURL.Host))
		if parsedURL.Scheme != "https" {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		isGrafana := strings.Contains(parsedURL.Host, "grafana.net")
		path := NormalizeOTLPPath(parsedURL.Path, "/v1/traces", isGrafana)
		opts = append(opts, otlptracehttp.WithURLPath(path))
	} else {
		host := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "http://"), "https://")
		opts = append(opts, otlptracehttp.WithEndpoint(host), otlptracehttp.WithInsecure())
	}

	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		log.Printf("[OpenTelemetry] Warning: failed to create OTLP trace exporter: %v", err)
		return func(context.Context) error { return nil }, nil
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	authPreview := "none"
	if auth, ok := cfg.Headers["Authorization"]; ok {
		if len(auth) > 16 {
			authPreview = auth[:10] + "..." + auth[len(auth)-4:]
		} else {
			authPreview = "configured"
		}
	}
	log.Printf("[OpenTelemetry] Tracing initialized: service.name=%s, deployment.environment=%s, service.version=%s, endpoint=%s, auth=%s",
		cfg.ServiceName, cfg.Environment, cfg.Version, cfg.Endpoint, authPreview)

	// Emit an immediate startup heartbeat span and flush so APM test connection checks succeed instantly
	go func() {
		time.Sleep(1 * time.Second)
		tr := tp.Tracer("choresync-system")
		_, span := tr.Start(context.Background(), "server.boot")
		span.SetAttributes(
			attribute.String("server.status", "online"),
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("deployment.environment", cfg.Environment),
			attribute.String("service.version", cfg.Version),
		)
		span.End()

		flushCtx, flushCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer flushCancel()
		if err := tp.ForceFlush(flushCtx); err != nil {
			log.Printf("[OpenTelemetry Error] Boot trace ForceFlush failed to %s: %v", cfg.Endpoint, err)
		} else {
			log.Printf("[OpenTelemetry] Boot trace successfully flushed to %s (service: %s, env: %s)", cfg.Endpoint, cfg.ServiceName, cfg.Environment)
		}
	}()

	return tp.Shutdown, nil
}

var (
	httpRequestsCounter    metric.Int64Counter
	httpDurationHistogram  metric.Float64Histogram
	dbErrorsCounter        metric.Int64Counter
	dbDurationHistogram    metric.Float64Histogram
	choresCompletedCounter metric.Int64Counter
)

// InitMeter initializes OpenTelemetry MeterProvider, instruments, and periodic metric exporter.
func InitMeter(ctx context.Context) (func(context.Context) error, error) {
	cfg := LoadConfigFromEnv()
	res := buildResource(ctx, cfg)

	opts := []otlpmetrichttp.Option{}
	parsedURL, err := url.Parse(cfg.Endpoint)
	if err == nil && parsedURL.Host != "" {
		opts = append(opts, otlpmetrichttp.WithEndpoint(parsedURL.Host))
		if parsedURL.Scheme != "https" {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		isGrafana := strings.Contains(parsedURL.Host, "grafana.net")
		path := NormalizeOTLPPath(parsedURL.Path, "/v1/metrics", isGrafana)
		opts = append(opts, otlpmetrichttp.WithURLPath(path))
	} else {
		host := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "http://"), "https://")
		opts = append(opts, otlpmetrichttp.WithEndpoint(host), otlpmetrichttp.WithInsecure())
	}

	if len(cfg.Headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(cfg.Headers))
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		log.Printf("[OpenTelemetry] Warning: failed to create OTLP metric exporter: %v", err)
		return func(context.Context) error { return nil }, nil
	}

	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(3*time.Second),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	meter := mp.Meter("github.com/choresync/backend")

	var meterErr error
	httpRequestsCounter, meterErr = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total count of HTTP requests processed"),
		metric.WithUnit("{request}"),
	)
	if meterErr != nil {
		log.Printf("[OpenTelemetry] Warning: failed to create http_requests_total counter: %v", meterErr)
	}

	httpDurationHistogram, meterErr = meter.Float64Histogram(
		"http_server_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
	)
	if meterErr != nil {
		log.Printf("[OpenTelemetry] Warning: failed to create http_server_request_duration_seconds histogram: %v", meterErr)
	}

	dbDurationHistogram, meterErr = meter.Float64Histogram(
		"db_query_duration_seconds",
		metric.WithDescription("Duration of database queries in seconds"),
		metric.WithUnit("s"),
	)
	if meterErr != nil {
		log.Printf("[OpenTelemetry] Warning: failed to create db_query_duration_seconds histogram: %v", meterErr)
	}

	dbErrorsCounter, meterErr = meter.Int64Counter(
		"db_query_errors_total",
		metric.WithDescription("Total count of failed database queries"),
		metric.WithUnit("{error}"),
	)
	if meterErr != nil {
		log.Printf("[OpenTelemetry] Warning: failed to create db_query_errors_total counter: %v", meterErr)
	}

	choresCompletedCounter, meterErr = meter.Int64Counter(
		"chores_completed_total",
		metric.WithDescription("Total count of completed chores"),
		metric.WithUnit("{chore}"),
	)
	if meterErr != nil {
		log.Printf("[OpenTelemetry] Warning: failed to create chores_completed_total counter: %v", meterErr)
	}

	if httpRequestsCounter != nil {
		httpRequestsCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("http.method", "BOOT"),
			attribute.String("http.route", "/server.boot"),
			attribute.Int("http.status_code", 200),
		))
	}

	log.Printf("[OpenTelemetry] Metrics initialized: service.name=%s, deployment.environment=%s, service.version=%s, interval=3s",
		cfg.ServiceName, cfg.Environment, cfg.Version)

	return mp.Shutdown, nil
}

// RecordHTTPRequest records status code and duration metrics for HTTP requests.
func RecordHTTPRequest(ctx context.Context, method, route string, statusCode int, duration float64) {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", route),
		attribute.Int("http.status_code", statusCode),
	}
	if httpRequestsCounter != nil {
		httpRequestsCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
	if httpDurationHistogram != nil {
		httpDurationHistogram.Record(ctx, duration, metric.WithAttributes(attrs...))
	}
}

// RecordDBQuery records duration and errors for database statements.
func RecordDBQuery(ctx context.Context, operation string, duration float64, err error) {
	if dbDurationHistogram != nil {
		dbDurationHistogram.Record(ctx, duration, metric.WithAttributes(
			attribute.String("db.operation", operation),
		))
	}
	if err != nil && dbErrorsCounter != nil {
		dbErrorsCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("db.operation", operation),
		))
	}
}

// RecordChoreCompleted records the completion of a chore for product telemetry.
func RecordChoreCompleted(ctx context.Context, householdID string) {
	if choresCompletedCounter != nil {
		choresCompletedCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("household.id", householdID),
		))
	}
}
