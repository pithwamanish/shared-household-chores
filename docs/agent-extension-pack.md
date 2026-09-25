# ChoreSync Agent Extension Pack Documentation

This repository includes a production-grade **Agent Extension Pack** adhering to the [Agent Plugins Open Standard](https://agent-plugins.org/) (backed by the Agent Application Interface Framework / AAIF and AWS). 

It packages ChoreSync repository instructions, reusable chore coordination workflows, specialist subagents, Model Context Protocol (MCP) tools, guardrail hooks, and least-privilege security boundaries into a standardized, cross-agent extension bundle compatible with Claude Code, Cursor, Antigravity/Gemini CLI, and Copilot.

---

## 1. Extension Pack Directory Layout

```text
household-chores/
├── plugin.json                              # Agent Plugins v1.0.0 root specification manifest
├── agent-capabilities/                      # Reusable procedural workflows (with SKILL.md)
│   ├── contract-audit/                      # OpenAPI 3.1 contract synchronization audit
│   │   ├── SKILL.md                         # Capability instructions & acceptance criteria
│   │   └── scripts/audit-contract.sh        # Contract drift verification script
│   ├── db-migration-runner/                 # PostgreSQL schema & DDL validation
│   │   ├── SKILL.md                         # Migration invariants & entity documentation
│   │   └── scripts/check-migrations.sh      # DDL verification script across 11 tables
│   └── cluster-health-prober/               # Multi-tier cluster health inspection
│       ├── SKILL.md                         # Service probe targets & ports
│       └── scripts/probe-cluster.sh         # Live probe script (API, Caddy, DB, Floci)
├── agent-hooks/                             # Event interceptors & security guardrails
│   ├── hooks.json                           # Hook declarations (pre/post tool execution)
│   ├── pre-tool-guardrail.sh                # Pre-execution safety enforcement (blocks drops/force-pushes)
│   └── post-tool-audit.sh                   # Post-execution audit logging to audit.log
├── mcp-server/                              # Model Context Protocol tools & stdio server
│   ├── mcp.json                             # MCP tool declarations and input schemas
│   └── server.py                            # Zero-dependency Python 3 JSON-RPC 2.0 MCP server
├── custom-agent/                            # Specialist subagents
│   └── specialist.md                        # ChoreSync Architecture & Governance Specialist
└── docs/                                    # Extension documentation & governance
    ├── agent-extension-pack.md              # This architecture guide & project instructions
    └── permissions.md                       # Security boundaries & least-privilege policy
```

---

## 2. Project Instructions for AI Agents

All AI agents interacting with this repository MUST adhere to these operating principles:

1. **Manifest Discovery**:
   - Inspect `plugin.json` at the start of any session to identify registered capabilities, MCP tools, and guardrails.
2. **Consult Security Policy**:
   - Review `docs/permissions.md` before executing any commands or introducing state changes.
3. **Use Reusable Capabilities**:
   - Prefer capabilities declared in `agent-capabilities/` (e.g. `audit-contract.sh`, `check-migrations.sh`, `probe-cluster.sh`) over ad-hoc host commands.
4. **Zero Host Runtime Mandate**:
   - Assume zero host runtimes. All tests, builds, and backend services execute inside containers or via lightweight inspection scripts.
5. **Comply with Guardrail Hooks**:
   - Tool executions are automatically validated by `agent-hooks/pre-tool-guardrail.sh`. Unsafe operations (such as `DROP DATABASE`, `TRUNCATE TABLE`, or force-pushes to `main`) are rejected immediately.

---

## 3. Registered Extension Components

### A. Reusable Workflows (`agent-capabilities/`)
1. **`contract-audit`** (`agent-capabilities/contract-audit/SKILL.md`):
   - Audits OpenAPI 3.1 contract endpoints, request/response models, and status codes against Go backend handlers and TypeScript frontend client.
   - Run via: `bash agent-capabilities/contract-audit/scripts/audit-contract.sh`.
2. **`db-migration-runner`** (`agent-capabilities/db-migration-runner/SKILL.md`):
   - Inspects PostgreSQL database schema, verifies table definitions across all 11 core tables (`households`, `members`, `chores`, `chore_completions`, `chore_swap_requests`, `reward_items`, `reward_redemptions`, `activity_logs`, `chore_comments`, `magic_links`, `password_reset_tokens`), and validates indices.
   - Run via: `bash agent-capabilities/db-migration-runner/scripts/check-migrations.sh`.
3. **`cluster-health-prober`** (`agent-capabilities/cluster-health-prober/SKILL.md`):
   - Probes running multi-tier cluster health across Go backend (`:8000/healthz`), Frontend (`:3000`), Caddy (`:8088`), and Floci cloud emulator (`:4566`).
   - Run via: `bash agent-capabilities/cluster-health-prober/scripts/probe-cluster.sh`.

### B. MCP Tools (`mcp-server/`)
The stdio MCP server in `mcp-server/server.py` exposes 4 project inspection tools via JSON-RPC 2.0:
- **`inspect_choresync_health`**: Probes container status and HTTP `/healthz` endpoints.
- **`inspect_openapi_contract`**: Extracts endpoint paths, operations, and schemas from `contracts/openapi.yaml`.
- **`inspect_db_schema`**: Analyzes table declarations, column lists, and constraints in `backend/internal/db/schema.sql`.
- **`inspect_cloud_emulator`**: Reports S3 bucket (`choresync-proofs`) and SQS queue (`choresync-reminders`) configuration on the Floci emulator.

### C. Specialist Subagent (`custom-agent/specialist.md`)
- **`chore-domain-specialist`**: Specialized agent persona for validating chore business logic (rotation rules, parent approval gates, peer swaps, points balance), OpenAPI sync, and local cloud emulator integration.

### D. Guardrails & Hooks (`agent-hooks/`)
- **`destructive-command-guardrail`** (`pre-tool-guardrail.sh`): Intercepts commands to prevent irreversible data loss or unapproved git overrides.
- **`audit-logger`** (`post-tool-audit.sh`): Records timestamp, tool invocations, and exit codes to `agent-hooks/audit.log`.

---

## 4. Verification & Testing

Verify full compliance of the extension pack:
```bash
verify-extension-pack
# or via Makefile
make ext-verify
```

Test the MCP server:
```bash
make ext-mcp-test
```
