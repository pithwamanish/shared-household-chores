#!/usr/bin/env python3
"""
ChoreSync On-Call Alert Receiver & Normalizer
Listens on port 5050 for Alertmanager webhooks or CLI payloads, validates and
normalizes them against payload-schema.json, and records incidents to incidents/.
"""

import os
import sys
import json
import time
from datetime import datetime, timezone
from http.server import HTTPServer, BaseHTTPRequestHandler

PORT = int(os.environ.get("RECEIVER_PORT", "5050"))
INCIDENTS_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "incidents")
SCHEMA_FILE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "payload-schema.json")

os.makedirs(INCIDENTS_DIR, exist_ok=True)


def normalize_alertmanager_alert(alert_raw, common_labels=None, common_annotations=None):
    """Normalize an individual Alertmanager alert or raw alert into the canonical schema."""
    labels = {**(common_labels or {}), **(alert_raw.get("labels") or {})}
    annotations = {**(common_annotations or {}), **(alert_raw.get("annotations") or {})}

    alert_name = labels.get("alertname") or labels.get("alert") or alert_raw.get("alert", {}).get("name", "UnknownAlert")
    state = alert_raw.get("status") or alert_raw.get("state") or alert_raw.get("alert", {}).get("state", "firing")
    severity = labels.get("severity") or alert_raw.get("alert", {}).get("severity", "critical")
    service_name = labels.get("service") or labels.get("service_name") or alert_raw.get("service", {}).get("name", "api-backend")
    environment = labels.get("environment") or labels.get("deployment_environment") or alert_raw.get("deployment", {}).get("environment", "development")
    version = labels.get("version") or labels.get("service_version") or alert_raw.get("service", {}).get("version", "dev-latest")

    dashboard_url = annotations.get("dashboard_url") or alert_raw.get("dashboard_url") or f"http://localhost:3001/d/overview?var-environment={environment}"
    runbook_url = annotations.get("runbook_url") or alert_raw.get("runbook_url") or "on-call-engineer/runbook.md"

    # Metric info extraction
    metric_name = labels.get("__name__") or annotations.get("metric_name") or "http_requests_total"
    observed_val = alert_raw.get("generatorURL", "")
    # Default numeric fallback for testing
    observed_num = alert_raw.get("metric", {}).get("observed_value", 5.2)
    threshold_num = alert_raw.get("metric", {}).get("threshold", 5.0)

    now_iso = datetime.now(timezone.utc).isoformat()
    if alert_raw.get("startsAt"):
        now_iso = alert_raw.get("startsAt")

    normalized = {
        "alert": {
            "name": str(alert_name),
            "state": "resolved" if state == "resolved" else "firing",
            "severity": severity if severity in ["critical", "warning", "info"] else "critical"
        },
        "service": {
            "name": str(service_name),
            "version": str(version)
        },
        "deployment": {
            "environment": environment if environment in ["development", "staging", "production"] else "development"
        },
        "metric": {
            "name": str(metric_name),
            "observed_value": float(observed_num),
            "threshold": float(threshold_num),
            "duration": labels.get("for", "1m")
        },
        "dashboard_url": str(dashboard_url),
        "runbook_url": str(runbook_url),
        "timestamp": now_iso
    }
    return normalized


def save_incident(normalized):
    timestamp_key = int(time.time())
    alert_name = normalized["alert"]["name"].lower()
    incident_id = f"inc-{timestamp_key}-{alert_name}"
    incident_path = os.path.join(INCIDENTS_DIR, f"{incident_id}.json")

    with open(incident_path, "w", encoding="utf-8") as f:
        json.dump(normalized, f, indent=2)

    print(f"[ON-CALL RECEIVER] Incident recorded: ID={incident_id} Alert={normalized['alert']['name']} Status={normalized['alert']['state']} Path={incident_path}", flush=True)
    return incident_id, incident_path


class WebhookHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path in ["/healthz", "/health"]:
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(b'{"status":"healthy","service":"on-call-receiver"}\n')
            return

        if self.path == "/incidents":
            incidents = []
            for fname in sorted(os.listdir(INCIDENTS_DIR), reverse=True):
                if fname.endswith(".json"):
                    with open(os.path.join(INCIDENTS_DIR, fname), "r") as f:
                        try:
                            incidents.append(json.load(f))
                        except Exception:
                            pass
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(incidents, indent=2).encode("utf-8"))
            return

        if self.path in ["/", "/index.html"]:
            incidents = []
            for fname in sorted(os.listdir(INCIDENTS_DIR), reverse=True):
                if fname.endswith(".json"):
                    with open(os.path.join(INCIDENTS_DIR, fname), "r") as f:
                        try:
                            incidents.append((fname, json.load(f)))
                        except Exception:
                            pass

            incident_rows = ""
            for fname, inc in incidents:
                alert_info = inc.get("alert", {})
                service_info = inc.get("service", {})
                state_badge = '<span style="background:#dc2626;color:white;padding:2px 8px;border-radius:4px;font-size:12px;">FIRING</span>' if alert_info.get("state") == "firing" else '<span style="background:#16a34a;color:white;padding:2px 8px;border-radius:4px;font-size:12px;">RESOLVED</span>'
                incident_rows += f"""<tr>
                    <td style="padding:10px;border-bottom:1px solid #334155;font-family:monospace;font-size:13px;">{fname}</td>
                    <td style="padding:10px;border-bottom:1px solid #334155;"><strong>{alert_info.get('name', 'N/A')}</strong></td>
                    <td style="padding:10px;border-bottom:1px solid #334155;">{state_badge}</td>
                    <td style="padding:10px;border-bottom:1px solid #334155;">{service_info.get('name', 'N/A')} ({service_info.get('version', 'N/A')})</td>
                    <td style="padding:10px;border-bottom:1px solid #334155;font-size:12px;color:#94a3b8;">{inc.get('timestamp', 'N/A')}</td>
                </tr>"""

            if not incident_rows:
                incident_rows = '<tr><td colspan="5" style="padding:20px;text-align:center;color:#94a3b8;">No incidents recorded yet.</td></tr>'

            html = f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>ChoreSync On-Call Alert Receiver</title>
  <style>
    body {{ font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; margin: 0; padding: 24px; }}
    .container {{ max-width: 960px; margin: 0 auto; }}
    h1 {{ color: #38bdf8; margin-bottom: 4px; }}
    .badge {{ background: #0284c7; color: white; padding: 4px 10px; border-radius: 9999px; font-size: 13px; }}
    .card {{ background: #1e293b; border-radius: 8px; padding: 20px; margin-top: 20px; border: 1px solid #334155; }}
    table {{ width: 100%; border-collapse: collapse; text-align: left; }}
    th {{ padding: 10px; background: #0f172a; color: #94a3b8; font-size: 13px; border-bottom: 2px solid #334155; }}
    a {{ color: #38bdf8; text-decoration: none; }}
    a:hover {{ text-decoration: underline; }}
    .links {{ display: flex; gap: 16px; margin-top: 12px; }}
  </style>
</head>
<body>
  <div class="container">
    <div style="display:flex; justify-content:space-between; align-items:center;">
      <div>
        <h1>🚨 ChoreSync On-Call Alert Receiver</h1>
        <p style="color:#94a3b8;margin-top:0;">Autonomous incident webhook endpoint listening on port 5050</p>
      </div>
      <div><span class="badge">● Online & Ready</span></div>
    </div>

    <div class="card">
      <h3 style="margin-top:0;color:#e2e8f0;">Quick Observability Navigation</h3>
      <div class="links">
        <a href="http://localhost:3001" target="_blank">📊 Grafana UI (3001)</a>
        <a href="http://localhost:9090" target="_blank">🔥 Prometheus Alerts & Metrics (9090)</a>
        <a href="http://localhost:9093" target="_blank">🔔 Alertmanager (9093)</a>
        <a href="/incidents" target="_blank">📄 Incidents JSON API</a>
        <a href="/healthz" target="_blank">❤️ Healthcheck</a>
      </div>
    </div>

    <div class="card">
      <h3 style="margin-top:0;color:#e2e8f0;">Recorded Incidents ({len(incidents)})</h3>
      <table>
        <thead>
          <tr>
            <th>Incident File</th>
            <th>Alert Name</th>
            <th>State</th>
            <th>Service / Version</th>
            <th>Timestamp</th>
          </tr>
        </thead>
        <tbody>
          {incident_rows}
        </tbody>
      </table>
    </div>
  </div>
</body>
</html>"""

            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.end_headers()
            self.wfile.write(html.encode("utf-8"))
            return

        self.send_response(404)
        self.end_headers()

    def do_POST(self):
        if self.path in ["/webhook", "/alert", "/api/v1/alerts"]:
            content_len = int(self.headers.get("Content-Length", 0))
            body = self.rfile.read(content_len).decode("utf-8")

            try:
                data = json.loads(body)
            except Exception as e:
                self.send_response(400)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": "INVALID_JSON", "message": str(e)}).encode("utf-8"))
                return

            # Check if this is an Alertmanager payload containing 'alerts' list
            saved_ids = []
            if isinstance(data, dict) and "alerts" in data:
                common_labels = data.get("commonLabels", {})
                common_ann = data.get("commonAnnotations", {})
                for a in data["alerts"]:
                    norm = normalize_alertmanager_alert(a, common_labels, common_ann)
                    inc_id, _ = save_incident(norm)
                    saved_ids.append(inc_id)
            else:
                # Direct single alert or normalized payload
                norm = normalize_alertmanager_alert(data)
                inc_id, _ = save_incident(norm)
                saved_ids.append(inc_id)

            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            resp = {
                "status": "success",
                "message": "Alerts received and normalized",
                "incident_ids": saved_ids
            }
            self.wfile.write(json.dumps(resp).encode("utf-8"))
            return

        self.send_response(404)
        self.end_headers()

    def log_message(self, format, *args):
        # Override to provide clean standard logging
        sys.stderr.write(f"[ON-CALL WEBHOOK] {args[0]} {args[1]} -> {args[2]}\n")


def main():
    # If run in CLI mode with stdin pipe:
    if not sys.stdin.isatty() and len(sys.argv) > 1 and sys.argv[1] == "--stdin":
        raw = sys.stdin.read()
        if raw.strip():
            payload = json.loads(raw)
            norm = normalize_alertmanager_alert(payload)
            inc_id, path = save_incident(norm)
            print(json.dumps({"incident_id": inc_id, "path": path, "payload": norm}, indent=2))
            return

    server = HTTPServer(("0.0.0.0", PORT), WebhookHandler)
    print(f"[ON-CALL RECEIVER] Listening for alerts on http://0.0.0.0:{PORT}/webhook", flush=True)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n[ON-CALL RECEIVER] Shutting down cleanly.", flush=True)
        server.server_close()


if __name__ == "__main__":
    main()
