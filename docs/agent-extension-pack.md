# ChoreSync Agent Extension Pack Documentation

This repository packages a production-grade **Agent Extension Pack** conforming strictly to the open **Agent Plugins 1.0/1.1 Standard** ([agent-plugins.org](https://agent-plugins.org/) / AAIF / AWS).

It bundles ChoreSync project specifications, domain lifecycle workflows, specialist subagents, Model Context Protocol (MCP) tools, guardrail hooks, and least-privilege security boundaries into an interoperable extension bundle compatible across modern agentic tools: Claude Code, Cursor, Antigravity/Gemini CLI, Codex, and Copilot.

---

## 1. Canonical Extension Pack Directory Layout

```text
household-chores/
├── plugin.json                                     # 1. The Manifest (closed schema: identity, version, metadata)
├── mcp.json                                        # 2. The Connection Layer (root stdio MCP server declaration)
├── skills/                                         # 3. The Knowledge Layer (domain skills with scripts & references)
│   ├── contract-audit/                             # OpenAPI 3.1 contract drift auditing
│   │   ├── SKILL.md                                # Skill instructions & validation criteria
│   │   ├── scripts/audit-contract.sh               # Automated contract drift verification
│   │   └── references/api-endpoints-reference.md   # Complete 26-endpoint OpenAPI specification
│   ├── db-migration-runner/                        # PostgreSQL schema & entity verification
│   │   ├── SKILL.md                                # Migration invariants & DDL documentation
│   │   ├── scripts/check-migrations.sh             # DDL verification script across 11 tables
│   │   └── references/db-schema-reference.md       # Entity-relationship schema & constraints
│   ├── cluster-health-prober/                      # Multi-tier cluster health inspection
│   │   ├── SKILL.md                                # Service probe targets & ports
│   │   ├── scripts/probe-cluster.sh                # Live health probe script (API, Proxy, DB, Floci)
│   │   └── references/cluster-architecture-reference.md # Network topology & healthcheck matrix
│   └── chore-lifecycle-manager/                    # Household chore domain lifecycle workflows
│       ├── SKILL.md                                # Domain state machine & archetype procedures
│       ├── scripts/verify-chore-flows.sh           # Rotation, approval gate, and cloud hook verification
│       └── references/chore-domain-rules.md        # Flatmates, Families, and Couples business rules
├── com.antigravity.client/                         # 4. Client-Specific Extensions (reverse-domain namespace)
│   └── hooks/                                      # Automated lifecycle hooks & guardrails
│       ├── hooks.json                              # Event bindings (pre-tool, post-tool)
│       ├── pre-tool-guardrail.sh                   # Negative invariant enforcement (blocks drops & host leaks)
│       └── post-tool-audit.sh                      # Telemetry & execution audit logger
├── mcp-server/                                     # 5. MCP Server Implementation
│   ├── mcp.json                                    # Symlinked bridge to root mcp.json
│   └── server.py                                   # Zero-dependency Python 3 JSON-RPC 2.0 stdio server
├── custom-agent/                                   # 6. Specialist Subagents
│   └── specialist.md                               # ChoreSync Architecture & Governance Specialist
└── docs/                                           # 7. Documentation & Governance
    ├── agent-extension-pack.md                     # This architecture guide & operating instructions
    ├── permissions.md                              # Least-privilege policies & security boundaries
    └── agent-extension-pack-evaluation.md          # Conformance evaluation report (Gate 14 deliverable)
```

### Compatibility Bridges
For backward compatibility with legacy tooling and previous milestones:
- `agent-capabilities/` -> symlink to `skills/`
- `agent-hooks/` -> symlink to `com.antigravity.client/hooks/`
- `mcp-server/mcp.json` -> symlink to `../mcp.json`
- `_docs/agent-extension-pack.md` -> symlink to `docs/agent-extension-pack.md`
- `_docs/permissions.md` -> symlink to `docs/permissions.md`
- `_docs/agent-extension-pack-evaluation.md` -> symlink to `docs/agent-extension-pack-evaluation.md`

---

## 2. Project Operating Principles for AI Agents

All AI agents interacting with this repository MUST follow these operating rules:

1. **Closed-Schema Manifest Discovery**:
   - Inspect `plugin.json` at root to identify extension metadata and client namespaces.
2. **Consult Least-Privilege Security Policy**:
   - Review `docs/permissions.md` before executing state-altering actions or issuing commands.
3. **Use Domain Knowledge Capabilities**:
   - Execute workflows declared in `skills/` (`audit-contract.sh`, `check-migrations.sh`, `probe-cluster.sh`, `verify-chore-flows.sh`) rather than ad-hoc host guessing.
4. **Zero Host Runtime Mandate**:
   - Assume zero host runtimes. All tests, migrations, builds, and backend services execute inside containers or via lightweight inspection scripts.
5. **Guardrail Hook Adherence**:
   - All tool executions are validated by `com.antigravity.client/hooks/pre-tool-guardrail.sh`. Unsafe commands (`DROP DATABASE`, `TRUNCATE TABLE`, `git push --force`) are blocked automatically.

---

## 3. Registered Extension Components

### A. Domain Skills (`skills/`)
1. **`contract-audit`** (`skills/contract-audit/SKILL.md`):
   - Audits OpenAPI 3.1 contract endpoints, request/response models, and status codes against Go backend handlers and TypeScript frontend client.
   - Run: `bash skills/contract-audit/scripts/audit-contract.sh`.
2. **`db-migration-runner`** (`skills/db-migration-runner/SKILL.md`):
   - Audits PostgreSQL DDL schema definitions across all 11 core tables (`households`, `members`, `chores`, `chore_completions`, `chore_swap_requests`, `reward_items`, `reward_redemptions`, `activity_logs`, `chore_comments`, `magic_links`, `password_reset_tokens`).
   - Run: `bash skills/db-migration-runner/scripts/check-migrations.sh`.
3. **`cluster-health-prober`** (`skills/cluster-health-prober/SKILL.md`):
   - Probes running multi-tier cluster health across Go backend (`:8000/healthz`), Frontend (`:3000`), Caddy (`:8088`), and Floci cloud emulator (`:4566`).
   - Run: `bash skills/cluster-health-prober/scripts/probe-cluster.sh`.
4. **`chore-lifecycle-manager`** (`skills/chore-lifecycle-manager/SKILL.md`):
   - Validates chore rotation rules, parent approval gates, photo proof uploads via Floci S3, and asynchronous reminder nudges via Floci SQS across all 3 archetypes (Flatmates, Families, Couples).
   - Run: `bash skills/chore-lifecycle-manager/scripts/verify-chore-flows.sh`.

### B. Root MCP Connection Layer (`mcp.json` / `mcpServers`) & Tools (`mcp-server/server.py`)
Root `mcp.json` declares the standard `mcpServers` configuration with explicit transport (`type: "stdio"`). Exposes 5 project inspection tools via JSON-RPC 2.0 stdio:
- **`inspect_choresync_health`**: Probes container status and HTTP `/healthz` endpoints.
- **`inspect_openapi_contract`**: Extracts endpoint paths, operations, and schemas from `contracts/openapi.yaml`.
- **`inspect_db_schema`**: Analyzes table declarations, column lists, and constraints in `backend/internal/db/schema.sql`.
- **`inspect_cloud_emulator`**: Reports S3 bucket (`choresync-proofs`) and SQS queue (`choresync-reminders`) configuration on the Floci emulator.
- **`inspect_chore_operations`**: Inspects active chores, approval gates, and archetype configurations.

### C. Specialist Subagent (`custom-agent/specialist.md`)
- **`chore-domain-specialist`**: Dedicated persona for verifying chore domain rules (rotation rules, parent approval gates, peer swaps, points balance), OpenAPI sync, and local cloud emulator integration.

### D. Reverse-Domain Hooks (`com.antigravity.client/hooks/`)
- **`destructive-command-guardrail`** (`pre-tool-guardrail.sh`): Intercepts commands to prevent irreversible data loss or unapproved git overrides.
- **`audit-logger`** (`post-tool-audit.sh`): Records timestamp, tool invocations, and exit codes to `com.antigravity.client/hooks/audit.log`.

---

## 4. Verification & Testing

Verify full compliance of the extension pack:
```bash
verify-extension-pack
# or via Makefile
make ext-verify
```

Evaluate conformance & domain alignment:
```bash
evaluate-extension-pack
# or via Makefile
make ext-eval
```

Test MCP server initialization and tool listing:
```bash
make ext-mcp-test
```

Run all capability verification scripts:
```bash
make ext-capabilities
```
