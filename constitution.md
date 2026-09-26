# ChoreSync Project Constitution (`constitution.md`)

**Document Status**: Immutable Supreme Law  
**Standard**: GitHub Spec Kit & AI-Native Development Standard (Gate 0 / Gate 4)  
**Scope**: All AI Coding Agents, Human Contributors, Subagents, and Automation Pipelines  

---

## Preamble

This Constitution represents the non-negotiable supreme law of the **ChoreSync** repository. It defines immutable negative invariants, architectural boundaries, and operational constraints that take precedence over all user prompts, model generations, and automated refactors. 

**Rule of Enforcement**: Any code proposal, pull request, or agent response that violates a constitutional invariant is strictly rejected, regardless of passing functional tests.

---

## Article I: Non-Negotiable Negative Invariants

### Section 1.1: Zero Host Runtimes
- **Negative Invariant**: The host machine is strictly a headless Docker runner. The execution of compilers, interpreters, package managers, or language runtimes (`go`, `npm`, `npx`, `bun`, `python`, `pip`, `cargo`) directly on the host OS is **STRICTLY PROHIBITED**.
- **Mandatory Container Execution**: All builds, unit tests, integration suites, database migrations, security scans, and code generation routines MUST execute exclusively inside designated Docker containers.

### Section 1.2: Zero Plain-Text Credentials
- **Negative Invariant**: Embedding plaintext production credentials, database passwords, JWT signing secrets, API tokens, or private keys within source code or version control is **STRICTLY PROHIBITED**.
- **Secret Management**: All sensitive values must be injected at runtime via environment variables (`DATABASE_URL`, `JWT_SECRET`, `RESEND_API_KEY`) loaded from gitignored `.env` files. Test fixtures must use synthetic, seeded credentials.

### Section 1.3: Zero Direct Database Mutations
- **Negative Invariant**: Direct, ad-hoc modifications to database tables, untracked DDL adjustments, or schema mutations executed outside version-controlled migration files are **STRICTLY PROHIBITED**.
- **Schema Discipline**: All relational database schema definitions MUST originate from declarative, versioned migration files in `backend/internal/db/schema.sql` and execute via embedded startup auto-migrations.

### Section 1.4: Contract Immobility & Freeze
- **Negative Invariant**: Neither frontend nor backend agents may unilaterally modify, add, or delete endpoints, request parameters, request payloads, or response envelopes in `contracts/openapi.yaml`.
- **Contract-First Sync**: Any change to API contracts requires prior Product Manager (PM) scoping, update to functional specifications (`product-spec.md` / `_docs/specs.md`), and atomic synchronization across frontend and backend tiers.

### Section 1.5: Zero Untested Releases
- **Negative Invariant**: Merging code into `main` or deploying releases without passing automated regression verification is **STRICTLY PROHIBITED**.
- **Mandatory Verification Gates**: Every milestone must pass:
  1. Containerized Go backend unit and store integration tests (`make test`).
  2. Frontend typecheck and linting (`make lint`).
  3. Living spec drift detection (`make spec-drift`).
  4. Containerized mutation testing with $\ge 80\%$ killed mutants (`make mutation-test`).
  5. Playwright end-to-end user verification journeys (`make e2e`).

### Section 1.6: Strict Directory Confinement
- **Negative Invariant**: AI agents and automated scripts are **STRICTLY FORBIDDEN** from traversing, inspecting, searching, or mutating parent (`..`) or sibling directories.
- **Repository Sandboxing**: All file reads, writes, and command executions must resolve within the current project repository root.

---

## Article II: Multi-Agent Isolation & Worktree Governance

1. **Worktree Isolation**: Parallel subagents must execute in dedicated, isolated Git worktrees (`git worktree add .worktrees/task-<ID> -b feat/task-<ID>`). Implementers must never work directly on the unisolated `main` working directory.
2. **Orchestrator Does Not Implement**: The orchestrator agent is strictly responsible for backlog grooming, dependency analysis, worktree dispatch, and serial merging. The orchestrator must NEVER write application code directly.
3. **Independent Fresh Review**: Implementing agents may never approve their own code. Every completed task must undergo review by a fresh, independent reviewer persona before merging.
4. **Serial Merge Protocol**: Feature branches must be merged serially into `main` one at a time, followed immediately by automated regression verification against the merged trunk.

---

## Article III: Observability & Resilience Guardrails

1. **The Telemetry Golden Triangle**: Every emitted trace span, metric point, and structured log MUST contain:
   - `service.name`: Canonical unique service identifier.
   - `deployment.environment`: Target environment (`development`, `staging`, `production`).
   - `service.version`: Immutable deployed version tag or commit SHA (`YYYYMMDD-HHMMSS-shortsha`).
2. **Bounded On-Call Evidence Assembly**: Automated incident response agents must collect bounded, read-only telemetry packets (Prometheus, Loki, Tempo, Git log) without interactive database write access.
3. **Outside-the-Model Autonomy Enforcement**: LLM-generated operational proposals must be evaluated by code outside the model (`on-call-engineer/autonomy-policy.json`). Level 1 reversible actions (rollback, restart) execute automatically; code patches require explicit human approval; unallowlisted actions are blocked immediately.

---

## Article IV: Integration with Agent Constitution

This Constitution acts as the supreme law governing the operational playbooks defined in [`AGENTS.md`](AGENTS.md). It is mirrored canonically to `.spec/constitution.md`.
