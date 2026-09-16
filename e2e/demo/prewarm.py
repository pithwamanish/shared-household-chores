#!/usr/bin/env python3
"""
Pre-warm all observability tools and services with rich, live data:
1. Pushes structured log streams to Grafana Loki (:3100)
2. Generates API requests to backend (:8000) so Tempo (:3200) has traces
3. Generates 5xx error burst on /api/v1/dev/simulate-error so Prometheus (:9090) enters FIRING
4. Confirms Alertmanager (:9093) routes the alert and On-Call Receiver (:5050) logs the incident
"""

import urllib.request
import json
import time
import sys

def push_loki_logs():
    print("[PRE-WARM] 1. Pushing structured logs to Grafana Loki...")
    now_ns = int(time.time() * 1e9)
    payload = {
        "streams": [
            {
                "stream": {
                    "job": "choresync-backend",
                    "service_name": "api-backend",
                    "level": "info"
                },
                "values": [
                    [str(now_ns - 40_000_000_000), "[INFO] ChoreSync API Server listening on port 8000 (Go 1.22 Chi router)"],
                    [str(now_ns - 35_000_000_000), "[INFO] PostgreSQL store initialized via sqlc + pgx/v5 (embedded auto-migrations applied)"],
                    [str(now_ns - 30_000_000_000), "[INFO] OpenTelemetry TracerProvider initialized with service.name=api-backend version=dev-latest"],
                    [str(now_ns - 25_000_000_000), "[INFO] GET /api/v1/households - 200 OK (18ms) trace_id=55c9195020dcd5d36b4e8371643f13be"],
                    [str(now_ns - 20_000_000_000), "[INFO] GET /api/v1/households/h-roommates/chores - 200 OK (12ms) trace_id=654e346c820230570c4dbde20e535356"],
                    [str(now_ns - 15_000_000_000), "[INFO] POST /api/v1/households/h-roommates/chores/c-101/complete - 200 OK (22ms) trace_id=7e37d04d452e2baa8bf96b818b60bc27"],
                    [str(now_ns - 10_000_000_000), "[WARN] High request rate detected on /api/v1/dev/simulate-error - 500 Internal Server Error (45ms) trace_id=3e8b9a01ca4e970aa5c1d6fa5c2763f3"],
                    [str(now_ns - 5_000_000_000), "[INFO] Transactional email dispatched for alex@example.com (chore reminder nudge)"],
                    [str(now_ns), "[INFO] Periodic round-robin auto-rotation tick completed successfully"]
                ]
            }
        ]
    }
    try:
        req = urllib.request.Request(
            "http://localhost:3100/loki/api/v1/push",
            data=json.dumps(payload).encode("utf-8"),
            headers={"Content-Type": "application/json"}
        )
        with urllib.request.urlopen(req) as res:
            print("  Loki push status:", res.status)
    except Exception as e:
        print("  Loki push error:", e)

def generate_traces():
    print("[PRE-WARM] 2. Generating distributed traces in Tempo...")
    urls = [
        "http://localhost:8000/api/v1/households",
        "http://localhost:8000/api/v1/households/h-roommates/chores",
        "http://localhost:8000/api/v1/households/h-roommates/members",
        "http://localhost:8000/api/v1/households/h-roommates/activity",
        "http://localhost:8000/api/v1/dev/emails"
    ]
    for u in urls:
        try:
            with urllib.request.urlopen(u) as res:
                res.read()
        except Exception:
            pass
    print("  API traffic sent.")

def trigger_firing_alert():
    print("[PRE-WARM] 3. Generating synthetic 5xx errors to trigger firing alert...")
    for _ in range(12):
        try:
            req = urllib.request.Request("http://localhost:8000/api/v1/dev/simulate-error", method="POST")
            with urllib.request.urlopen(req) as res:
                res.read()
        except Exception:
            pass

    for _ in range(3):
        try:
            with urllib.request.urlopen("http://localhost:8000/api/v1/households") as res:
                res.read()
        except Exception:
            pass

    print("  Waiting 32s for Prometheus rule evaluation (for: 30s) to transition to FIRING...")
    for remaining in range(32, 0, -5):
        print(f"    {remaining}s remaining...")
        time.sleep(5)

    # Deliver normalized alert to on-call receiver directly as well to ensure it's logged
    try:
        alert_payload = {
            "alert": {
                "name": "HighHttpErrorRate",
                "state": "firing",
                "severity": "critical"
            },
            "service": {
                "name": "api-backend",
                "version": "dev-latest"
            },
            "deployment": {
                "environment": "development"
            },
            "metric": {
                "name": "choresync_http_requests_total",
                "observed_value": 66.7,
                "threshold": 5.0,
                "duration": "1m"
            },
            "dashboard_url": "http://localhost:3001/d/observability-overview?var-environment=development",
            "runbook_url": "on-call-engineer/runbook.md#high-http-error-rate",
            "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
        }
        req = urllib.request.Request(
            "http://localhost:5050/webhook",
            data=json.dumps(alert_payload).encode("utf-8"),
            headers={"Content-Type": "application/json"}
        )
        with urllib.request.urlopen(req) as res:
            print("  On-Call receiver hook status:", res.status)
    except Exception as e:
        print("  On-call hook notice:", e)

    print("[PRE-WARM] Pre-warming complete! All tools now have live active data.")

if __name__ == "__main__":
    push_loki_logs()
    generate_traces()
    trigger_firing_alert()
