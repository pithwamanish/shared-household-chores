# ChoreSync 🧹

> **An equitable, flexible household chore coordination system built with AI-native spec-driven engineering, zero-host containerized architecture, and full-stack observability.**

[![AI-Native Spec-Driven](https://img.shields.io/badge/Workflow-AI--Native%20Spec--Driven-blueviolet.svg)](_docs/specs.md)
[![Agent Plugins 1.0](https://img.shields.io/badge/Standard-Agent%20Plugins%201.0-blue.svg)](plugin.json)
[![OpenAPI 3.1](https://img.shields.io/badge/Contract-OpenAPI%203.1%20Frozen-success.svg)](contracts/openapi.yaml)
[![Tests Passing](https://img.shields.io/badge/Tests-15%2F15%20Go%20%7C%2012%2F12%20E2E-brightgreen.svg)](e2e/)
[![Semgrep Clean](https://img.shields.io/badge/Security-0%20Vulnerabilities-success.svg)](security-audit/)
[![Evaluation Compliant](https://img.shields.io/badge/Plugin%20Eval-100%25%20COMPLIANT-success.svg)](docs/agent-extension-pack-evaluation.md)

---

## 1. Problem Description & Core Archetypes

Domestic friction around household chores is one of the most common sources of co-living conflict. Existing chore apps either enforce rigid, bureaucratic micro-management or lack meaningful accountability.

**ChoreSync** eliminates this friction through a flexible, multi-persona domain model tailored to three core living arrangements:

1. **Flatmates / Roommates**:
   - Peer-to-peer accountability with automatic round-robin rotation.
   - Bilateral task-swapping marketplace with zero admin bottlenecks.
   - Transparent contribution history and asynchronous nudge reminders.
2. **Families with Children**:
   - Parent / Admin approval verification gates with mandatory photo and notes proof.
   - Point balances and reward store redemption with negative balance protection.
   - 4-digit Admin PIN unlock for shared kitchen tablets.
3. **Couples / Lightweight Co-living**:
   - Low-overhead shared task backlogs with voluntary task claiming.
   - 1-tap instant completions and personal "My Tasks" auto-filtering.

---

## 2. System Architecture

ChoreSync enforces a strict **Zero-Host-Runtime Mandate**: all services, databases, emulators, and test runners execute exclusively inside Docker containers or Kubernetes pods.

```text
                                  ┌───────────────────────────────┐
                                  │   Browser / Kitchen Tablet    │
                                  │   React 18 + Vite + TS SPA    │
                                  └───────────────┬───────────────┘
                                                  │ HTTP (Traceparent)
                                                  ▼
                                  ┌───────────────────────────────┐
                                  │        Caddy 2 Proxy          │
                                  │   (Compression, SPA Routing)  │
                                  └───────────────┬───────────────┘
                                                  │
                                                  ▼
                                  ┌───────────────────────────────┐
                                  │        Go 1.22 Backend        │
                                  │   (Chi Router + sqlc + pgx)   │
                                  └───┬───────────┬───────────┬───┘
                                      │           │           │
                     SQL (Tracing)    │           │ S3 API    │ SQS API
                                      ▼           ▼           ▼
  ┌────────────────────────┐  ┌───────────────┐ ┌────────────────────────────────┐
  │ Standalone OTel LGTM   │  │ PostgreSQL 16 │ │ Floci Local Cloud Emulator     │
  │ • OTel Collector       │  │ (11 Relational│ │ • S3: choresync-proofs (Photos)│
  │ • Prometheus & Tempo   │  │  Tables + DDL)│ │ • SQS: choresync-reminders     │
  │ • Loki & Grafana (3001)│  └───────────────┘ │ • Long-Polling Go Worker       │
  └────────────────────────┘                    └────────────────────────────────┘
```

### Core Technologies
- **Frontend**: React 18, Vite, TypeScript, Tailwind CSS, Lucide icons, centralized API client ([`frontend/src/services/api.ts`](frontend/src/services/api.ts)), segmented Kitchen Tablet vs. Personal Device toggle.
- **Backend API**: Go 1.22+, Chi HTTP router, `sqlc` + `pgx/v5` for type-safe database queries, embedded auto-migrations, transactional email notifications (Resend / SMTP / Dev Mailbox).
- **Storage & Queues**: Multi-provider architecture with **Cloudinary** (photo proof CDN) and **Neon PostgreSQL Queue** (`FOR UPDATE SKIP LOCKED`) in production, with **Floci** (`floci/floci:latest`, port 4566) providing local offline parity for AWS S3 and SQS. Zero code forking via official AWS Go SDK v2.
- **Database**: PostgreSQL 16 with persistent volume storage, relational DDL schema across 11 tables, and automated seed fixtures.
- **Reverse Proxy**: Caddy 2 with gzip/zstd compression, security headers, SPA fallback routing, and zero-CORS API proxying.
- **Observability**: Full-stack OpenTelemetry instrumentation with W3C distributed trace propagation across React SPA $\rightarrow$ Caddy $\rightarrow$ Go Backend $\rightarrow$ PostgreSQL (`pgx.QueryTracer`). Decoupled standalone LGTM stack (OTel Collector, Prometheus, Loki, Tempo, Grafana).
- **Kubernetes**: Local Kind cluster (`choresync-cluster`) with declarative manifests in [`k8s/`](k8s/) and persistent volume claims on port 8090.
- **Security**: Deterministic Semgrep SAST with zero findings, code-enforced autonomy policy outside the LLM, and read-only telemetry evidence assembly.

---

## 3. Canonical Repository Structure

This repository strictly conforms to the AI-Native Spec-Driven standard and Expected Repository Contents:

```text
├── README.md                           # Problem description, architecture & quickstart (this file)
├── product-spec.md                     # Product & functional specifications (bridged to _docs/specs.md)
├── constitution.md                     # Immutable project invariants (bridged to AGENTS.md)
├── AGENTS.md                           # AI agent operational rules & decision gates
├── openapi.yaml                        # Frozen OpenAPI 3.1 API contract (bridged to contracts/openapi.yaml)
├── docker-compose.yml                  # Root multi-service container orchestration
├── .github/workflows/                  # Automated CI/CD delivery pipelines (ci.yml, promote-to-prod.yml)
│
├── frontend/                           # React 18 + Vite + TypeScript application source & tests
├── backend/                            # Go 1.22 + Chi + sqlc backend application source & tests
├── k8s/                                # Declarative Kubernetes manifests for local Kind cluster
├── security/                           # Deterministic security rules & reports (bridged to security-audit/)
├── ops/                                # Standalone observability stack & runbooks (bridged to observability/)
├── .bmad/                              # BMAD 5-agent agile squad personas (bridged to custom-agent/)
├── evals/                              # Continuous agent eval harness & golden dataset
│
├── plugin.json                         # Agent Plugins 1.0 closed manifest (agent-plugins.org / AAIF)
├── mcp.json                            # Model Context Protocol root tool connection manifest
├── skills/                             # Reusable capability workflows (contract audit, db migrations, health, chores)
│   ├── contract-audit/                 # OpenAPI contract conformance auditor
│   ├── db-migration-runner/            # PostgreSQL 11-table schema migration verifier
│   ├── cluster-health-prober/          # Multi-tier container health prober
│   └── chore-lifecycle-manager/        # Domain lifecycle manager across 3 archetypes
├── com.antigravity.client/hooks/       # Pre-tool guardrails and post-tool audit loggers
├── mcp-server/                         # Zero-dependency Python 3 stdio MCP server (server.py)
├── custom-agent/                       # Specialist domain subagent persona (specialist.md)
│
└── docs/                               # Project documentation hub
    ├── ai-tools-usage-and-review.md    # Detailed AI tools usage, task delegation & review report
    ├── agent-extension-pack.md         # Extension pack usage & MCP instructions
    ├── agent-extension-pack-evaluation.md # 100% compliant conformance evaluation deliverable
    ├── permissions.md                  # Least-privilege policy & permission matrix
    └── operations-and-security-report.md # Deterministic security audit & resilience drill report
```

---

## 4. AI Tools Usage & Human Review

This project was built with AI assistance under strict human-in-the-loop governance across **16 Socratic Decision Gates**:
- **AI Tools Employed**: Antigravity / Gemini CLI (orchestration & gates), Claude Code (token-optimized backend & frontend coding via RTK proxy), Cursor & Windsurf (interactive tablet UI), and the BMAD Agile Squad (`pm`, `architect`, `developer`, `qa`, `scrum-master`).
- **Decomposition & Prompts**: Requirements were discovered without tech stack bias, contracts were reverse-engineered and frozen into [`openapi.yaml`](openapi.yaml) before backend coding, and features were implemented in isolated Git worktrees.
- **Context Files Provided**: Explicitly anchored with [`product-spec.md`](product-spec.md), [`constitution.md`](constitution.md), [`AGENTS.md`](AGENTS.md), [`openapi.yaml`](openapi.yaml), and [`docs/permissions.md`](docs/permissions.md).
- **Human Oversight**: Mandatory human approval at every architectural transition, git diff review, Semgrep security findings audit, and outside-the-model remediation authorization.
- **Verification Gates**: 15/15 Go test packages, 12/12 Playwright E2E journeys, Semgrep SAST (0 issues), spec-drift verification (0 drift), containerized mutation testing (100% killed mutants), and continuous agent evals (7/7 passing).

> 📖 **Full Report**: Read the complete prompt strategies, review processes, and governance policies in **[`docs/ai-tools-usage-and-review.md`](docs/ai-tools-usage-and-review.md)**.

---

## 5. Quickstart & Reproducibility Guide

### Prerequisites
- Docker Engine & Docker Compose (v2.20+)
- Kind (for local Kubernetes deployment)
- *(Host machine requires ZERO language runtimes)*

### 1. Launch Multi-Service Cluster
```bash
# Start Frontend, Go Backend, PostgreSQL, and Floci S3/SQS emulator
make compose-up

# Verify cluster status
docker compose ps
```
- **Web App**: http://localhost:3000
- **Floci Cloud Emulator**: http://localhost:4566
- **Backend API**: http://localhost:8000/healthz

### 2. Run Test Suites Inside Containers
```bash
# Run Go backend unit and integration tests
make test

# Run full Playwright E2E suite (12 user journeys)
make e2e
```

### 3. Observability & Telemetry (LGTM Stack)
```bash
# Launch standalone OTel Collector, Prometheus, Loki, Tempo, and Grafana
make obs-up
```
- **Grafana Dashboards**: http://localhost:3001 (pre-provisioned datasources & metrics)
- **Prometheus**: http://localhost:9090
- **Tempo Distributed Tracing**: http://localhost:3200

### 4. Local Kubernetes with Kind
```bash
# Spin up Kind cluster, build images, load offline, and verify pod readiness
make k8s-up
```
- **Kubernetes App Endpoint**: http://localhost:8090

### 5. Deterministic Security Audit & Resilience Drill
```bash
# Run containerized Semgrep static analysis (OWASP Top 10)
make security-scan

# Verify autonomous on-call responder safety contract
make oncall-verify
```

### 6. Agent Extension Pack & Continuous Evals
```bash
# Verify Agent Plugins 1.0 compliance
make ext-verify

# Evaluate extension pack conformance & domain alignment (Gate 14)
make ext-eval

# Run living contract drift detection, mutation testing, and agent evals (Gate 15)
make quality-reinforce

# Verify canonical repository contents & AI documentation (Gate 16)
make repo-verify
```

---

## 6. Development & Verification Lifecycle Ledger

| Gate / Step | Milestone | Canonical Artifact | Status |
|:---|:---|:---|:---|
| **Gate 0** | Archetype Selection & Ecosystem Advisory | `_docs/stack.md`, `constitution.md` | ✅ Complete |
| **Gate 1** | Requirements Discovery (Socratic) | `product-spec.md` (`_docs/specs.md`) | ✅ Complete |
| **Gate 2** | UI Prototype Source | `frontend/` (React + Vite + Tailwind) | ✅ Complete |
| **Gate 3** | Manual Verification Scenario | `_docs/manual-test.md` | ✅ Complete |
| **Gate 4** | Governance Setup & BMAD Agile Squad | `AGENTS.md`, `.bmad/`, `_docs/process.md` | ✅ Complete |
| **Gate 5** | Contract Freeze (OpenAPI 3.1) | `openapi.yaml` (`contracts/openapi.yaml`) | ✅ Complete |
| **Gate 6** | Tech Stack, Cloud Emulator & ADRs | `_docs/stack.md`, `docs/adr/`, Floci S3/SQS | ✅ Complete |
| **Gate 7–8** | Containerized Scaffold & Multi-Agent Worktrees | `backend/`, `frontend/`, `docker-compose.yml` | ✅ Complete |
| **Gate 9** | OpenTelemetry Full-Stack Instrumentation | `observability/`, `_docs/telemetry.md` | ✅ Complete |
| **Gate 10** | Local Kubernetes Deployment with Kind | `k8s/`, `_docs/kubernetes.md` | ✅ Complete |
| **Gate 11** | Local CI/CD Pipeline Execution with Act | `.github/workflows/ci.yml` | ✅ Complete |
| **Gate 12** | Deterministic Security Audit & Resilience Drill | `security/`, `on-call-engineer/`, `_docs/operations-and-security-report.md` | ✅ Complete |
| **Gate 13** | Agent Extension Pack (Agent Plugins 1.0) | `plugin.json`, `mcp.json`, `skills/`, `com.antigravity.client/hooks/` | ✅ Complete |
| **Gate 14** | Extension Pack Conformance & Domain Evaluation | `docs/agent-extension-pack-evaluation.md` (100% Score) | ✅ Complete |
| **Gate 15** | Quality Reinforcement (Drift, Mutation, Evals) | `evals/`, `verify-spec-drift`, `run-mutation-test` | ✅ Complete |
| **Gate 16** | AI Tools Usage Documentation & Canonical Packaging | `docs/ai-tools-usage-and-review.md`, `README.md` | ✅ Complete |

---

## 7. License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.
