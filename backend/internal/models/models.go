package models

// HouseholdMode represents living arrangement modes.
const (
	HouseholdModeFlatmate = "flatmate"
	HouseholdModeFamily   = "family"
	HouseholdModeCasual   = "casual"
)

// MemberRole represents a member's role within the household.
const (
	MemberRoleAdmin  = "admin"
	MemberRoleMember = "member"
	MemberRoleChild  = "child"
)

// ChoreCategory represents the category of the chore.
const (
	ChoreCategoryCleaning    = "cleaning"
	ChoreCategoryKitchen     = "kitchen"
	ChoreCategoryYard        = "yard"
	ChoreCategoryPets        = "pets"
	ChoreCategoryMaintenance = "maintenance"
	ChoreCategoryOther       = "other"
)

// ChoreAssignmentType represents how the chore is assigned.
const (
	ChoreAssignmentDirect     = "direct"
	ChoreAssignmentRoundRobin = "round_robin"
	ChoreAssignmentOpenPool   = "open_pool"
)

// ChoreRecurrenceType represents recurrence cadence.
const (
	ChoreRecurrenceNone       = "none"
	ChoreRecurrenceDaily      = "daily"
	ChoreRecurrenceWeekly     = "weekly"
	ChoreRecurrenceMonthly    = "monthly"
	ChoreRecurrenceCustomDays = "custom_days"
)

// ChoreStatus represents the lifecycle state of a chore.
const (
	ChoreStatusUnassigned      = "unassigned"
	ChoreStatusAssigned        = "assigned"
	ChoreStatusInProgress      = "in_progress"
	ChoreStatusPendingApproval = "pending_approval"
	ChoreStatusCompleted       = "completed"
	ChoreStatusOverdue         = "overdue"
)

// CompletionStatus represents approval status for a chore completion.
const (
	CompletionStatusApproved        = "approved"
	CompletionStatusRejected        = "rejected"
	CompletionStatusPendingApproval = "pending_approval"
)

// SwapStatus represents the state of a chore swap request.
const (
	SwapStatusPending   = "pending"
	SwapStatusAccepted  = "accepted"
	SwapStatusRejected  = "rejected"
	SwapStatusCancelled = "cancelled"
)

// RedemptionStatus represents status of a reward redemption.
const (
	RedemptionStatusRequested = "requested"
	RedemptionStatusFulfilled = "fulfilled"
	RedemptionStatusCancelled = "cancelled"
)

// ActivityEventType represents activity event types.
const (
	ActivityChoreCreated      = "chore_created"
	ActivityChoreAssigned     = "chore_assigned"
	ActivityChoreCompleted    = "chore_completed"
	ActivityApprovalRequested = "approval_requested"
	ActivityChoreApproved     = "chore_approved"
	ActivityChoreRejected     = "chore_rejected"
	ActivitySwapProposed      = "swap_proposed"
	ActivitySwapAccepted      = "swap_accepted"
	ActivityChoreNudge        = "chore_nudge"
	ActivityRewardRedeemed    = "reward_redeemed"
	ActivityRewardFulfilled   = "reward_fulfilled"
)

// ChoreRecurrenceRule contains specific recurrence settings.
type ChoreRecurrenceRule struct {
	Days []string `json:"days,omitempty"`
}

// HouseholdSettings holds configuration for a household.
type HouseholdSettings struct {
	DefaultMode string `json:"default_mode"`
	AllowSwaps  bool   `json:"allow_swaps"`
	Timezone    string `json:"timezone"`
	AdminPIN    string `json:"admin_pin,omitempty"`
}

// HouseholdSettingsUpdate represents mutable settings.
type HouseholdSettingsUpdate struct {
	DefaultMode *string `json:"default_mode,omitempty"`
	AllowSwaps  *bool   `json:"allow_swaps,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
	AdminPIN    *string `json:"admin_pin,omitempty"`
}

// Household represents a household entity.
type Household struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	InviteCode string            `json:"invite_code"`
	Settings   HouseholdSettings `json:"settings"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
}

