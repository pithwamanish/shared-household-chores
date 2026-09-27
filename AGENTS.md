# ChoreSync: AI-Native Agent Constitution & Guidelines (`AGENTS.md`)

Welcome to the **ChoreSync** repository. This document is the canonical constitution and operational playbook for all AI coding agents (Claude Code, Cursor, Antigravity, Copilot, Windsurf) and human contributors working on this codebase.

---

## 1. Project Mission & System Overview

**ChoreSync** is an equitable, flexible household chore coordination system designed to eliminate domestic friction and manage shared responsibilities across three key living arrangements:
1. **Flatmates / Roommates**: Peer-to-peer accountability, round-robin auto-rotation, task swapping marketplace, and transparent contribution history.
2. **Families with Children**: Admin/Parent verification gates, chore assignment, photo/notes proof verification, and gamified points/rewards.
3. **Couples / Lightweight Co-living**: Low-overhead shared task backlogs, voluntary task claiming, one-tap instant completions, and gentle nudges.

- **Product Specifications**: [`_docs/specs.md`](_docs/specs.md)
- **Manual Verification Journey**: [`_docs/manual-test.md`](_docs/manual-test.md)
- **Workflow State Ledger**: [`_docs/state.md`](_docs/state.md)

---

## 2. Repository Structure & Sandbox Boundaries

All agents must strictly respect domain boundaries to avoid cross-layer drift, file collisions, and architectural confusion:

```
household-chores/
├── AGENTS.md                  # Canonical constitution & agent guidelines (this file)
├── CLAUDE.md / GEMINI.md      # Tool pointer bridges to AGENTS.md
├── Makefile                   # Standardized development & verification targets
├── contracts/                 # Contract-First Source of Truth
│   └── openapi.yaml           # Frozen OpenAPI 3.1 specification (Step 5)
├── _docs/                     # Specifications, architecture, testing & agent guides
│   ├── specs.md               # Step 1: Core specification & entity models
│   ├── manual-test.md         # Step 3: End-to-end 6-step verification scenario
│   ├── state.md               # Step progression & lifecycle ledger
│   ├── process.md             # Step 4: Multi-agent execution graph & worktree rules
│   ├── task-template.md       # Step 4: Task grooming & issue template
│   ├── stack.md               # Step 6: Tech stack & free-tier deployment decision
│   ├── tasks.md               # Step 8: Granular task backlog & statuses
│   └── team/                  # Subagent personas (pm.md, swe.md, qa.md, sre.md)
├── frontend/                  # React + Vite + TypeScript application
│   ├── src/
│   │   ├── components/        # UI views (Board, Modals, Swaps, Rewards, Activity)
│   │   ├── services/          # API client / service abstraction layer
│   │   └── types/             # Domain and DTO TypeScript interfaces
│   └── package.json
├── backend/                   # Application backend service (scaffolded in Step 7)
│   ├── src/                   # Routers, business logic, persistence store
│   └── tests/                 # Unit and integration test suites
└── e2e/                       # Playwright end-to-end test suite (Step 9)
```

### Boundary & Operational Constraints
1. **Constitutional Supreme Law**: All agents must strictly respect the negative invariants defined in [`constitution.md`](constitution.md) (zero host runtimes, zero plaintext credentials, zero direct DB mutations, contract immobility).
2. **Repository Boundary**: Confine all file reads, writes, and terminal commands strictly to this repository. Never inspect, search, or read parent (`..`) or sibling directories.
3. **Git & Worktree Isolation**: Make focused, regular commits explaining the architectural rationale. Parallel feature work must execute in isolated Git worktrees (`.worktrees/task-<ID>`).
4. **Process & Orchestration**: Follow the multi-agent orchestration graph defined in [`_docs/process.md`](_docs/process.md).
5. **Frontend Isolation**: Frontend agents MUST NOT alter backend source files, database schemas, or deployment configs.
6. **Backend Isolation**: Backend agents MUST NOT edit frontend components or styling.
7. **Contract Immobility**: Neither Frontend nor Backend agents may alter [`contracts/openapi.yaml`](contracts/openapi.yaml) unilaterally. Changes to the contract require prior PM alignment and contract synchronization.
8. **Documentation Integrity**: Do not delete existing comments, architectural decisions, or docstrings unless explicitly requested.

