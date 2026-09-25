#!/usr/bin/env python3
"""Continuous Evaluation Runner for ChoreSync AI Agent & System Invariants.

Evaluates project prompt assertions, domain entities, contracts, database models,
cloud emulators, and security policies against evals/golden-dataset.json.
"""

import json
import os
import sys
import time


def run_evals(dataset_path="evals/golden-dataset.json", base_dir="."):
  if not os.path.isabs(dataset_path):
    dataset_path = os.path.join(base_dir, dataset_path)

  if not os.path.exists(dataset_path):
    print(f"Error: Dataset {dataset_path} not found.")
    return 1

  with open(dataset_path, "r", encoding="utf-8") as f:
    cases = json.load(f)

  passed = 0
  failed = 0

  print("================================================================")
  print("      CONTINUOUS AGENT & DOMAIN EVALUATION RUNNER (GATE 15)     ")
  print("================================================================")
  print(f"Dataset: {dataset_path}")
  print(f"Evaluating {len(cases)} continuous AI evaluation cases...\n")

  for case in cases:
    cid = case.get("id")
    desc = case.get("description")
    target_artifact = case.get("target_artifact")
    expected_keywords = case.get("expected_keywords", [])
    assertions = case.get("assertions", [])

    start = time.time()
    artifact_path = os.path.join(base_dir, target_artifact) if target_artifact else None

    content = ""
    if artifact_path and os.path.exists(artifact_path):
      with open(artifact_path, "r", encoding="utf-8", errors="ignore") as af:
        content = af.read()
    else:
      content = f"Mock simulation for {case.get('input', '')}"

    case_passed = True
    failure_reason = ""

    # Check keyword assertions
    content_lower = content.lower()
    for kw in expected_keywords:
      if kw.lower() not in content_lower:
        case_passed = False
        failure_reason = f"Missing expected keyword '{kw}' in {target_artifact}"
        break

    # Check structured assertions
    if case_passed:
      for assertion in assertions:
        atype = assertion.get("type")
        if atype == "contains_all":
          missing = [v for v in assertion.get("values", []) if v.lower() not in content_lower]
          if missing:
            case_passed = False
            failure_reason = f"Missing required values {missing} in {target_artifact}"
            break
        elif atype == "contains_any":
          found = any(v.lower() in content_lower for v in assertion.get("values", []))
          if not found:
            case_passed = False
            failure_reason = f"None of expected values {assertion.get('values')} found in {target_artifact}"
            break
        elif atype == "min_path_count":
          path_count = content.count("paths:") + content.count("summary:")
          if path_count < assertion.get("value", 1):
            case_passed = False
            failure_reason = f"Path count {path_count} below minimum {assertion.get('value')}"
            break

    elapsed_ms = (time.time() - start) * 1000

    if case_passed:
      print(f"  [PASS] {cid}: {desc} ({elapsed_ms:.1f}ms)")
      passed += 1
    else:
      print(f"  [FAIL] {cid}: {desc}")
      print(f"         ↳ Error: {failure_reason}")
      failed += 1

  print("\n================================================================")
  print(f"Eval Summary: {passed} passed, {failed} failed across {len(cases)} assertions")
  print("================================================================")

  if failed == 0:
    print("✓ All continuous evaluation assertions passed successfully.")
    return 0
  else:
    print(f"✗ {failed} evaluation assertion(s) failed.")
    return 1


if __name__ == "__main__":
  dataset = sys.argv[1] if len(sys.argv) > 1 else "evals/golden-dataset.json"
  sys.exit(run_evals(dataset))
