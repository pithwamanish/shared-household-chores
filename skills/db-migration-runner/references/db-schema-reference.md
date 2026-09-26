# ChoreSync PostgreSQL 16 Relational Schema Reference

This reference documents the relational data models, constraints, and migration invariants defined in `backend/internal/db/schema.sql` and managed via `sqlc` + `pgx/v5`.

---

## 1. Relational Entities (11 Tables)

```text
  ┌─────────────────────────────────────────────────────────────┐
  │                        households                           │
  │  (id, name, invite_code, timezone, created_at, updated_at)  │
  └──────────────────────────────┬──────────────────────────────┘
                                 │ 1:N
        ┌────────────────────────┼────────────────────────┐
        ▼                        ▼                        ▼
  ┌───────────┐            ┌───────────┐            ┌───────────┐
  │  members  │            │  chores   │            │  rewards  │
  └─────┬─────┘            └─────┬─────┘            └─────┬─────┘
        │                        │                        │
        │ 1:N                    │ 1:N                    │ 1:N
        ▼                        ▼                        ▼
  ┌───────────┐            ┌───────────┐            ┌───────────┐
  │activity_  │            │chore_     │            │reward_    │
  │  logs     │            │completions│            │redemptions│
  └───────────┘            └───────────┘            └───────────┘
```

### Table Definitions:

1. **`households`**:
   - `id VARCHAR(64) PRIMARY KEY`, `name VARCHAR(255)`, `invite_code VARCHAR(32) UNIQUE`, `mode VARCHAR(32)` (`flatmates`, `family`, `couples`).
2. **`members`**:
   - `id VARCHAR(64) PRIMARY KEY`, `household_id VARCHAR(64) REFERENCES households(id) ON DELETE CASCADE`, `name VARCHAR(255)`, `email VARCHAR(255)`, `role VARCHAR(32)` (`admin`, `member`, `child`), `avatar_url TEXT`, `points_balance INT DEFAULT 0`, `password_hash TEXT`.
3. **`chores`**:
   - `id VARCHAR(64) PRIMARY KEY`, `household_id VARCHAR(64) REFERENCES households(id) ON DELETE CASCADE`, `title VARCHAR(255)`, `description TEXT`, `points INT DEFAULT 10`, `status VARCHAR(32)` (`pending`, `completed`, `pending_approval`), `assigned_to VARCHAR(64) REFERENCES members(id)`, `recurrence VARCHAR(32)` (`none`, `daily`, `weekly`, `monthly`), `rotation_type VARCHAR(32)` (`none`, `round_robin`), `requires_approval BOOLEAN DEFAULT FALSE`, `due_date TIMESTAMPTZ`.
4. **`chore_completions`**:
   - `id VARCHAR(64) PRIMARY KEY`, `chore_id VARCHAR(64) REFERENCES chores(id) ON DELETE CASCADE`, `completed_by VARCHAR(64) REFERENCES members(id)`, `proof_photo_url TEXT`, `notes TEXT`, `status VARCHAR(32)` (`approved`, `pending_approval`), `approved_by VARCHAR(64)`, `created_at TIMESTAMPTZ`.
5. **`chore_swap_requests`**:
   - `id VARCHAR(64) PRIMARY KEY`, `chore_id VARCHAR(64) REFERENCES chores(id) ON DELETE CASCADE`, `proposer_id VARCHAR(64) REFERENCES members(id)`, `target_id VARCHAR(64) REFERENCES members(id)`, `status VARCHAR(32)` (`pending`, `accepted`, `rejected`), `created_at TIMESTAMPTZ`.
6. **`reward_items`**:
   - `id VARCHAR(64) PRIMARY KEY`, `household_id VARCHAR(64) REFERENCES households(id) ON DELETE CASCADE`, `title VARCHAR(255)`, `description TEXT`, `points_cost INT`, `icon VARCHAR(64)`.
7. **`reward_redemptions`**:
   - `id VARCHAR(64) PRIMARY KEY`, `household_id VARCHAR(64) REFERENCES households(id) ON DELETE CASCADE`, `reward_id VARCHAR(64) REFERENCES reward_items(id)`, `member_id VARCHAR(64) REFERENCES members(id)`, `points_spent INT`, `status VARCHAR(32)`.
8. **`activity_logs`**:
   - `id VARCHAR(64) PRIMARY KEY`, `household_id VARCHAR(64) REFERENCES households(id) ON DELETE CASCADE`, `actor_id VARCHAR(64) REFERENCES members(id)`, `action VARCHAR(64)`, `details JSONB`, `created_at TIMESTAMPTZ`.
9. **`chore_comments`**:
   - `id VARCHAR(64) PRIMARY KEY`, `chore_id VARCHAR(64) REFERENCES chores(id) ON DELETE CASCADE`, `member_id VARCHAR(64) REFERENCES members(id)`, `message TEXT`, `created_at TIMESTAMPTZ`.
10. **`magic_links`**:
    - `token VARCHAR(128) PRIMARY KEY`, `email VARCHAR(255)`, `expires_at TIMESTAMPTZ`, `used BOOLEAN DEFAULT FALSE`.
11. **`password_reset_tokens`**:
    - `token VARCHAR(128) PRIMARY KEY`, `member_id VARCHAR(64) REFERENCES members(id) ON DELETE CASCADE`, `expires_at TIMESTAMPTZ`, `used BOOLEAN DEFAULT FALSE`.

---

## 2. Invariants & Indices

- **Cascading Deletes**: `ON DELETE CASCADE` on all foreign keys referencing `households(id)` ensures complete tenant isolation without orphaned data leaks.
- **Indices**:
  - `CREATE INDEX idx_chores_household ON chores(household_id);`
  - `CREATE INDEX idx_members_household ON members(household_id);`
  - `CREATE INDEX idx_activity_household ON activity_logs(household_id);`
- **Auto-Migrations**: Handled idempotently at Go backend startup via `backend/internal/store/postgres.go` using `CREATE TABLE IF NOT EXISTS`.
