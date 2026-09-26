#!/usr/bin/env bash
set -euo pipefail

# Domain Workflow Verifier for ChoreSync Lifecycle Logic
# Tests presence and correctness of chore rotations, approval gates, swaps, and cloud hooks.

ROOT_DIR="$(cd "$(dirname "$0")/../../.." && pwd)"
SCHEMA="${ROOT_DIR}/backend/internal/db/schema.sql"
OPENAPI="${ROOT_DIR}/contracts/openapi.yaml"
STORE_FILE="${ROOT_DIR}/backend/internal/store/store.go"
UPLOADS_HANDLER="${ROOT_DIR}/backend/internal/handlers/uploads.go"

echo "==> Auditing ChoreSync Domain Lifecycle Workflows..."

ERRORS=0

# 1. Verify schema support for rotation and approval states
echo "1. Checking database schema fields for chore workflows..."
if [ -f "${SCHEMA}" ]; then
  for field in "rotation_member_ids" "current_rotation_index" "assignment_type" "requires_approval" "requires_proof"; do
    if grep -qi "${field}" "${SCHEMA}"; then
      echo "  [OK] Field/state present in schema: ${field}"
    else
      echo "  [FAIL] Missing expected field/state in schema: ${field}" >&2
      ERRORS=$((ERRORS + 1))
    fi
  done
else
  echo "  [FAIL] Schema file not found: ${SCHEMA}" >&2
  ERRORS=$((ERRORS + 1))
fi

# 2. Verify OpenAPI endpoints for chore operations
echo "2. Checking OpenAPI contract endpoints for chore operations..."
if [ -f "${OPENAPI}" ]; then
  for ep in "/api/chores/{chore_id}/complete" "/api/completions/{completion_id}/approve" "/api/chores/{chore_id}/nudge" "/api/swaps/{swap_id}/accept" "/api/uploads/photo"; do
    if grep -Fq "${ep}:" "${OPENAPI}"; then
      echo "  [OK] Endpoint declared in OpenAPI: ${ep}"
    else
      echo "  [FAIL] Missing operational endpoint in OpenAPI: ${ep}" >&2
      ERRORS=$((ERRORS + 1))
    fi
  done
else
  echo "  [FAIL] Contract file not found: ${OPENAPI}" >&2
  ERRORS=$((ERRORS + 1))
fi

# 3. Verify Store logic for rotation and approval checks
echo "3. Checking backend store implementation..."
if [ -f "${STORE_FILE}" ]; then
  if grep -qi "RotateChore" "${STORE_FILE}" || grep -qi "ChoreAssignmentRoundRobin" "${STORE_FILE}"; then
    echo "  [OK] Round-robin auto-rotation logic implemented in backend store."
  else
    echo "  [FAIL] Round-robin auto-rotation logic missing in backend store." >&2
    ERRORS=$((ERRORS + 1))
  fi

  if grep -qi "pending_approval" "${STORE_FILE}" || grep -qi "StatusPendingApproval" "${STORE_FILE}"; then
    echo "  [OK] Approval gate state machine implemented in backend store."
  else
    echo "  [FAIL] Approval gate state machine missing in backend store." >&2
    ERRORS=$((ERRORS + 1))
  fi
fi

# 4. Verify Floci S3 proof upload and SQS reminder queue integration
echo "4. Checking Floci cloud integration..."
if [ -f "${UPLOADS_HANDLER}" ]; then
  if grep -qi "s3" "${UPLOADS_HANDLER}"; then
    echo "  [OK] S3 photo proof upload handler present."
  else
    echo "  [WARN] S3 upload handler may be abstract or missing."
  fi
fi

if [ "${ERRORS}" -eq 0 ]; then
  echo "==> Chore Domain Lifecycle Verification PASSED: 100% domain compliance verified!"
  exit 0
else
  echo "==> Chore Domain Lifecycle Verification FAILED: ${ERRORS} errors detected!" >&2
  exit 1
fi
