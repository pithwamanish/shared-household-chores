# Extension Pack Security & Permissions Policy

This document defines the security boundaries, authorization tiers, and permission constraints for AI agents interacting with the ChoreSync project.

---

## 1. Principle of Least Privilege

Autonomous AI agents and subagents operate under **strict read-preferred, bounded-write** constraints:
- **Default Stance**: Read-only observation, schema inspection, and bounded analysis.
- **Write Actions**: Strictly confined to files inside the repository root (`household-chores/`).
- **Execution**: Restricted to containerized environments (`docker compose`, `docker run`) or allowlisted inspection commands. Host runtime modification or direct execution of non-sandboxed tools is prohibited.

---

## 2. Permission Tiers

| Tier | Scope | Permissions Granted | Authorization Requirement |
|:---|:---|:---|:---|
| **Tier 1 (Read-Only)** | File inspection, git log, OTel telemetry, metrics, MCP tool queries | Unrestricted within project bounds | Fully Autonomous |
| **Tier 2 (Safe Non-Destructive)** | Test execution, container startup (`docker compose up`), local build | Container execution permitted | Autonomous with Audit Log |
| **Tier 3 (Modifying Source)** | Code edits, test creation, configuration updates | Scoped to project files | Bounded Autonomy (Git-tracked) |
| **Tier 4 (Privileged / Destructive)** | Database schema dropping, hard resets, force pushing | **STRICTLY BLOCKED** | Requires Explicit Human Approval |

---

## 3. Allowlisted & Forbidden Commands

### Allowlisted Commands
- `docker compose ps` / `docker compose logs` / `docker compose up -d` / `docker compose down`
- `docker compose run --rm <service> <test_cmd>`
- `git status` / `git diff` / `git log`
- `curl -s http://localhost:<PORT>/healthz`
- `python3 mcp-server/server.py`
- `make ext-verify` / `make ext-mcp-test`
- `make k8s-verify` / `make oncall-verify`

### Forbidden / Blocked Commands (Enforced by `agent-hooks/pre-tool-guardrail.sh`)
- Destructive filesystem deletions: `rm -rf /` or `rm -rf ~`
- Unparameterized database drops: `DROP DATABASE`, `TRUNCATE TABLE`, `DELETE FROM * WHERE 1=1`
- Git overrides: `git push --force` or `git push -f` to protected branches (`main`, `master`)
- Arbitrary secret exfiltration: direct dumping of environment files containing production secrets

---

## 4. MCP Server Security & Tool Permissions

All Model Context Protocol tools declared in `mcp-server/mcp.json` operate with read-only inspection scopes:
1. **`inspect_choresync_health`**: Probes local HTTP health endpoints and Docker container state; does not mutate containers.
2. **`inspect_openapi_contract`**: Reads `contracts/openapi.yaml`; does not write or modify contracts unilaterally.
3. **`inspect_db_schema`**: Parses `backend/internal/db/schema.sql` statically; does not issue live write queries or alter database state.
4. **`inspect_cloud_emulator`**: Returns Floci S3 and SQS configuration metadata without writing to production cloud credentials.

---

## 5. Audit Trail & Compliance

Every tool invocation and guardrail outcome is logged to `agent-hooks/audit.log` via `agent-hooks/post-tool-audit.sh` with timestamp, tool name, and exit status.
