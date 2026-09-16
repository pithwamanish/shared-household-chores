# Observability Stack (LGTM + OpenTelemetry Collector)

This directory contains a standalone, decoupled Docker Compose project for full-stack observability across **ChoreSync** services.

---

## 1. Architecture Overview

```
                          ┌──────────────────────────┐
                          │     Web Browser / SPA    │
                          └─────────────┬────────────┘
                                        │ HTTP (4318)
                                        ▼
┌──────────────────┐      ┌──────────────────────────┐
│  Go Backend API  ├─────►│ OpenTelemetry Collector │
└──────────────────┘ HTTP │  - Receivers: gRPC/HTTP  │
                     (4318)│  - Processors: batch/mem │
                          └──────┬──────┬──────┬─────┘
                                 │      │      │
                ┌────────────────┘      │      └────────────────┐
                │ OTLP (4317)           │ Prometheus (8889)     │ HTTP Push (3100)
                ▼                       ▼                       ▼
      ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
      │   Grafana Tempo  │    │    Prometheus    │    │   Grafana Loki   │
      │ (Distributed Traces)│  │ (Metrics Store)  │    │  (Log Streams)   │
      └─────────┬────────┘    └─────────┬────────┘    └─────────┬────────┘
                │                       │                       │
                └────────────────►┌─────┴──────┐◄───────────────┘
                                  │   Grafana  │
                                  │ (Port 3001)│
                                  └────────────┘
```

### Components

| Service | Container Name | Host Port | Internal Port | Description |
| :--- | :--- | :--- | :--- | :--- |
| **OpenTelemetry Collector** | `observability-otel-collector` | `4317` (gRPC), `4318` (HTTP), `8889` (Prom metrics), `8888` (health) | `4317`, `4318`, `8889`, `8888` | Ingests OTLP traces, metrics, and logs; batches, enriches, and fans out to backends. |
| **Prometheus** | `observability-prometheus` | `9090` | `9090` | Time-series metrics engine scraping OTel Collector exporter. |
| **Grafana Loki** | `observability-loki` | `3100` | `3100` | Log aggregation system indexing structured log streams. |
| **Grafana Tempo** | `observability-tempo` | `3200` | `3200` | High-volume distributed trace storage and query engine. |
| **Grafana** | `observability-grafana` | `3001` | `3000` | Unified visualization UI with pre-provisioned datasources and cross-drilldown navigation. |

---

## 2. Quick Start

### Starting the Observability Stack

From repository root:
```bash
docker compose -f observability/docker-compose.yml up -d
```
Or navigate into the directory:
```bash
cd observability && docker compose up -d
```

### Checking Service Health & Logs

```bash
docker compose -f observability/docker-compose.yml ps
docker compose -f observability/docker-compose.yml logs -f otel-collector
```

### Stopping the Observability Stack

```bash
docker compose -f observability/docker-compose.yml down
```

To purge volumes (Prometheus metrics, Loki logs, Tempo traces, Grafana state):
```bash
docker compose -f observability/docker-compose.yml down -v
```

---

## 3. Accessing Services

- **Grafana**: [http://localhost:3001](http://localhost:3001) (Credentials: `admin` / `admin`)
  - **Datasources**: Pre-provisioned with Prometheus, Loki, and Tempo.
  - **Dashboards**: Pre-provisioned `Observability Overview` under the *Observability* folder.
- **Prometheus UI**: [http://localhost:9090](http://localhost:9090)
  - Targets status: [http://localhost:9090/targets](http://localhost:9090/targets)
- **Tempo HTTP Endpoint**: [http://localhost:3200](http://localhost:3200)
- **Loki Ready Check**: [http://localhost:3100/ready](http://localhost:3100/ready)
- **OTel Collector Metrics**: [http://localhost:8889/metrics](http://localhost:8889/metrics)

---

## 4. Connecting the Application Stack

The observability stack runs on its own isolated bridge network `observability-net` and exposes standard ports on `localhost`.

### A. Via Host Ports (Default)
Any application running on the host or inside a container reaching `localhost` / `host.docker.internal` can publish to the collector:
- **Traces/Metrics/Logs OTLP HTTP**: `http://localhost:4318` (or `http://host.docker.internal:4318`)
- **Traces/Metrics/Logs OTLP gRPC**: `localhost:4317`

### B. Via Docker Bridge Network
To connect the ChoreSync application stack directly over container networking:
1. Ensure the observability stack is running (`docker compose -f observability/docker-compose.yml up -d`).
2. Attach the backend container to `observability-net`:
   ```yaml
   networks:
     default:
     observability-net:
       external: true
   ```
3. Set the backend endpoint:
   ```bash
   OTEL_EXPORTER_OTLP_ENDPOINT=http://observability-otel-collector:4318
   ```

---

## 5. Pre-Configured Cross-Navigation in Grafana

The Grafana provisioning file (`grafana/provisioning/datasources/datasources.yaml`) links the three telemetry pillars:

1. **Traces → Logs**: When inspecting a trace in Tempo, clicking a span reveals a direct link to query Loki logs matching that service and timestamp range.
2. **Traces → Metrics**: Click a span to view associated service latency and rate metrics in Prometheus.
3. **Logs → Traces**: Any Loki log line containing `trace_id=...` or `TraceID=...` renders a clickable link taking you straight into Tempo to inspect the full trace waterfall.
