# ChoreSync Cluster & Services Architecture Reference

This reference details the network topology, service tiers, ports, and healthcheck probes for the ChoreSync system.

---

## 1. Multi-Tier Services Topology

```text
  Client Traffic (Browser / Mobile / Tablet)
         │
         ▼
  ┌─────────────────────────────────────────────────────────────┐
  │         Caddy 2 Reverse Proxy (Port 8088 / Prod)            │
  │     - Static SPA Frontend fallback (React + Vite)           │
  │     - Zero-CORS API reverse proxy routing (/api/ -> 8000)   │
  │     - Automatic Gzip & Zstandard compression                │
  └──────────────┬──────────────────────────────┬───────────────┘
                 │ /api/*                       │ /uploads/*
                 ▼                              ▼
  ┌─────────────────────────────┐ ┌─────────────────────────────┐
  │  Go 1.22 Chi Backend API    │ │  Floci Cloud Emulator (4566)│
  │  - Port 8000 (Internal)     │ │  - S3: choresync-proofs     │
  │  - Stateless HMAC JWT Auth  │ │  - SQS: choresync-reminders │
  │  - SQS Background Worker    │ │  - AWS SDK Go v2 integration│
  └──────────────┬──────────────┘ └─────────────────────────────┘
                 │
                 ▼
  ┌─────────────────────────────┐
  │  PostgreSQL 16 Database     │
  │  - Port 5432                │
  │  - 11 Relational Tables     │
  │  - Auto-migrations with DDL │
  └─────────────────────────────┘
```

---

## 2. Port & Endpoint Matrix

| Service Tier | Container Name | Host Port | Internal Port | Healthcheck / Probe |
|:---|:---|:---|:---|:---|
| **Frontend Dev** | `choresync-frontend` | `3000` | `3000` | `GET http://localhost:3000` |
| **Caddy Proxy** | `choresync-proxy` | `8088` | `80` | `GET http://localhost:8088/healthz` (200 OK) |
| **Go Backend API** | `choresync-backend` | `8000` (or `8080`) | `8000` | `GET http://localhost:8000/healthz` (JSON `{"status":"ok"}`) |
| **PostgreSQL DB** | `choresync-db` | `5432` | `5432` | `pg_isready -U choresync -d choresync` |
| **Floci Cloud** | `choresync-floci` | `4566` | `4566` | `GET http://localhost:4566/` |
| **Observability (OTel)**| `otel-collector` | `4317` (gRPC), `4318` (HTTP)| `4317/4318` | `GET http://localhost:13133/` |
| **Prometheus** | `prometheus` | `9090` | `9090` | `GET http://localhost:9090/-/healthy` |
| **Grafana** | `grafana` | `3001` | `3000` | `GET http://localhost:3001/api/health` |
| **Tempo (Traces)** | `tempo` | `3200` | `3200` | `GET http://localhost:3200/ready` |
| **Alertmanager** | `alertmanager` | `9093` | `9093` | `GET http://localhost:9093/-/healthy` |
| **Kubernetes Kind**| `choresync-cluster`| `8090` | `80` | `GET http://localhost:8090/healthz` |

---

## 3. Deployment Modes

1. **Local Development**: `docker compose up -d` (Vite HMR on `:3000`, Go Air live-reload on `:8000`, Postgres on `:5432`, Floci on `:4566`).
2. **Production Staging**: `docker compose -f docker-compose.prod.yml up -d` (Caddy on `:8088`, compiled Go binary, production assets).
3. **Local Kubernetes**: `make k8s-up` (Deployments, Services, 1Gi PVC, probes verified on `:8090`).
