# ChoreSync Agent Extension Pack: 6-Step Interactive Demo Script

This document details the concrete, reproducible **6-Step Interactive Demo Script** for the ChoreSync Agent Extension Pack, conforming strictly to the **Agent Plugins 1.0 Open Standard** ([agent-plugins.org](https://agent-plugins.org/) / AAIF / AWS) and **Zoomcamp Module 5** (`05-agent-capabilities/README.md`).

This script validates that an AI coding agent correctly leverages instructions, reusable skills, specialized personas, MCP tools, and safety guardrails under human review.

---

## The 6-Step Verification Workflow

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                   6-STEP INTERACTIVE VERIFICATION WORKFLOW                  │
├─────────────────────────────────────────────────────────────────────────────┤
│ Step 1: Agent Reads Project Instructions (AGENTS.md, constitution.md)       │
│ Step 2: Reusable Workflow Invocation (skills/ & agent-capabilities/)        │
│ Step 3: Specialized Subagent Execution (custom-agent/specialist.md, .bmad/) │
│ Step 4: Model Context Protocol (MCP) Tool Call (mcp-server/server.py)       │
│ Step 5: Guardrail & Hook Enforcement (pre-tool-guardrail.sh, audit log)     │
│ Step 6: Human Review of Final Diff (docs/permissions.md security audit)     │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

### Step 1: Agent Reads Project Instructions

- **Objective**: The AI coding assistant accesses the repository constitution and internalizes project architecture, zero-host-runtime mandate, and domain boundaries.
- **Verification Target**: [`AGENTS.md`](../AGENTS.md) and [`constitution.md`](../constitution.md).
- **Execution**:
  1. The agent inspects `AGENTS.md` and discovers:
     - Project mission: Household chore coordination for Flatmates, Families, and Couples.
     - Zero-host-runtime mandate: All tools, tests, and builds run in Docker containers.
     - Contract-first principle: All API modifications must mirror `contracts/openapi.yaml`.
  2. The agent verifies negative invariants: Zero direct DB alterations, zero plaintext secrets.
- **Expected Outcome**: The agent confirms understanding and constrains execution to containerized commands and approved domain directories.

---

### Step 2: Reusable Workflow Invocation

- **Objective**: The agent invokes deterministic domain capability workflows declared in [`skills/`](../skills/) rather than hallucinating ad-hoc shell commands.
- **Verification Target**: Reusable skill workflows:
  - `contract-audit`: `bash skills/contract-audit/scripts/audit-contract.sh`
  - `db-migration-runner`: `bash skills/db-migration-runner/scripts/check-migrations.sh`
  - `cluster-health-prober`: `bash skills/cluster-health-prober/scripts/probe-cluster.sh`
  - `chore-lifecycle-manager`: `bash skills/chore-lifecycle-manager/scripts/verify-chore-flows.sh`
- **Execution**:
  ```bash
  # Run contract drift audit capability
  bash skills/contract-audit/scripts/audit-contract.sh

  # Run chore lifecycle manager capability
  bash skills/chore-lifecycle-manager/scripts/verify-chore-flows.sh
  ```
- **Expected Outcome**: All 4 automated capability scripts execute successfully, confirming 0 contract drift, 11 PostgreSQL schema tables verified, and domain state transitions compliant across Flatmates, Families, and Couples.

---

### Step 3: Specialized Subagent Execution

- **Objective**: Execute a complex architectural task in an isolated context window using the dedicated domain specialist subagent persona.
- **Verification Target**: [`custom-agent/specialist.md`](../custom-agent/specialist.md) (or `.bmad/` squad).
- **Execution**:
  1. Orchestrator invokes the `ChoreSync Architecture & Governance Specialist` persona:
     ```markdown
     Role: ChoreSync Architecture & Governance Specialist
     Scope: Auditing chore rotation, parent approval gates, and Floci cloud emulator integration.
     ```
  2. The subagent operates within bounded tool permissions (read-only inspection of OpenAPI contracts, DDL schema, and container status).
  3. The subagent reports architectural audit findings back to the main agent without context rot.
- **Expected Outcome**: Focused architectural review completed with zero memory pollution in the parent agent session.

---

### Step 4: Model Context Protocol (MCP) Tool Call

- **Objective**: The agent interacts with the local MCP server via standard JSON-RPC 2.0 stdio to inspect application state without direct unconstrained shell access.
- **Verification Target**: Root [`mcp.json`](../mcp.json) and [`mcp-server/server.py`](../mcp-server/server.py).
- **Execution**:
  1. The agent queries tools list:
     ```bash
     echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | \
       docker run --rm -i -v $(pwd):/app -w /app python:3.11-alpine python3 mcp-server/server.py
     ```
  2. The agent executes `inspect_chore_operations`:
     ```bash
     echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "inspect_chore_operations", "arguments": {"archetype": "flatmates"}}}' | \
       docker run --rm -i -v $(pwd):/app -w /app python:3.11-alpine python3 mcp-server/server.py
     ```
- **Expected Outcome**: Server returns structured JSON tool descriptions and executes `inspect_chore_operations`, returning rotation rules, swap policies, and active chores.

---

### Step 5: Guardrail & Hook Enforcement

- **Objective**: Prove that the extension pack's automated client hooks intercept and block destructive operations violating project negative invariants.
- **Verification Target**: [`com.antigravity.client/hooks/pre-tool-guardrail.sh`](../com.antigravity.client/hooks/pre-tool-guardrail.sh) and [`com.antigravity.client/hooks/post-tool-audit.sh`](../com.antigravity.client/hooks/post-tool-audit.sh).
- **Execution**:
  1. Simulate an unauthorized destructive operation:
     ```bash
     echo "DROP DATABASE choresync;" | bash com.antigravity.client/hooks/pre-tool-guardrail.sh
     ```
  2. Test a bare-metal host execution attempt:
     ```bash
     echo "rm -rf /" | bash com.antigravity.client/hooks/pre-tool-guardrail.sh
     ```
- **Expected Outcome**:
  - The guardrail hook intercepts the forbidden command, outputs `GUARDRAIL BLOCKED: Destructive command intercepted by agent-hooks!`, and exits with code `1`.
  - The attempt is logged in `com.antigravity.client/hooks/audit.log` by `post-tool-audit.sh`.

---

### Step 6: Human Review of Final Diff

- **Objective**: Enforce the human-in-the-loop security gate before committing or deploying code changes.
- **Verification Target**: [`docs/permissions.md`](permissions.md) and `git diff`.
- **Execution**:
  1. The developer inspects the proposed git diff:
     ```bash
     rtk git diff
     ```
  2. The developer audits changes against the least-privilege matrix in `docs/permissions.md`:
     - Are contracts modified without PM approval? (Denied)
     - Are raw database credentials exposed? (Denied)
     - Are tests executed inside containers? (Allowed)
  3. Once validated, the human signs off and merges the changes.
- **Expected Outcome**: Human operator verifies compliance against security boundaries and completes the promotion cycle.

---

## 2. Automated Demo Verification

You can verify the entire Extension Pack workflow automatically using the project's Makefile:

```bash
# 1. Run MCP tool tests
make ext-mcp-test

# 2. Run domain capability workflows
make ext-capabilities

# 3. Verify extension pack integrity (Zoomcamp Module 5)
make ext-verify

# 4. Run full semantic conformance evaluation
make ext-eval
```
