# ChoreSync: AI Tools Usage, Review & Verification Report

This document details how AI tools, coding assistants, agentic IDEs, and subagents were utilized, guided, reviewed, and verified throughout the design, development, and operationalization of **ChoreSync**.

---

## 1. Executive Summary & AI-Native Methodology

ChoreSync was built using the **AI-Native Spec-Driven Development Methodology**, synthesizing principles from the **BMAD Method**, **GitHub Spec Kit**, **OpenSpec**, and **Agent Plugins 1.0 (AAIF / AWS)** standard. 

Rather than relying on uncontrolled prompt-and-pray generation, the system advanced through **16 Socratic Decision Gates**:
- Requirements were elicited and frozen into functional domain specifications before selecting any tech stack.
- The API contract ([`contracts/openapi.yaml`](contracts/openapi.yaml)) was reverse-engineered and frozen before backend implementation.
- Every architectural choice (cloud hosting, database persistence, observability stack, local cloud emulator) was evaluated with trade-off matrices and confirmed at dedicated decision gates.
- Code was verified using containerized tests, deterministic SAST, living contract drift detection, mutation testing, and continuous agent evaluations.

---

## 2. AI Tools & Agents Employed

| Tool / Assistant | Role & Application Scope | Execution Context & Tooling |
|---|---|---|
| **Antigravity / Gemini CLI** | Architectural orchestration, Socratic gate transitions, container orchestration, Kubernetes manifests, and evaluation gates. | Meta-environment terminal agent with system prompt guardrails. |
| **Claude Code** | High-precision backend Go and React TypeScript coding, refactoring, and test authoring. | Terminal CLI operating through **RTK (Rust Token Killer)** token proxy (60–90% token reduction). |
| **Cursor & Windsurf** | Component styling, interactive tablet mode UI, and rapid frontend iteration. | Agentic IDE linked directly to [`AGENTS.md`](AGENTS.md) via `.cursorrules` and `.windsurfrules`. |
| **BMAD Agile Squad** | Specialized agent personas defined in [`.bmad/`](.bmad/): Product Manager (`pm.md`), System Architect (`architect.md`), Lead Developer (`developer.md`), QA Engineer (`qa.md`), and Scrum Master (`scrum-master.md`). | Context-isolated subagent personas with defined boundaries. |
| **Domain Specialist** | Domain expert subagent defined in [`custom-agent/specialist.md`](custom-agent/specialist.md) enforcing living arrangement rules across Flatmates, Families, and Couples. | Specialist subagent with domain-bounded tools. |
| **ChoreSync MCP Server** | Model Context Protocol server exposing 5 tools for cluster health, database schema inspection, contract verification, and chore lifecycle operations. | Zero-dependency Python 3 `stdio` server ([`mcp-server/server.py`](mcp-server/server.py)) configured in [`mcp.json`](mcp.json). |

---

## 3. Prompts & Task Delegation Strategy

### A. Spec-Driven Task Decomposition
Monolithic prompts (e.g. *"build a full household chore app"*) were strictly prohibited. Work was partitioned into isolated, verifiable milestones:

1. **Gate 1 (Requirements Discovery)**: Functional requirements only; zero tech stack discussion.
2. **Gate 2 & 3 (UI Prototype & Validation Journey)**: Scaffolded UI mock and 6-step manual test journey ([`_docs/manual-test.md`](_docs/manual-test.md)).
3. **Gate 5 (Contract Freeze)**: Generated [`contracts/openapi.yaml`](contracts/openapi.yaml) (26 endpoints, 46 schemas) before backend implementation.
4. **Gate 6 (Tech Stack & Cloud Emulator)**: Evaluated Render, Fly.io, Koyeb, Supabase, Neon, Floci, and MinIO. Selected Go + Chi + `sqlc` + PostgreSQL on Neon, with **Floci** (`floci/floci:latest`) as local S3/SQS emulator.
5. **Gate 7–11 (Implementation, CI/CD, Observability, Kubernetes)**: Multi-agent tasks executed in isolated git worktrees (`feat/task-<ID>`).
6. **Gate 12–15 (Security, Extension Pack, Quality Reinforcement)**: Semgrep SAST, Agent Plugins 1.0 packaging, mutation testing, and continuous eval harness.

### B. Concrete Prompting Templates

- **Socratic Requirements Discovery Prompt**:
  > *"Analyze the household chore domain across three distinct personas: Flatmates, Families with Children, and Couples. Elicit business logic, entities, verification gates, and anti-goals into `product-spec.md`. Do NOT propose or select any programming language, framework, database, or cloud host at this stage."*

- **Contract Reverse-Engineering Prompt**:
  > *"Analyze the confirmed frontend API service calls and data types. Reverse-engineer a complete OpenAPI 3.1 contract specification in `contracts/openapi.yaml`. Ensure all request/response bodies, query parameters, enum types, and error envelopes (`RESOURCE_NOT_FOUND`, `UNAUTHORIZED`, etc.) are exhaustively declared."*

- **Zero-Host Containerized Implementation Prompt**:
  > *"Implement the Go backend handlers strictly adhering to `contracts/openapi.yaml`. Assume the host machine has ZERO runtimes (no Go, no Node, no compilers). All tests, migrations, and builds must be executed exclusively via `docker compose` or ephemeral container runs."*

- **Local Cloud Emulator Parity Prompt**:
  > *"Integrate Floci (`floci/floci:latest`) on port 4566 to provide 100% cloud component coverage for AWS S3 and SQS. Implement S3 photo proof upload and SQS reminder nudges using official AWS Go SDK v2 with zero code forking between local emulator and production cloud."*

---

## 4. Context Files Provided to AI Agents

To prevent architectural drift and hallucinations, every AI interaction was anchored by explicit context files:

