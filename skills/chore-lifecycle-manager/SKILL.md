---
name: chore-lifecycle-manager
description: >-
  Manages, validates, and simulates the complete ChoreSync domain lifecycle across
  Flatmates, Families, and Couples archetypes: round-robin auto-rotation, parent
  approval gates, photo proof uploads via Floci S3, and asynchronous reminder nudges via Floci SQS.
---

# Chore Lifecycle Manager Capability Workflow

Use this capability workflow to execute, verify, and simulate ChoreSync's domain-specific business rules, entity transitions, and cloud integrations.

---

## 1. Domain Operations & Lifecycle States

```text
  [ Created / Pending ] ───────► [ In Progress ]
           │                            │
           │ (Claim / Assign)           │ (Complete)
           ▼                            ▼
  [ Peer Swap Pending ]          Requires Approval?
     │              │               ├── YES ──► [ Pending Approval ] ──► (Admin Approves) ──► [ Completed & Points Credited ]
     ▼ (Accept)     ▼ (Reject)      └── NO  ──► [ Completed & Points Credited ]
  [ Reassigned ]  [ Remains ]
```

### Supported Archetypes:
1. **Flatmates / Roommates**:
   - Round-robin auto-rotation upon chore completion (advances assignee automatically).
   - Peer-to-peer chore swap marketplace with accept/reject semantics.
   - Transparent contribution history and activity log feed.
2. **Families with Children**:
   - Mandatory parent/admin approval gate (`requires_approval: true`).
   - Completion proof attachment (photo proof upload via Floci S3 bucket `choresync-proofs`).
   - Gamified point balances and reward redemption catalog.
3. **Couples / Lightweight Co-living**:
   - Voluntary chore claiming from shared backlog.
   - 1-tap instant completions.
   - Gentle asynchronous reminder nudges queued to Floci SQS `choresync-reminders`.

---

## 2. Automated Verification Script

Run the automated chore domain verification script:

```bash
bash skills/chore-lifecycle-manager/scripts/verify-chore-flows.sh
# or legacy path
bash agent-capabilities/chore-lifecycle-manager/scripts/verify-chore-flows.sh
```

---

## 3. Key Invariants

1. **Non-Negative Points**: Member points balance cannot drop below zero upon reward redemption.
2. **Approval Gate Isolation**: Points are never credited to a member until the parent/admin approves a completion with `requires_approval: true`.
3. **Swap Responsibility**: When a swap is proposed, current assignee responsibility remains unchanged until the recipient formally accepts.
4. **Cloud Decoupling**: SQS reminder messages trigger asynchronous background worker dispatch without blocking HTTP response latency.
