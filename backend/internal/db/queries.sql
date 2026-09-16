-- name: GetHouseholds :many
SELECT * FROM households ORDER BY created_at ASC;

-- name: GetHousehold :one
SELECT * FROM households WHERE id = $1 LIMIT 1;

-- name: GetHouseholdByInviteCode :one
SELECT * FROM households WHERE invite_code = $1 LIMIT 1;

-- name: CreateHousehold :one
INSERT INTO households (
    id, name, invite_code, default_mode, allow_swaps, timezone, admin_pin, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: UpdateHouseholdSettings :one
UPDATE households
SET default_mode = $2, allow_swaps = $3, timezone = $4, admin_pin = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetMembersByHousehold :many
SELECT * FROM members WHERE household_id = $1 ORDER BY created_at ASC;

-- name: GetMember :one
SELECT * FROM members WHERE id = $1 LIMIT 1;

-- name: GetMemberByEmail :one
SELECT * FROM members WHERE email = $1 LIMIT 1;

-- name: CreateMember :one
INSERT INTO members (
    id, household_id, name, email, avatar_url, role, points_balance, total_points_earned, streak, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: UpdateMemberStats :one
UPDATE members
SET points_balance = $2, total_points_earned = $3, streak = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetChoresByHousehold :many
SELECT * FROM chores WHERE household_id = $1 ORDER BY created_at DESC;

-- name: GetChore :one
SELECT * FROM chores WHERE id = $1 LIMIT 1;

-- name: CreateChore :one
INSERT INTO chores (
    id, household_id, title, description, category, effort_points, assignment_type,
    current_assignee_id, rotation_member_ids, current_rotation_index, recurrence_type,
    recurrence_rule, due_date, requires_approval, requires_proof, status, created_by,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
) RETURNING *;

-- name: UpdateChore :one
UPDATE chores SET
    title = $2,
    description = $3,
    category = $4,
    effort_points = $5,
    assignment_type = $6,
    current_assignee_id = $7,
    rotation_member_ids = $8,
    current_rotation_index = $9,
    recurrence_type = $10,
    recurrence_rule = $11,
    due_date = $12,
    requires_approval = $13,
    requires_proof = $14,
    status = $15,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteChore :exec
DELETE FROM chores WHERE id = $1;

-- name: UpdateChoreAssigneeAndStatus :one
UPDATE chores
SET current_assignee_id = $2, status = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateChoreRotation :one
UPDATE chores
SET current_assignee_id = $2, current_rotation_index = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetCompletion :one
SELECT * FROM chore_completions WHERE id = $1 LIMIT 1;

-- name: GetPendingApprovalsByHousehold :many
SELECT * FROM chore_completions
WHERE household_id = $1 AND status = 'pending_approval'
ORDER BY completed_at DESC;

-- name: CreateCompletion :one
INSERT INTO chore_completions (
    id, chore_id, household_id, member_id, status, proof_notes, proof_photo_url,
    approved_by, rejection_reason, points_awarded, completed_at, verified_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: UpdateCompletionApproval :one
UPDATE chore_completions
SET status = $2, approved_by = $3, rejection_reason = $4, verified_at = $5
WHERE id = $1
RETURNING *;

-- name: GetSwapsByHousehold :many
SELECT * FROM chore_swap_requests WHERE household_id = $1 ORDER BY created_at DESC;

-- name: GetSwap :one
SELECT * FROM chore_swap_requests WHERE id = $1 LIMIT 1;

-- name: CreateSwap :one
INSERT INTO chore_swap_requests (
    id, household_id, chore_id, requester_id, target_member_id, status, reason, created_at, resolved_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: UpdateSwapStatus :one
UPDATE chore_swap_requests
SET status = $2, resolved_at = $3
WHERE id = $1
RETURNING *;

-- name: GetRewardsByHousehold :many
SELECT * FROM reward_items WHERE household_id = $1 ORDER BY points_cost ASC;

-- name: GetReward :one
SELECT * FROM reward_items WHERE id = $1 LIMIT 1;

-- name: CreateReward :one
INSERT INTO reward_items (
    id, household_id, title, description, points_cost, icon, is_active, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetRedemptionsByHousehold :many
SELECT * FROM reward_redemptions WHERE household_id = $1 ORDER BY requested_at DESC;

-- name: GetRedemption :one
SELECT * FROM reward_redemptions WHERE id = $1 LIMIT 1;

-- name: CreateRedemption :one
INSERT INTO reward_redemptions (
    id, household_id, member_id, reward_item_id, points_spent, status, requested_at, fulfilled_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateRedemptionStatus :one
UPDATE reward_redemptions
SET status = $2, fulfilled_at = $3
WHERE id = $1
RETURNING *;

-- name: GetActivitiesByHousehold :many
SELECT * FROM activity_logs
WHERE household_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateActivityLog :one
INSERT INTO activity_logs (
    id, household_id, recipient_id, actor_id, event_type, entity_type, entity_id, message, read, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetCommentsByChore :many
SELECT * FROM chore_comments WHERE chore_id = $1 ORDER BY created_at ASC;

-- name: CreateComment :one
INSERT INTO chore_comments (
    id, chore_id, member_id, message, created_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetMagicLink :one
SELECT * FROM magic_links WHERE token = $1 LIMIT 1;

-- name: UpsertMagicLink :one
INSERT INTO magic_links (token, member_id, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (token) DO UPDATE SET created_at = EXCLUDED.created_at
RETURNING *;

-- name: TruncateAll :exec
TRUNCATE TABLE chore_comments, activity_logs, reward_redemptions, reward_items, chore_swap_requests, chore_completions, chores, members, households, magic_links CASCADE;
