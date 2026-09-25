#!/usr/bin/env bash
set -euo pipefail

# Pre-Tool Execution Guardrail Hook
# Intercepts agent actions to prevent irreversible actions or data loss.

CMD_INPUT="${*:-}"
if [ -z "${CMD_INPUT}" ] && [ ! -t 0 ]; then
  CMD_INPUT="$(cat)"
fi

# Disallowed destructive patterns
FORBIDDEN_PATTERNS=(
  "rm -rf /"
  "rm -rf ~"
  "DROP DATABASE"
  "drop database"
  "TRUNCATE TABLE"
  "truncate table"
  "DELETE FROM .* WHERE 1=1"
  "git push --force origin main"
  "git push -f origin main"
  "git push --force origin master"
)

for pattern in "${FORBIDDEN_PATTERNS[@]}"; do
  if echo "${CMD_INPUT}" | grep -qiE "${pattern}"; then
    echo "==================================================================" >&2
    echo "GUARDRAIL BLOCKED: Destructive command intercepted by agent-hooks!" >&2
    echo "Matched pattern: ${pattern}" >&2
    echo "Action rejected under docs/permissions.md security policy." >&2
    echo "==================================================================" >&2
    exit 1
  fi
done

exit 0
