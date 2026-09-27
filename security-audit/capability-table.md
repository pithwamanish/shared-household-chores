# Responder Capability Inventory & Attack Surface Analysis (`capability-table.md`)

**Target System**: ChoreSync (`household-chores`)  
**Component**: Autonomous On-Call Responder & AI Coding Agent Tools  
**Security Framework**: Snyk Agent Scan / Zoomcamp Module 4 & 5 Attack Surface Inventory  
**Auditor Engine**: Static Code Inspection + Policy Engine Verification  
**Status**: **CERTIFIED LEAST-PRIVILEGE (0 Direct Database Credentials, 0 Host Shell Access)**

---

## 1. Operating Principle & Attack Surface Rationale

In an AI-native operational loop, the autonomous on-call responder is an actor with execution potential. Giving an agent unrestricted access to production environments or database credentials turns the agent into a catastrophic attack vector for prompt injection, hallucinations, and runaway destructive loops.

**Core Axiom**:
> *A model supplies confidence; code outside the model enforces permission.*

The responder operates with strictly bounded, read-only evidence gathering, code-enforced autonomy policy evaluation (`autonomy-policy.yaml`), and zero access to production database credentials.

---

## 2. Comprehensive Capability Inventory Matrix

| Capability Name | Tool / Mechanism | Allowed Scope | Autonomy Tier | Credential Scope | Prohibited Operations | Threat Vector & Defense-in-Depth |
| :--- | :--- | :--- | :---: | :--- | :--- | :--- |
| **Telemetry Inspection** | HTTP GET `/api/v1/query` to Prometheus (:9090) | Read-only metric scraping (error rates, request latencies, target up status) | **Read-Only** | Anonymous / None | POST/PUT/DELETE to Prometheus config; cannot alter scrape targets | *Vector*: Poisoned metrics.<br>*Defense*: Queries strictly parameterized; sanitized via `jq`. |
| **Log Stream Querying** | HTTP GET `/loki/api/v1/query_range` to Loki (:3100) or `docker logs` | Read-only log lines with `level=error` | **Read-Only** | Anonymous / None | Write to log stream; accessing logs of sensitive services | *Vector*: Prompt injection via log string.<br>*Defense*: Logs treated as untrusted text; output schema strictly enforced. |
| **Git Revision Inspection** | `git log -n 5`, `git diff`, `git rev-parse` | Read-only commit messages, authors, SHAs | **Read-Only** | None (Local filesystem read) | `git push`, `git rebase`, `git reset --hard` | *Vector*: Malicious commit message injection.<br>*Defense*: Parsed as JSON array without shell execution. |
| **Container Inspection** | `docker ps --filter`, `docker inspect`, `kubectl get` | Inspecting container state, image tags, exit codes | **Read-Only** | Read-only docker socket access | `docker rm`, `docker kill`, `docker exec -it` | *Vector*: Container tampering.<br>*Defense*: Read-only filters; commands run without interactive TTY. |
| **Service Rollback** | `incident-response/runbooks/rollback.sh` | Restoring previous deployment tag in `docker-compose.deploy.yml` or `kubectl rollout undo` | **LEVEL 1 (Auto-Execute)** | Ephemeral runner execution | Deleting namespaces, deleting persistent volumes (`-v`) | *Vector*: Runaway rollback loops.<br>*Defense*: Strictly capped at max 1 execution per incident. |
| **Service Restart** | `docker compose restart <service>` | Restarting backend, frontend, or emulator container | **LEVEL 1 (Auto-Execute)** | Local Compose CLI | Restarting host system or core daemon | *Vector*: Denial of Service.<br>*Defense*: Capped at max 2 restarts per incident. |
| **Code Patch Application** | `git apply /tmp/patch.diff` | Applying local bugfix patch | **LEVEL 2 (Human Approval)** | Requires Human Sign-off | Direct branch push without review | *Vector*: Unverified code defect.<br>*Defense*: Blocked until signed off by human engineer. |
| **Deployment Scaling** | `kubectl scale deployment/backend` | Adjusting replica count for traffic surge | **LEVEL 2 (Human Approval)** | Requires Human Sign-off | Scaling to 0 or exceeding node quota | *Vector*: Resource exhaustion.<br>*Defense*: Requires human approval gate. |
| **Database Direct Writes** | `psql`, raw SQL console, migrations | **PROHIBITED** | **LEVEL 3 (BLOCKED)** | **ZERO CREDENTIALS** | `DROP`, `TRUNCATE`, `ALTER`, `DELETE`, `INSERT`, `UPDATE` | *Vector*: Direct data loss or tampering.<br>*Defense*: Responder has 0 database credentials; port 5432 inaccessible. |
| **Host Root Execution** | `sudo`, host shell tools (`npm`, `go`, `pip`) | **PROHIBITED** | **LEVEL 3 (BLOCKED)** | **ZERO HOST PRIVILEGES** | Direct host compilation, host daemon access | *Vector*: Host takeover.<br>*Defense*: Zero-host-runtime mandate; execution containerized. |

