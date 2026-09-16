# OpenTelemetry Full-Stack Instrumentation Specification (`_docs/telemetry.md`)

This document defines the distributed tracing, metrics, and structured observability architecture across ChoreSync's frontend, web server / reverse proxy, backend API, and database layers.

---

## 1. Golden Triangle Resource Attributes (Mandatory Invariants)

Every telemetry span and metric emitted across ChoreSync contains the following standardized resource attributes:

| Resource Attribute | Semantic Convention | Description | Development | Production |
| :--- | :--- | :--- | :--- | :--- |
| **Service Name** | `service.name` | Canonical logical identifier of the emitting tier | `web-frontend`, `api-backend`, `caddy-proxy` | `web-frontend`, `api-backend`, `caddy-proxy` |
| **Environment** | `deployment.environment` | Deployment lifecycle stage | `development` | `production` |
| **Deployed Version** | `service.version` | Immutable release tag or Git commit SHA | `dev-latest` | `YYYYMMDD-HHMMSS-shortsha` / `prod-latest` |

---

## 2. Distributed Tracing Topology & Context Propagation

ChoreSync enforces the standard **W3C Distributed Trace Context specification** (`traceparent` and `tracestate` headers) across the complete request lifecycle:

```
[Browser / React SPA]
  │ (OpenTelemetry WebTracer + FetchInstrumentation)
  │ Injects W3C `traceparent: 00-{traceId}-{spanId}-01`
  ▼
[Web Server / Ingress Proxy: Caddy 2]
  │ Passes through `traceparent`, `tracestate`, and injects `X-Request-ID`
  │ Reverse proxies to Backend API over internal container network
  ▼
[Backend API: Go 1.22 + Chi Router]
  │ (OpenTelemetry Go SDK + HTTP Server Middleware)
  │ Extracts W3C `traceparent` from headers and creates child Server Span
  │ Sets `http.method`, `http.target`, `http.route`, `http.status_code`, `client.address`
  │ Injects `traceparent` into HTTP response headers for downstream caller correlation
  ▼
[Database Tier: PostgreSQL 16 + pgx/v5]
  │ (OpenTelemetry PGXQueryTracer)
  │ Automatically creates child Client Span for each executed SQL statement
  │ Sets `db.system="postgresql"`, `db.name="choresync"`, `db.statement=<SQL>`, `net.peer.name="postgres"`
```

---

## 3. Tier-by-Tier Implementation Details

### 3.1 Web Frontend (`frontend/src/telemetry.ts`)
- **Library**: `@opentelemetry/sdk-trace-web`, `@opentelemetry/sdk-trace-base`, `@opentelemetry/exporter-trace-otlp-http`, `@opentelemetry/instrumentation-fetch`, `@opentelemetry/instrumentation-document-load`.
- **Initialization**: Imported at the very top of `frontend/src/main.tsx` before React mounting.
- **Header Propagation**: `FetchInstrumentation` configured with `propagateTraceHeaderCorsUrls` to automatically inject W3C `traceparent` into all API requests (`/api/*`).
- **Exporter**: OTLP HTTP JSON/Protobuf exporter targeting `VITE_OTEL_EXPORTER_URL` (default: `${window.location.protocol}//${window.location.hostname}:4318/v1/traces`).

### 3.2 Web Server / Reverse Proxy (`frontend/Caddyfile`)
- Configured to pass through and propagate incoming W3C trace headers:
  ```caddy
  handle /api/* {
      reverse_proxy {$BACKEND_URL:backend:8000} {
          header_up Host {upstream_hostport}
          header_up X-Real-IP {remote_host}
          header_up X-Forwarded-For {remote_host}
          header_up X-Forwarded-Proto {scheme}
          header_up traceparent {header.traceparent}
          header_up tracestate {header.tracestate}
          header_up X-Request-ID {uuid}
      }
  }
  ```

### 3.3 Backend API (`backend/internal/telemetry/`)
- **Initialization** (`telemetry.go`): Configures `sdktrace.NewTracerProvider` with `resource.New` (attaching `service.name`, `deployment.environment`, `service.version`), OTLP HTTP trace exporter, and global `propagation.NewCompositeTextMapPropagator(TraceContext{}, Baggage{})`.
- **HTTP Middleware** (`middleware.go`): Intercepts HTTP requests on Chi router:
  - Extracts W3C `traceparent`.
  - Creates server span: `HTTP {METHOD} {PATH}`.
  - Injects `traceparent` into response headers.
  - Excludes `/healthz`, `/health`, and `/metrics` from tracing to prevent log pollution.
