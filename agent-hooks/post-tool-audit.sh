#!/usr/bin/env bash
set -euo pipefail

# Post-Tool Execution Audit & Guardrail Hook
# Records execution timestamp, command, and status to local audit log.
# Flags and audits if any prohibited or blocked operations escaped pre-tool guardrails.

LOG_FILE="$(dirname "$0")/audit.log"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
CMD_NAME="${1:-unknown}"
EXIT_STATUS="${2:-0}"

# Audit and flag any forbidden or prohibited operations
if echo "${CMD_NAME}" | grep -qiE "rm -rf|drop table|truncate|--privileged"; then
  echo "[${TIMESTAMP}] [GUARDRAIL-ALERT] Prohibited pattern observed: ${CMD_NAME} exit=${EXIT_STATUS}" >> "${LOG_FILE}"
else
  echo "[${TIMESTAMP}] tool=${CMD_NAME} exit=${EXIT_STATUS}" >> "${LOG_FILE}"
fi
exit 0
