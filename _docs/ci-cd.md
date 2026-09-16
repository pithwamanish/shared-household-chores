# Automated CI/CD Pipeline (`_docs/ci-cd.md`)

This document outlines the **Continuous Integration (CI)** and **Continuous Deployment (CD)** pipeline configured for ChoreSync using **GitHub Actions**, **Docker Compose**, and **Render**.

---

## 1. Pipeline Architecture

```
                  +----------------------------------------------+
                  |  Git Push / PR to main (or manual dispatch)  |
                  +----------------------------------------------+
                                          |
        +------------------+--------------+------------------+
        |                  |                                 |
        v                  v                                 v
+---------------+  +---------------+                 +---------------+
|  backend-ci   |  |  frontend-ci  |                 |contract-audit |
| - Go 1.22     |  | - Node.js 22  |                 | - OpenAPI 3.1 |
| - Postgres 16 |  | - Typecheck   |                 |   syntax &    |
| - Unit tests  |  | - Lint        |                 |   schema      |
| - Race check  |  | - Vite build  |                 |   validation  |
+---------------+  +---------------+                 +---------------+
        \                  /                                 /
         \________________/_________________________________/
                                  |
                                  v
                      +-----------------------+
                      |        e2e-ci         |
                      | - Docker Compose up   |
                      | - Healthcheck polling |
                      | - 11/11 Playwright    |
                      |   E2E test journeys   |
                      | - Artifacts on fail   |
                      +-----------------------+
                                  |
                                  v (Only on push to main)
================================================================================
                    TWO-STAGE CONTAINER REGISTRY PIPELINE
================================================================================
                                  |
                                  v
                      +-----------------------+
                      |   1. BUILD STAGE      |
                      |   (build-and-push)    |
                      | - Package Backend &   |
                      |   Frontend Docker     |
                      | - Tag: YYYYMMDD-      |
                      |   HHMMSS-shortsha     |
                      | - Tag: dev-latest     |
                      | - Push to GHCR        |
                      +-----------------------+
                                  |
                                  v
                      +-----------------------+
                      |   2. DEPLOY STAGE     |
                      |     (deploy-dev)      |
                      | - Pull from GHCR      |
                      | - Zero compiler/build |
                      | - Serve via           |
                      |   docker-compose.     |
                      |   deploy.yml          |
                      | - Verify /healthz     |
                      +-----------------------+
                                  |
                                  v
================================================================================
          MANUAL PROD PROMOTION WORKFLOW (.github/workflows/promote-to-prod.yml)
================================================================================
                                  |
                                  v (Manual workflow_dispatch with confirmation)
                      +-----------------------+
                      | Pull Deployed Dev     |
                      | Image from GHCR       |
                      +-----------------------+
                                  |
                                  v
                      +-----------------------+
                      | Re-tag for Production |
                      | - YYYYMMDD-HHMMSS-    |
                      |   shortsha            |
                      | - prod-latest         |
                      | - Push to GHCR        |
                      +-----------------------+
                                  |
                                  v
                      +-----------------------+
                      | Serve Promoted Image  |
                      | in Production         |
                      | - docker-compose.     |
                      |   deploy.yml          |
                      | - Live /healthz check |
                      +-----------------------+
```

---

## 2. CI/CD Jobs Overview

The workflows are defined in [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) and [`.github/workflows/promote-to-prod.yml`](../.github/workflows/promote-to-prod.yml):

### 2.1 Automated CI/CD Pipeline (`ci.yml`)

