#!/usr/bin/env python3
"""
Generate professional spoken audio narration for the ChoreSync architecture demo
using Google Text-to-Speech (gTTS) and concatenate with exact scene timings.
"""

import os
import subprocess
from gtts import gTTS

AUDIO_DIR = "/app/demo/audio"
os.makedirs(AUDIO_DIR, exist_ok=True)

SCRIPTS = [
    (1, "Welcome to ChoreSync. In this walkthrough, we explore the behind the scenes architecture: our Full-Stack application, Standalone LGTM Observability Stack, Actionable Alerts, and Two-Stage Container CI/CD pipeline."),
    (2, "Here is the ChoreSync frontend. In addition to standard chore management, notice the Kitchen Tablet toggle designed for wall mounted shared tablets, with 4-digit Admin PIN unlock, and our in-memory Dev Mailbox capturing transactional emails like magic links and chore reminders without external SMTP spam."),
    (3, "Next, we examine the telemetry flow. ChoreSync implements end-to-end W3C distributed tracing. The React Web SDK injects traceparent headers, Caddy reverse proxy forwards them, Go Chi middleware creates server spans, and the PostgreSQL driver auto-instruments database queries."),
    (4, "Here is the live Grafana Observability Dashboard. Notice that all scrape targets are UP, the OpenTelemetry Collector span ingestion rate is actively tracked, and container logs are streaming into the unified dashboard."),
    (5, "Now let us inspect distributed tracing in Grafana Tempo. Each trace captures the full HTTP request lifecycle, showing exact database queries and statement durations executed by our PostgreSQL pgx tracer."),
    (6, "In Grafana Loki, structured container logs from the Go API, OpenTelemetry Collector, and PostgreSQL are aggregated. LogQL queries automatically correlate log lines to distributed trace IDs in Tempo."),
    (7, "Moving to Prometheus, we view the three symptom-based alert rules: High HTTP Error Rate, High Request Latency, and Database Query Errors, each equipped with non-arbitrary thresholds and runbook links."),
    (8, "Here we observe a live synthetic failure scenario. A burst of 500 errors pushes the error rate above the 5 percent threshold, causing Prometheus to transition the alert state from pending to firing."),
    (9, "In Alertmanager, active firing alerts are grouped by service and environment, and dispatched directly to the on-call webhook receiver."),
    (10, "Here is the On-Call Webhook Receiver dashboard. Incoming alerts are validated and normalized according to our strict schema, ready for autonomous on-call agent remediation."),
    (11, "Now let us review our Two-Stage Container Delivery and Promotion pipeline. Images are built and stamped with immutable timestamp tags in Stage 1, served via pre-built compose in Stage 2, and promoted to production using SHA256 digest immutability."),
    (12, "All twelve quality gates, unit tests, PostgreSQL store tests, eleven Playwright end-to-end journeys, and autonomous incident verification suites are passing 100 percent.")
]

def generate():
    print("[AUDIO] Generating speech narration clips via gTTS...")
    for idx, text in SCRIPTS:
        out_path = os.path.join(AUDIO_DIR, f"scene_{idx}.mp3")
        print(f"  Generating Scene {idx}...")
        tts = gTTS(text=text, lang="en", tld="com", slow=False)
        tts.save(out_path)
    print("[AUDIO] All 12 audio narration clips generated successfully!")

if __name__ == "__main__":
    generate()