- **Database Query Tracing** (`db.go`): Implements `pgx.QueryTracer` interface (`TraceQueryStart` and `TraceQueryEnd`) and attaches to `pgxpool.Config.ConnConfig.Tracer` in `backend/internal/store/postgres.go`.

---

## 4. Standalone Observability Stack (`observability/`)

In addition to the development collector, a decoupled production-grade observability stack runs in its own Compose project within [`observability/`](observability/docker-compose.yml):

- **OpenTelemetry Collector** (`observability-otel-collector`):
  - Ingests traces, metrics, and logs over OTLP gRPC (`4319` host -> `4317` container) and HTTP (`4318`).
  - Configured via [`observability/otel-collector-config.yaml`](observability/otel-collector-config.yaml).
  - Routes traces to **Tempo**, metrics to **Prometheus** exporter (`:8889`), and logs to **Loki** push API (`http://loki:3100/loki/api/v1/push`).
- **Prometheus** (`observability-prometheus`):
  - Scrapes OTel Collector metrics endpoint on port `8889` and internal telemetry on `8888`.
  - UI accessible at `http://localhost:9090`.
- **Grafana Loki** (`observability-loki`):
  - Log aggregation store running on port `3100`.
  - Ingests structured JSON/OTLP logs tagged with `service.name`, `deployment.environment`, `service.version`, and trace correlation IDs.
- **Grafana Tempo** (`observability-tempo`):
  - High-throughput distributed tracing backend running on port `3200`.
  - Ingests OTLP spans directly from OTel Collector via internal gRPC (`tempo:4317`).
- **Grafana** (`observability-grafana`):
  - Unified dashboard and explorer running on port `3001` (avoiding port 3000 collision with frontend).
  - Pre-provisioned datasources linking Tempo traces to Loki logs (`tracesToLogsV2`), Tempo to Prometheus metrics (`tracesToMetrics`), and Loki `trace_id` regex derived fields linking to Tempo.
  - Pre-provisioned overview dashboard under `Observability/Observability Overview`.

### Running the Observability Stack

```bash
# Start the observability stack
make obs-up
# or: docker compose -f observability/docker-compose.yml up -d

# View collector logs
make obs-logs
# or: docker compose -f observability/docker-compose.yml logs -f otel-collector

# Stop the stack
make obs-down
# or: docker compose -f observability/docker-compose.yml down
```

---

## 5. Active Container Verification Procedure

To verify live telemetry emission:

1. **Launch Stack**:
   ```bash
   docker compose up --build -d
   ```

2. **Trigger API Traffic**:
   ```bash
   curl -i http://localhost:8000/api/v1/households
   ```
   *Expected output: Response headers contain `Traceparent: 00-{traceId}-{spanId}-01`.*

3. **Verify Excluded Healthcheck**:
   ```bash
   curl -i http://localhost:8000/healthz
   ```
   *Expected output: Returns 200 OK without generating trace spans.*

4. **Inspect Collector Container Logs**:
   ```bash
   docker compose logs otel-collector
   # or for standalone stack:
   docker compose -f observability/docker-compose.yml logs otel-collector
   ```
   *Verification checklist:*
   - [x] Spans appear with `InstrumentationScope github.com/choresync/backend/http` and `github.com/choresync/backend/db`
   - [x] Resource attributes contain `service.name: api-backend` (or `web-frontend`)
   - [x] Resource attributes contain `deployment.environment: development`
   - [x] Resource attributes contain `service.version: dev-latest`
   - [x] Database child spans contain `db.system: postgresql`, `db.name: choresync`, and SQL statements
   - [x] Traces queryable in Tempo (`http://localhost:3200/api/traces/{traceId}`)
   - [x] Logs queryable in Loki (`http://localhost:3100/loki/api/v1/query_range`)
   - [x] Metrics scraped in Prometheus (`http://localhost:9090/api/v1/query?query=up`)
   - [x] Grafana UI active at `http://localhost:3001` with pre-configured datasources
