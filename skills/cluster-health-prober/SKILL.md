---
name: cluster-health-prober
description: >-
  Probes running multi-tier cluster health across Go backend, Caddy reverse proxy,
  PostgreSQL database, and Floci local cloud emulator (S3 & SQS).
---

# Cluster Health Prober Capability Workflow

Use this capability workflow to verify live multi-tier cluster health across the ChoreSync stack.

---

## 1. Automated Cluster Probe Script

Run the automated cluster probe script:

```bash
bash agent-capabilities/cluster-health-prober/scripts/probe-cluster.sh
```

## 2. Monitored Tiers

1. **Go Backend API**:
   - Port `8080` (Direct) or `8088` (via Caddy)
   - Endpoint: `/healthz` or `/api/v1/chores`
2. **Caddy Reverse Proxy**:
   - Port `8088` (Prod) or `3000` (Dev)
   - Checks static assets and proxy pass-through to API
3. **PostgreSQL Relational DB**:
   - Port `5432`
   - Readiness: `pg_isready -U choresync -d choresync`
4. **Floci Cloud Emulator**:
   - Port `4566`
   - S3 bucket: `choresync-proofs`
   - SQS queue: `choresync-reminders`
