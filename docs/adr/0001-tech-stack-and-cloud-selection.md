# 1. Tech Stack, Cloud Architecture, and Local Cloud Emulator Selection

Date: 2026-09-25

## Status

Accepted

## Context

ChoreSync requires a reliable, lightweight, multi-tenant backend service implementing the 26 endpoints and 46 data schemas defined in the frozen OpenAPI 3.1 contract (`contracts/openapi.yaml`). The application supports three primary living arrangements:
1. Flatmates / Roommates (peer-to-peer accountability, auto-rotation, chore swaps)
2. Families with Children (admin approval gates, photo proof verification, gamified points/rewards)
3. Couples / Lightweight Co-living (voluntary task claiming, gentle nudges)

Key constraints and requirements:
- **Zero-Host-Runtime Mandate**: All local development, CI testing, and evaluation runs must execute inside ephemeral Docker containers without host compilers.
- **Local-Cloud-First & Cloud-Parity**: 100% of required cloud components (S3 object storage for photo proofs, SQS queueing for asynchronous chore reminder nudges) must be fully testable offline with zero remote billing dependencies.
- **Resource Efficiency**: High density and minimal cold-boot latency suitable for free-tier cloud deployment (Render, Fly.io, Neon).
- **Stateless Resilience**: Stateless authentication with HMAC-SHA256 JWT, transactional email fallback, and persistent PostgreSQL storage via embedded migrations.

We evaluated three potential backend architectures:
1. **Golang 1.22+ + Chi Router + `sqlc` + `pgx/v5`** with Floci local cloud emulator.
2. **Node.js 22 + TypeScript + Express + Prisma**.
3. **Python 3.12 + FastAPI + SQLAlchemy + LocalStack**.

## Decision

We selected **Option 1: Golang 1.22+ with Chi Router and PostgreSQL (`sqlc` + `pgx/v5`)**, paired with **Floci (`floci/floci:latest`)** as the local cloud emulator and **React 18 + Vite + TypeScript** for the frontend.

Key architectural selections:
- **Backend**: Go 1.22+ with `go-chi/chi/v5` router for native `net/http` compatibility, zero reflection overhead, sub-50ms cold starts, and <25MB runtime RAM consumption.
- **Data Persistence**: PostgreSQL 16 relational database with strictly-typed queries generated via `sqlc` and `jackc/pgx/v5`, embedded DDL auto-migrations across 11 tables, and thread-safe in-memory store for unit test suites.
- **Local Cloud Emulator**: Floci on port 4566 emulating AWS S3 (`choresync-proofs` bucket) and AWS SQS (`choresync-reminders` queue). Long-polling background worker in Go with AWS SDK Go v2, maintaining 100% unified Compose and Kubernetes manifests between local emulator and real cloud.
- **Observability**: Full-stack OpenTelemetry instrumentation with distributed W3C `traceparent` context propagation, decoupled standalone LGTM stack (OTel Collector, Prometheus, Loki, Tempo, Grafana).
- **Delivery**: Containerized multi-stage Docker builds, two-stage build/deploy pipeline stamping immutable `YYYYMMDD-HHMMSS-shortsha` tags, and local Kubernetes deployment with Kind.

## Consequences

### Positive
- Ultra-low memory consumption (15–25MB RAM), allowing multiple replicas and observability sidecars to run within free-tier limits.
- High developer iteration speed: Air hot-reloading for Go backend (<1s recompile) and Vite HMR for React frontend.
- Zero code forking: Standard AWS SDK Go v2 interacts with Floci locally and AWS in production purely through `AWS_ENDPOINT_URL` environment variables.
- Type safety: Zero contract drift through automated alignment against `contracts/openapi.yaml`.

### Negative / Trade-offs
- Static typing and code generation require running `sqlc` and schema migrations rather than dynamic ORM migrations.
- Ephemeral Docker execution requires caching module dependencies to avoid repeated network downloads during test runs.