// HouseholdCreateRequest represents household creation payload.
type HouseholdCreateRequest struct {
	Name        string  `json:"name"`
	DefaultMode *string `json:"default_mode,omitempty"`
	AllowSwaps  *bool   `json:"allow_swaps,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
}

// HouseholdJoinRequest represents joining via invite code.
type HouseholdJoinRequest struct {
	InviteCode string `json:"invite_code"`
}

// Member represents a household member.
type Member struct {
	ID                string  `json:"id"`
	HouseholdID       string  `json:"household_id"`
	Name              string  `json:"name"`
	Email             *string `json:"email,omitempty"`
	PasswordHash      string  `json:"-"`
	AvatarURL         *string `json:"avatar_url"`
	Role              string  `json:"role"`
	PointsBalance     int     `json:"points_balance"`
	TotalPointsEarned int     `json:"total_points_earned"`
	Streak            int     `json:"streak"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// MemberCreateRequest represents member creation payload.
type MemberCreateRequest struct {
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	Email     *string `json:"email,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// Chore represents a household chore task.
type Chore struct {
	ID                   string               `json:"id"`
	HouseholdID          string               `json:"household_id"`
	Title                string               `json:"title"`
	Description          *string              `json:"description"`
	Category             string               `json:"category"`
	EffortPoints         int                  `json:"effort_points"`
	AssignmentType       string               `json:"assignment_type"`
	CurrentAssigneeID    *string              `json:"current_assignee_id"`
	RotationMemberIDs    []string             `json:"rotation_member_ids,omitempty"`
	CurrentRotationIndex *int                 `json:"current_rotation_index"`
	RecurrenceType       string               `json:"recurrence_type"`
	RecurrenceRule       *ChoreRecurrenceRule `json:"recurrence_rule,omitempty"`
	DueDate              *string              `json:"due_date"`
	RequiresApproval     bool                 `json:"requires_approval"`
	RequiresProof        bool                 `json:"requires_proof"`
	Status               string               `json:"status"`
	CreatedBy            string               `json:"created_by"`
	CreatedAt            string               `json:"created_at"`
	UpdatedAt            string               `json:"updated_at"`
}

// ChoreCreateRequest represents chore creation payload.
type ChoreCreateRequest struct {
	HouseholdID          string               `json:"household_id"`
	Title                string               `json:"title"`
	Description          *string              `json:"description,omitempty"`
	Category             string               `json:"category"`
	EffortPoints         int                  `json:"effort_points"`
	AssignmentType       string               `json:"assignment_type"`
	CurrentAssigneeID    *string              `json:"current_assignee_id,omitempty"`
	RotationMemberIDs    []string             `json:"rotation_member_ids,omitempty"`
	CurrentRotationIndex *int                 `json:"current_rotation_index,omitempty"`
	RecurrenceType       string               `json:"recurrence_type"`
	RecurrenceRule       *ChoreRecurrenceRule `json:"recurrence_rule,omitempty"`
	DueDate              *string              `json:"due_date,omitempty"`
	RequiresApproval     *bool                `json:"requires_approval,omitempty"`
	RequiresProof        *bool                `json:"requires_proof,omitempty"`
	Status               *string              `json:"status,omitempty"`
	CreatedBy            string               `json:"created_by"`
}

// ChoreUpdateRequest represents chore update payload.
type ChoreUpdateRequest struct {
	Title                *string              `json:"title,omitempty"`
	Description          *string              `json:"description,omitempty"`
	Category             *string              `json:"category,omitempty"`
	EffortPoints         *int                 `json:"effort_points,omitempty"`
	AssignmentType       *string              `json:"assignment_type,omitempty"`
	CurrentAssigneeID    *string              `json:"current_assignee_id,omitempty"`
	RotationMemberIDs    *[]string            `json:"rotation_member_ids,omitempty"`
	CurrentRotationIndex *int                 `json:"current_rotation_index,omitempty"`
	RecurrenceType       *string              `json:"recurrence_type,omitempty"`
	RecurrenceRule       *ChoreRecurrenceRule `json:"recurrence_rule,omitempty"`
	DueDate              *string              `json:"due_date,omitempty"`
	RequiresApproval     *bool                `json:"requires_approval,omitempty"`
	RequiresProof        *bool                `json:"requires_proof,omitempty"`
	Status               *string              `json:"status,omitempty"`
}

// ChoreClaimRequest represents chore claiming payload.
type ChoreClaimRequest struct {
	MemberID string `json:"member_id"`
}

// ChoreCompleteRequest represents chore completion submission.
type ChoreCompleteRequest struct {
	MemberID       string  `json:"member_id"`
	ProofNotes     *string `json:"proof_notes,omitempty"`
	ProofPhotoURL  *string `json:"proof_photo_url,omitempty"`
}

// ChoreCompletion represents a completion record.
type ChoreCompletion struct {
	ID              string  `json:"id"`
	ChoreID         string  `json:"chore_id"`
	HouseholdID     string  `json:"household_id"`
	MemberID        string  `json:"member_id"`
	Status          string  `json:"status"`
	ProofNotes      *string `json:"proof_notes"`
	ProofPhotoURL   *string `json:"proof_photo_url"`
	ApprovedBy      *string `json:"approved_by"`
	RejectionReason *string `json:"rejection_reason"`
	PointsAwarded   int     `json:"points_awarded"`
	CompletedAt     string  `json:"completed_at"`
	VerifiedAt      *string `json:"verified_at"`
}

// ChoreCompleteResponse represents the response when completing a chore.
type ChoreCompleteResponse struct {
	Chore         Chore            `json:"chore"`
	Completion    *ChoreCompletion `json:"completion,omitempty"`
	PointsAwarded int              `json:"pointsAwarded"`
}

// PendingApprovalItem represents an item in the approval queue.
type PendingApprovalItem struct {
	Completion ChoreCompletion `json:"completion"`
	Chore      Chore           `json:"chore"`
	Member     Member          `json:"member"`
}

// ApproveCompletionRequest represents an admin approval payload.
type ApproveCompletionRequest struct {
	AdminMemberID string `json:"admin_member_id"`
}

// RejectCompletionRequest represents an admin rejection payload.
type RejectCompletionRequest struct {
	AdminMemberID string `json:"admin_member_id"`
	Reason        string `json:"reason"`
}

// ChoreSwapRequest represents a swap request entity.
type ChoreSwapRequest struct {
	ID             string  `json:"id"`
	HouseholdID    string  `json:"household_id"`
	ChoreID        string  `json:"chore_id"`
	RequesterID    string  `json:"requester_id"`
	TargetMemberID *string `json:"target_member_id"`
	Status         string  `json:"status"`
	Reason         *string `json:"reason"`
	CreatedAt      string  `json:"created_at"`
	ResolvedAt     *string `json:"resolved_at"`
}

// ChoreSwapCreateRequest represents the request on /api/chores/{chore_id}/swap.
type ChoreSwapCreateRequest struct {
	RequesterID    string  `json:"requester_id"`
	TargetMemberID *string `json:"target_member_id,omitempty"`
	Reason         *string `json:"reason,omitempty"`
}

// SwapCreateRequest represents the request on /api/swaps.
type SwapCreateRequest struct {
	ChoreID        string  `json:"chore_id"`
	RequesterID    string  `json:"requester_id"`
	TargetMemberID *string `json:"target_member_id,omitempty"`
	Reason         *string `json:"reason,omitempty"`
}

// SwapAcceptRequest represents swap acceptance.
type SwapAcceptRequest struct {
	AcceptorMemberID string `json:"acceptor_member_id"`
}

// SwapItem represents a swap with full relational context.
type SwapItem struct {
	Swap         ChoreSwapRequest `json:"swap"`
	Chore        Chore            `json:"chore"`
	Requester    Member           `json:"requester"`
	TargetMember *Member          `json:"targetMember"`
}

// RewardItem represents a redeemable catalog reward.
type RewardItem struct {
	ID          string  `json:"id"`
	HouseholdID string  `json:"household_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	PointsCost  int     `json:"points_cost"`
	Icon        *string `json:"icon,omitempty"`
	IsActive    bool    `json:"is_active"`
}

// RewardItemCreateRequest represents reward creation payload.
type RewardItemCreateRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	PointsCost  int     `json:"points_cost"`
	Icon        *string `json:"icon,omitempty"`
}

