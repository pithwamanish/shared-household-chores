# Task Template

Use this template when grooming tasks in `_docs/tasks.md` or opening issues for subagents.

---

## Objective
<!-- High-level outcome of this task, e.g. "Implement round-robin chore auto-rotation logic in the backend" -->

## Context
<!-- References to _docs/specs.md sections and contracts/openapi.yaml endpoints/schemas -->
- Specs: [`_docs/specs.md`](specs.md)
- Contract: [`contracts/openapi.yaml`](../contracts/openapi.yaml)
- Manual Test Journey: [`_docs/manual-test.md`](manual-test.md)

## Scope & File Boundaries
- **Assigned Role**: SWE / QA / SRE
- **Branch / Worktree**: `feat/task-<ID>` / `.worktrees/task-<ID>`
- **Files to create or modify**:
  - `backend/src/...`
  - `backend/tests/...`
  - (No modification to `frontend/` or `contracts/openapi.yaml` without PM approval)

## Subtasks
- [ ] Subtask 1: Detailed action item
- [ ] Subtask 2: Detailed action item
- [ ] Subtask 3: Unit and integration tests

## Acceptance Criteria
- [ ] Conforms strictly to `contracts/openapi.yaml` (zero payload drift)
- [ ] Universal JSON error envelope returned on failure conditions
- [ ] No direct host runs: Tests executed inside container (`docker compose run --rm ...` or SDK container)
- [ ] Active Running Containers Gate: `docker compose ps` confirms services are `Up`
- [ ] Live endpoint verification: Health and API endpoints return 200 OK over HTTP
- [ ] Relevant steps in `_docs/manual-test.md` pass without regression
- [ ] Code passes type-checking and linter checks with zero warnings/errors
- [ ] CI/CD Pipeline Gate: All GitHub Actions workflow checks pass (`backend-ci`, `frontend-ci`, `contract-audit`, `docker-prod-build`, `e2e-ci`)
