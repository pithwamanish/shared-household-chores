package store

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/choresync/backend/internal/models"
)
type MemoryStore struct {
	mu          sync.RWMutex
	households  map[string]models.Household
	members     map[string]models.Member
	chores      map[string]models.Chore
	completions map[string]models.ChoreCompletion
	swaps       map[string]models.ChoreSwapRequest
	rewards     map[string]models.RewardItem
	redemptions map[string]models.RewardRedemption
	activities  []models.ActivityLog
	comments       []models.ChoreComment
	magicLinks     map[string]magicLinkEntry
	passwordResets map[string]passwordResetEntry
}

type magicLinkEntry struct {
	memberID  string
	createdAt time.Time
	expiresAt time.Time
}

type passwordResetEntry struct {
	memberID  string
	expiresAt time.Time
}

func NewStore() Store {
	return NewMemoryStore()
}

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{}
	s.Reset()
	return s
}

func (s *MemoryStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	households, members, chores, completions, swaps, rewards, redemptions, activities, comments := getInitialData()

	now := time.Now().UTC()
	demoExpiry := now.Add(365 * 24 * time.Hour)
	s.magicLinks = make(map[string]magicLinkEntry)
	s.magicLinks["MAGIC-SARAH"] = magicLinkEntry{memberID: "m-sarah", createdAt: now, expiresAt: demoExpiry}
	s.magicLinks["MAGIC-DAVID"] = magicLinkEntry{memberID: "m-david-fam", createdAt: now, expiresAt: demoExpiry}
	s.magicLinks["MAGIC-ALEX"] = magicLinkEntry{memberID: "m-alex", createdAt: now, expiresAt: demoExpiry}

	s.passwordResets = make(map[string]passwordResetEntry)

	s.households = make(map[string]models.Household, len(households))
	for _, h := range households {
		s.households[h.ID] = h
	}

	s.members = make(map[string]models.Member, len(members))
	for _, m := range members {
		s.members[m.ID] = m
	}

	s.chores = make(map[string]models.Chore, len(chores))
	for _, c := range chores {
		s.chores[c.ID] = c
	}

	s.completions = make(map[string]models.ChoreCompletion, len(completions))
	for _, comp := range completions {
		s.completions[comp.ID] = comp
	}

	s.swaps = make(map[string]models.ChoreSwapRequest, len(swaps))
	for _, sw := range swaps {
		s.swaps[sw.ID] = sw
	}

	s.rewards = make(map[string]models.RewardItem, len(rewards))
	for _, r := range rewards {
		s.rewards[r.ID] = r
	}

	s.redemptions = make(map[string]models.RewardRedemption, len(redemptions))
	for _, red := range redemptions {
		s.redemptions[red.ID] = red
	}

	s.activities = make([]models.ActivityLog, len(activities))
	copy(s.activities, activities)

	s.comments = make([]models.ChoreComment, len(comments))
	copy(s.comments, comments)
}

func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// ----------------- HOUSEHOLDS -----------------

func (s *MemoryStore) GetHouseholds() []models.Household {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]models.Household, 0, len(s.households))
	for _, h := range s.households {
		list = append(list, h)
	}
	return list
}

