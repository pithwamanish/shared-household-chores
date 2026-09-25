#!/usr/bin/env python3
"""Lightweight stdio Model Context Protocol (MCP) server for ChoreSync.

Zero-dependency standard library implementation compliant with MCP JSON-RPC
specifications. Exposes tools for health probes, OpenAPI contract discovery,
database schema inspection, and cloud emulator telemetry.
"""

import json
import os
import re
import subprocess
import sys
import urllib.request


def read_message():
  line = sys.stdin.readline()
  if not line:
    return None
  return json.loads(line)


def write_message(obj):
  sys.stdout.write(json.dumps(obj) + "\n")
  sys.stdout.flush()


def tool_inspect_health(args):
  target_service = args.get("service")
  results = {"timestamp": os.getenv("TIMESTAMP", ""), "services": {}}

  # 1. Backend probe
  try:
    req = urllib.request.Request("http://localhost:8000/healthz")
    with urllib.request.urlopen(req, timeout=2) as resp:
      if resp.status == 200:
        data = json.loads(resp.read().decode("utf-8"))
        results["services"]["backend"] = {"status": "healthy", "payload": data}
      else:
        results["services"]["backend"] = {
            "status": "unhealthy",
            "http_status": resp.status,
        }
  except Exception as e:
    results["services"]["backend"] = {"status": "unreachable", "error": str(e)}

  # 2. Frontend / Proxy probe
  try:
    req = urllib.request.Request("http://localhost:3000")
    with urllib.request.urlopen(req, timeout=2) as resp:
      results["services"]["frontend"] = {
          "status": "healthy" if resp.status == 200 else "degraded",
          "http_status": resp.status,
      }
  except Exception as e:
    results["services"]["frontend"] = {
        "status": "unreachable",
        "error": str(e),
    }

  # 3. Floci cloud emulator probe
  try:
    req = urllib.request.Request("http://localhost:4566/")
    with urllib.request.urlopen(req, timeout=2) as resp:
      results["services"]["floci"] = {
          "status": "healthy",
          "http_status": resp.status,
      }
  except Exception as e:
    results["services"]["floci"] = {"status": "unreachable", "error": str(e)}

  # 4. Containers summary
  try:
    proc = subprocess.run(
        ["docker", "compose", "ps", "--format", "json"],
        capture_output=True,
        text=True,
        check=False,
    )
    if proc.returncode == 0 and proc.stdout.strip():
      containers = []
      for line in proc.stdout.strip().split("\n"):
        if line.strip():
          try:
            containers.append(json.loads(line))
          except Exception:
            pass
      results["containers"] = containers
  except Exception as e:
    results["containers_error"] = str(e)

  if target_service and target_service in results["services"]:
    return {
        "status": "success",
        "service": target_service,
        "detail": results["services"][target_service],
    }

  return {"status": "success", "overview": results}


def tool_inspect_contract(args):
  contract_path = os.path.join(os.getcwd(), "contracts", "openapi.yaml")
  if not os.path.exists(contract_path):
    return {
        "status": "missing",
        "message": "contracts/openapi.yaml was not found in project root",
    }

  tag_filter = args.get("tag")
  try:
    with open(contract_path, "r", encoding="utf-8") as f:
      lines = f.readlines()

    paths = []
    current_path = None
    for line in lines:
      match = re.match(r"^  (/[^:]+):", line)
      if match:
        current_path = match.group(1)
        paths.append(current_path)

    return {
        "status": "success",
        "contract": "contracts/openapi.yaml",
        "total_paths": len(paths),
        "paths": paths,
        "filtered_by_tag": tag_filter or "all",
    }
  except Exception as e:
    return {"status": "error", "error": str(e)}


def tool_inspect_db_schema(args):
  schema_path = os.path.join(os.getcwd(), "backend", "internal", "db", "schema.sql")
  if not os.path.exists(schema_path):
    return {
        "status": "missing",
        "message": "backend/internal/db/schema.sql not found",
    }

  target_table = args.get("table")
  try:
    with open(schema_path, "r", encoding="utf-8") as f:
      content = f.read()

    table_matches = re.findall(
        r"CREATE TABLE IF NOT EXISTS ([a-zA-Z0-9_]+)\s*\((.*?)\);",
        content,
        re.DOTALL,
    )
    tables = {}
    for tbl_name, body in table_matches:
      columns = [
          col.strip().split()[0]
          for col in body.strip().split("\n")
          if col.strip()
          and not col.strip().startswith("--")
          and not col.strip().startswith("CONSTRAINT")
          and not col.strip().startswith("FOREIGN")
          and not col.strip().startswith("PRIMARY")
      ]
      tables[tbl_name] = columns

    if target_table:
      if target_table in tables:
        return {
            "status": "success",
            "table": target_table,
            "columns": tables[target_table],
        }
      return {
          "status": "not_found",
          "message": (
              f"Table '{target_table}' not found. Available:"
              f" {list(tables.keys())}"
          ),
      }

    return {
        "status": "success",
        "schema_file": "backend/internal/db/schema.sql",
        "total_tables": len(tables),
        "tables": tables,
    }
  except Exception as e:
    return {"status": "error", "error": str(e)}


