---
name: db-migration-runner
description: >-
  Inspects PostgreSQL relational schema, verifies table definitions and foreign keys,
  and validates migration status across ChoreSync entities.
---

# Database Migration & Schema Runner Capability

Use this capability workflow to verify PostgreSQL relational schema definitions, check table integrity, and inspect migration status.

---

## 1. Automated Schema Verification

Run the automated schema and migration inspection script:

```bash
bash agent-capabilities/db-migration-runner/scripts/check-migrations.sh
```

## 2. Monitored Entities

The ChoreSync database includes the following relational entities:
- `households` (Core multi-tenant boundary)
- `members` (Household occupants with roles and points)
- `chores` (Tasks, recurrence rules, approval status, assignees)
- `swaps` (Peer task exchange proposals and status)
- `rewards` (Gamified redemption catalog)
- `redemptions` (Claimed rewards history)
- `activity_logs` (Audit timeline of chore events)
- `password_reset_tokens` (Auth reset tokens with expiration)
- `magic_link_tokens` (Single-use 15m expiration auth tokens)
- `proof_photos` / upload metadata

---

## 3. Migration Invariants

1. **Idempotence**: Auto-migration logic in Go backend (`backend/internal/store/postgres.go`) creates tables `IF NOT EXISTS`.
2. **Type Safety**: Queries are generated via `sqlc` with strict type mappings in `backend/internal/db/`.
3. **Data Integrity**: Cascading deletes on household boundaries prevent orphaned chores or members.
