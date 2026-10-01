# Production Cloud Deployment & Hardening Guide (`_docs/deployment.md`)

This guide provides complete instructions for deploying **ChoreSync** to a generous free-tier cloud architecture using **Render** and **Neon Serverless PostgreSQL**, with production multi-stage Docker builds and **Caddy 2** reverse proxy.

---

## 1. Cloud Architecture & Free-Tier Blueprint

```
                     +---------------------------------------+
                     |             End User                  |
                     +---------------------------------------+
                                         |
                                         v HTTPS
                     +---------------------------------------+
                     |       Caddy 2 Web Server & Proxy      |
                     |  - Gzip / Zstd Compression            |
                     |  - Hardened Security Headers          |
                     |  - Static Asset Cache (1 Year)        |
                     |  - Client-Side SPA Fallback           |
                     +---------------------------------------+
                                  /             \
                   Static Assets /               \ /api/* & /healthz
                                v                 v
        +----------------------------+   +----------------------------+
        | React + Vite Production SPA|   | Statically-Compiled Go API |
        | (Pure static HTML/JS/CSS)  |   | (nobody:nobody unprivileged)|
        +----------------------------+   +----------------------------+
                                                          |
                                                          | TLS / pgx pool
                                                          v
                                         +----------------------------+
                                         |  Neon Serverless Postgres  |
                                         |  - Auto Schema Migrations  |
                                         |  - Auto Seed Initialization|
                                         |  - Scale-to-Zero Compute   |
                                         +----------------------------+
```

### Free-Tier Resource Allocation
| Component | Provider & Tier | Cost | Limits & Tradeoffs |
| :--- | :--- | :--- | :--- |
| **Go Backend** | Render Free Web Service | **$0/mo** | 512 MB RAM, 0.1 CPU. Sleeps after 15 min inactivity (~30s cold start on wake). |
| **Relational DB** | Neon Serverless Postgres | **$0/mo** | 0.5 GB storage, serverless autoscaling, scale-to-zero compute. |
| **Photo Proof Storage**| Cloudinary Free Tier (or S3/Floci) | **$0/mo** | 25 GB/month bandwidth and storage with global CDN. |
| **Background Queue** | Neon PostgreSQL Queue (or SQS/Floci) | **$0/mo** | ACID `FOR UPDATE SKIP LOCKED` worker utilizing existing database connection. |
| **Frontend & Proxy** | Render Free Docker Web Service (or Static Site) | **$0/mo** | Fast Caddy 2 reverse proxy with HTTP/2 & HTTP/3. |
| **Email Service** | Resend Free Tier (or In-Memory Dev Outbox) | **$0/mo** | 3,000 emails/month (100 emails/day) with custom domain. |

---

## 2. Neon Serverless PostgreSQL Provisioning

