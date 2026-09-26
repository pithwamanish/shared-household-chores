# ChoreSync Domain Rules & Operational Logic Reference

This document formalizes the domain rules, state machines, and business logic governing chore management in ChoreSync across all supported household archetypes.

---

## 1. Household Archetypes & Operational Modes

ChoreSync adapts its workflow dynamically based on the household's living arrangement:

### A. Flatmates / Roommates Mode
- **Core Goal**: Zero domestic friction, equitable chore balance, and accountability.
- **Round-Robin Rotation**: Chores configured with `rotation_type: "round_robin"` automatically advance to the next roommate in `members` upon completion.
- **Peer Swap Marketplace**: Any roommate can propose a task swap (`POST /api/v1/swaps`). The proposer remains responsible until the target roommate accepts (`POST /api/v1/swaps/{id}/accept`).
- **Transparency Feed**: All chore completions, swaps, and nudge reminders are published to `activity_logs`.

### B. Families with Children Mode
- **Core Goal**: Teaching responsibility, parent verification, and gamified incentives.
- **Approval Gate**: Chores configured with `requires_approval: true` enter status `pending_approval` upon completion.
- **Photo Proof Verification**: Children attach a photo proof via Floci S3 (`POST /api/v1/uploads/photo`) before submitting.
- **Gamified Economy**: Parents/Admins sign off on the completion (`POST /api/v1/chores/{id}/approve`), which credits points to the child's balance. Children redeem rewards (`POST /api/v1/rewards/{id}/redeem`) from the household catalog.
- **Non-Negative Balance Invariant**: Points balances cannot be decremented below zero.

### C. Couples / Lightweight Mode
- **Core Goal**: Low-overhead task backlogs and minimal friction.
- **Shared Backlog**: Chores can remain unassigned in a shared pool for voluntary claiming (`POST /api/v1/chores/{id}/claim`).
- **One-Tap Completion**: Instant completion without mandatory photo proof or approval gates.
- **Gentle Nudges**: Either partner can trigger an asynchronous reminder nudge (`POST /api/v1/chores/{id}/nudge`) dispatched through Floci SQS.

---

## 2. Chore State Transition Diagram

```text
               ┌────────────────────────┐
               │        Created         │
               └───────────┬────────────┘
                           │
                           ▼
               ┌────────────────────────┐
        ┌─────►│        Pending         │◄────┐
        │      └───────────┬────────────┘     │
        │                  │                  │
 (Swap Rejected)           │ (Complete)  (Swap Proposed)
        │                  ▼                  │
        │           Requires Approval?        │
        │             ├── YES ──┐             │
        │             └── NO ─┐ │             │
        │                     │ │             │
        │                     │ ▼             │
        │       ┌───────────────────────┐     │
        │       │   Pending Approval    │     │
        │       └───────────┬───────────┘     │
        │                   │                 │
        │                   ▼ (Parent Admin)  │
        │       ┌───────────────────────┐     │
        │       │       Approved        │     │
        │       └───────────┬───────────┘     │
        │                   │                 │
        │                   ▼                 │
        │       ┌───────────────────────┐     │
        │       │       Completed       │     │
        │       │  (Points Credited)    │     │
        │       └───────────┬───────────┘     │
        │                   │                 │
        │     (If Recurring)│                 │
        └───────────────────┴─────────────────┘
```

---

## 3. Asynchronous Cloud Integrations (Floci S3 & SQS)

1. **Photo Proof Storage (AWS S3)**:
   - S3 Bucket: `choresync-proofs`
   - Zero-code-forking AWS Go SDK v2 uploads multipart form data.
   - Max file size: 10MB; supports `image/jpeg`, `image/png`, `image/webp`.
2. **Chore Reminder Queue (AWS SQS)**:
   - SQS Queue: `choresync-reminders`
   - Endpoint: `POST /api/v1/chores/{id}/nudge` returns `202 Accepted` immediately.
   - Background Go worker long-polls messages with 20-second timeout, logs telemetry, and dispatches mock emails without degrading API response latency.
