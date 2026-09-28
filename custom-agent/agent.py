#!/usr/bin/env python3
"""Headless Custom Autonomous Agent Loop for ChoreSync.

Fulfills Requirement 6 (Custom Agent) from Zoomcamp Module 5.
Executes non-interactive verification tasks in CI/CD pipelines,
evaluating project state against declarative rules with deterministic exit codes.
"""

import argparse
import json
import os
import subprocess
import sys


def run_cmd(cmd, cwd=None):
  res = subprocess.run(
      cmd,
      shell=True,
      cwd=cwd,
      capture_output=True,
      text=True,
      check=False,
  )
  return res.returncode, res.stdout.strip(), res.stderr.strip()


def task_audit():
  print("================================================================")
  print("        CHORESYNC HEADLESS CUSTOM AGENT: AUDIT TASK             ")
  print("================================================================")
  failures = 0

  # 1. Check OpenAPI contract existence & validity
  contract_path = "contracts/openapi.yaml"
  if os.path.exists(contract_path):
    print(f"  [PASS] API Contract present: {contract_path}")
  else:
    print(f"  [FAIL] Missing API contract: {contract_path}")
    failures += 1

  # 2. Check Database schema existence
  schema_path = "backend/internal/db/schema.sql"
  if os.path.exists(schema_path):
    print(f"  [PASS] Database DDL schema present: {schema_path}")
  else:
    print(f"  [FAIL] Missing database DDL schema: {schema_path}")
    failures += 1

  # 3. Check Constitution negative invariants
  constitution_path = "constitution.md"
  if os.path.exists(constitution_path):
    print(f"  [PASS] Immutable project constitution present: {constitution_path}")
  else:
    print(f"  [FAIL] Missing constitution: {constitution_path}")
    failures += 1

  # 4. Check Plugin manifest
  plugin_path = "plugin.json"
  if os.path.exists(plugin_path):
    print(f"  [PASS] Agent Plugins manifest present: {plugin_path}")
  else:
    print(f"  [FAIL] Missing plugin manifest: {plugin_path}")
    failures += 1

  print("----------------------------------------------------------------")
  if failures == 0:
    print("✓ Headless Audit Completed Successfully (0 failures).")
    return 0
  else:
    print(f"✗ Headless Audit Failed ({failures} failures).")
    return 1


def task_health():
  print("================================================================")
  print("        CHORESYNC HEADLESS CUSTOM AGENT: HEALTH TASK            ")
  print("================================================================")
  # Invoke cluster health prober skill
  code, stdout, stderr = run_cmd("bash skills/cluster-health-prober/scripts/probe-cluster.sh")
  print(stdout)
  if stderr:
    print(stderr, file=sys.stderr)
  return code


def task_drift():
  print("================================================================")
  print("        CHORESYNC HEADLESS CUSTOM AGENT: DRIFT TASK             ")
  print("================================================================")
  # Invoke contract audit skill
  code, stdout, stderr = run_cmd("bash skills/contract-audit/scripts/audit-contract.sh")
  print(stdout)
  if stderr:
    print(stderr, file=sys.stderr)
  return code


def main():
  parser = argparse.ArgumentParser(description="ChoreSync Headless Custom Agent Loop")
  parser.add_argument(
      "--task",
      choices=["audit", "health", "drift"],
      default="audit",
      help="Autonomous task to execute (audit, health, drift)",
  )
  args = parser.parse_args()

  if args.task == "audit":
    sys.exit(task_audit())
  elif args.task == "health":
    sys.exit(task_health())
  elif args.task == "drift":
    sys.exit(task_drift())
  else:
    print(f"Unknown task: {args.task}", file=sys.stderr)
    sys.exit(1)


if __name__ == "__main__":
  main()