| Stage | Job | Trigger | Runtime / Environment | What It Validates |
| :--- | :--- | :--- | :--- | :--- |
| **Verification** | **`backend-ci`** | Push, PR | Go 1.22 + PostgreSQL 16 Alpine service | Code formatting (`go vet`), unit tests, in-memory store tests, and real PostgreSQL integration tests (`go test -v -race ./...`). |
| **Verification** | **`frontend-ci`** | Push, PR | Node.js 22 | TypeScript compilation (`tsc --noEmit`), lint checks, and Vite production bundle generation (`npm run build`). |
| **Verification** | **`contract-audit`** | Push, PR | Python 3 | Verifies [`contracts/openapi.yaml`](../contracts/openapi.yaml) presence and parses syntax against OpenAPI 3.1 YAML specification. |
| **Verification** | **`e2e-ci`** | Push, PR | Docker Compose + Playwright container | Boots full stack (`postgres`, `backend`, `frontend`), waits for `/healthz`, and runs all 11 end-to-end user journeys in headless Chromium. |
| **Stage 1: Build** | **`build-and-push`** | Push to `main`, dispatch | Docker Buildx + GHCR | Multi-stage Docker image build. Tags each image using mandatory **`YYYYMMDD-HHMMSS-shortsha`** pattern (e.g. `20260818-163457-83242da`) and `dev-latest`. Pushes images to GitHub Container Registry. |
| **Stage 2: Deploy** | **`deploy-dev`** | Push to `main`, dispatch | Docker Compose Deploy Stack | Pulls pre-built images from GHCR (zero compilation), serves stack via `docker-compose.deploy.yml`, and validates live HTTP `/healthz` response. |

### 2.2 Manual Production Promotion Workflow (`promote-to-prod.yml`)

Production follows the **"Build Once, Promote Everywhere"** paradigm. Images are never compiled from source during production promotion:

| Step | Action | Description |
| :--- | :--- | :--- |
| **1. Trigger** | `workflow_dispatch` | Triggered manually by an authorized engineer with confirmation (`confirm_promotion: promote`). |
| **2. Pull Dev Image** | Pull currently deployed Dev image | Pulls `ghcr.io/<repo>/backend:dev-latest` (or user-specified `YYYYMMDD-HHMMSS-shortsha` tag) and frontend image. |
| **3. Re-Tag & Push** | Re-tag for Production | Re-tags the identical binary image with timestamped `YYYYMMDD-HHMMSS-shortsha` and `prod-latest`, then pushes to GHCR. |
| **4. Deploy Prod** | Pull & Serve Promoted Image | Serves the promoted images via `docker-compose.deploy.yml` with zero build overhead. |
| **5. Health Verification** | Live `/healthz` check | Polls the live HTTP health check until `200 OK` is received. |
| **6. Audit Log** | Immutable Step Summary | Records image digests, tags, commit SHA, and promoter username in GitHub Actions run summary. |

---

## 3. Running the Pipeline Locally

Developers can test and run both stages locally using `make`:

### 3.1 Local CI Verification
```bash
make ci-local
```
Runs `verify` (frontend lint + Go test suite), `prod-build` (Docker build check), and `e2e` (Playwright suite).

### 3.2 Two-Stage Build and Deploy
```bash
# Stage 1: Build and tag images with YYYYMMDD-HHMMSS-shortsha pattern
make docker-build-tag
# or: make build-image

# Stage 2: Pull and serve pre-built images with zero compilation
make deploy-dev
# or: make deploy

# Teardown deploy stack
make deploy-down
```

### 3.3 Manual Production Promotion
```bash
# Pull deployed dev image, re-tag for prod, and serve in production
make promote-prod
```

---

## 4. CD Configuration & Webhooks

1. **GitHub Container Registry (GHCR)**:
   - Images are automatically published to `ghcr.io/<owner>/choresync/backend` and `ghcr.io/<owner>/choresync/frontend`.
   - Each build produces immutable timestamped tags (`YYYYMMDD-HHMMSS-shortsha`) and floating channel tags (`dev-latest`, `prod-latest`).
2. **Cloud Webhook Integration (Render / PaaS)**:
   - Optional webhook secrets (`RENDER_DEV_DEPLOY_HOOK_URL`, `RENDER_PROD_DEPLOY_HOOK_URL`) can be configured in GitHub Secrets.
   - When configured, the deploy stage dispatches the webhook to roll out the updated container registry images to cloud infrastructure.