1. **Create Free Account**:
   - Navigate to [neon.tech](https://neon.tech) and sign up with GitHub or Google.
2. **Create Project**:
   - Project Name: `choresync`
   - Database Name: `neondb` (default)
   - Region: Select `US East (Ohio / us-east-2)` or `US West (Oregon)` to match your Render deployment region for lowest latency.
3. **Obtain Connection String**:
   - In the Neon Dashboard, copy the connection string:
     ```
     postgresql://[username]:[password]@[endpoint].neon.tech/neondb?sslmode=require
     ```
4. **Automated Schema Initialization**:
   - You do **not** need to run manual migration scripts or install external migration CLIs.
   - When the Go backend boots with `DATABASE_URL`, its embedded migration engine automatically executes `internal/db/schema.sql`, verifies existing tables, applies incremental schema patches, and populates initial demo households if the database is empty.

---

## 3. Render 1-Click Deployment (Blueprint)

ChoreSync includes a canonical Render Blueprint configuration in [`render.yaml`](../render.yaml).

### Step-by-Step Deployment
1. **Push to GitHub**:
   - Ensure your repository with `render.yaml`, `backend/Dockerfile`, and `frontend/Dockerfile.prod` is pushed to GitHub.
2. **Connect Blueprint on Render**:
   - Log into [dashboard.render.com](https://dashboard.render.com/).
   - Click **"New +"** in the top navigation and select **"Blueprint"**.
   - Connect your GitHub repository.
3. **Configure Environment Parameters**:
   - Render will parse `render.yaml` and discover two web services:
     - `choresync-backend`
     - `choresync-frontend`
   - When prompted for **`DATABASE_URL`**, paste your Neon connection string (`postgresql://...sslmode=require`).
   - `JWT_SECRET` is automatically generated with a secure random key.
   - When prompted for **`BACKEND_URL`** on `choresync-frontend`, provide the backend HTTPS URL: `https://choresync-backend.onrender.com` (Caddy reverse proxies `/api/*` with `header_up Host` preventing loops).
   - `STORAGE_PROVIDER` defaults to `cloudinary` (supply `CLOUDINARY_URL` or API keys from Cloudinary dashboard) or falls back to `mock`.
   - `QUEUE_PROVIDER` defaults to `neon` (ACID-safe background job queue reusing `DATABASE_URL` with zero extra setup).
   - If using real email dispatch, set `EMAIL_PROVIDER=resend` and enter your `RESEND_API_KEY`. Otherwise, leave default `EMAIL_PROVIDER=mock`.
4. **Deploy**:
   - Click **"Apply"**. Render will build both Docker images and launch the services.
5. **Access Your Application**:
   - Once deployment completes, your frontend will be live at `https://choresync-frontend.onrender.com`.

---

## 4. Alternative: Render Free Static Site (Zero Cold Start Frontend)

If you prefer your frontend to **never sleep** on Render:
1. In Render Dashboard, click **"New +"** -> **"Static Site"**.
2. Connect your repo:
   - **Root Directory**: `frontend`
   - **Build Command**: `npm ci && npm run build`
   - **Publish Directory**: `dist`
3. Add Environment Variable:
   - `VITE_API_URL`: `https://choresync-backend.onrender.com`
4. Add Rewrite Rule under **Redirects/Rewrites**:
   - Source: `/*`
   - Destination: `/index.html`
   - Action: `Rewrite`
5. Result: Frontend loads instantly from global CDN without sleeping; only backend wakes on first API request.

---

## 5. Local Production Stack Verification

You can run and test the exact production multi-stage build locally using Docker Compose:

```bash
# Launch production cluster with Caddy reverse proxy on port 8088
docker compose -f docker-compose.prod.yml up -d

# Verify running containers
docker compose -f docker-compose.prod.yml ps

# Check Caddy reverse proxy health check
curl -s http://localhost:8088/healthz

# Check REST API through reverse proxy
curl -s http://localhost:8088/api/v1/households

# Open in browser
open http://localhost:8088
```

To stop the production cluster:
```bash
docker compose -f docker-compose.prod.yml down
```

---

## 6. Production Hardening Checklist

- [x] **Unprivileged Container User**: Backend runs under `nobody:nobody` user with non-root privileges.
- [x] **Statically Linked Binary**: Built with `CGO_ENABLED=0 GOOS=linux -ldflags="-s -w"` in minimal Alpine runtime (~19.7 MB image).
- [x] **Secure Reverse Proxy**: Caddy 2 manages connection timeouts, buffer limits, and injects `X-Content-Type-Options`, `X-Frame-Options`, and `Referrer-Policy`.
- [x] **SSL / TLS Termination**: Full HTTPS encryption automatically provisioned by Render and Caddy.
- [x] **Neon Connection Pooling**: Go `pgxpool` configured with `MaxConns=10`, `MinConns=2`, and `MaxConnIdleTime=30m` optimized for Neon serverless scale-to-zero architecture.
- [x] **Resilient Token Invalidation**: Magic links strictly single-use and expire in 15 minutes; password resets expire in 60 minutes.

---

## 7. Two-Stage Build & Deploy Architecture (Registry-Backed)

To decouple image compilation from runtime serving and eliminate compiler overhead on production targets, deployment is divided into two distinct stages:

```
[Source Code] ──► [Stage 1: BUILD] ──► [Registry: ghcr.io] ──► [Stage 2: DEPLOY] ──► [SERVE]
                   - Multi-stage build  - Tag: YYYYMMDD-HHMMSS-shortsha - Pull pre-built image
                   - Zero runtime host  - dev-latest & prod-latest       - Zero build tooling
```

### 1. Tagging Specification: `YYYYMMDD-HHMMSS-shortsha`
Every image published to the container registry is stamped with an immutable UTC timestamp and short Git commit SHA:
```bash
# Example format: 20260818-163457-83242da
SHORT_SHA=$(git rev-parse --short=7 HEAD)
TIMESTAMP=$(date -u +'%Y%m%d-%H%M%S')
IMAGE_TAG="${TIMESTAMP}-${SHORT_SHA}"
```

### 2. Stage 1: Build & Push
Executed automatically on push to `main` via `.github/workflows/ci.yml`:
1. Compiles statically-linked Go binary in Alpine builder.
2. Compiles React + TypeScript frontend into static dist via Vite and packages with Caddy 2.
3. Tags and publishes to GitHub Container Registry (`ghcr.io`):
   - `ghcr.io/<repo>/backend:YYYYMMDD-HHMMSS-shortsha` & `dev-latest`
   - `ghcr.io/<repo>/frontend:YYYYMMDD-HHMMSS-shortsha` & `dev-latest`

### 3. Stage 2: Deploy (Pull & Serve)
Executed without rebuilding code from source:
```bash
# Pull and serve dev images using docker-compose.deploy.yml
make deploy-dev

# Or pull explicitly:
docker pull ghcr.io/<repo>/backend:dev-latest
docker pull ghcr.io/<repo>/frontend:dev-latest
docker compose -f docker-compose.deploy.yml up -d
```

### 4. Manual Production Promotion Workflow
Production deployments follow the **"Build once, promote everywhere"** invariant. The production environment NEVER rebuilds from source code. Instead, the manual promotion workflow (`.github/workflows/promote-to-prod.yml`) pulls the currently deployed dev image to production:

1. **Trigger**: In GitHub Actions, navigate to **"Promote Dev Image to Prod"** and click **"Run workflow"**.
2. **Pull Dev Image**: Workflow pulls `ghcr.io/<repo>/backend:dev-latest` (or a specific `YYYYMMDD-HHMMSS-shortsha` tag).
3. **Immutability Verification**: Captures SHA256 image digest ensuring byte-for-byte fidelity.
4. **Re-tag for Prod**: Tags the verified image with the mandatory `YYYYMMDD-HHMMSS-shortsha` pattern (e.g. `20260818-163457-83242da`), `${TIMESTAMP}-${SHORT_SHA}-prod`, and `prod-latest`.
5. **Serve to Prod**: Deploys the pre-built image and executes live HTTP health checks before completing rollout.