func (s *MemoryStore) GetHousehold(id string) (*models.Household, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	h, ok := s.households[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &h, nil
}

func (s *MemoryStore) CreateHousehold(req models.HouseholdCreateRequest) models.Household {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339)
	id := generateID("h")
	inviteCode := strings.ToUpper(fmt.Sprintf("%s-%04d", strings.ReplaceAll(req.Name, " ", "")[:min(4, len(req.Name))], time.Now().Nanosecond()%10000))

	mode := models.HouseholdModeFlatmate
	if req.DefaultMode != nil && *req.DefaultMode != "" {
		mode = *req.DefaultMode
	}

	allowSwaps := true
	if req.AllowSwaps != nil {
		allowSwaps = *req.AllowSwaps
	}

	tz := "America/New_York"
	if req.Timezone != nil && *req.Timezone != "" {
		tz = *req.Timezone
	}

	h := models.Household{
		ID:         id,
		Name:       req.Name,
		InviteCode: inviteCode,
		Settings: models.HouseholdSettings{
			DefaultMode: mode,
			AllowSwaps:  allowSwaps,
			Timezone:    tz,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.households[id] = h
	return h
}

func (s *MemoryStore) JoinHousehold(inviteCode string) (*models.Household, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cleanCode := strings.TrimSpace(strings.ToUpper(inviteCode))
	for _, h := range s.households {
		if strings.ToUpper(h.InviteCode) == cleanCode {
			return &h, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) UpdateHouseholdSettings(id string, update models.HouseholdSettingsUpdate) (*models.Household, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	h, ok := s.households[id]
	if !ok {
		return nil, ErrNotFound
	}

	if update.DefaultMode != nil {
		h.Settings.DefaultMode = *update.DefaultMode
	}
	if update.AllowSwaps != nil {
		h.Settings.AllowSwaps = *update.AllowSwaps
	}
	if update.Timezone != nil {
		h.Settings.Timezone = *update.Timezone
	}
	if update.AdminPIN != nil {
		h.Settings.AdminPIN = *update.AdminPIN
	}
	h.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	s.households[id] = h
	return &h, nil
}

// ----------------- MEMBERS -----------------

func (s *MemoryStore) GetMembers(householdID string) []models.Member {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]models.Member, 0)
	for _, m := range s.members {
		if m.HouseholdID == householdID {
			list = append(list, m)
		}
	}
	return list
}

func (s *MemoryStore) GetMember(id string) (*models.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, ok := s.members[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &m, nil
}

func (s *MemoryStore) AddMember(householdID string, req models.MemberCreateRequest) (*models.Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.households[householdID]; !ok {
		return nil, ErrNotFound
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id := generateID("m")

	avatar := req.AvatarURL
	if avatar == nil || *avatar == "" {
		url := fmt.Sprintf("https://api.dicebear.com/7.x/bottts/svg?seed=%s", req.Name)
		avatar = &url
	}

	m := models.Member{
		ID:                id,
		HouseholdID:       householdID,
		Name:              strings.TrimSpace(req.Name),
		Email:             req.Email,
		AvatarURL:         avatar,
		Role:              req.Role,
		PointsBalance:     0,
		TotalPointsEarned: 0,
		Streak:            0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	s.members[id] = m
	return &m, nil
}

// ----------------- CHORES -----------------

func (s *MemoryStore) GetChores(householdID, status, category, search, view, assigneeID string) []models.Chore {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	list := make([]models.Chore, 0)

	for id, chore := range s.chores {
		if chore.HouseholdID != householdID {
			continue
		}

		// Dynamically flag overdue tasks
		if chore.DueDate != nil && chore.Status != models.ChoreStatusCompleted && chore.Status != models.ChoreStatusPendingApproval {
			if due, err := time.Parse(time.RFC3339, *chore.DueDate); err == nil {
				if due.Before(now) && chore.Status != models.ChoreStatusOverdue {
					chore.Status = models.ChoreStatusOverdue
					chore.UpdatedAt = now.Format(time.RFC3339)
					s.chores[id] = chore
				}
			}
		}

		// Filtering
		if status != "" && chore.Status != status {
			continue
		}
		if category != "" && chore.Category != category {
			continue
		}
		if search != "" {
			term := strings.ToLower(search)
			titleMatches := strings.Contains(strings.ToLower(chore.Title), term)
			descMatches := chore.Description != nil && strings.Contains(strings.ToLower(*chore.Description), term)
			if !titleMatches && !descMatches {
				continue
			}
		}
		if view != "" {
			switch view {
			case "my":
				if assigneeID == "" || chore.CurrentAssigneeID == nil || *chore.CurrentAssigneeID != assigneeID {
					continue
				}
			case "open":
				if chore.AssignmentType != models.ChoreAssignmentOpenPool || chore.CurrentAssigneeID != nil {
					continue
				}
			case "approval":
				if chore.Status != models.ChoreStatusPendingApproval {
					continue
				}
			case "overdue":
				if chore.Status != models.ChoreStatusOverdue {
					continue
				}
			}
		}

		list = append(list, chore)
	}

	return list
}

func (s *MemoryStore) GetChore(id string) (*models.Chore, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chore, ok := s.chores[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &chore, nil
}

func (s *MemoryStore) CreateChore(req models.ChoreCreateRequest) (*models.Chore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.households[req.HouseholdID]; !ok {
		return nil, ErrNotFound
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id := generateID("c")

	initialStatus := models.ChoreStatusAssigned
	if req.Status != nil && *req.Status != "" {
		initialStatus = *req.Status
	}

	currentAssigneeID := req.CurrentAssigneeID
	currentRotIndex := req.CurrentRotationIndex

	if req.AssignmentType == models.ChoreAssignmentOpenPool {
		initialStatus = models.ChoreStatusUnassigned
		currentAssigneeID = nil
	} else if req.AssignmentType == models.ChoreAssignmentRoundRobin {
		rotList := req.RotationMemberIDs
		idx := 0
		if currentRotIndex != nil {
			idx = *currentRotIndex
		}
		if len(rotList) > 0 {
			chosen := rotList[idx%len(rotList)]
			currentAssigneeID = &chosen
			initialStatus = models.ChoreStatusAssigned
		} else {
			currentAssigneeID = nil
			initialStatus = models.ChoreStatusUnassigned
		}
	}

	reqApproval := false
	if req.RequiresApproval != nil {
		reqApproval = *req.RequiresApproval
	}

	reqProof := false
	if req.RequiresProof != nil {
		reqProof = *req.RequiresProof
	}

	chore := models.Chore{
		ID:                   id,
		HouseholdID:          req.HouseholdID,
		Title:                req.Title,
		Description:          req.Description,
		Category:             req.Category,
		EffortPoints:         req.EffortPoints,
		AssignmentType:       req.AssignmentType,
		CurrentAssigneeID:    currentAssigneeID,
		RotationMemberIDs:    req.RotationMemberIDs,
		CurrentRotationIndex: currentRotIndex,
		RecurrenceType:       req.RecurrenceType,
		RecurrenceRule:       req.RecurrenceRule,
		DueDate:              req.DueDate,
		RequiresApproval:     reqApproval,
		RequiresProof:        reqProof,
		Status:               initialStatus,
		CreatedBy:            req.CreatedBy,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	s.chores[id] = chore

	// Log activity
	s.addActivityLocked(req.HouseholdID, req.CreatedBy, nil, models.ActivityChoreCreated, "chore", id,
		fmt.Sprintf("Created new chore \"%s\" (%d pts)", chore.Title, chore.EffortPoints))

	return &chore, nil
}

func (s *MemoryStore) UpdateChore(id string, req models.ChoreUpdateRequest) (*models.Chore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chore, ok := s.chores[id]
	if !ok {
		return nil, ErrNotFound
	}

	if req.Title != nil {
		chore.Title = *req.Title
	}
	if req.Description != nil {
		chore.Description = req.Description
	}
	if req.Category != nil {
		chore.Category = *req.Category
	}
	if req.EffortPoints != nil {
		chore.EffortPoints = *req.EffortPoints
	}
	if req.AssignmentType != nil {
		chore.AssignmentType = *req.AssignmentType
	}
	if req.CurrentAssigneeID != nil {
		chore.CurrentAssigneeID = req.CurrentAssigneeID
	}
	if req.RotationMemberIDs != nil {
		chore.RotationMemberIDs = *req.RotationMemberIDs
	}
	if req.CurrentRotationIndex != nil {
		chore.CurrentRotationIndex = req.CurrentRotationIndex
	}
	if req.RecurrenceType != nil {
		chore.RecurrenceType = *req.RecurrenceType
	}
	if req.RecurrenceRule != nil {
		chore.RecurrenceRule = req.RecurrenceRule
	}
	if req.DueDate != nil {
		chore.DueDate = req.DueDate
	}
	if req.RequiresApproval != nil {
		chore.RequiresApproval = *req.RequiresApproval
	}
	if req.RequiresProof != nil {
		chore.RequiresProof = *req.RequiresProof
	}
	if req.Status != nil {
		chore.Status = *req.Status
	}

	chore.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.chores[id] = chore
	return &chore, nil
}

func (s *MemoryStore) DeleteChore(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.chores[id]; !ok {
		return ErrNotFound
	}
	delete(s.chores, id)
	return nil
}

func (s *MemoryStore) ClaimChore(choreID, memberID string) (*models.Chore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chore, ok := s.chores[choreID]
	if !ok {
		return nil, ErrNotFound
	}
	member, ok := s.members[memberID]
	if !ok {
		return nil, ErrNotFound
	}

	chore.CurrentAssigneeID = &memberID
	chore.Status = models.ChoreStatusAssigned
	chore.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.chores[choreID] = chore

	s.addActivityLocked(chore.HouseholdID, memberID, nil, models.ActivityChoreAssigned, "chore", chore.ID,
		fmt.Sprintf("%s claimed \"%s\" from the Open Pool", member.Name, chore.Title))

	return &chore, nil
}

func (s *MemoryStore) CompleteChore(choreID, memberID string, proofNotes, proofPhotoURL *string) (*models.Chore, *models.ChoreCompletion, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chore, ok := s.chores[choreID]
	if !ok {
		return nil, nil, 0, ErrNotFound
	}
	member, ok := s.members[memberID]
	if !ok {
		return nil, nil, 0, ErrNotFound
	}
	if chore.HouseholdID != member.HouseholdID {
		return nil, nil, 0, fmt.Errorf("%w: member '%s' does not belong to household '%s'", ErrForbidden, memberID, chore.HouseholdID)
	}
	if chore.Status == models.ChoreStatusPendingApproval {
		return nil, nil, 0, fmt.Errorf("%w: chore '%s' is already awaiting verification", ErrBadRequest, choreID)
	}
	if chore.Status == models.ChoreStatusCompleted {
		return nil, nil, 0, fmt.Errorf("%w: chore '%s' is already completed", ErrBadRequest, choreID)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	compID := generateID("comp")

	if chore.RequiresApproval {
		// Requires parent/admin approval
		chore.Status = models.ChoreStatusPendingApproval
		chore.UpdatedAt = now
		s.chores[choreID] = chore

		comp := models.ChoreCompletion{
			ID:            compID,
			ChoreID:       chore.ID,
			HouseholdID:   chore.HouseholdID,
			MemberID:      memberID,
			Status:        models.CompletionStatusPendingApproval,
			ProofNotes:    proofNotes,
			ProofPhotoURL: proofPhotoURL,
			PointsAwarded: chore.EffortPoints,
			CompletedAt:   now,
		}
		s.completions[compID] = comp

		s.addActivityLocked(chore.HouseholdID, memberID, nil, models.ActivityApprovalRequested, "chore", chore.ID,
			fmt.Sprintf("%s submitted \"%s\" for verification", member.Name, chore.Title))

		return &chore, &comp, 0, nil
	}

	// Instant completion
	chore.Status = models.ChoreStatusCompleted
	chore.UpdatedAt = now
	s.chores[choreID] = chore

	// Award points and update streak
	member.PointsBalance += chore.EffortPoints
	member.TotalPointsEarned += chore.EffortPoints
	member.Streak += 1
	member.UpdatedAt = now
	s.members[memberID] = member

	comp := models.ChoreCompletion{
		ID:            compID,
		ChoreID:       chore.ID,
		HouseholdID:   chore.HouseholdID,
		MemberID:      memberID,
		Status:        models.CompletionStatusApproved,
		ProofNotes:    proofNotes,
		ProofPhotoURL: proofPhotoURL,
		PointsAwarded: chore.EffortPoints,
		CompletedAt:   now,
		VerifiedAt:    &now,
	}
	s.completions[compID] = comp

	s.addActivityLocked(chore.HouseholdID, memberID, nil, models.ActivityChoreCompleted, "chore", chore.ID,
		fmt.Sprintf("%s completed \"%s\" (+%d pts)", member.Name, chore.Title, chore.EffortPoints))

	// Advance recurrence and rotation
	s.advanceRecurrenceLocked(chore)

	return &chore, &comp, chore.EffortPoints, nil
}

// RotateChore advances a round-robin chore to the next member in the sequence.
func (s *MemoryStore) RotateChore(householdID, choreID string) (*models.Chore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chore, ok := s.chores[choreID]
	if !ok {
		return nil, ErrNotFound
	}
	if householdID != "" && chore.HouseholdID != householdID {
		return nil, ErrNotFound
	}

	if chore.AssignmentType != models.ChoreAssignmentRoundRobin {
		return nil, fmt.Errorf("%w: chore '%s' is not configured for round-robin rotation (assignment_type: %s)", ErrBadRequest, choreID, chore.AssignmentType)
	}

	if len(chore.RotationMemberIDs) == 0 {
		return nil, fmt.Errorf("%w: chore '%s' has no rotation members configured", ErrBadRequest, choreID)
	}

	nextIndex := 0
	if chore.CurrentRotationIndex != nil {
		nextIndex = (*chore.CurrentRotationIndex + 1) % len(chore.RotationMemberIDs)
	}
	chore.CurrentRotationIndex = &nextIndex
	nextAssignee := chore.RotationMemberIDs[nextIndex]
	chore.CurrentAssigneeID = &nextAssignee
	chore.Status = models.ChoreStatusAssigned
	chore.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	s.chores[choreID] = chore

	assigneeName := nextAssignee
	if m, ok := s.members[nextAssignee]; ok {
		assigneeName = m.Name
	}
	s.addActivityLocked(chore.HouseholdID, chore.CreatedBy, &nextAssignee, models.ActivityChoreAssigned, "chore", choreID,
		fmt.Sprintf("Chore \"%s\" rotated to %s", chore.Title, assigneeName))

	return &chore, nil
}

// ----------------- APPROVALS -----------------

func (s *MemoryStore) GetPendingApprovals(householdID string) []models.PendingApprovalItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]models.PendingApprovalItem, 0)
	for _, comp := range s.completions {
		if comp.HouseholdID == householdID && comp.Status == models.CompletionStatusPendingApproval {
			chore, okC := s.chores[comp.ChoreID]
			member, okM := s.members[comp.MemberID]
			if okC && okM {
				list = append(list, models.PendingApprovalItem{
					Completion: comp,
					Chore:      chore,
					Member:     member,
				})
			}
		}
	}
	return list
}

func (s *MemoryStore) ApproveChore(completionID, adminMemberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	comp, ok := s.completions[completionID]
	if !ok {
		return ErrNotFound
	}
	if comp.Status != models.CompletionStatusPendingApproval {
		return fmt.Errorf("%w: completion '%s' is not awaiting approval (current status: %s)", ErrBadRequest, completionID, comp.Status)
	}

	admin, okA := s.members[adminMemberID]
	if !okA {
		return ErrNotFound
	}
	if admin.Role != models.MemberRoleAdmin || admin.HouseholdID != comp.HouseholdID {
		return fmt.Errorf("%w: member '%s' does not have admin privileges in household '%s'", ErrForbidden, adminMemberID, comp.HouseholdID)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	comp.Status = models.CompletionStatusApproved
	comp.ApprovedBy = &adminMemberID
	comp.VerifiedAt = &now
	s.completions[completionID] = comp

	chore, okC := s.chores[comp.ChoreID]
	if okC {
		chore.Status = models.ChoreStatusCompleted
		chore.UpdatedAt = now
		s.chores[comp.ChoreID] = chore
	}

	member, okM := s.members[comp.MemberID]
	if okM && okC {
		member.PointsBalance += chore.EffortPoints
		member.TotalPointsEarned += chore.EffortPoints
		member.Streak += 1
		member.UpdatedAt = now
		s.members[comp.MemberID] = member

		s.addActivityLocked(comp.HouseholdID, adminMemberID, &comp.MemberID, models.ActivityChoreApproved, "chore", chore.ID,
			fmt.Sprintf("%s approved %s's \"%s\" (+%d pts)", admin.Name, member.Name, chore.Title, chore.EffortPoints))

		s.advanceRecurrenceLocked(chore)
	}

	return nil
}

func (s *MemoryStore) RejectChore(completionID, adminMemberID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	comp, ok := s.completions[completionID]
	if !ok {
		return ErrNotFound
	}
	if comp.Status != models.CompletionStatusPendingApproval {
		return fmt.Errorf("%w: completion '%s' is not awaiting approval (current status: %s)", ErrBadRequest, completionID, comp.Status)
	}

	admin, okA := s.members[adminMemberID]
	if !okA {
		return ErrNotFound
	}
	if admin.Role != models.MemberRoleAdmin || admin.HouseholdID != comp.HouseholdID {
		return fmt.Errorf("%w: member '%s' does not have admin privileges in household '%s'", ErrForbidden, adminMemberID, comp.HouseholdID)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	comp.Status = models.CompletionStatusRejected
	comp.ApprovedBy = &adminMemberID
	comp.RejectionReason = &reason
	comp.VerifiedAt = &now
	s.completions[completionID] = comp

	chore, okC := s.chores[comp.ChoreID]
	if okC {
		// Revert to assigned for rework (or unassigned if no current assignee)
		if chore.CurrentAssigneeID != nil {
			chore.Status = models.ChoreStatusAssigned
		} else {
			chore.Status = models.ChoreStatusUnassigned
		}
		chore.UpdatedAt = now
		s.chores[comp.ChoreID] = chore
	}

	member, okM := s.members[comp.MemberID]
	memberName := "Member"
	if okM {
		memberName = member.Name
	}
	choreTitle := "chore"
	if okC {
		choreTitle = chore.Title
	}

	s.addActivityLocked(comp.HouseholdID, adminMemberID, &comp.MemberID, models.ActivityChoreRejected, "chore", comp.ChoreID,
		fmt.Sprintf("%s requested rework for %s's \"%s\": %s", admin.Name, memberName, choreTitle, reason))

	return nil
}

// ----------------- SWAPS -----------------

func (s *MemoryStore) GetSwaps(householdID string) []models.SwapItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]models.SwapItem, 0)
	for _, sw := range s.swaps {
		if sw.HouseholdID == householdID {
			chore, okC := s.chores[sw.ChoreID]
			reqMember, okR := s.members[sw.RequesterID]
			if okC && okR {
				var target *models.Member
				if sw.TargetMemberID != nil {
					if tm, okT := s.members[*sw.TargetMemberID]; okT {
						target = &tm
					}
				}
				list = append(list, models.SwapItem{
					Swap:         sw,
					Chore:        chore,
					Requester:    reqMember,
					TargetMember: target,
				})
			}
		}
	}
	return list
}

func (s *MemoryStore) CreateSwap(householdID, choreID, requesterID string, targetMemberID, reason *string) (*models.ChoreSwapRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chore, okC := s.chores[choreID]
	if !okC {
		return nil, ErrNotFound
	}
	if householdID != "" && chore.HouseholdID != householdID {
		return nil, ErrNotFound
	}

	household, okH := s.households[chore.HouseholdID]
	if !okH {
		return nil, ErrNotFound
	}
	if !household.Settings.AllowSwaps {
		return nil, fmt.Errorf("%w: swaps are disabled for this household", ErrBadRequest)
	}

	requester, okR := s.members[requesterID]
	if !okR {
		return nil, ErrNotFound
	}
	if requester.HouseholdID != chore.HouseholdID {
		return nil, fmt.Errorf("%w: requester does not belong to the chore's household", ErrForbidden)
	}

	if chore.CurrentAssigneeID == nil || *chore.CurrentAssigneeID != requesterID {
		return nil, fmt.Errorf("%w: only the currently assigned member can propose a chore swap", ErrBadRequest)
	}

	if chore.Status == models.ChoreStatusCompleted {
		return nil, fmt.Errorf("%w: cannot swap an already completed chore", ErrBadRequest)
	}

	for _, existing := range s.swaps {
		if existing.ChoreID == chore.ID && existing.Status == models.SwapStatusPending {
			return nil, fmt.Errorf("%w: chore '%s' already has an active pending swap proposal", ErrBadRequest, chore.Title)
		}
	}

	if targetMemberID != nil {
		target, okT := s.members[*targetMemberID]
		if !okT {
			return nil, ErrNotFound
		}
		if target.HouseholdID != chore.HouseholdID {
			return nil, fmt.Errorf("%w: target member does not belong to the household", ErrBadRequest)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id := generateID("swap")

	sw := models.ChoreSwapRequest{
		ID:             id,
		HouseholdID:    chore.HouseholdID,
		ChoreID:        chore.ID,
		RequesterID:    requesterID,
		TargetMemberID: targetMemberID,
		Status:         models.SwapStatusPending,
		Reason:         reason,
		CreatedAt:      now,
	}
	s.swaps[id] = sw

	targetDesc := "on the Household Swap Board"
	if targetMemberID != nil {
		if target, okT := s.members[*targetMemberID]; okT {
			targetDesc = fmt.Sprintf("with %s", target.Name)
		}
	}

	s.addActivityLocked(chore.HouseholdID, requesterID, targetMemberID, models.ActivitySwapProposed, "swap", id,
		fmt.Sprintf("%s offered to swap \"%s\" %s", requester.Name, chore.Title, targetDesc))

	return &sw, nil
}

func (s *MemoryStore) AcceptSwap(swapID, acceptorMemberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sw, okS := s.swaps[swapID]
	if !okS {
		return ErrNotFound
	}
	if sw.Status != models.SwapStatusPending {
		return fmt.Errorf("%w: swap proposal is not pending (status: %s)", ErrBadRequest, sw.Status)
	}

	chore, okC := s.chores[sw.ChoreID]
	if !okC {
		return ErrNotFound
	}
	acceptor, okA := s.members[acceptorMemberID]
	if !okA {
		return ErrNotFound
	}
	if acceptor.HouseholdID != sw.HouseholdID {
		return fmt.Errorf("%w: member does not belong to household '%s'", ErrForbidden, sw.HouseholdID)
	}
	if sw.TargetMemberID != nil && *sw.TargetMemberID != acceptorMemberID {
		return fmt.Errorf("%w: swap proposal is targeted to another member", ErrForbidden)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	chore.CurrentAssigneeID = &acceptorMemberID
	chore.Status = models.ChoreStatusAssigned
	chore.UpdatedAt = now
	s.chores[sw.ChoreID] = chore

	sw.Status = models.SwapStatusAccepted
	sw.ResolvedAt = &now
	s.swaps[swapID] = sw

	s.addActivityLocked(sw.HouseholdID, acceptorMemberID, &sw.RequesterID, models.ActivitySwapAccepted, "swap", sw.ID,
		fmt.Sprintf("%s accepted the swap for \"%s\"", acceptor.Name, chore.Title))

	return nil
}

func (s *MemoryStore) RejectSwap(swapID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sw, okS := s.swaps[swapID]
	if !okS {
		return ErrNotFound
	}
	if sw.Status != models.SwapStatusPending {
		return fmt.Errorf("%w: swap proposal is not pending (status: %s)", ErrBadRequest, sw.Status)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	sw.Status = models.SwapStatusRejected
	sw.ResolvedAt = &now
	s.swaps[swapID] = sw
	return nil
}

func (s *MemoryStore) CancelSwap(swapID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sw, okS := s.swaps[swapID]
	if !okS {
		return ErrNotFound
	}
	if sw.Status != models.SwapStatusPending {
		return fmt.Errorf("%w: swap proposal is not pending (status: %s)", ErrBadRequest, sw.Status)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	sw.Status = models.SwapStatusCancelled
	sw.ResolvedAt = &now
	s.swaps[swapID] = sw
	return nil
}

// ----------------- REWARDS -----------------

func (s *MemoryStore) GetRewards(householdID string) []models.RewardItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]models.RewardItem, 0)
	for _, r := range s.rewards {
		if r.HouseholdID == householdID && r.IsActive {
			list = append(list, r)
		}
	}
	return list
}

func (s *MemoryStore) GetReward(rewardID string) (*models.RewardItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.rewards[rewardID]
	if !ok {
		return nil, ErrNotFound
	}
	return &r, nil
}

func (s *MemoryStore) CreateReward(householdID string, req models.RewardItemCreateRequest) (*models.RewardItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.households[householdID]; !ok {
		return nil, ErrNotFound
	}

	id := generateID("rew")
	icon := req.Icon
	if icon == nil || *icon == "" {
		defIcon := "Gift"
		icon = &defIcon
	}

	r := models.RewardItem{
		ID:          id,
		HouseholdID: householdID,
		Title:       req.Title,
		Description: req.Description,
		PointsCost:  req.PointsCost,
		Icon:        icon,
		IsActive:    true,
	}

	s.rewards[id] = r
	return &r, nil
}

func (s *MemoryStore) RedeemReward(rewardID, memberID string) (*models.RewardRedemption, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reward, okR := s.rewards[rewardID]
	if !okR {
		return nil, ErrNotFound
	}
	member, okM := s.members[memberID]
	if !okM {
		return nil, ErrNotFound
	}

	if !reward.IsActive {
		return nil, fmt.Errorf("%w: reward '%s' is inactive", ErrBadRequest, reward.Title)
	}

	if reward.HouseholdID != member.HouseholdID {
		return nil, fmt.Errorf("%w: member does not belong to the household offering this reward", ErrForbidden)
	}

	if member.PointsBalance < reward.PointsCost {
		return nil, fmt.Errorf("%w: insufficient points: need %d pts, have %d pts", ErrBadRequest, reward.PointsCost, member.PointsBalance)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	member.PointsBalance -= reward.PointsCost
	member.UpdatedAt = now
	s.members[memberID] = member

	redID := generateID("red")
	red := models.RewardRedemption{
		ID:           redID,
		HouseholdID:  reward.HouseholdID,
		MemberID:     memberID,
		RewardItemID: rewardID,
		PointsSpent:  reward.PointsCost,
		Status:       models.RedemptionStatusRequested,
		RequestedAt:  now,
	}
	s.redemptions[redID] = red

	s.addActivityLocked(reward.HouseholdID, memberID, nil, models.ActivityRewardRedeemed, "reward", reward.ID,
		fmt.Sprintf("%s redeemed \"%s\" for %d pts", member.Name, reward.Title, reward.PointsCost))

	return &red, nil
}

func (s *MemoryStore) FulfillReward(redemptionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	red, ok := s.redemptions[redemptionID]
	if !ok {
		return ErrNotFound
	}
	if red.Status == models.RedemptionStatusFulfilled {
		return fmt.Errorf("%w: redemption is already fulfilled", ErrBadRequest)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	red.Status = models.RedemptionStatusFulfilled
	red.FulfilledAt = &now
	s.redemptions[redemptionID] = red

	rewardTitle := "Reward"
	if r, okR := s.rewards[red.RewardItemID]; okR {
		rewardTitle = r.Title
	}
	memberName := "Member"
	if m, okM := s.members[red.MemberID]; okM {
		memberName = m.Name
	}

	s.addActivityLocked(red.HouseholdID, "system", &red.MemberID, models.ActivityRewardFulfilled, "reward", red.ID,
		fmt.Sprintf("Reward \"%s\" was fulfilled for %s!", rewardTitle, memberName))

	return nil
}

func (s *MemoryStore) GetRedemptions(householdID string) []models.RewardRedemptionItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]models.RewardRedemptionItem, 0)
	for _, red := range s.redemptions {
		if red.HouseholdID == householdID {
			reward, okR := s.rewards[red.RewardItemID]
			member, okM := s.members[red.MemberID]
			if okR && okM {
				list = append(list, models.RewardRedemptionItem{
					Redemption: red,
					Reward:     reward,
					Member:     member,
				})
			}
		}
	}
	return list
}

// ----------------- ACTIVITIES & NUDGES -----------------

func (s *MemoryStore) GetActivitiesFiltered(householdID string, entityType, eventType string, limit, offset int) []models.ActivityLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]models.ActivityLog, 0)
	for i := len(s.activities) - 1; i >= 0; i-- {
		act := s.activities[i]
		if act.HouseholdID != householdID {
			continue
		}
		if entityType != "" && act.EntityType != entityType {
			continue
		}
		if eventType != "" && act.EventType != eventType {
			continue
		}
		filtered = append(filtered, act)
	}

	if offset < 0 {
		offset = 0
	}
	if offset >= len(filtered) {
		return []models.ActivityLog{}
	}

	sliced := filtered[offset:]
	if limit > 0 && limit < len(sliced) {
		sliced = sliced[:limit]
	}

	return sliced
}

func (s *MemoryStore) GetActivities(householdID string) []models.ActivityLog {
	return s.GetActivitiesFiltered(householdID, "", "", 0, 0)
}

func (s *MemoryStore) SendNudge(choreID, senderMemberID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chore, okC := s.chores[choreID]
	if !okC {
		return "", ErrNotFound
	}
	sender, okS := s.members[senderMemberID]
	if !okS {
		return "", ErrNotFound
	}

	var recipientName string
	var recipientID *string

	if chore.CurrentAssigneeID != nil {
		recipientID = chore.CurrentAssigneeID
		if a, okA := s.members[*chore.CurrentAssigneeID]; okA {
			recipientName = a.Name
		}
	}

	var msg string
	if recipientName != "" {
		msg = fmt.Sprintf("Friendly nudge from %s to %s: \"%s\" is due soon!", sender.Name, recipientName, chore.Title)
	} else {
		msg = fmt.Sprintf("Friendly reminder: Open chore \"%s\" is ready for someone to claim!", chore.Title)
	}

	s.addActivityLocked(chore.HouseholdID, senderMemberID, recipientID, models.ActivityChoreNudge, "chore", chore.ID, msg)
	return msg, nil
}

func (s *MemoryStore) GetComments(choreID string) []models.ChoreCommentWithMember {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]models.ChoreCommentWithMember, 0)
	for _, comm := range s.comments {
		if comm.ChoreID == choreID {
			var memberPtr *models.Member
			if m, ok := s.members[comm.MemberID]; ok {
				memberPtr = &m
			}
			list = append(list, models.ChoreCommentWithMember{
				Comment: comm,
				Member:  memberPtr,
			})
		}
	}
	return list
}

func (s *MemoryStore) AddComment(choreID, memberID, message string) (*models.ChoreComment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.chores[choreID]; !ok {
		return nil, ErrNotFound
	}
	if _, ok := s.members[memberID]; !ok {
		return nil, ErrNotFound
	}

	comm := models.ChoreComment{
		ID:        generateID("comm"),
		ChoreID:   choreID,
		MemberID:  memberID,
		Message:   strings.TrimSpace(message),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	s.comments = append(s.comments, comm)
	return &comm, nil
}

// ----------------- INTERNAL HELPERS -----------------

func (s *MemoryStore) addActivityLocked(householdID, actorID string, recipientID *string, eventType, entityType, entityID, message string) {
	act := models.ActivityLog{
		ID:          generateID("act"),
		HouseholdID: householdID,
		RecipientID: recipientID,
		ActorID:     actorID,
		EventType:   eventType,
		EntityType:  entityType,
		EntityID:    entityID,
		Message:     message,
		Read:        false,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	s.activities = append(s.activities, act)
}

func (s *MemoryStore) advanceRecurrenceLocked(chore models.Chore) {
	if chore.RecurrenceType == models.ChoreRecurrenceNone {
		return
	}

	baseDate := time.Now().UTC()
	if chore.DueDate != nil {
		if parsed, err := time.Parse(time.RFC3339, *chore.DueDate); err == nil {
			baseDate = parsed
		}
	}

	var nextDate time.Time
	switch chore.RecurrenceType {
	case models.ChoreRecurrenceDaily:
		nextDate = baseDate.AddDate(0, 0, 1)
	case models.ChoreRecurrenceWeekly, models.ChoreRecurrenceCustomDays:
		nextDate = baseDate.AddDate(0, 0, 7)
	case models.ChoreRecurrenceMonthly:
		nextDate = baseDate.AddDate(0, 1, 0)
	default:
		return
	}

	nextDueDate := nextDate.Format(time.RFC3339)
	nextAssigneeID := chore.CurrentAssigneeID
	nextIndex := 0
	if chore.CurrentRotationIndex != nil {
		nextIndex = *chore.CurrentRotationIndex
	}

	status := models.ChoreStatusAssigned
	if chore.AssignmentType == models.ChoreAssignmentRoundRobin && len(chore.RotationMemberIDs) > 0 {
		nextIndex = (nextIndex + 1) % len(chore.RotationMemberIDs)
		chosen := chore.RotationMemberIDs[nextIndex]
		nextAssigneeID = &chosen
	} else if chore.AssignmentType == models.ChoreAssignmentOpenPool {
		nextAssigneeID = nil
		status = models.ChoreStatusUnassigned
	}

	now := time.Now().UTC().Format(time.RFC3339)
	nextChoreID := generateID("c")

	nextChore := models.Chore{
		ID:                   nextChoreID,
		HouseholdID:          chore.HouseholdID,
		Title:                chore.Title,
		Description:          chore.Description,
		Category:             chore.Category,
		EffortPoints:         chore.EffortPoints,
		AssignmentType:       chore.AssignmentType,
		CurrentAssigneeID:    nextAssigneeID,
		RotationMemberIDs:    chore.RotationMemberIDs,
		CurrentRotationIndex: &nextIndex,
		RecurrenceType:       chore.RecurrenceType,
		RecurrenceRule:       chore.RecurrenceRule,
		DueDate:              &nextDueDate,
		RequiresApproval:     chore.RequiresApproval,
		RequiresProof:        chore.RequiresProof,
		Status:               status,
		CreatedBy:            chore.CreatedBy,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	s.chores[nextChoreID] = nextChore

	targetDesc := "Placed in Open Pool"
	if nextAssigneeID != nil {
		if m, ok := s.members[*nextAssigneeID]; ok {
			targetDesc = fmt.Sprintf("Assigned to %s", m.Name)
		}
	}

	s.addActivityLocked(chore.HouseholdID, chore.CreatedBy, nextAssigneeID, models.ActivityChoreCreated, "chore", nextChoreID,
		fmt.Sprintf("Scheduled next instance for \"%s\". %s", chore.Title, targetDesc))
}

// ----------------- AUTH & PIN -----------------

// VerifyAdminPIN verifies the 4-digit admin PIN for sensitive household actions.
func (s *MemoryStore) VerifyAdminPIN(householdID, pin string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	h, ok := s.households[householdID]
	if !ok {
		return false, ErrNotFound
	}

	cleanPIN := strings.TrimSpace(pin)
	if h.Settings.AdminPIN == "" {
		return cleanPIN == "1234", nil
	}

	return h.Settings.AdminPIN == cleanPIN, nil
}

// GetMemberByEmail looks up a member and household by email address.
func (s *MemoryStore) GetMemberByEmail(email string) (*models.Member, *models.Household, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	for _, m := range s.members {
		if m.Email != nil && strings.ToLower(*m.Email) == cleanEmail {
			h := s.households[m.HouseholdID]
			memCopy := m
			return &memCopy, &h, nil
		}
	}
	return nil, nil, ErrNotFound
}

// CreateMagicLink generates a login token for the member with the given email.
func (s *MemoryStore) CreateMagicLink(email string) (string, *models.Member, *models.Household, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	var foundMember *models.Member
	for _, m := range s.members {
		if m.Email != nil && strings.ToLower(*m.Email) == cleanEmail {
			memCopy := m
			foundMember = &memCopy
			break
		}
	}

	if foundMember == nil {
		return "", nil, nil, ErrNotFound
	}

	h, ok := s.households[foundMember.HouseholdID]
	if !ok {
		return "", nil, nil, ErrNotFound
	}

	token := fmt.Sprintf("MAGIC-%06d", time.Now().UnixNano()%1000000)
	now := time.Now().UTC()
	s.magicLinks[token] = magicLinkEntry{
		memberID:  foundMember.ID,
		createdAt: now,
		expiresAt: now.Add(15 * time.Minute),
	}

	return token, foundMember, &h, nil
}

// VerifyMagicLink verifies a token and returns the member and household.
func (s *MemoryStore) VerifyMagicLink(token string) (*models.Member, *models.Household, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanToken := strings.TrimSpace(strings.ToUpper(token))
	entry, ok := s.magicLinks[cleanToken]
	actualKey := cleanToken
	if !ok {
		rawToken := strings.TrimSpace(token)
		entry, ok = s.magicLinks[rawToken]
		if !ok {
			return nil, nil, errors.New("INVALID_OR_EXPIRED_TOKEN")
		}
		actualKey = rawToken
	}

	now := time.Now().UTC()
	isDemoToken := actualKey == "MAGIC-SARAH" || actualKey == "MAGIC-DAVID" || actualKey == "MAGIC-ALEX"

	// 1. Time-Based Expiration check (15 minutes TTL for generated tokens)
	if !isDemoToken && (now.After(entry.expiresAt) || now.Sub(entry.createdAt) > 15*time.Minute) {
		delete(s.magicLinks, actualKey)
		return nil, nil, errors.New("INVALID_OR_EXPIRED_TOKEN")
	}

	// 2. Single-Use Invalidation: Delete token immediately after login (preserve static demo tokens)
	if !isDemoToken {
		delete(s.magicLinks, actualKey)
	}

	m, okM := s.members[entry.memberID]
	if !okM {
		return nil, nil, ErrNotFound
	}

	h, okH := s.households[m.HouseholdID]
	if !okH {
		return nil, nil, ErrNotFound
	}

	memCopy := m
	hCopy := h
	return &memCopy, &hCopy, nil
}

// GetDemoMember returns member and household for demo fast-switching.
func (s *MemoryStore) GetDemoMember(memberID string) (*models.Member, *models.Household, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, okM := s.members[memberID]
	if !okM {
		return nil, nil, ErrNotFound
	}

	h, okH := s.households[m.HouseholdID]
	if !okH {
		return nil, nil, ErrNotFound
	}

	memCopy := m
	hCopy := h
	return &memCopy, &hCopy, nil
}

// VerifyPassword checks if the provided password matches the member's credentials.
func (s *MemoryStore) VerifyPassword(memberID, password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, ok := s.members[memberID]
	if !ok {
		return false
	}

	// Default demo password is "password123" for seed accounts
	if m.PasswordHash == "" {
		return password == "password123"
	}

	err := bcrypt.CompareHashAndPassword([]byte(m.PasswordHash), []byte(password))
	return err == nil
}

// RegisterMember registers a new member and either creates a new household or attaches to an existing one.
func (s *MemoryStore) RegisterMember(req models.RegisterRequest) (*models.Member, *models.Household, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	// Check duplicate email
	for _, m := range s.members {
		if m.Email != nil && strings.ToLower(*m.Email) == cleanEmail {
			return nil, nil, fmt.Errorf("email '%s' is already registered", cleanEmail)
		}
	}

	var household models.Household
	role := req.Role
	if role == "" {
		role = models.MemberRoleAdmin
	}

	// Option A: Join existing household via InviteCode
	if strings.TrimSpace(req.InviteCode) != "" {
		cleanInvite := strings.ToUpper(strings.TrimSpace(req.InviteCode))
		var foundH *models.Household
		for _, h := range s.households {
			if strings.ToUpper(h.InviteCode) == cleanInvite {
				hCopy := h
				foundH = &hCopy
				break
			}
		}
		if foundH == nil {
			return nil, nil, fmt.Errorf("invalid invite code '%s'", req.InviteCode)
		}
		household = *foundH
		if req.Role == "" {
			role = models.MemberRoleMember
		}
	} else {
		// Option B: Create new household
		hName := strings.TrimSpace(req.HouseholdName)
		if hName == "" {
			hName = fmt.Sprintf("%s's Home", req.Name)
		}
		newHID := fmt.Sprintf("h-%d", time.Now().UnixNano()%1000000)
		inviteCode := fmt.Sprintf("HOME-%04d", time.Now().UnixNano()%10000)
		household = models.Household{
			ID:         newHID,
			Name:       hName,
			InviteCode: inviteCode,
			Settings: models.HouseholdSettings{
				DefaultMode: models.HouseholdModeFlatmate,
				AllowSwaps:  true,
				Timezone:    "America/New_York",
				AdminPIN:    "1234",
			},
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		s.households[newHID] = household
		role = models.MemberRoleAdmin
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newMID := fmt.Sprintf("m-%d", time.Now().UnixNano()%1000000)
	avatar := "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=120&auto=format&fit=crop&q=80"
	nowStr := time.Now().UTC().Format(time.RFC3339)

	newMember := models.Member{
		ID:                newMID,
		HouseholdID:       household.ID,
		Name:              req.Name,
		Email:             &cleanEmail,
		PasswordHash:      string(hash),
		AvatarURL:         &avatar,
		Role:              role,
		PointsBalance:     0,
		TotalPointsEarned: 0,
		Streak:            0,
		CreatedAt:         nowStr,
		UpdatedAt:         nowStr,
	}
	s.members[newMID] = newMember

	memCopy := newMember
	hCopy := household
	return &memCopy, &hCopy, nil
}

func (s *MemoryStore) CreatePasswordResetToken(email string) (string, *models.Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	var target *models.Member
	for _, m := range s.members {
		if m.Email != nil && strings.ToLower(strings.TrimSpace(*m.Email)) == cleanEmail {
			target = &m
			break
		}
	}
	if target == nil {
		return "", nil, ErrNotFound
	}

	token := fmt.Sprintf("RST-%d", time.Now().UnixNano()%100000000)
	s.passwordResets[token] = passwordResetEntry{
		memberID:  target.ID,
		expiresAt: time.Now().Add(1 * time.Hour),
	}

	memCopy := *target
	return token, &memCopy, nil
}

func (s *MemoryStore) ResetPasswordWithToken(token, newPassword string) (*models.Member, *models.Household, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.passwordResets[token]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, nil, ErrForbidden
	}

	member, ok := s.members[entry.memberID]
	if !ok {
		return nil, nil, ErrNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	member.PasswordHash = string(hash)
	member.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.members[entry.memberID] = member

	delete(s.passwordResets, token)

	household, ok := s.households[member.HouseholdID]
	if !ok {
		return nil, nil, ErrNotFound
	}

	memCopy := member
	hCopy := household
	return &memCopy, &hCopy, nil
}

// UpdatePassword updates a member's password hash in memory.
func (s *MemoryStore) UpdatePassword(memberID, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	member, ok := s.members[memberID]
	if !ok {
		return ErrNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	member.PasswordHash = string(hash)
	member.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.members[memberID] = member
	return nil
}



