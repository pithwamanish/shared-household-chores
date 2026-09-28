# ChoreSync Model Context Protocol (MCP) Server

This directory contains the reference **Model Context Protocol (MCP)** server implementation for ChoreSync, adhering to the open MCP specification and the **Agent Plugins 1.0** standard.

---

## 1. Overview & Transport Protocol

The ChoreSync MCP server is a lightweight, zero-dependency Python 3 implementation running over the standard **stdio** transport using JSON-RPC 2.0.

- **Transport**: `stdio`
- **Location**: `mcp-server/server.py`
- **Root Declaration**: Declared in [`mcp.json`](../mcp.json) at the repository root:
  ```json
  {
    "$schema": "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json",
    "mcpServers": {
      "choresync-tools": {
        "type": "stdio",
        "command": "python3",
        "args": ["mcp-server/server.py"],
        "env": {
          "CHORESYNC_ENV": "development",
          "FLOCI_ENDPOINT": "http://localhost:4566",
          "BACKEND_URL": "http://localhost:8000"
        },
        "cwd": "${PLUGIN_ROOT}"
      }
    }
  }
  ```

---

## 2. Exposed MCP Tools

The server exposes 5 bounded inspection tools designed for AI coding assistants:

| Tool Name | Purpose | Parameters |
| :--- | :--- | :--- |
| **`inspect_choresync_health`** | Probes HTTP health on backend (`:8000`), frontend (`:3000`), and Floci (`:4566`), plus active container states. | `service` (optional: `backend`, `frontend`, `floci`) |
| **`inspect_openapi_contract`** | Parses `contracts/openapi.yaml` and returns structured endpoint paths, operations, parameters, and response schemas. | `tag` (optional filter) |
| **`inspect_db_schema`** | Reads PostgreSQL DDL definitions from `backend/internal/db/schema.sql` and summarizes tables, columns, and foreign keys. | `table` (optional table name) |
| **`inspect_cloud_emulator`** | Inspects Floci local cloud emulator configuration for S3 buckets (`choresync-proofs`) and SQS queues (`choresync-reminders`). | None |
| **`inspect_chore_operations`** | Summarizes ChoreSync business rules across Flatmates, Families, and Couples archetypes, rotation indexes, and approval gates. | `household_id`, `archetype` (optional) |

---

## 3. Running & Verifying the MCP Server

You can verify the MCP server directly using containerized execution without host runtime dependencies:

### Initialize Request
```bash
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}' | \
  docker run --rm -i -v $(pwd):/app -w /app python:3.11-alpine python3 mcp-server/server.py
```

### List Tools Request
```bash
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}' | \
  docker run --rm -i -v $(pwd):/app -w /app python:3.11-alpine python3 mcp-server/server.py
```

### Call Tool Request
```bash
echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "inspect_openapi_contract", "arguments": {"tag": "chores"}}}' | \
  docker run --rm -i -v $(pwd):/app -w /app python:3.11-alpine python3 mcp-server/server.py
```

Or via Makefile target:
```bash
make ext-mcp-test
```

---

## 4. Security & Least-Privilege Design

1. **Read-Only Inspection**: All MCP tools operate strictly in a read-only capacity. They cannot mutate databases, destroy files, or alter code.
2. **Filesystem Containment**: All relative paths are constrained to `${PLUGIN_ROOT}`.
3. **Zero Plaintext Secrets**: No credentials, tokens, or database passwords are hardcoded in the server. Configuration is supplied via environment variables.
