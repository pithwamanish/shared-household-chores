#!/usr/bin/env bash
set -euo pipefail

# Post-Tool Execution Audit Hook
# Records execution timestamp, command, and status to local audit log.

LOG_FILE="$(dirname "$0")/audit.log"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
CMD_NAME="${1:-unknown}"
EXIT_STATUS="${2:-0}"

echo "[${TIMESTAMP}] tool=${CMD_NAME} exit=${EXIT_STATUS}" >> "${LOG_FILE}"
exit 0