// RewardRedeemRequest represents a member redeeming a reward.
type RewardRedeemRequest struct {
	MemberID string `json:"member_id"`
}

// RewardRedemption represents a reward redemption record.
type RewardRedemption struct {
	ID           string  `json:"id"`
	HouseholdID  string  `json:"household_id"`
	MemberID     string  `json:"member_id"`
	RewardItemID string  `json:"reward_item_id"`
	PointsSpent  int     `json:"points_spent"`
	Status       string  `json:"status"`
	RequestedAt  string  `json:"requested_at"`
	FulfilledAt  *string `json:"fulfilled_at"`
}

// RewardRedemptionItem represents a redemption with full relational context.
type RewardRedemptionItem struct {
	Redemption RewardRedemption `json:"redemption"`
	Reward     RewardItem       `json:"reward"`
	Member     Member           `json:"member"`
}

// ActivityLog represents an audit trail event.
type ActivityLog struct {
	ID          string  `json:"id"`
	HouseholdID string  `json:"household_id"`
	RecipientID *string `json:"recipient_id"`
	ActorID     string  `json:"actor_id"`
	EventType   string  `json:"event_type"`
	EntityType  string  `json:"entity_type"`
	EntityID    string  `json:"entity_id"`
	Message     string  `json:"message"`
	Read        bool    `json:"read"`
	CreatedAt   string  `json:"created_at"`
}

