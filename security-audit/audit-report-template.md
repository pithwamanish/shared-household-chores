# Security Audit & Static Analysis Report

## 1. Executive Summary
- **Target Repository**: `{{PROJECT_NAME}}`
- **Audit Date**: `{{TIMESTAMP}}`
- **Auditor Engine**: Semgrep (Deterministic SAST) paired with Headless Agent Review
- **Overall Status**: [PASS / REQUIRE_REMEDIATION / CRITICAL_FAIL]

---

## 2. Deterministic Static Security Scan (Semgrep)
- **Scanner Image**: `returntocorp/semgrep:latest`
- **Rule Packs**: `semgrep-rules.yml` + `p/security-audit`
- **Total Files Scanned**: `{{FILES_SCANNED}}`
- **Findings Summary**:
  - Critical / Errors: `{{ERROR_COUNT}}`
  - Warnings: `{{WARNING_COUNT}}`

### Detailed Scanner Findings
| Rule ID | File Path | Line | Severity | Description |
| :--- | :--- | :--- | :--- | :--- |
| `{{RULE_ID}}` | `{{FILE}}` | `{{LINE}}` | `{{SEVERITY}}` | `{{DESCRIPTION}}` |

---

## 3. Headless AI Model Security Review
- **Reviewer Agent Model**: `{{MODEL_NAME}}`
- **Scope**: Business logic vulnerabilities, authentication/authorization flows, secret propagation.
- **Model Evaluation**:
  - [ ] No hardcoded production credentials detected
  - [ ] API endpoints enforce authentication and rate limiting
  - [ ] Database queries are parameterized
  - [ ] CORS policies restrict untrusted origins

---

## 4. Responder Capabilities & Credential Inventory
Inventory of permissions granted to the automated on-call responder:
- **Permissions**: Read-only allowlisted queries (Prometheus, Loki, Git log)
- **Database Access**: ZERO direct database credentials
- **Production Secrets**: Inaccessible to AI model
- **Action Execution**: Strictly mediated by `on-call-engineer/autonomy-policy.json`

---

## 5. Human Validation & Remediation Trail
- **Human Reviewer**: `{{REVIEWER_NAME}}`
- **Sign-off Date**: `{{SIGNOFF_DATE}}`
- **Action Items**:
  - [ ] Fix identified Semgrep warnings
  - [ ] Ensure all API routes are covered by authentication middleware
