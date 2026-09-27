# Security Audit Brief & Threat Model (`audit-brief.md`)

**Target System**: ChoreSync (`household-chores`)  
**Scope**: Full-Stack Application (Go backend, React TypeScript frontend, Python tooling, Docker/Kubernetes infrastructure, AI Agent Extension Pack, and On-Call Responders)  
**Governance Standard**: AI-Dev-Tools Zoomcamp Module 4 (DevOps, Observability & Security)  

---

## 1. Audit Objectives & Core Principles

Modern software delivery pairs automated CI/CD with autonomous AI agents. This introduces two distinct security boundaries:
1. **Application Security (Deterministic SAST)**: Ensuring the codebase has zero OWASP Top 10 vulnerabilities (zero SQL injections, zero hardcoded production credentials, zero unvalidated shell commands).
2. **AI Responder Security (Attack Surface Management)**: Treating the automated incident responder as attack surface. Ensuring the AI responder operates under strict least-privilege, has zero write access to production databases, and is constrained by code-enforced autonomy policies outside the model.

**The Operating Principle**:
> *The model may reason; the system must observe, authorize, verify, and remember.*

---

## 2. Methodology & Multi-Layer Defense

Our security auditing process integrates three complementary layers:

```
┌─────────────────────────────────┐
│ 1. Deterministic Scanner        │  Containerized Semgrep SAST
│    (Semgrep inside Docker)      │  Scans 80+ files against OWASP Top 10 & secret rules
└──────────────┬──────────────────┘  Zero host runtimes required
               │
               ▼
┌─────────────────────────────────┐
│ 2. AI Model Review              │  Headless model security audit
│    (PR & Incident Boundary)     │  Evaluates logic flaws, auth bypass, and data flow
└──────────────┬──────────────────┘  Model proposals treated as untrusted suggestions
               │
               ▼
┌─────────────────────────────────┐
│ 3. Human Validation &           │  Human security engineer reviews scanner findings,
│    Disposition Workflow         │  signs off on capability tables, and assigns disposition
└─────────────────────────────────┘  (remediated, false_positive, accepted_risk)
```

---

## 3. Threat Model & Audit Boundaries

### A. OWASP Top 10 Application Threats
- **A01: Broken Access Control**: Verify multi-tenant isolation, household ownership guards, and admin verification PIN checks.
- **A02: Cryptographic Failures**: Verify bcrypt password hashing, HMAC-SHA256 JWT signing, and 15-minute expiration on magic link tokens.
- **A03: Injection**:
  - *SQL Injection*: Enforce type-safe parameterized queries via `sqlc` with `pgx/v5`. Strictly forbid `fmt.Sprintf` or string concatenation in SQL queries.
  - *Command Injection*: Ensure background workers and scripts invoke deterministic binary paths with arguments, avoiding `sh -c` or `os.system` interpolation.
- **A05: Security Misconfiguration**: Prohibit default passwords in production containers; enforce Caddy security headers (CSP, HSTS, X-Content-Type-Options).
- **A07: Identification and Authentication Failures**: Validate single-use invalidation of magic links upon login.

### B. AI Coding Agent & Responder Attack Surface (Snyk Agent Scan Alignment)
- **Prompt Injection & Indirect Manipulation**: An attacker injecting malicious payloads via chore titles or log strings to manipulate the on-call model into executing destructive actions.
  - *Defense*: The responder only consumes bounded, structured evidence JSON; actions are strictly evaluated against `autonomy-policy.yaml` with hard-coded forbidden pattern regexes.
- **Unauthorized Privilege Escalation**: A model attempting to modify production database tables or cluster namespaces.
  - *Defense*: Responders have **ZERO** direct database write credentials and zero root shell access.
- **Unverified Code Fixes**: An agent deploying speculative patches directly to production.
  - *Defense*: Level 1 is restricted to reversible operational actions (rollback, service restart). Any code patch requires Level 2 human approval.

---

## 4. Acceptance Criteria & Gate Passing Conditions

1. **Zero Critical / Error Findings**: Semgrep scan must report 0 errors.
2. **100% Finding Disposition**: All findings recorded in `security-audit/runs/` must have a valid human disposition (`remediated`, `false_positive`, or `accepted_risk`).
3. **Capability Inventory Maintenance**: [`capability-table.md`](capability-table.md) must be kept up to date for all active agent tools and responder scripts.
4. **Zero Host Tools Dependency**: All scans and audits must execute inside official container images (`returntocorp/semgrep:latest`).
