package store

import (
	"errors"

	"github.com/choresync/backend/internal/models"
)

var (
	ErrNotFound   = errors.New("RESOURCE_NOT_FOUND")
	ErrBadRequest = errors.New("BAD_REQUEST")
	ErrForbidden  = errors.New("FORBIDDEN")
)

// Store defines the persistent storage interface for ChoreSync.
type Store interface {
	Reset()
	GetHouseholds() []models.Household
	GetHousehold(id string) (*models.Household, error)
	CreateHousehold(req models.HouseholdCreateRequest) models.Household
	JoinHousehold(inviteCode string) (*models.Household, error)
	UpdateHouseholdSettings(id string, update models.HouseholdSettingsUpdate) (*models.Household, error)
	GetMembers(householdID string) []models.Member
	GetMember(id string) (*models.Member, error)
	AddMember(householdID string, req models.MemberCreateRequest) (*models.Member, error)
	GetChores(householdID, status, category, search, view, assigneeID string) []models.Chore
	GetChore(id string) (*models.Chore, error)
	CreateChore(req models.ChoreCreateRequest) (*models.Chore, error)
	UpdateChore(id string, req models.ChoreUpdateRequest) (*models.Chore, error)
	DeleteChore(id string) error
	ClaimChore(choreID, memberID string) (*models.Chore, error)
	CompleteChore(choreID, memberID string, proofNotes, proofPhotoURL *string) (*models.Chore, *models.ChoreCompletion, int, error)
	RotateChore(householdID, choreID string) (*models.Chore, error)
	GetPendingApprovals(householdID string) []models.PendingApprovalItem
	ApproveChore(completionID, adminMemberID string) error
	RejectChore(completionID, adminMemberID, reason string) error
	GetSwaps(householdID string) []models.SwapItem
	CreateSwap(householdID, choreID, requesterID string, targetMemberID, reason *string) (*models.ChoreSwapRequest, error)
	AcceptSwap(swapID, acceptorMemberID string) error
	RejectSwap(swapID string) error
	CancelSwap(swapID string) error
	GetRewards(householdID string) []models.RewardItem
	GetReward(rewardID string) (*models.RewardItem, error)
	CreateReward(householdID string, req models.RewardItemCreateRequest) (*models.RewardItem, error)
	RedeemReward(rewardID, memberID string) (*models.RewardRedemption, error)
	FulfillReward(redemptionID string) error
	GetRedemptions(householdID string) []models.RewardRedemptionItem
	GetActivitiesFiltered(householdID string, entityType, eventType string, limit, offset int) []models.ActivityLog
	GetActivities(householdID string) []models.ActivityLog
	SendNudge(choreID, senderMemberID string) (string, error)
	GetComments(choreID string) []models.ChoreCommentWithMember
	AddComment(choreID, memberID, message string) (*models.ChoreComment, error)
	VerifyAdminPIN(householdID, pin string) (bool, error)
	GetMemberByEmail(email string) (*models.Member, *models.Household, error)
	CreateMagicLink(email string) (string, *models.Member, *models.Household, error)
	VerifyMagicLink(token string) (*models.Member, *models.Household, error)
	GetDemoMember(memberID string) (*models.Member, *models.Household, error)
	VerifyPassword(memberID, password string) bool
	RegisterMember(req models.RegisterRequest) (*models.Member, *models.Household, error)
	CreatePasswordResetToken(email string) (string, *models.Member, error)
	ResetPasswordWithToken(token, newPassword string) (*models.Member, *models.Household, error)
}
