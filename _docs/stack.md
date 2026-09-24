# Technology Stack & Free-Tier Deployment Evaluation (`_docs/stack.md`)

**Project**: ChoreSync (Household Chore Coordination System)  
**Status**: ✅ Confirmed by User (Golang Architecture Selected)  
**Milestone**: Step 6 of AI-Native Spec-Driven Lifecycle  

---

## 1. Executive Summary & Context

ChoreSync requires a reliable, lightweight backend service implementing the 26 endpoints and 46 data schemas defined in [`contracts/openapi.yaml`](file:///home/manish/AI-Practice/DataTalks/ai-dev-tools-zoomcamp-2026/modules/01-ai-native-workflow/household-chores/contracts/openapi.yaml). The frontend is built with **React 19 + TypeScript + Vite**. 

Per the user's architectural selection in Step 6, the backend will be implemented in **Golang**. Go offers superior performance characteristics for free-tier cloud environments: ultra-low RAM footprint (~15–25MB vs ~100MB+ in Node/Python), near-instant cold boot (<50ms), and compact multi-stage Docker images (~20MB).

---

## 2. Architecture Comparison Matrix

| Dimension | Option 1: Golang + Chi / Koyeb (Confirmed) | Option 2: Go on Serverless (Cloud Run) | Option 3: TypeScript / Node PaaS |
| :--- | :--- | :--- | :--- |
| **Target Runtime** | **Golang 1.22+ + Chi (`go-chi/chi/v5`)** | **Golang 1.22+ (Cloud Run Container)** | Node.js 22 (TypeScript) + Express |
| **Compute Hosting** | **Koyeb** (Free Eco) or **Fly.io** | **GCP Cloud Run** (2M free req/mo) | Render / Koyeb |
| **Database** | **In-Memory** (Step 7) → **Neon Postgres** (`sqlc` + `pgx`) | **In-Memory** → **Neon Postgres** | In-Memory → Neon Postgres (Prisma) |
| **Observability** | **BetterStack** (Free Uptime & Logs) or **HyperDX** | **Google Cloud Logging** or **Grafana Cloud** | BetterStack / HyperDX |
| **RAM Footprint** | **~15–25 MB** (95% free headroom on 512MB) | **~15–25 MB** (Fits comfortably in 128MB tier) | ~80–120 MB |
| **Cold Start** | 0s on Koyeb; **<50ms** on spin-up | **<100ms** container boot | ~1–3s on Cloud Run; ~40s on Render |
| **Monthly Cost** | **$0.00 / month** | **$0.00 / month** | **$0.00 / month** |
| **Image Size** | **~15–25 MB** (Multi-stage `scratch`/`alpine`) | **~15–25 MB** | ~150–250 MB |

---

## 3. Deep-Dive: The Golang Backend Architecture

### Framework: Chi (`go-chi/chi/v5`)
* **Why Chi**:
  - 100% compliant with standard library `net/http.Handler`.
  - Zero external dependency bloat; extremely lightweight and fast.
  - Native integration with `oapi-codegen` to generate strictly-typed Go server interfaces and request/response models directly from [`contracts/openapi.yaml`](file:///home/manish/AI-Practice/DataTalks/ai-dev-tools-zoomcamp-2026/modules/01-ai-native-workflow/household-chores/contracts/openapi.yaml).
  - Robust standard middleware ecosystem (Logger, Recoverer, CORS, RequestID, Timeout).

### Persistence & Data Layer Strategy
1. **Step 7 (Scaffold & Test Harness)**:
   - **In-Memory Store**: Thread-safe in-memory repository implementing domain storage interfaces using Go structs with `sync.RWMutex`.
   - Complete support for all 26 endpoints with seed test fixtures matching [`_docs/manual-test.md`](file:///home/manish/AI-Practice/DataTalks/ai-dev-tools-zoomcamp-2026/modules/01-ai-native-workflow/household-chores/_docs/manual-test.md).
   - Instant unit test execution (<0.2s for entire test suite via `go test ./...`).
2. **Production Deployment**:
   - **`sqlc` + `pgx/v5`**: Compile-time type-safe SQL query generation for **Neon Serverless Postgres** (or pure Go `modernc.org/sqlite` without CGO).
   - Zero reflection overhead, optimal query performance, and compile-time verification of SQL queries against schema migrations.

---

## 4. Quotas, Failure Modes & Mitigations Table

| Failure Mode / Constraint | Affected Option | Impact | Prevention / Mitigation Strategy |
| :--- | :--- | :--- | :--- |
| **RAM Exhaustion** | Node/Python stacks | Container OOM killed on 512MB tiers | **Eliminated by Golang**: Go process consumes ~15–25MB, leaving >90% headroom. |
| **Cold Start Delays** | PaaS containers on sleep | 30–50s delay on first request | Go compiles to native binary; boots in <50ms. Koyeb Eco instance prevents sleeping altogether. |
| **Database Connection Spikes** | Postgres on serverless | Exhausts connection limit | Use Neon's pooled connection string (`?sslmode=require`) and configure `pgxpool.Config{MaxConns: 5}`. |
| **Contract Drift** | Hand-crafted models | API changes break client | Run `oapi-codegen` against `contracts/openapi.yaml` during CI/CD to verify Go interfaces. |

---

## 5. Architectural Decision Summary (Confirmed)

1. **Language & Runtime**: **Golang (Go 1.22+)**
2. **HTTP Routing**: **Chi (`go-chi/chi/v5`)** conforming to `contracts/openapi.yaml`
3. **Persistence**:
   - **Step 7**: Thread-safe in-memory repository (`sync.RWMutex`)
   - **Production**: Neon Serverless Postgres via `sqlc` + `pgx/v5`
4. **Cloud Target**:
   - Frontend: **Vercel / Cloudflare Pages** (Free CDN edge)
   - Backend: **Koyeb Free Eco** (512MB RAM, always active) or **GCP Cloud Run** (scale-to-zero)
   - Observability: **BetterStack** uptime monitor + `/healthz` endpoint

---

## 6. Verification & Approval Checklist

- [x] Evaluated 3 distinct architectural options with exact free tier parameters.
- [x] Evaluated Golang trade-offs, frameworks (Chi vs Gin vs Fiber), and memory footprint.
- [x] Documented potential failure modes (RAM, cold starts, connection pooling).
- [x] Identified migration path from Step 7 (In-Memory Go store) to Production (`sqlc` + Neon Postgres).
- [x] **Human Confirmation**: **CONFIRMED** — Golang selected for backend implementation. (Step 7 held until user signals to start).

---

## 7. Local Cloud Emulator Integration (Floci: S3 + SQS)

Per the AI-Native Spec-Driven Development Rule (Gate 6: Local Cloud First), ChoreSync integrates a local cloud emulator to achieve 100% cloud component coverage during offline development, local container testing, and automated CI pipelines before promoting to production cloud infrastructure.

### 7.1 Emulator Selection: Floci vs. Alternatives

| Emulator | Memory (RAM) | Boot Time | Services Emulated | Local Parity Verdict |
| :--- | :--- | :--- | :--- | :--- |
| **Floci (`floci/floci:latest`)** | **<50 MB** | **<0.5s** | S3, SQS, DynamoDB, SNS | **Selected**: Ultra-lightweight Quarkus native binary, near-zero host overhead, instant CI startup. |
| MinIO (`minio/minio`) | ~100–150 MB | ~1.5s | S3 only | Excellent for S3 alone, but lacks queueing (requires separate RabbitMQ/SQS emulator). |
| LocalStack (`localstack/localstack`) | ~800MB–1.5GB | ~15–30s | Full AWS suite | Excessive resource footprint for local development and lightweight CI runners. |

### 7.2 Emulated Cloud Components in ChoreSync

1. **S3 Object Storage (`choresync-proofs` bucket)**:
   - **Use Case**: Secure storage and retrieval of chore completion photo proofs submitted by household members.
   - **Endpoints**: `POST /api/v1/uploads/photo` (multipart image upload with 10MB limit and format validation) and `GET /api/v1/uploads/photo/{key}` (direct streaming retrieval).
   - **Local Bucket**: Auto-created on backend startup via the cloud client manager (`choresync-proofs`).

2. **SQS Message Queue (`choresync-reminders` queue)**:
   - **Use Case**: Decoupled asynchronous processing of chore reminder nudges, offloading email/push notifications from the HTTP request cycle.
   - **Endpoints**: `POST /api/v1/chores/{chore_id}/nudge` returns immediate response with `"sqs_queued": true`.
   - **Background Worker**: Long-polling background worker in the Go backend consumes messages from `choresync-reminders` and processes reminder dispatches.

### 7.3 Zero-Code-Forking & Environment Parity

The application uses official **AWS SDK for Go v2** (`github.com/aws/aws-sdk-go-v2`) for all cloud interactions. The codebase contains zero conditional environment branches (no `if env == "production"`):

- **Local / CI**:
  ```bash
  AWS_ENDPOINT_URL=http://floci:4566
  AWS_REGION=us-east-1
  AWS_ACCESS_KEY_ID=test
  AWS_SECRET_ACCESS_KEY=test
  S3_BUCKET_PROOFS=choresync-proofs
  SQS_QUEUE_REMINDERS=choresync-reminders
  ```
- **Production (AWS S3 & SQS)**:
  Configure `AWS_REGION`, standard AWS IAM role / credentials, and omit `AWS_ENDPOINT_URL` to route seamlessly to real AWS cloud services.

### 7.4 Common Orchestration Layer
- **Docker Compose**: Orchestrated via `choresync-floci` service on port `4566` in `docker-compose.yml`, `docker-compose.prod.yml`, and `docker-compose.deploy.yml`.
- **Kubernetes (`k8s/`)**: Declarative manifests in `k8s/floci-deployment.yaml` and `k8s/floci-service.yaml`, bundled via `k8s/kustomization.yaml` and injected via `k8s/configmap.yaml`.