| Context File | Purpose & Role |
|---|---|
| [`product-spec.md`](product-spec.md) *(bridged to `_docs/specs.md`)* | Canonical source of truth for personas, entities (Household, Member, Chore, Swap, Reward), business logic, and anti-goals. |
| [`constitution.md`](constitution.md) *(bridged to `AGENTS.md`)* | Project supreme law declaring negative invariants, boundary rules, zero-host execution mandate, and decision gate sequencing. |
| [`contracts/openapi.yaml`](contracts/openapi.yaml) *(bridged to `openapi.yaml`)* | Frozen API contract. Neither frontend nor backend agents could modify this file unilaterally. |
| [`docs/permissions.md`](docs/permissions.md) | Least-privilege matrix defining read-only operations, Level 1 safe commands, and Level 2 human-authorized boundaries. |
| [`_docs/manual-test.md`](_docs/manual-test.md) | 6-step manual end-to-end verification journey used to cross-check UI and backend integration. |
| [`on-call-engineer/autonomy-policy.json`](on-call-engineer/autonomy-policy.json) | Machine-readable autonomy policy evaluated by code outside the LLM during incident response. |

---

## 5. Human Review & Oversight Process

### A. Human Decision Gates
The human developer acted as the final authority at every critical juncture:
1. **Gate 0 (Archetype Selection)**: Human confirmed the Full-Stack Web App archetype and local cloud strategy.
2. **Gate 1 (Spec Review)**: Human reviewed the 3 living arrangements (Flatmates, Families, Couples) and approved the domain specification.
3. **Gate 5 (Contract Review)**: Human approved the 26 OpenAPI endpoints before any backend code was written.
4. **Gate 6 (Stack Approval)**: Human approved Go 1.22 + Chi + PostgreSQL on Neon and Floci local emulator.
5. **Gate 12 (Security Sign-Off)**: Human reviewed Semgrep SAST results and validated the autonomy policy authorization script.
6. **Gate 14 (Plugin Evaluation Sign-Off)**: Human reviewed the 100% compliant Agent Extension Pack evaluation deliverable.

### B. Code Review & Verification Workflow
- **Git Diffs**: Every change was inspected using `git diff` before staging and committing.
- **Contract Drift Audits**: Verified with `verify-spec-drift` to ensure zero drift between code routes and `contracts/openapi.yaml`.
- **Anti-Tautology Verification**: Guarded against false-positive "always green" test suites using mutation testing (`run-mutation-test`).

---

## 6. Verification Gates & Guardrails

| Verification Gate | Enforcement Mechanism | Pass Criteria | Result |
|---|---|---|---|
| **Zero-Host-Runtime Mandate** | System prompt & Makefile targets | Zero compiler/toolchain dependencies on the host OS | ✅ 100% containerized |
| **Safety Interceptor Hook** | Bash pre-tool hook ([`com.antigravity.client/hooks/pre-tool-guardrail.sh`](com.antigravity.client/hooks/pre-tool-guardrail.sh)) | Destructive commands (`rm -rf`, `DROP TABLE`, unallowlisted binaries) blocked | ✅ Active & verified |
| **Deterministic SAST** | Containerized Semgrep (`security-audit/semgrep-rules.yml`) | Zero OWASP vulnerabilities, zero SQL injections, zero hardcoded credentials across 83 files | ✅ 0 issues found |
| **Backend Integration Suite** | Dockerized Go test runner against PostgreSQL 16 container | 15/15 unit and integration test packages passing | ✅ 15/15 passing (100%) |
| **Playwright E2E Suite** | Containerized Playwright testing all 12 user journeys | 12/12 journeys passing (including Floci S3/SQS and tablet mode) | ✅ 12/12 passing (100%) |
| **Spec & Contract Drift** | Automated AST route auditor (`verify-spec-drift`) | 100% alignment between OpenAPI paths and Go Chi handler routes | ✅ 0 drift (100%) |
| **Mutation Testing** | Containerized fault injector (`run-mutation-test`) | Intentionally injected logic mutations must fail the test suite ($\ge 80\%$) | ✅ 100% killed (2/2) |
| **Continuous Agent Evals** | Automated evaluation harness (`evals/eval_runner.py`) | 7/7 domain and constitutional assertions passing against golden dataset | ✅ 7/7 passing (100%) |
| **Extension Pack Conformance** | Semantic & syntactic plugin evaluator (`evaluate-extension-pack`) | Agent Plugins 1.0 schema, MCP tool signatures, and hook invariants | ✅ 44/44 passing (100%) |
| **Repository Contents Verifier** | Canonical repository compliance auditor (`verify-repo-contents`) | 21/21 expected core items, bridges, and documentation criteria | ✅ 21/21 passing (100%) |

---

## 7. AI Tool & Data Security Policy

- **Zero Secret Exposure**: No production API keys, database credentials, or private tokens are embedded in source code, committed to Git, or passed to AI models. All secrets are injected at runtime via environment variables loaded from `.gitignore`d `.env` files.
- **Credential Scrubbing**: Test fixtures use synthetic, seeded credentials (e.g. `alex@example.com`, bcrypt hashed passwords) with strict multi-tenant isolation.
- **Bounded Telemetry Boundaries**: On-call responder subagents access telemetry (Prometheus, Loki, Tempo) exclusively through read-only allowlisted endpoints via [`on-call-engineer/scripts/collect-evidence`](on-call-engineer/scripts/collect-evidence). Direct database write credentials are never provided to autonomous responders.
- **Outside-the-Model Action Authorization**: Remediation actions proposed by LLMs cannot be executed directly; they must pass through deterministic policy evaluation ([`on-call-engineer/scripts/authorize-action`](on-call-engineer/scripts/authorize-action)) enforcing Level 1 auto-approval vs Level 2 human escalation.