// ChoreNudgeRequest represents a nudge trigger request.
type ChoreNudgeRequest struct {
	SenderMemberID string `json:"sender_member_id"`
}

// ChoreNudgeResponse represents the nudge result.
type ChoreNudgeResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	EmailDispatched bool   `json:"email_dispatched,omitempty"`
	SQSQueued       bool   `json:"sqs_queued,omitempty"`
}

// ChoreComment represents a comment on a chore.
type ChoreComment struct {
	ID        string `json:"id"`
	ChoreID   string `json:"chore_id"`
	MemberID  string `json:"member_id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// ChoreCommentWithMember represents a comment with author context.
type ChoreCommentWithMember struct {
	Comment ChoreComment `json:"comment"`
	Member  *Member      `json:"member"`
}

// ChoreCommentCreateRequest represents chore comment payload.
type ChoreCommentCreateRequest struct {
	MemberID string `json:"member_id"`
	Message  string `json:"message"`
}

// SuccessResponse represents a standard success confirmation.
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents the universal error response schema.
type ErrorResponse struct {
	Error   string         `json:"error"`
	Message string         `json:"message"`
	Code    int            `json:"code"`
	Details map[string]any `json:"details"`
}

// AuthTokenResponse represents the response when an auth token is issued.
type AuthTokenResponse struct {
	AccessToken string     `json:"access_token"`
	TokenType   string     `json:"token_type"`
	ExpiresIn   int64      `json:"expires_in"`
	Member      *Member    `json:"member"`
	Household   *Household `json:"household"`
}

// MagicLinkRequest represents requesting a magic login link.
type MagicLinkRequest struct {
	Email string `json:"email"`
}

// MagicLinkVerifyRequest represents verifying a magic link token.
type MagicLinkVerifyRequest struct {
	Token string `json:"token"`
}

// DemoLoginRequest represents minting a demo token for a persona.
type DemoLoginRequest struct {
	MemberID string `json:"member_id"`
}

// VerifyPINRequest represents verifying the household admin PIN.
type VerifyPINRequest struct {
	PIN string `json:"pin"`
}

// LoginRequest represents email + password login credentials.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents new user registration and household creation/joining.
type RegisterRequest struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	HouseholdName string `json:"household_name,omitempty"`
	InviteCode    string `json:"invite_code,omitempty"`
	Role          string `json:"role,omitempty"`
}

// ForgotPasswordRequest represents requesting a password reset email.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordResponse represents the response after requesting password reset.
type ForgotPasswordResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

// ResetPasswordRequest represents resetting the password with a token.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// DevEmail represents an outgoing email captured in the development/testing buffer.
type DevEmail struct {
	ID        string `json:"id"`
	To        string `json:"to"`
	ToName    string `json:"to_name"`
	From      string `json:"from"`
	Subject   string `json:"subject"`
	TextBody  string `json:"text_body"`
	HTMLBody  string `json:"html_body"`
	Type      string `json:"type"` // "magic_link", "password_reset", "chore_reminder", "chore_assigned"
	SentAt    string `json:"sent_at"`
	Provider  string `json:"provider"`
	Token     string `json:"token,omitempty"`
}

