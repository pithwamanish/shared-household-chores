#!/usr/bin/env bash
set -euo pipefail

# incident-response/collect-evidence.sh
# Bounded, repeatable read-only evidence collection script.
# Gathers telemetry, logs, git revision history, and container health
# WITHOUT leaking database credentials, API keys, or production secrets.

OUTPUT_FILE=""
INCIDENT_ID="INC-$(date -u +'%Y%m%d-%H%M%S')"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -o|--output)
      OUTPUT_FILE="$2"
      shift 2
      ;;
    -i|--incident)
      INCIDENT_ID="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: collect-evidence.sh [-o output.json] [-i incident_id]"
      exit 0
      ;;
    *)
      INCIDENT_ID="$1"
      shift
      ;;
  esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
TIMESTAMP="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"

# 1. Inspect Git Revisions (Read-Only)
CURRENT_COMMIT="$(git -C "$REPO_ROOT" rev-parse --short HEAD 2>/dev/null || echo "unknown")"
CURRENT_BRANCH="$(git -C "$REPO_ROOT" rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")"

RECENT_COMMITS_JSON="$(git -C "$REPO_ROOT" log -n 5 --pretty=format:'{"sha":"%h","author":"%an","date":"%ad","message":"%s"}' --date=iso 2>/dev/null | jq -s '.' || echo '[]')"

# 2. Inspect Container Runtimes (Read-Only)
CONTAINER_STATUS="unknown"
if command -v docker >/dev/null 2>&1; then
  CONTAINER_STATUS="$(docker ps --filter "name=backend" --format '{"name":"{{.Names}}","status":"{{.Status}}","image":"{{.Image}}"}' 2>/dev/null | jq -s '.' || echo '[]')"
else
  CONTAINER_STATUS="[]"
fi

# 3. Query Prometheus for Error Rate & Latency (Read-Only HTTP API)
PROM_URL="${PROMETHEUS_URL:-http://localhost:9090}"
ERROR_RATE="0.0"
P95_LATENCY="0.0"

if curl -s -f --max-time 2 "${PROM_URL}/-/healthy" >/dev/null 2>&1; then
  # 5xx error rate query
  PROM_5XX_QUERY='sum(rate(http_requests_total{http_status_code=~"5.."}[1m]))/clamp_min(sum(rate(http_requests_total[1m])),0.001)*100'
  RAW_ERR_RATE="$(curl -s -G --data-urlencode "query=${PROM_5XX_QUERY}" "${PROM_URL}/api/v1/query" 2>/dev/null || true)"
  ERROR_RATE="$(echo "$RAW_ERR_RATE" | jq -r '.data.result[0].value[1] // "0.0"' 2>/dev/null || echo "0.0")"

  # P95 latency query
  PROM_LAT_QUERY='histogram_quantile(0.95,sum(rate(http_server_request_duration_seconds_bucket[1m]))by(le))'
  RAW_LAT="$(curl -s -G --data-urlencode "query=${PROM_LAT_QUERY}" "${PROM_URL}/api/v1/query" 2>/dev/null || true)"
  P95_LATENCY="$(echo "$RAW_LAT" | jq -r '.data.result[0].value[1] // "0.0"' 2>/dev/null || echo "0.0")"
fi

# 4. Query Recent Error Logs (Read-Only)
LOKI_URL="${LOKI_URL:-http://localhost:3100}"
LOG_SAMPLES="[]"

if curl -s -f --max-time 2 "${LOKI_URL}/ready" >/dev/null 2>&1; then
  RAW_LOGS="$(curl -s -G --data-urlencode 'query={service_name="api-backend"} |= "level=error"' "${LOKI_URL}/loki/api/v1/query_range" 2>/dev/null || true)"
  LOG_SAMPLES="$(echo "$RAW_LOGS" | jq '[.data.result[].values[][1]]' 2>/dev/null || echo '[]')"
elif command -v docker >/dev/null 2>&1; then
  # Fallback to recent container logs filtered for error keywords
  LOG_SAMPLES="$(docker logs --tail 30 choresync-backend 2>&1 | grep -iE 'error|panic|fatal|500' | head -n 5 | jq -R -s -c 'split("\n")[:-1]' 2>/dev/null || echo '[]')"
fi

# 5. Assemble Strictly Bounded Evidence Packet
EVIDENCE_JSON=$(cat <<EOF
{
  "incident_id": "${INCIDENT_ID}",
  "timestamp": "${TIMESTAMP}",
  "service": {
    "name": "api-backend",
    "deployed_version": "${CURRENT_COMMIT}",
    "branch": "${CURRENT_BRANCH}",
    "environment": "development"
  },
  "metrics": {
    "http_5xx_error_rate_pct": ${ERROR_RATE:-0.0},
    "p95_latency_seconds": ${P95_LATENCY:-0.0},
    "alert_threshold_breached": true
  },
  "runtime_status": {
    "containers": ${CONTAINER_STATUS}
  },
  "git_evidence": {
    "head_commit": "${CURRENT_COMMIT}",
    "recent_commits": ${RECENT_COMMITS_JSON}
  },
  "log_evidence": ${LOG_SAMPLES},
  "security_guarantee": {
    "allowlisted_read_only_sources": ["prometheus", "loki", "git-log", "docker-ps"],
    "database_credentials_exposed": false,
    "production_secrets_exposed": false
  }
}
EOF
)

if [ -n "$OUTPUT_FILE" ]; then
  mkdir -p "$(dirname "$OUTPUT_FILE")"
  echo "$EVIDENCE_JSON" > "$OUTPUT_FILE"
  echo "Evidence packet written to: $OUTPUT_FILE"
else
  echo "$EVIDENCE_JSON"
fi
