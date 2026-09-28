# Specialist Subagents & Custom Agent Loops (`custom-agent/`)

This directory houses the specialist subagent personas and the headless autonomous custom agent loop for ChoreSync, bridging **Zoomcamp Module 5** and the **BMAD 5-Agent Squad Architecture**.

---

## 1. Decision Matrix: Interactive vs. Custom Agents

When designing agent workflows, use the following operational matrix:

| Dimension | Interactive Coding Agent | Custom Autonomous Agent (`agent.py`) |
| :--- | :--- | :--- |
| **Tooling** | Antigravity, Claude Code, Cursor, Windsurf | Python script runner, Docker runner, headless CLI |
| **Best Suited For** | Interactive feature implementation, vibe-coding, refactoring, exploratory debugging | Non-interactive CI/CD checks, scheduled contract drift audits, structured prompt evals |
| **Context Window** | Conversational, shared with human developer | Bounded, single-objective, deterministic exit criteria |
| **Human in Loop** | Continuous back-and-forth review | Evaluates against declarative policies (`docs/permissions.md`) |
| **Execution Trigger**| Human prompt in terminal or IDE | Webhook, cron, GitHub Action, automated test step |

---

## 2. Subagent Directory & Personas

Subagent persona definitions provide scoped system prompts to prevent LLM context rot during large epics:

1. **`specialist.md`**: ChoreSync Architecture & Governance Specialist. Enforces contract freeze, relational DDL invariants, and cloud emulator parity.
2. **`pm.md`**: Product Manager Agent. Audits acceptance criteria and keeps `contracts/openapi.yaml` in sync.
3. **`architect.md`**: System Architect Agent. Owns contract freezes and Architecture Decision Records (`docs/adr/`).
4. **`developer.md` / `swe.md`**: Software Engineer Agent. Implements features adhering strictly to `contracts/openapi.yaml` inside container worktrees.
5. **`qa.md`**: QA & Verification Agent. Executes manual test scenarios and Playwright test suites in isolated contexts.
6. **`scrum-master.md`**: Workflow Coordinator. Maintains Kanban state in `_docs/tasks.md` and gate ledgers.
7. **`sre.md`**: Site Reliability Engineer Agent. Manages Docker Compose, Kubernetes manifests (`k8s/`), and observability.

---

## 3. Parallel Subagents via Git Worktrees

To prevent context rot during multi-file parallel tasks:
1. The Orchestrator stays in the primary branch.
2. Subagents are invoked in isolated Git worktrees:
   ```bash
   git worktree add .worktrees/task-api -b feat/task-api
   # Subagent performs scoped task inside .worktrees/task-api
   git worktree remove .worktrees/task-api
   ```
3. Subagents deliver isolated git diffs for serial review and merge.

---

## 4. Headless Autonomous Loop (`custom-agent/agent.py`)

For headless automation in CI/CD or background maintenance, run the custom agent loop:

```bash
docker run --rm -v $(pwd):/app -w /app python:3.11-alpine python3 custom-agent/agent.py --task audit
```

Supported tasks:
- `--task audit`: Audits OpenAPI contracts and PostgreSQL schemas.
- `--task health`: Probes multi-service container health.
- `--task drift`: Verifies zero spec drift against implementation.
