# ChoreSync Behind-the-Scenes Architecture & Observability Demo Video

This document describes the automated video demonstration covering the non-obvious, behind-the-scenes architectural tiers of ChoreSync: Full-Stack Observability (LGTM), W3C Distributed Tracing, Actionable Symptom Alerts (Gate 11), Autonomous On-Call Remediation (Gate 12), and Two-Stage Container Delivery (Gate 8).

---

## 1. How to Watch the Video

### Option A: Interactive Web Player (Recommended)
Open your browser to:
* **Local**: [http://localhost:3000/demo.html](http://localhost:3000/demo.html)
* **LAN**: `http://<HOST_IP>:3000/demo.html` (e.g. `http://192.168.0.105:3000/demo.html`)

The interactive player features direct timestamp jumping, video scrubbing, and quick links to live observability services.

### Option B: Local Media Players
The raw high-definition recording is stored in the repository at:
* [`_docs/videos/choresync_architecture_observability_demo.webm`](videos/choresync_architecture_observability_demo.webm)
* [`frontend/public/choresync_demo.webm`](../frontend/public/choresync_demo.webm)
* [`e2e/recordings/choresync_architecture_observability_demo.webm`](../e2e/recordings/choresync_architecture_observability_demo.webm)

Compatible with VLC Media Player, Google Chrome, Mozilla Firefox, Microsoft Edge, and Apple Safari.

---

## 2. Re-Recording the Demo

To trigger a fresh recording from scratch using the containerized Playwright engine:

```bash
make demo-record
```

This launches a headless Chromium instance in 720p HD, navigates through all presentation slides and live UI pages, injects the on-screen HUD, simulates live errors, and writes the output to `_docs/videos/`.

---

## 3. Video Chapters & Architectural Topics

| Chapter | Time | Topic Demonstrated | Architectural Context & Behind-the-Scenes Insight |
| :---: | :---: | :--- | :--- |
| **01** | `0:00` | **Architecture Overview** | Core Go 1.22 Chi API, PostgreSQL 16 store (`sqlc` + `pgx/v5`), React 18 frontend, and standalone LGTM observability stack. |
| **02** | `0:07` | **Frontend Kiosk & Dev Mailbox** | Demonstrates the Shared Kitchen Tablet touch bar, Admin PIN verification, and the in-memory transactional Dev Mailbox modal capturing magic links and chore reminders without external SMTP spam. |
| **03** | `0:18` | **W3C Distributed Tracing Flow** | Context propagation journey: React Web SDK ➔ Caddy reverse proxy ➔ Go Chi API ➔ PostgreSQL database (`pgx.QueryTracer`). |
| **04** | `0:25` | **Grafana Observability Dashboard** | Live Golden Signals dashboard showing Prometheus exporter scrape health (all UP), OTel Collector span throughput, and Loki log streams. |
| **05** | `0:33` | **Distributed Tracing in Tempo** | TraceQL query engine, trace waterfall visualization, and exact parameterized SQL query statement latencies. |
| **06** | `0:40` | **Log Aggregation in Loki** | Structured container log exploration with LogQL and automatic TraceID correlation linking logs to Tempo traces. |
| **07** | `0:47` | **Actionable Alerts in Prometheus** | Symptom-based alert rules: `HighHttpErrorRate` (> 5%), `HighRequestLatency` (> 1s), and `DatabaseQueryErrors` (> 0.05/s) with dashboard and runbook URLs. |
| **08** | `0:53` | **Live Incident Simulation** | Synthetic error burst hitting `/api/v1/dev/simulate-error`, pushing error rate to 66.7%, and watching Prometheus transition from PENDING to FIRING. |
| **09** | `1:02` | **Alertmanager Notification Routing** | Grouped webhook dispatch to `http://on-call-receiver:5050/webhook`. |
| **10** | `1:07` | **On-Call Receiver Normalization** | Standardized payload normalization against `payload-schema.json` and autonomous agent incident preparation. |
| **11** | `1:14` | **Two-Stage Container Delivery** | "Build Once, Promote Everywhere": immutable timestamp tags (`YYYYMMDD-HHMMSS-shortsha`), zero-compiler deploy compose, and SHA256 digest promotion. |
| **12** | `1:22` | **Final Quality Sign-Off** | 100% green verification matrix across Go unit tests, PostgreSQL store tests, 11 Playwright E2E journeys, and incident remediation suites. |

---

## 4. Key Invariants Visualized

1. **The Telemetry Golden Triangle**: Every trace span, metric point, and log contains `service.name`, `deployment.environment`, and `service.version`.
2. **The Reproduction Invariant**: Autonomous on-call agents are prohibited from modifying application code without first establishing a reproducible failing test case.
3. **Decoupled Delivery**: Image packaging is decoupled from runtime deployment, eliminating compiler tools from production containers.
