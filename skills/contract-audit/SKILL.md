---
name: contract-audit
description: >-
  Audits OpenAPI 3.1 contract endpoints, request/response models, and status codes
  against Go backend route handlers and frontend TypeScript API client methods.
---

# Contract Audit Capability Workflow

Use this capability workflow to verify contract-first synchronization between `contracts/openapi.yaml`, the Go backend router (`backend/internal/handlers/`), and the React TypeScript client (`frontend/src/services/api.ts`).

---

## 1. Automated Verification Script

Run the automated contract audit script:

```bash
bash agent-capabilities/contract-audit/scripts/audit-contract.sh
```

## 2. Manual Inspection Steps

1. **Verify Contract Paths**:
   Confirm that all 26 paths in `contracts/openapi.yaml` (including `/api/v1/chores`, `/api/v1/swaps`, `/api/v1/rewards`, `/api/v1/uploads/photo`, `/api/v1/auth/login`) are registered in `backend/cmd/server/main.go`.

2. **Verify Schema Types**:
   Ensure TypeScript interfaces in `frontend/src/types/index.ts` strictly mirror the schemas defined in `contracts/openapi.yaml`.

3. **Verify Error Responses**:
   Ensure all error responses conform to the universal error envelope:
   ```json
   {
     "error": "RESOURCE_NOT_FOUND",
     "message": "Chore with ID c-123 was not found",
     "code": 404,
     "details": {}
   }
   ```