---

## 3. Development Lifecycle (Spec-Driven Architecture)

Every agent operates within the comprehensive AI-Native Spec-Driven Lifecycle tracked in [`_docs/state.md`](_docs/state.md):

| Gate / Step | Milestone | Canonical Artifact | Status |
| :--- | :--- | :--- | :--- |
| **Gate 0** | Track Selection & Archetype Confirmation | `AGENTS.md`, `constitution.md` | ✅ Complete |
| **Pre-Step 1**| Working Backwards PRFAQ | `PRFAQ.md` / `_docs/prfaq.md` | ✅ Complete |
| **Gate 1 (Step 1)** | Scoping & Functional Spec | `_docs/specs.md` (`product-spec.md`) | ✅ Complete |
| **Gate 2 (Step 2)** | Frontend Prototype & Service Mock | `frontend/` | ✅ Complete |
| **Gate 3 (Step 3)** | Manual Verification Journey | `_docs/manual-test.md` | ✅ Complete |
| **Gate 4 (Step 4)** | Governance Bundle & BMAD Squad | `AGENTS.md`, `.bmad/`, `_docs/process.md` | ✅ Complete |
| **Gate 5 (Step 5)** | Contract Freeze (OpenAPI 3.1) | `contracts/openapi.yaml` | ✅ Complete |
| **Gate 6 (Step 6)** | Stack Decision & Nygard ADR | `_docs/stack.md`, `docs/adr/` | ✅ Complete |
| **Gate 7 (Step 7)** | Backend Foundation & In-Memory Store | `backend/` | ✅ Complete |
| **Gate 8 (Step 8)** | Multi-Agent Task Orchestration | `_docs/tasks.md` + Worktrees | ✅ Complete |
| **Gate 9 (Step 9)** | Playwright E2E Verification Suite | `e2e/` (Containerized) | ✅ Complete |
| **Gate 10** | Local Kubernetes Deployment (Kind) | `k8s/`, `k8s/verify` | ✅ Complete |
| **Gate 11** | Local CI/CD (Act & Kind Rollout) | `.github/workflows/ci.yml` | ✅ Complete |
| **Gate 12** | Deterministic Security (Semgrep) & On-Call | `security-audit/`, `on-call-engineer/` | ✅ Complete |
| **Gate 13** | Agent Extension Pack (Agent Plugins 1.0) | `plugin.json`, `mcp.json`, `skills/` | ✅ Complete |
| **Gate 14** | Plugin Conformance Evaluation | `_docs/agent-extension-pack-evaluation.md` | ✅ Complete |
| **Gate 15** | Living Spec Drift & Mutation Testing | `verify-spec-drift`, `run-mutation-test` | ✅ Complete |
| **Gate 16** | AI Tools Usage & Canonical Repository | `docs/ai-tools-usage-and-review.md`, `README.md` | ✅ Complete |

---

## 4. Multi-Agent Roles & Responsibilities

When executing multi-agent workflows, agents adopt the following roles:

### 1. Orchestrator / Coordinator
- Reads `_docs/tasks.md` and coordinates subagent assignment.
- Dispatches tasks into isolated Git worktrees.
- Resolves dependencies and conducts final merges into `main`.
- **Golden Rule**: The orchestrator coordinates and verifies; it NEVER writes code directly.

### 2. Product Manager (PM) Agent (`_docs/team/pm.md`)
- Audits task acceptance criteria against `_docs/specs.md`.
- Maintains `contracts/openapi.yaml` sync across frontend and backend.
- Manages `_docs/tasks.md` status progression (`TODO` → `IN_PROGRESS` → `READY_FOR_QA` → `DONE`).

### 3. Software Engineer (SWE) Agent (`_docs/team/swe.md`)
- Works in an isolated git worktree branch: `feat/task-<ID>`.
- Implements features adhering strictly to `contracts/openapi.yaml` schemas.
- Writes corresponding unit and integration tests.
- Verifies local functionality before handing off to QA.

