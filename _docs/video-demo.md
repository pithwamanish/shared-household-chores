# ChoreSync Behind-the-Scenes Architecture & Observability Demo Video

This document describes the automated video demonstration covering the non-obvious, behind-the-scenes architectural tiers of ChoreSync: Full-Stack Observability (LGTM), W3C Distributed Tracing, Actionable Symptom Alerts (Gate 11), Autonomous On-Call Remediation (Gate 12), and Two-Stage Container Delivery (Gate 8).

---

## 1. How to Watch the Video

### Option A: Interactive Web Player (Recommended)
Open your browser to:
* **Local**: [http://localhost:3000/demo.html](http://localhost:3000/demo.html)
* **LAN**: `http://<HOST_IP>:3000/demo.html` (e.g. `http://192.168.0.105:3000/demo.html`)

The interactive player features direct timestamp jumping, video scrubbing, spoken audio narration, and quick links to live observability services.

### Option B: Local Media Players
Both MP4 and WebM formats with synchronized voice narration are stored in the repository:
* **MP4 (Universal - H.264 + AAC)**: [`_docs/videos/choresync_architecture_observability_demo.mp4`](videos/choresync_architecture_observability_demo.mp4) (6.1 MB)
* **WebM (Modern Web - VP8 + Opus)**: [`_docs/videos/choresync_architecture_observability_demo.webm`](videos/choresync_architecture_observability_demo.webm) (13.1 MB)
* Also served statically from [`frontend/public/`](../frontend/public/)

Compatible with VLC Media Player, QuickTime, Windows Media Player, Google Chrome, Mozilla Firefox, Microsoft Edge, and Apple Safari.

---

## 2. Re-Recording the Demo

To trigger a fresh recording from scratch:

```bash
make demo-record
```

This pipeline automatically:
1. Pre-warms live data across all observability tools (Loki logs, Tempo traces, and firing Prometheus alert)
2. Records the full-featured browser walkthrough in an isolated Playwright container with exact narration timings
3. Multiplexes the video with spoken voice narration audio into both WebM and MP4 formats.

---

## 3. Video Chapters & Architectural Topics (Spoken Narration Included)

| Chapter | Time | Topic Demonstrated | Architectural Context & Behind-the-Scenes Insight |
| :---: | :---: | :--- | :--- |
| **01** | `0:00` | **Architecture Overview** | Core Go 1.22 Chi API, PostgreSQL 16 store (`sqlc` + `pgx/v5`), React 18 frontend, and standalone LGTM observability stack. |
| **02** | `0:18` | **Frontend Kiosk & Dev Mailbox** | Demonstrates the Shared Kitchen Tablet touch bar, Admin PIN verification, and the in-memory transactional Dev Mailbox modal capturing magic links and chore reminders without external SMTP spam. |
| **03** | `0:42` | **W3C Distributed Tracing Flow** | Context propagation journey: React Web SDK ➔ Caddy reverse proxy ➔ Go Chi API ➔ PostgreSQL database (`pgx.QueryTracer`). |
| **04** | `1:06` | **Grafana Observability Dashboard** | Live Golden Signals dashboard showing Prometheus exporter scrape health (all UP), OTel Collector span throughput, and live Loki container logs. |
| **05** | `1:22` | **Distributed Tracing in Tempo** | TraceQL query engine, trace waterfall visualization with database query statement latencies and nested spans. |
| **06** | `1:40` | **Log Aggregation in Loki** | Structured container log exploration with LogQL and automatic TraceID correlation linking logs to Tempo traces. |
| **07** | `1:58` | **Actionable Alerts in Prometheus** | Symptom-based alert rules: `HighHttpErrorRate` (> 5%), `HighRequestLatency` (> 1s), and `DatabaseQueryErrors` (> 0.05/s) with dashboard and runbook URLs. |
| **08** | `2:14` | **Live Incident Simulation** | Active `HighHttpErrorRate` alert in red FIRING state with observed error rate (66.7%), threshold (5%), and annotations. |
| **09** | `2:29` | **Alertmanager Notification Routing** | Grouped webhook dispatch to `http://on-call-receiver:5050/webhook`. |
| **10** | `2:39` | **On-Call Receiver Normalization** | Live incident dashboard (:5050) displaying the normalized payload against `payload-schema.json` with firing badge. |
| **11** | `2:52` | **Two-Stage Container Delivery** | "Build Once, Promote Everywhere": immutable timestamp tags (`YYYYMMDD-HHMMSS-shortsha`), zero-compiler deploy compose, and SHA256 digest promotion. |
| **12** | `3:13` | **Final Quality Sign-Off** | 100% green verification matrix across Go unit tests, PostgreSQL store tests, 11 Playwright E2E journeys, and incident remediation suites. |

---

## 4. Key Invariants Visualized

1. **The Telemetry Golden Triangle**: Every trace span, metric point, and log contains `service.name`, `deployment.environment`, and `service.version`.
2. **The Reproduction Invariant**: Autonomous on-call agents are prohibited from modifying application code without first establishing a reproducible failing test case.
3. **Decoupled Delivery**: Image packaging is decoupled from runtime deployment, eliminating compiler tools from production containers.
