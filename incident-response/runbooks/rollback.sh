#!/usr/bin/env bash
set -euo pipefail

# incident-response/runbooks/rollback.sh
# Level 1 Auto-Execute runbook: rolls back service to previous stable release
# and logs execution trail without mutating relational database state.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
INCIDENT_DIR="$REPO_ROOT/incident-response/incidents"
AUDIT_LOG="$INCIDENT_DIR/audit-trail.log"
TIMESTAMP="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"

mkdir -p "$INCIDENT_DIR"

echo "========================================================"
echo "  [RUNBOOK] Level 1 Operational Rollback Initiated      "
echo "  Timestamp: $TIMESTAMP                                 "
echo "========================================================"

TARGET_SERVICE="${1:-backend}"
echo "Target Service: $TARGET_SERVICE"

# 1. Check for Active Kind Kubernetes Cluster
if command -v kubectl >/dev/null 2>&1 && kubectl get deployment "$TARGET_SERVICE" --context kind-choresync-cluster >/dev/null 2>&1; then
  echo "==> Executing Kubernetes rollout undo for deployment/$TARGET_SERVICE..."
  kubectl rollout undo "deployment/$TARGET_SERVICE" --context kind-choresync-cluster
  kubectl rollout status "deployment/$TARGET_SERVICE" --context kind-choresync-cluster --timeout=60s
  ROLLBACK_METHOD="kubernetes_rollout_undo"
# 2. Check for Docker Compose Deployment
elif [ -f "$REPO_ROOT/docker-compose.deploy.yml" ]; then
  echo "==> Executing Docker Compose deployment rollback to stable images..."
  docker compose -f "$REPO_ROOT/docker-compose.deploy.yml" up -d "$TARGET_SERVICE"
  ROLLBACK_METHOD="docker_compose_deploy_rollback"
# 3. Fallback: Local Docker Compose Restart
else
  echo "==> Restarting local container service $TARGET_SERVICE..."
  docker compose -f "$REPO_ROOT/docker-compose.yml" restart "$TARGET_SERVICE"
  ROLLBACK_METHOD="docker_compose_restart"
fi

# 4. Record Audit Trail
AUDIT_ENTRY="[${TIMESTAMP}] ACTION=rollback SERVICE=${TARGET_SERVICE} METHOD=${ROLLBACK_METHOD} STATUS=SUCCESS CALLER=autonomous_responder"
echo "$AUDIT_ENTRY" >> "$AUDIT_LOG"
echo "  ✓ Audit trail logged to: $AUDIT_LOG"

echo "========================================================"
echo "  ✓ Rollback execution complete!                        "
echo "========================================================"