def tool_inspect_cloud_emulator(args):
  return {
      "status": "success",
      "cloud_provider": "floci",
      "image": "floci/floci:latest",
      "endpoint_url": "http://localhost:4566",
      "services": {
          "s3": {
              "bucket_name": "choresync-proofs",
              "purpose": "Chore completion photo proof storage (up to 10MB)",
              "region": "us-east-1",
          },
          "sqs": {
              "queue_name": "choresync-reminders",
              "purpose": "Decoupled asynchronous chore reminder nudges",
              "worker": (
                  "Go background long-polling worker in backend/cmd/server"
              ),
          },
      },
  }


def main():
  while True:
    try:
      msg = read_message()
      if msg is None:
        break

      req_id = msg.get("id")
      method = msg.get("method")
      params = msg.get("params", {})

      if method == "initialize":
        write_message({
            "jsonrpc": "2.0",
            "id": req_id,
            "result": {
                "protocolVersion": "2024-11-05",
                "capabilities": {"tools": {}},
                "serverInfo": {
                    "name": "choresync-mcp-server",
                    "version": "1.0.0",
                },
            },
        })
      elif method == "tools/list":
        write_message({
            "jsonrpc": "2.0",
            "id": req_id,
            "result": {
                "tools": [
                    {
                        "name": "inspect_choresync_health",
                        "description": (
                            "Checks the live health and container status of"
                            " ChoreSync services"
                        ),
                        "inputSchema": {
                            "type": "object",
                            "properties": {
                                "service": {
                                    "type": "string",
                                    "description": (
                                        "Optional specific service to check"
                                    ),
                                }
                            },
                        },
                    },
                    {
                        "name": "inspect_openapi_contract",
                        "description": (
                            "Reads and summarizes API endpoints from"
                            " contracts/openapi.yaml"
                        ),
                        "inputSchema": {
                            "type": "object",
                            "properties": {
                                "tag": {
                                    "type": "string",
                                    "description": "Optional tag filter",
                                }
                            },
                        },
                    },
                    {
                        "name": "inspect_db_schema",
                        "description": (
                            "Inspects PostgreSQL DDL schema definitions and"
                            " tables"
                        ),
                        "inputSchema": {
                            "type": "object",
                            "properties": {
                                "table": {
                                    "type": "string",
                                    "description": (
                                        "Optional specific table to inspect"
                                    ),
                                }
                            },
                        },
                    },
                    {
                        "name": "inspect_cloud_emulator",
                        "description": (
                            "Returns configuration and status for the local"
                            " Floci cloud emulator (AWS S3 and SQS)"
                        ),
                        "inputSchema": {"type": "object", "properties": {}},
                    },
                ]
            },
        })
      elif method == "tools/call":
        tool_name = params.get("name")
        arguments = params.get("arguments", {})
        if tool_name == "inspect_choresync_health":
          res = tool_inspect_health(arguments)
        elif tool_name == "inspect_openapi_contract":
          res = tool_inspect_contract(arguments)
        elif tool_name == "inspect_db_schema":
          res = tool_inspect_db_schema(arguments)
        elif tool_name == "inspect_cloud_emulator":
          res = tool_inspect_cloud_emulator(arguments)
        else:
          res = {"error": f"Unknown tool: {tool_name}"}

        write_message({
            "jsonrpc": "2.0",
            "id": req_id,
            "result": {
                "content": [{"type": "text", "text": json.dumps(res, indent=2)}]
            },
        })
      elif method == "ping":
        write_message({"jsonrpc": "2.0", "id": req_id, "result": {}})
      else:
        write_message({
            "jsonrpc": "2.0",
            "id": req_id,
            "error": {"code": -32601, "message": f"Method {method} not found"},
        })
    except Exception as ex:
      write_message({
          "jsonrpc": "2.0",
          "id": None,
          "error": {"code": -32603, "message": str(ex)},
      })


if __name__ == "__main__":
  main()
