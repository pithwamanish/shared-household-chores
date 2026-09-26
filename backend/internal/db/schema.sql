-- ChoreSync PostgreSQL Database Schema (Step 6/7 Persistence Migration)

CREATE TABLE IF NOT EXISTS households (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    invite_code TEXT UNIQUE NOT NULL,
    default_mode TEXT NOT NULL DEFAULT 'flatmate',
    allow_swaps BOOLEAN NOT NULL DEFAULT true,
    timezone TEXT NOT NULL DEFAULT 'America/New_York',
    admin_pin TEXT NOT NULL DEFAULT '1234',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS members (
    id TEXT PRIMARY KEY,
    household_id TEXT NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    email TEXT,
    password_hash TEXT NOT NULL DEFAULT '',
    avatar_url TEXT,
    role TEXT NOT NULL DEFAULT 'member',
    points_balance INTEGER NOT NULL DEFAULT 0,
    total_points_earned INTEGER NOT NULL DEFAULT 0,
    streak INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE members ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS chores (
    id TEXT PRIMARY KEY,
    household_id TEXT NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL,
    effort_points INTEGER NOT NULL DEFAULT 10,
    assignment_type TEXT NOT NULL DEFAULT 'direct',
    current_assignee_id TEXT REFERENCES members(id) ON DELETE SET NULL,
    rotation_member_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    current_rotation_index INTEGER,
    recurrence_type TEXT NOT NULL DEFAULT 'none',
    recurrence_rule JSONB,
    due_date TIMESTAMPTZ,
    requires_approval BOOLEAN NOT NULL DEFAULT false,
    requires_proof BOOLEAN NOT NULL DEFAULT false,
    status TEXT NOT NULL DEFAULT 'unassigned',
    created_by TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chore_completions (
    id TEXT PRIMARY KEY,
    chore_id TEXT NOT NULL REFERENCES chores(id) ON DELETE CASCADE,
    household_id TEXT NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    member_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending_approval',
    proof_notes TEXT,
    proof_photo_url TEXT,
    approved_by TEXT REFERENCES members(id) ON DELETE SET NULL,
    rejection_reason TEXT,
    points_awarded INTEGER NOT NULL DEFAULT 0,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS chore_swap_requests (
    id TEXT PRIMARY KEY,
    household_id TEXT NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    chore_id TEXT NOT NULL REFERENCES chores(id) ON DELETE CASCADE,
    requester_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    target_member_id TEXT REFERENCES members(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS reward_items (
    id TEXT PRIMARY KEY,
    household_id TEXT NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    points_cost INTEGER NOT NULL DEFAULT 50,
    icon TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reward_redemptions (
    id TEXT PRIMARY KEY,
    household_id TEXT NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    member_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    reward_item_id TEXT NOT NULL REFERENCES reward_items(id) ON DELETE CASCADE,
    points_spent INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'requested',
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    fulfilled_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS activity_logs (
    id TEXT PRIMARY KEY,
    household_id TEXT NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    recipient_id TEXT REFERENCES members(id) ON DELETE SET NULL,
    actor_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chore_comments (
    id TEXT PRIMARY KEY,
    chore_id TEXT NOT NULL REFERENCES chores(id) ON DELETE CASCADE,
    member_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS magic_links (
    token TEXT PRIMARY KEY,
    member_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '15 minutes'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    token TEXT PRIMARY KEY,
    member_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reminder_jobs (
    id TEXT PRIMARY KEY,
    chore_id TEXT NOT NULL,
    chore_title TEXT NOT NULL,
    due_date TEXT NOT NULL DEFAULT '',
    assignee_email TEXT NOT NULL DEFAULT '',
    assignee_name TEXT NOT NULL DEFAULT '',
    sender_name TEXT NOT NULL DEFAULT '',
    household_name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    enqueued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indices for rapid lookup & foreign key query paths
CREATE INDEX IF NOT EXISTS idx_members_household_id ON members(household_id);
CREATE INDEX IF NOT EXISTS idx_chores_household_id ON chores(household_id);
CREATE INDEX IF NOT EXISTS idx_chores_current_assignee_id ON chores(current_assignee_id);
CREATE INDEX IF NOT EXISTS idx_chore_completions_household_id ON chore_completions(household_id);
CREATE INDEX IF NOT EXISTS idx_chore_completions_chore_id ON chore_completions(chore_id);
CREATE INDEX IF NOT EXISTS idx_chore_swap_requests_household_id ON chore_swap_requests(household_id);
CREATE INDEX IF NOT EXISTS idx_reward_items_household_id ON reward_items(household_id);
CREATE INDEX IF NOT EXISTS idx_reward_redemptions_household_id ON reward_redemptions(household_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_household_id ON activity_logs(household_id);
CREATE INDEX IF NOT EXISTS idx_chore_comments_chore_id ON chore_comments(chore_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_member_id ON password_reset_tokens(member_id);
CREATE INDEX IF NOT EXISTS idx_reminder_jobs_status_enqueued ON reminder_jobs(status, enqueued_at);