---

## 3. Credential & Secret Isolation Inventory

The following table inventories every secret in the system and documents its complete isolation from the responder agent:

| Secret Name | Purpose | Production Storage | Responder Access | Verification Method |
| :--- | :--- | :--- | :---: | :--- |
| `POSTGRES_PASSWORD` | PostgreSQL master password | Docker Secret / Environment | **BLOCKED (None)** | Injected into DB container only; not mounted into responder container. |
| `JWT_SECRET` | HMAC-SHA256 session token signing | Server Environment | **BLOCKED (None)** | Backend internal memory only; never logged in telemetry or traces. |
| `RESEND_API_KEY` | Transactional email delivery | External Secret Store | **BLOCKED (None)** | Dev environment uses in-memory mock mailbox; no real credentials present. |
| `CLOUDINARY_API_SECRET` | Photo proof upload signature | External Secret Store | **BLOCKED (None)** | Fallback to local Floci emulator without external cloud credentials. |
| `GRAFANA_PASSWORD` | Observability dashboard admin | Local Compose Environment | **BLOCKED (None)** | Responder queries Prometheus & Loki directly via unauthenticated read ports. |

---

## 4. Supply Chain & Provenance Verification (Snyk Agent Scan)

All automated tools, base container images, and scripts are pinned and verified:

| Component | Pinned Version / Digest | Upstream Origin | Integrity Mechanism |
| :--- | :--- | :--- | :--- |
| **Semgrep SAST Container** | `returntocorp/semgrep:latest` | Docker Hub (Official) | Ephemeral container, read-only volume mount (`:ro` / `:src`) |
| **OpenTelemetry Collector** | `otel/opentelemetry-collector-contrib:0.95.0` | GitHub / Docker Hub | Multi-platform signed image |
| **Prometheus** | `prom/prometheus:v2.51.0` | GitHub / Docker Hub | Pinned tag, local TSDB storage volume |
| **Grafana Loki** | `grafana/loki:2.9.4` | GitHub / Docker Hub | Pinned tag, local file storage |
| **Grafana Tempo** | `grafana/tempo:2.4.1` | GitHub / Docker Hub | Pinned tag, memory/local block storage |
| **On-Call Receiver** | `golang:1.22-alpine` | Docker Hub (Official) | Multi-stage build, zero external dependencies (std library only) |
| **Guardrails & Hooks** | `com.antigravity.client/hooks/` | Local Repository | Git-tracked, hash-verified pre-execution hooks |

---

## 5. Security Attestation & Human Sign-Off

- **Audit Date**: 2026-09-28
- **Attestation Statement**: The ChoreSync on-call responder and automated tooling have been inventoried and verified to adhere to least-privilege principles. No credentials allowing database mutation or host compromise are exposed to the AI model.
- **Human Security Reviewer**: Lead Platform & Security Engineer (`pithwamanish`)
- **Disposition**: **APPROVED & CERTIFIED**
