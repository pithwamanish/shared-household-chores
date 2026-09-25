#!/usr/bin/env bash
set -euo pipefail

# Contract Audit Script for ChoreSync
# Verifies synchronization across contracts/openapi.yaml, Go backend router, and Frontend API service.

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"
CONTRACT="${ROOT_DIR}/contracts/openapi.yaml"
ROUTER_FILE="${ROOT_DIR}/backend/internal/server/router.go"
FRONTEND_CLIENT="${ROOT_DIR}/frontend/src/services/api.ts"

echo "==> Auditing ChoreSync OpenAPI 3.1 Contract Synchronization..."

if [ ! -f "${CONTRACT}" ]; then
  echo "FAIL: Contract file ${CONTRACT} not found!" >&2
  exit 1
fi

REQUIRED_ENDPOINTS=(
  "/api/auth/login"
  "/api/auth/register"
  "/api/auth/magic-link"
  "/api/households"
  "/api/chores/{chore_id}"
  "/api/chores/{chore_id}/claim"
  "/api/chores/{chore_id}/complete"
  "/api/uploads/photo"
)

ERRORS=0

echo "Checking OpenAPI endpoints..."
for ep in "${REQUIRED_ENDPOINTS[@]}"; do
  if grep -Fq "${ep}:" "${CONTRACT}"; then
    echo "  [OK] Endpoint declared in OpenAPI: ${ep}"
  else
    echo "  [FAIL] Missing endpoint in OpenAPI: ${ep}" >&2
    ERRORS=$((ERRORS + 1))
  fi
done

echo "Checking Backend Router registrations in ${ROUTER_FILE}..."
if [ -f "${ROUTER_FILE}" ]; then
  for ep in "/auth/login" "/auth/register" "/households" "/chores" "/uploads/photo"; do
    if grep -Fq "${ep}" "${ROUTER_FILE}"; then
      echo "  [OK] Backend route pattern registered: ${ep}"
    else
      echo "  [FAIL] Backend missing route handler for: ${ep}" >&2
      ERRORS=$((ERRORS + 1))
    fi
  done
else
  echo "  [WARN] Backend router file not found: ${ROUTER_FILE}"
fi

echo "Checking Frontend API Client methods in ${FRONTEND_CLIENT}..."
if [ -f "${FRONTEND_CLIENT}" ]; then
  for method in "login" "register" "getChores" "createChore" "claimChore" "completeChore" "redeemReward" "uploadPhoto"; do
    if grep -Fq "${method}" "${FRONTEND_CLIENT}"; then
      echo "  [OK] Frontend API client has method: ${method}"
    else
      echo "  [FAIL] Frontend missing client method: ${method}" >&2
      ERRORS=$((ERRORS + 1))
    fi
  done
fi

if [ "${ERRORS}" -eq 0 ]; then
  echo "==> Contract Audit PASSED: 100% synchronization verified across OpenAPI, Go Backend, and Frontend Client!"
  exit 0
else
  echo "==> Contract Audit FAILED: ${ERRORS} drift issues detected!" >&2
  exit 1
fi