### 4. QA & Verification Agent (`_docs/team/qa.md`)
- Operates in a fresh, independent context to prevent hallucination / context rot.
- Validates the manual test scenario (`_docs/manual-test.md`) and automated test suites.
- If verification passes: Marks task `PASSED` and notifies Orchestrator to merge.
- If verification fails: Rejects task with reproducible execution logs and error details.

---

## 5. Architectural & Coding Conventions

### 5.1 Contract-First Principle
- All API interactions (endpoints, parameters, request payloads, response envelopes, status codes) must strictly mirror `contracts/openapi.yaml`.
- Error responses must follow the universal JSON error schema:
  ```json
  {
    "error": "RESOURCE_NOT_FOUND",
    "message": "Chore with ID c-123 was not found in household h-roommates",
    "code": 404,
    "details": {}
  }
  ```

### 5.2 Business Logic Rules
- **Chore Recurrence & Rotation**: Advancing a round-robin rotation or generating recurring chore instances must be idempotent.
- **Approval Gate Workflow**: When `requires_approval` is `true`, marking a chore complete moves status to `pending_approval`. Points are credited **only** upon Admin approval.
- **Peer Swaps**: A swap proposal shifts chore status or assigns pending flags without breaking current assignee responsibility until the recipient formally accepts.
- **Gamification Integrity**: Point balances cannot drop below zero.

### 5.3 Code Quality & Tooling
- **TypeScript**: Strict mode enabled; no implicit `any`; prefer interfaces and explicit return types.
- **Component Design**: Modular, atomic UI components with clear props and accessible labels.
- **State Management**: Predictable optimistic updates with graceful rollback on API failure.
- **Token Efficiency**: AI agents utilize token-optimized tools internally when running shell commands.

---

## 6. Git & Worktree Conventions

- **Branch Naming**:
  - Features: `feat/task-<ID>-<short-description>`
  - Fixes: `fix/task-<ID>-<short-description>`
  - Docs: `docs/<description>`
- **Worktree Pattern for Parallel Agents**:
  ```bash
  git worktree add .worktrees/task-<ID> -b feat/task-<ID>
  # Implement, test, verify
  git worktree remove .worktrees/task-<ID>
  ```
- **Commit Message Format**: Follow Conventional Commits:
  - `feat(chores): implement round-robin auto-rotation logic`
  - `fix(swaps): prevent double claim when swap is pending`
  - `test(e2e): add playwright scenario for approval gate`
  - `docs(specs): update rotation index semantics`

---

## 7. Verification Gates

Before any task or pull request is declared complete, it must pass the following verification gates:
1. **Contract Check**: Zero drift between `contracts/openapi.yaml` and implementation.
2. **Lint & Typecheck**: No TypeScript or linter errors (`npm run lint` or `npm run typecheck`).
3. **Unit Tests**: All unit test suites pass (Go: `go test ./...` / Frontend: `npm test`).
4. **End-to-End**: User verification journey passes without regression (`_docs/manual-test.md`).

---

## 8. Fast Iteration Protocol (Speed & Efficiency Rules)

To eliminate unnecessary waiting during bug fixes, enhancements, and feature development, all agents MUST follow these performance rules:
1. **Live Hot-Reloading Over Rebuilds**:
   - Frontend changes (`frontend/src/`) are instantly reflected via Vite HMR with polling watch. Never execute `npm run build` or rebuild the frontend container for routine UI changes.
   - Backend changes (`backend/internal/`, `backend/cmd/`) are recompiled and restarted in `< 1s` via Air (`.air.toml`). Never run `docker compose up --build backend` for Go code edits.
2. **Targeted Verification (Default for Small Changes)**:
   - For bug fixes and minor logic edits, run *only* the specific test affected:
     - Unit: `docker run --rm -v $(pwd)/backend:/app -w /app golang:1.22-alpine go test -run "<TestName>" ./...`
     - E2E: `docker compose run --rm e2e npx playwright test -g "<ScenarioName>"`
   - Reserve full regression runs (`docker compose run --rm e2e`) strictly for final milestone sign-offs.
3. **Fail Fast**:
   - When a test assertion fails, inspect the trace/logs directly and patch surgically. Never rerun unrelated test journeys until the targeted fix is verified.

