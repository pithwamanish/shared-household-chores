package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/choresync/backend/internal/db"
	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/telemetry"
)

// PostgresStore implements the Store interface backed by PostgreSQL and pgx/v5.
type PostgresStore struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// NewPostgresStore initializes a connection pool, runs migrations, and seeds if empty.
func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	// Attach OpenTelemetry query tracer for distributed DB spans
	config.ConnConfig.Tracer = telemetry.NewPGXQueryTracer()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Apply database DDL migrations
	if _, err := pool.Exec(ctx, db.SchemaSQL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to run schema migrations: %w", err)
	}

	// Ensure migration additions for magic_links (expires_at column)
	_, _ = pool.Exec(ctx, `
		ALTER TABLE magic_links ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '15 minutes');
	`)

	ps := &PostgresStore{
		pool: pool,
		q:    db.New(pool),
	}

	// Check if initial seeding is needed
	households, err := ps.q.GetHouseholds(ctx)
	if err != nil || len(households) == 0 {
		ps.Seed()
	}

	return ps, nil
}

func (p *PostgresStore) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

// ----------------- TYPE CONVERSION HELPERS -----------------

func textToPg(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func strPtrFromPg(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func intToPg(i *int) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*i), Valid: true}
}

func intPtrFromPg(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	val := int(i.Int32)
	return &val
}

func timeToPg(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func timePtrToPg(s *string) pgtype.Timestamptz {
	if s == nil || *s == "" {
		return pgtype.Timestamptz{Valid: false}
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func timeFromPg(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

func (p *PostgresStore) householdFromDB(h db.Household) models.Household {
	return models.Household{
		ID:         h.ID,
		Name:       h.Name,
		InviteCode: h.InviteCode,
		Settings: models.HouseholdSettings{
			DefaultMode: h.DefaultMode,
			AllowSwaps:  h.AllowSwaps,
			Timezone:    h.Timezone,
			AdminPIN:    h.AdminPin,
		},
		CreatedAt: h.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: h.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func (p *PostgresStore) memberFromDB(m db.Member) models.Member {
	return models.Member{
		ID:                m.ID,
		HouseholdID:       m.HouseholdID,
		Name:              m.Name,
		Email:             strPtrFromPg(m.Email),
		AvatarURL:         strPtrFromPg(m.AvatarUrl),
		Role:              m.Role,
		PointsBalance:     int(m.PointsBalance),
		TotalPointsEarned: int(m.TotalPointsEarned),
		Streak:            int(m.Streak),
		CreatedAt:         m.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:         m.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func (p *PostgresStore) choreFromDB(c db.Chore) models.Chore {
	rotationMembers := make([]string, 0)
	if len(c.RotationMemberIds) > 0 {
		_ = json.Unmarshal(c.RotationMemberIds, &rotationMembers)
	}
	if rotationMembers == nil {
		rotationMembers = make([]string, 0)
	}

	var recRule *models.ChoreRecurrenceRule
	if len(c.RecurrenceRule) > 0 {
		_ = json.Unmarshal(c.RecurrenceRule, &recRule)
	}

	return models.Chore{
		ID:                   c.ID,
		HouseholdID:          c.HouseholdID,
		Title:                c.Title,
		Description:          strPtrFromPg(c.Description),
		Category:             c.Category,
		EffortPoints:         int(c.EffortPoints),
		AssignmentType:       c.AssignmentType,
		CurrentAssigneeID:    strPtrFromPg(c.CurrentAssigneeID),
		RotationMemberIDs:    rotationMembers,
		CurrentRotationIndex: intPtrFromPg(c.CurrentRotationIndex),
		RecurrenceType:       c.RecurrenceType,
		RecurrenceRule:       recRule,
		DueDate:              timeFromPg(c.DueDate),
		RequiresApproval:     c.RequiresApproval,
		RequiresProof:        c.RequiresProof,
		Status:               c.Status,
		CreatedBy:            c.CreatedBy,
		CreatedAt:            c.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:            c.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func (p *PostgresStore) completionFromDB(c db.ChoreCompletion) models.ChoreCompletion {
	return models.ChoreCompletion{
		ID:              c.ID,
		ChoreID:         c.ChoreID,
		HouseholdID:     c.HouseholdID,
		MemberID:        c.MemberID,
		Status:          c.Status,
		ProofNotes:      strPtrFromPg(c.ProofNotes),
		ProofPhotoURL:   strPtrFromPg(c.ProofPhotoUrl),
		ApprovedBy:      strPtrFromPg(c.ApprovedBy),
		RejectionReason: strPtrFromPg(c.RejectionReason),
		PointsAwarded:   int(c.PointsAwarded),
		CompletedAt:     c.CompletedAt.Time.Format(time.RFC3339),
		VerifiedAt:      timeFromPg(c.VerifiedAt),
	}
}

func (p *PostgresStore) swapFromDB(s db.ChoreSwapRequest) models.ChoreSwapRequest {
	return models.ChoreSwapRequest{
		ID:             s.ID,
		ChoreID:        s.ChoreID,
		HouseholdID:    s.HouseholdID,
		RequesterID:    s.RequesterID,
		TargetMemberID: strPtrFromPg(s.TargetMemberID),
		Reason:         strPtrFromPg(s.Reason),
		Status:         s.Status,
		CreatedAt:      s.CreatedAt.Time.Format(time.RFC3339),
		ResolvedAt:     timeFromPg(s.ResolvedAt),
	}
}

func (p *PostgresStore) rewardFromDB(r db.RewardItem) models.RewardItem {
	return models.RewardItem{
		ID:          r.ID,
		HouseholdID: r.HouseholdID,
		Title:       r.Title,
		Description: r.Description,
		PointsCost:  int(r.PointsCost),
		Icon:        strPtrFromPg(r.Icon),
		IsActive:    r.IsActive,
	}
}

func (p *PostgresStore) redemptionFromDB(red db.RewardRedemption) models.RewardRedemption {
	return models.RewardRedemption{
		ID:           red.ID,
		RewardItemID: red.RewardItemID,
		HouseholdID:  red.HouseholdID,
		MemberID:     red.MemberID,
		PointsSpent:  int(red.PointsSpent),
		Status:       red.Status,
		RequestedAt:  red.RequestedAt.Time.Format(time.RFC3339),
		FulfilledAt:  timeFromPg(red.FulfilledAt),
	}
}

func (p *PostgresStore) activityFromDB(a db.ActivityLog) models.ActivityLog {
	return models.ActivityLog{
		ID:          a.ID,
		HouseholdID: a.HouseholdID,
		ActorID:     a.ActorID,
		RecipientID: strPtrFromPg(a.RecipientID),
		EventType:   a.EventType,
		EntityType:  a.EntityType,
		EntityID:    a.EntityID,
		Message:     a.Message,
		Read:        a.Read,
		CreatedAt:   a.CreatedAt.Time.Format(time.RFC3339),
	}
}

// ----------------- SEEDING & RESET -----------------

// Reset truncates all tables and re-seeds the standard fixture data.
func (p *PostgresStore) Reset() {
	ctx := context.Background()
	_ = p.q.TruncateAll(ctx)
	p.Seed()
}

// Seed populates the database with initial fixtures.
func (p *PostgresStore) Seed() {
	ctx := context.Background()
	households, members, chores, completions, swaps, rewards, redemptions, activities, comments := getInitialData()

	for _, h := range households {
		cTime, _ := time.Parse(time.RFC3339, h.CreatedAt)
		uTime, _ := time.Parse(time.RFC3339, h.UpdatedAt)
		_, _ = p.q.CreateHousehold(ctx, db.CreateHouseholdParams{
			ID:          h.ID,
			Name:        h.Name,
			InviteCode:  h.InviteCode,
			DefaultMode: h.Settings.DefaultMode,
			AllowSwaps:  h.Settings.AllowSwaps,
			Timezone:    h.Settings.Timezone,
			AdminPin:    h.Settings.AdminPIN,
			CreatedAt:   timeToPg(cTime),
			UpdatedAt:   timeToPg(uTime),
		})
	}

	for _, m := range members {
		cTime, _ := time.Parse(time.RFC3339, m.CreatedAt)
		uTime, _ := time.Parse(time.RFC3339, m.UpdatedAt)
		_, _ = p.q.CreateMember(ctx, db.CreateMemberParams{
			ID:                m.ID,
			HouseholdID:       m.HouseholdID,
			Name:              m.Name,
			Email:             textToPg(m.Email),
			AvatarUrl:         textToPg(m.AvatarURL),
			Role:              m.Role,
			PointsBalance:     int32(m.PointsBalance),
			TotalPointsEarned: int32(m.TotalPointsEarned),
			Streak:            int32(m.Streak),
			CreatedAt:         timeToPg(cTime),
			UpdatedAt:         timeToPg(uTime),
		})
	}

	for _, c := range chores {
		cTime, _ := time.Parse(time.RFC3339, c.CreatedAt)
		uTime, _ := time.Parse(time.RFC3339, c.UpdatedAt)
		rotJSON, _ := json.Marshal(c.RotationMemberIDs)
		var recJSON []byte
		if c.RecurrenceRule != nil {
			recJSON, _ = json.Marshal(c.RecurrenceRule)
		}

		_, _ = p.q.CreateChore(ctx, db.CreateChoreParams{
			ID:                   c.ID,
			HouseholdID:          c.HouseholdID,
			Title:                c.Title,
			Description:          textToPg(c.Description),
			Category:             c.Category,
			EffortPoints:         int32(c.EffortPoints),
			AssignmentType:       c.AssignmentType,
			CurrentAssigneeID:    textToPg(c.CurrentAssigneeID),
			RotationMemberIds:    rotJSON,
			CurrentRotationIndex: intToPg(c.CurrentRotationIndex),
			RecurrenceType:       c.RecurrenceType,
			RecurrenceRule:       recJSON,
			DueDate:              timePtrToPg(c.DueDate),
			RequiresApproval:     c.RequiresApproval,
			RequiresProof:        c.RequiresProof,
			Status:               c.Status,
			CreatedBy:            c.CreatedBy,
			CreatedAt:            timeToPg(cTime),
			UpdatedAt:            timeToPg(uTime),
		})
	}

	for _, comp := range completions {
		compTime, _ := time.Parse(time.RFC3339, comp.CompletedAt)
		_, _ = p.q.CreateCompletion(ctx, db.CreateCompletionParams{
			ID:              comp.ID,
			ChoreID:         comp.ChoreID,
			HouseholdID:     comp.HouseholdID,
			MemberID:        comp.MemberID,
			Status:          comp.Status,
			ProofNotes:      textToPg(comp.ProofNotes),
			ProofPhotoUrl:   textToPg(comp.ProofPhotoURL),
			ApprovedBy:      textToPg(comp.ApprovedBy),
			RejectionReason: textToPg(comp.RejectionReason),
			PointsAwarded:   int32(comp.PointsAwarded),
			CompletedAt:     timeToPg(compTime),
			VerifiedAt:      timePtrToPg(comp.VerifiedAt),
		})
	}

	for _, sw := range swaps {
		cTime, _ := time.Parse(time.RFC3339, sw.CreatedAt)
		_, _ = p.q.CreateSwap(ctx, db.CreateSwapParams{
			ID:             sw.ID,
			HouseholdID:    sw.HouseholdID,
			ChoreID:        sw.ChoreID,
			RequesterID:    sw.RequesterID,
			TargetMemberID: textToPg(sw.TargetMemberID),
			Status:         sw.Status,
			Reason:         textToPg(sw.Reason),
			CreatedAt:      timeToPg(cTime),
			ResolvedAt:     timePtrToPg(sw.ResolvedAt),
		})
	}

	for _, r := range rewards {
		now := time.Now().UTC()
		_, _ = p.q.CreateReward(ctx, db.CreateRewardParams{
			ID:          r.ID,
			HouseholdID: r.HouseholdID,
			Title:       r.Title,
			Description: r.Description,
			PointsCost:  int32(r.PointsCost),
			Icon:        textToPg(r.Icon),
			IsActive:    r.IsActive,
			CreatedAt:   timeToPg(now),
		})
	}

	for _, red := range redemptions {
		cTime, _ := time.Parse(time.RFC3339, red.RequestedAt)
		_, _ = p.q.CreateRedemption(ctx, db.CreateRedemptionParams{
			ID:           red.ID,
			HouseholdID:  red.HouseholdID,
			MemberID:     red.MemberID,
			RewardItemID: red.RewardItemID,
			PointsSpent:  int32(red.PointsSpent),
			Status:       red.Status,
			RequestedAt:  timeToPg(cTime),
			FulfilledAt:  timePtrToPg(red.FulfilledAt),
		})
	}

	for _, a := range activities {
		cTime, _ := time.Parse(time.RFC3339, a.CreatedAt)
		_, _ = p.q.CreateActivityLog(ctx, db.CreateActivityLogParams{
			ID:          a.ID,
			HouseholdID: a.HouseholdID,
			RecipientID: textToPg(a.RecipientID),
			ActorID:     a.ActorID,
			EventType:   a.EventType,
			EntityType:  a.EntityType,
			EntityID:    a.EntityID,
			Message:     a.Message,
			Read:        a.Read,
			CreatedAt:   timeToPg(cTime),
		})
	}

	for _, com := range comments {
		cTime, _ := time.Parse(time.RFC3339, com.CreatedAt)
		_, _ = p.q.CreateComment(ctx, db.CreateCommentParams{
			ID:        com.ID,
			ChoreID:   com.ChoreID,
			MemberID:  com.MemberID,
			Message:   com.Message,
			CreatedAt: timeToPg(cTime),
		})
	}

	// Insert default magic links
	now := time.Now().UTC()
	_, _ = p.q.UpsertMagicLink(ctx, db.UpsertMagicLinkParams{Token: "MAGIC-SARAH", MemberID: "m-sarah", CreatedAt: timeToPg(now)})
	_, _ = p.q.UpsertMagicLink(ctx, db.UpsertMagicLinkParams{Token: "MAGIC-DAVID", MemberID: "m-david-fam", CreatedAt: timeToPg(now)})
	_, _ = p.q.UpsertMagicLink(ctx, db.UpsertMagicLinkParams{Token: "MAGIC-ALEX", MemberID: "m-alex", CreatedAt: timeToPg(now)})
}

// ----------------- HOUSEHOLDS -----------------

func (p *PostgresStore) GetHouseholds() []models.Household {
	ctx := context.Background()
	rows, err := p.q.GetHouseholds(ctx)
	if err != nil {
		return []models.Household{}
	}
	result := make([]models.Household, len(rows))
	for i, r := range rows {
		result[i] = p.householdFromDB(r)
	}
	return result
}

func (p *PostgresStore) GetHousehold(id string) (*models.Household, error) {
	ctx := context.Background()
	r, err := p.q.GetHousehold(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	h := p.householdFromDB(r)
	return &h, nil
}

func (p *PostgresStore) CreateHousehold(req models.HouseholdCreateRequest) models.Household {
	ctx := context.Background()
	id := generateID("h")
	inviteCode := strings.ToUpper(strings.ReplaceAll(req.Name, " ", "-")) + "-JOIN"

	defaultMode := models.HouseholdModeFlatmate
	if req.DefaultMode != nil {
		defaultMode = *req.DefaultMode
	}

	allowSwaps := true
	if req.AllowSwaps != nil {
		allowSwaps = *req.AllowSwaps
	}

	timezone := "America/New_York"
	if req.Timezone != nil {
		timezone = *req.Timezone
	}

	now := time.Now().UTC()
	r, _ := p.q.CreateHousehold(ctx, db.CreateHouseholdParams{
		ID:          id,
		Name:        req.Name,
		InviteCode:  inviteCode,
		DefaultMode: defaultMode,
		AllowSwaps:  allowSwaps,
		Timezone:    timezone,
		AdminPin:    "1234",
		CreatedAt:   timeToPg(now),
		UpdatedAt:   timeToPg(now),
	})

	return p.householdFromDB(r)
}

func (p *PostgresStore) JoinHousehold(inviteCode string) (*models.Household, error) {
	ctx := context.Background()
	r, err := p.q.GetHouseholdByInviteCode(ctx, inviteCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	h := p.householdFromDB(r)
	return &h, nil
}

func (p *PostgresStore) UpdateHouseholdSettings(id string, update models.HouseholdSettingsUpdate) (*models.Household, error) {
	ctx := context.Background()
	current, err := p.q.GetHousehold(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	defaultMode := current.DefaultMode
	if update.DefaultMode != nil {
		defaultMode = *update.DefaultMode
	}

	allowSwaps := current.AllowSwaps
	if update.AllowSwaps != nil {
		allowSwaps = *update.AllowSwaps
	}

	timezone := current.Timezone
	if update.Timezone != nil {
		timezone = *update.Timezone
	}

	adminPin := current.AdminPin
	if update.AdminPIN != nil {
		adminPin = *update.AdminPIN
	}

	r, err := p.q.UpdateHouseholdSettings(ctx, db.UpdateHouseholdSettingsParams{
		ID:          id,
		DefaultMode: defaultMode,
		AllowSwaps:  allowSwaps,
		Timezone:    timezone,
		AdminPin:    adminPin,
	})
	if err != nil {
		return nil, err
	}

	h := p.householdFromDB(r)
	return &h, nil
}

// ----------------- MEMBERS -----------------

func (p *PostgresStore) GetMembers(householdID string) []models.Member {
	ctx := context.Background()
	rows, err := p.q.GetMembersByHousehold(ctx, householdID)
	if err != nil {
		return []models.Member{}
	}
	result := make([]models.Member, len(rows))
	for i, r := range rows {
		result[i] = p.memberFromDB(r)
	}
	return result
}

func (p *PostgresStore) GetMember(id string) (*models.Member, error) {
	ctx := context.Background()
	r, err := p.q.GetMember(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	m := p.memberFromDB(r)
	return &m, nil
}

func (p *PostgresStore) AddMember(householdID string, req models.MemberCreateRequest) (*models.Member, error) {
	ctx := context.Background()
	if _, err := p.GetHousehold(householdID); err != nil {
		return nil, ErrNotFound
	}

	id := generateID("m")
	now := time.Now().UTC()
	r, err := p.q.CreateMember(ctx, db.CreateMemberParams{
		ID:                id,
		HouseholdID:       householdID,
		Name:              req.Name,
		Email:             textToPg(req.Email),
		AvatarUrl:         textToPg(req.AvatarURL),
		Role:              req.Role,
		PointsBalance:     0,
		TotalPointsEarned: 0,
		Streak:            0,
		CreatedAt:         timeToPg(now),
		UpdatedAt:         timeToPg(now),
	})
	if err != nil {
		return nil, err
	}

	m := p.memberFromDB(r)
	return &m, nil
}

// ----------------- CHORES -----------------

func (p *PostgresStore) GetChores(householdID, status, category, search, view, assigneeID string) []models.Chore {
	ctx := context.Background()
	rows, err := p.q.GetChoresByHousehold(ctx, householdID)
	if err != nil {
		return []models.Chore{}
	}

	result := make([]models.Chore, 0)
	for _, r := range rows {
		c := p.choreFromDB(r)

		if status != "" && c.Status != status {
			continue
		}
		if category != "" && c.Category != category {
			continue
		}
		if assigneeID != "" && (c.CurrentAssigneeID == nil || *c.CurrentAssigneeID != assigneeID) {
			continue
		}
		if search != "" {
			query := strings.ToLower(search)
			titleMatch := strings.Contains(strings.ToLower(c.Title), query)
			descMatch := c.Description != nil && strings.Contains(strings.ToLower(*c.Description), query)
			if !titleMatch && !descMatch {
				continue
			}
		}
		result = append(result, c)
	}
	return result
}

func (p *PostgresStore) GetChore(id string) (*models.Chore, error) {
	ctx := context.Background()
	r, err := p.q.GetChore(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	c := p.choreFromDB(r)
	return &c, nil
}

func (p *PostgresStore) CreateChore(req models.ChoreCreateRequest) (*models.Chore, error) {
	ctx := context.Background()
	if _, err := p.GetHousehold(req.HouseholdID); err != nil {
		return nil, ErrNotFound
	}

	id := generateID("c")
	now := time.Now().UTC()

	effortPoints := 10
	if req.EffortPoints > 0 {
		effortPoints = req.EffortPoints
	}

	assignmentType := models.ChoreAssignmentDirect
	if req.AssignmentType != "" {
		assignmentType = req.AssignmentType
	}

	status := models.ChoreStatusAssigned
	if req.Status != nil {
		status = *req.Status
	} else if req.CurrentAssigneeID == nil {
		status = models.ChoreStatusUnassigned
	}

	requiresApproval := false
	if req.RequiresApproval != nil {
		requiresApproval = *req.RequiresApproval
	}

	requiresProof := false
	if req.RequiresProof != nil {
		requiresProof = *req.RequiresProof
	}

	rotJSON, _ := json.Marshal(req.RotationMemberIDs)
	var recJSON []byte
	if req.RecurrenceRule != nil {
		recJSON, _ = json.Marshal(req.RecurrenceRule)
	}

	r, err := p.q.CreateChore(ctx, db.CreateChoreParams{
		ID:                   id,
		HouseholdID:          req.HouseholdID,
		Title:                req.Title,
		Description:          textToPg(req.Description),
		Category:             req.Category,
		EffortPoints:         int32(effortPoints),
		AssignmentType:       assignmentType,
		CurrentAssigneeID:    textToPg(req.CurrentAssigneeID),
		RotationMemberIds:    rotJSON,
		CurrentRotationIndex: intToPg(req.CurrentRotationIndex),
		RecurrenceType:       req.RecurrenceType,
		RecurrenceRule:       recJSON,
		DueDate:              timePtrToPg(req.DueDate),
		RequiresApproval:     requiresApproval,
		RequiresProof:        requiresProof,
		Status:               status,
		CreatedBy:            req.CreatedBy,
		CreatedAt:            timeToPg(now),
		UpdatedAt:            timeToPg(now),
	})
	if err != nil {
		return nil, err
	}

	chore := p.choreFromDB(r)

	// Log activity
	p.addActivity(req.HouseholdID, req.CreatedBy, req.CurrentAssigneeID, models.ActivityChoreCreated, "chore", id, fmt.Sprintf("Created chore '%s'", req.Title))

	return &chore, nil
}

func (p *PostgresStore) UpdateChore(id string, req models.ChoreUpdateRequest) (*models.Chore, error) {
	ctx := context.Background()
	current, err := p.q.GetChore(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	c := p.choreFromDB(current)

	if req.Title != nil {
		c.Title = *req.Title
	}
	if req.Description != nil {
		c.Description = req.Description
	}
	if req.Category != nil {
		c.Category = *req.Category
	}
	if req.EffortPoints != nil {
		c.EffortPoints = *req.EffortPoints
	}
	if req.AssignmentType != nil {
		c.AssignmentType = *req.AssignmentType
	}
	if req.CurrentAssigneeID != nil {
		c.CurrentAssigneeID = req.CurrentAssigneeID
	}
	if req.RotationMemberIDs != nil {
		c.RotationMemberIDs = *req.RotationMemberIDs
	}
	if req.CurrentRotationIndex != nil {
		c.CurrentRotationIndex = req.CurrentRotationIndex
	}
	if req.RecurrenceType != nil {
		c.RecurrenceType = *req.RecurrenceType
	}
	if req.RecurrenceRule != nil {
		c.RecurrenceRule = req.RecurrenceRule
	}
	if req.DueDate != nil {
		c.DueDate = req.DueDate
	}
	if req.RequiresApproval != nil {
		c.RequiresApproval = *req.RequiresApproval
	}
	if req.RequiresProof != nil {
		c.RequiresProof = *req.RequiresProof
	}
	if req.Status != nil {
		c.Status = *req.Status
	}

	rotJSON, _ := json.Marshal(c.RotationMemberIDs)
	var recJSON []byte
	if c.RecurrenceRule != nil {
		recJSON, _ = json.Marshal(c.RecurrenceRule)
	}

	r, err := p.q.UpdateChore(ctx, db.UpdateChoreParams{
		ID:                   c.ID,
		Title:                c.Title,
		Description:          textToPg(c.Description),
		Category:             c.Category,
		EffortPoints:         int32(c.EffortPoints),
		AssignmentType:       c.AssignmentType,
		CurrentAssigneeID:    textToPg(c.CurrentAssigneeID),
		RotationMemberIds:    rotJSON,
		CurrentRotationIndex: intToPg(c.CurrentRotationIndex),
		RecurrenceType:       c.RecurrenceType,
		RecurrenceRule:       recJSON,
		DueDate:              timePtrToPg(c.DueDate),
		RequiresApproval:     c.RequiresApproval,
		RequiresProof:        c.RequiresProof,
		Status:               c.Status,
	})
	if err != nil {
		return nil, err
	}

	updated := p.choreFromDB(r)
	return &updated, nil
}

func (p *PostgresStore) DeleteChore(id string) error {
	ctx := context.Background()
	if _, err := p.q.GetChore(ctx, id); err != nil {
		return ErrNotFound
	}
	return p.q.DeleteChore(ctx, id)
}

func (p *PostgresStore) ClaimChore(choreID, memberID string) (*models.Chore, error) {
	ctx := context.Background()
	_, err := p.q.GetChore(ctx, choreID)
	if err != nil {
		return nil, ErrNotFound
	}

	if _, err := p.q.GetMember(ctx, memberID); err != nil {
		return nil, ErrNotFound
	}

	r, err := p.q.UpdateChoreAssigneeAndStatus(ctx, db.UpdateChoreAssigneeAndStatusParams{
		ID:                choreID,
		CurrentAssigneeID: textToPg(&memberID),
		Status:            models.ChoreStatusInProgress,
	})
	if err != nil {
		return nil, err
	}

	c := p.choreFromDB(r)
	p.addActivity(c.HouseholdID, memberID, nil, models.ActivityChoreAssigned, "chore", choreID, fmt.Sprintf("Claimed chore '%s'", c.Title))
	return &c, nil
}

func (p *PostgresStore) CompleteChore(choreID, memberID string, proofNotes, proofPhotoURL *string) (*models.Chore, *models.ChoreCompletion, int, error) {
	ctx := context.Background()
	current, err := p.q.GetChore(ctx, choreID)
	if err != nil {
		return nil, nil, 0, ErrNotFound
	}

	chore := p.choreFromDB(current)
	now := time.Now().UTC()

	completionID := generateID("comp")

	if chore.RequiresApproval {
		// Needs parent/admin verification
		compRecord, err := p.q.CreateCompletion(ctx, db.CreateCompletionParams{
			ID:              completionID,
			ChoreID:         choreID,
			HouseholdID:     chore.HouseholdID,
			MemberID:        memberID,
			Status:          models.CompletionStatusPendingApproval,
			ProofNotes:      textToPg(proofNotes),
			ProofPhotoUrl:   textToPg(proofPhotoURL),
			ApprovedBy:      pgtype.Text{Valid: false},
			RejectionReason: pgtype.Text{Valid: false},
			PointsAwarded:   int32(chore.EffortPoints),
			CompletedAt:     timeToPg(now),
			VerifiedAt:      pgtype.Timestamptz{Valid: false},
		})
		if err != nil {
			return nil, nil, 0, err
		}

		upChore, _ := p.q.UpdateChoreAssigneeAndStatus(ctx, db.UpdateChoreAssigneeAndStatusParams{
			ID:                choreID,
			CurrentAssigneeID: textToPg(chore.CurrentAssigneeID),
			Status:            models.ChoreStatusPendingApproval,
		})

		choreUpdated := p.choreFromDB(upChore)
		completion := p.completionFromDB(compRecord)

		p.addActivity(chore.HouseholdID, memberID, nil, models.ActivityApprovalRequested, "chore", choreID, fmt.Sprintf("Submitted chore '%s' for approval", chore.Title))

		return &choreUpdated, &completion, 0, nil
	}

	// Instantly completed
	compRecord, err := p.q.CreateCompletion(ctx, db.CreateCompletionParams{
		ID:              completionID,
		ChoreID:         choreID,
		HouseholdID:     chore.HouseholdID,
		MemberID:        memberID,
		Status:          models.CompletionStatusApproved,
		ProofNotes:      textToPg(proofNotes),
		ProofPhotoUrl:   textToPg(proofPhotoURL),
		ApprovedBy:      textToPg(&memberID),
		RejectionReason: pgtype.Text{Valid: false},
		PointsAwarded:   int32(chore.EffortPoints),
		CompletedAt:     timeToPg(now),
		VerifiedAt:      timeToPg(now),
	})
	if err != nil {
		return nil, nil, 0, err
	}

	// Credit member points & streak
	if mem, err := p.q.GetMember(ctx, memberID); err == nil {
		_, _ = p.q.UpdateMemberStats(ctx, db.UpdateMemberStatsParams{
			ID:                memberID,
			PointsBalance:     mem.PointsBalance + int32(chore.EffortPoints),
			TotalPointsEarned: mem.TotalPointsEarned + int32(chore.EffortPoints),
			Streak:            mem.Streak + 1,
		})
	}

	// Mark completed
	_, _ = p.q.UpdateChoreAssigneeAndStatus(ctx, db.UpdateChoreAssigneeAndStatusParams{
		ID:                choreID,
		CurrentAssigneeID: textToPg(chore.CurrentAssigneeID),
		Status:            models.ChoreStatusCompleted,
	})

	p.advanceRecurrence(chore)

	upChore, _ := p.q.GetChore(ctx, choreID)
	choreUpdated := p.choreFromDB(upChore)
	completion := p.completionFromDB(compRecord)

	p.addActivity(chore.HouseholdID, memberID, nil, models.ActivityChoreCompleted, "chore", choreID, fmt.Sprintf("Completed chore '%s' (+%d pts)", chore.Title, chore.EffortPoints))

	return &choreUpdated, &completion, chore.EffortPoints, nil
}

func (p *PostgresStore) RotateChore(householdID, choreID string) (*models.Chore, error) {
	ctx := context.Background()
	current, err := p.q.GetChore(ctx, choreID)
	if err != nil {
		return nil, ErrNotFound
	}

	chore := p.choreFromDB(current)
	if chore.AssignmentType != models.ChoreAssignmentRoundRobin || len(chore.RotationMemberIDs) == 0 {
		return nil, errors.New("CHORE_NOT_ROUND_ROBIN")
	}

	currIdx := 0
	if chore.CurrentRotationIndex != nil {
		currIdx = *chore.CurrentRotationIndex
	}
	nextIdx := (currIdx + 1) % len(chore.RotationMemberIDs)
	nextAssigneeID := chore.RotationMemberIDs[nextIdx]

	r, err := p.q.UpdateChoreRotation(ctx, db.UpdateChoreRotationParams{
		ID:                   choreID,
		CurrentAssigneeID:    textToPg(&nextAssigneeID),
		CurrentRotationIndex: intToPg(&nextIdx),
	})
	if err != nil {
		return nil, err
	}

	updated := p.choreFromDB(r)
	p.addActivity(householdID, nextAssigneeID, nil, models.ActivityChoreAssigned, "chore", choreID, fmt.Sprintf("Chore rotated to member %s", nextAssigneeID))
	return &updated, nil
}

func (p *PostgresStore) GetPendingApprovals(householdID string) []models.PendingApprovalItem {
	ctx := context.Background()
	rows, err := p.q.GetPendingApprovalsByHousehold(ctx, householdID)
	if err != nil {
		return []models.PendingApprovalItem{}
	}

	items := make([]models.PendingApprovalItem, 0)
	for _, r := range rows {
		comp := p.completionFromDB(r)
		choreObj, err := p.GetChore(comp.ChoreID)
		if err != nil {
			continue
		}
		memObj, err := p.GetMember(comp.MemberID)
		if err != nil {
			continue
		}

		items = append(items, models.PendingApprovalItem{
			Completion: comp,
			Chore:      *choreObj,
			Member:     *memObj,
		})
	}
	return items
}

func (p *PostgresStore) ApproveChore(completionID, adminMemberID string) error {
	ctx := context.Background()
	compRow, err := p.q.GetCompletion(ctx, completionID)
	if err != nil {
		return ErrNotFound
	}

	comp := p.completionFromDB(compRow)
	if comp.Status != models.CompletionStatusPendingApproval {
		return errors.New("COMPLETION_NOT_PENDING")
	}

	adminMem, err := p.q.GetMember(ctx, adminMemberID)
	if err != nil || adminMem.Role != models.MemberRoleAdmin {
		return ErrForbidden
	}

	now := time.Now().UTC()
	_, err = p.q.UpdateCompletionApproval(ctx, db.UpdateCompletionApprovalParams{
		ID:              completionID,
		Status:          models.CompletionStatusApproved,
		ApprovedBy:      textToPg(&adminMemberID),
		RejectionReason: pgtype.Text{Valid: false},
		VerifiedAt:      timeToPg(now),
	})
	if err != nil {
		return err
	}

	// Credit points to completing member
	if m, err := p.q.GetMember(ctx, comp.MemberID); err == nil {
		_, _ = p.q.UpdateMemberStats(ctx, db.UpdateMemberStatsParams{
			ID:                comp.MemberID,
			PointsBalance:     m.PointsBalance + int32(comp.PointsAwarded),
			TotalPointsEarned: m.TotalPointsEarned + int32(comp.PointsAwarded),
			Streak:            m.Streak + 1,
		})
	}

	// Mark chore completed and advance recurrence
	if choreObj, err := p.GetChore(comp.ChoreID); err == nil {
		_, _ = p.q.UpdateChoreAssigneeAndStatus(ctx, db.UpdateChoreAssigneeAndStatusParams{
			ID:                choreObj.ID,
			CurrentAssigneeID: textToPg(choreObj.CurrentAssigneeID),
			Status:            models.ChoreStatusCompleted,
		})
		p.advanceRecurrence(*choreObj)
	}

	p.addActivity(comp.HouseholdID, adminMemberID, &comp.MemberID, models.ActivityChoreApproved, "chore", comp.ChoreID, fmt.Sprintf("Approved completion (+%d pts)", comp.PointsAwarded))
	return nil
}

func (p *PostgresStore) RejectChore(completionID, adminMemberID, reason string) error {
	ctx := context.Background()
	compRow, err := p.q.GetCompletion(ctx, completionID)
	if err != nil {
		return ErrNotFound
	}

	comp := p.completionFromDB(compRow)
	if comp.Status != models.CompletionStatusPendingApproval {
		return errors.New("COMPLETION_NOT_PENDING")
	}

	adminMem, err := p.q.GetMember(ctx, adminMemberID)
	if err != nil || adminMem.Role != models.MemberRoleAdmin {
		return ErrForbidden
	}

	now := time.Now().UTC()
	_, err = p.q.UpdateCompletionApproval(ctx, db.UpdateCompletionApprovalParams{
		ID:              completionID,
		Status:          models.CompletionStatusRejected,
		ApprovedBy:      textToPg(&adminMemberID),
		RejectionReason: textToPg(&reason),
		VerifiedAt:      timeToPg(now),
	})
	if err != nil {
		return err
	}

	// Reset chore back to assigned
	if choreObj, err := p.GetChore(comp.ChoreID); err == nil {
		_, _ = p.q.UpdateChoreAssigneeAndStatus(ctx, db.UpdateChoreAssigneeAndStatusParams{
			ID:                choreObj.ID,
			CurrentAssigneeID: textToPg(choreObj.CurrentAssigneeID),
			Status:            models.ChoreStatusAssigned,
		})
	}

	p.addActivity(comp.HouseholdID, adminMemberID, &comp.MemberID, models.ActivityChoreRejected, "chore", comp.ChoreID, fmt.Sprintf("Rejected completion: %s", reason))
	return nil
}

// ----------------- SWAPS -----------------

func (p *PostgresStore) GetSwaps(householdID string) []models.SwapItem {
	ctx := context.Background()
	rows, err := p.q.GetSwapsByHousehold(ctx, householdID)
	if err != nil {
		return []models.SwapItem{}
	}

	items := make([]models.SwapItem, 0)
	for _, r := range rows {
		sw := p.swapFromDB(r)
		choreObj, err := p.GetChore(sw.ChoreID)
		if err != nil {
			continue
		}
		reqMember, err := p.GetMember(sw.RequesterID)
		if err != nil {
			continue
		}

		var targetMember *models.Member
		if sw.TargetMemberID != nil {
			targetMember, _ = p.GetMember(*sw.TargetMemberID)
		}

		items = append(items, models.SwapItem{
			Swap:         sw,
			Chore:        *choreObj,
			Requester:    *reqMember,
			TargetMember: targetMember,
		})
	}
	return items
}

func (p *PostgresStore) CreateSwap(householdID, choreID, requesterID string, targetMemberID, reason *string) (*models.ChoreSwapRequest, error) {
	ctx := context.Background()
	hh, err := p.GetHousehold(householdID)
	if err != nil {
		return nil, ErrNotFound
	}
	if !hh.Settings.AllowSwaps {
		return nil, errors.New("SWAPS_DISABLED")
	}

	chore, err := p.GetChore(choreID)
	if err != nil {
		return nil, ErrNotFound
	}
	if chore.CurrentAssigneeID == nil || *chore.CurrentAssigneeID != requesterID {
		return nil, errors.New("NOT_CHORE_ASSIGNEE")
	}
	if chore.Status == models.ChoreStatusCompleted {
		return nil, errors.New("CHORE_ALREADY_COMPLETED")
	}

	id := generateID("sw")
	now := time.Now().UTC()
	r, err := p.q.CreateSwap(ctx, db.CreateSwapParams{
		ID:             id,
		HouseholdID:    householdID,
		ChoreID:        choreID,
		RequesterID:    requesterID,
		TargetMemberID: textToPg(targetMemberID),
		Status:         models.SwapStatusPending,
		Reason:         textToPg(reason),
		CreatedAt:      timeToPg(now),
		ResolvedAt:     pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		return nil, err
	}

	sw := p.swapFromDB(r)
	p.addActivity(householdID, requesterID, targetMemberID, models.ActivitySwapProposed, "swap", id, fmt.Sprintf("Proposed swap for chore '%s'", chore.Title))
	return &sw, nil
}

func (p *PostgresStore) AcceptSwap(swapID, acceptorMemberID string) error {
	ctx := context.Background()
	r, err := p.q.GetSwap(ctx, swapID)
	if err != nil {
		return ErrNotFound
	}

	sw := p.swapFromDB(r)
	if sw.Status != models.SwapStatusPending {
		return errors.New("SWAP_NOT_PENDING")
	}

	chore, err := p.GetChore(sw.ChoreID)
	if err != nil {
		return ErrNotFound
	}

	// Reassign chore to acceptor
	_, err = p.q.UpdateChoreAssigneeAndStatus(ctx, db.UpdateChoreAssigneeAndStatusParams{
		ID:                chore.ID,
		CurrentAssigneeID: textToPg(&acceptorMemberID),
		Status:            models.ChoreStatusAssigned,
	})
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	// Update swap status
	_, err = p.q.UpdateSwapStatus(ctx, db.UpdateSwapStatusParams{
		ID:         swapID,
		Status:     models.SwapStatusAccepted,
		ResolvedAt: timeToPg(now),
	})
	if err != nil {
		return err
	}

	p.addActivity(sw.HouseholdID, acceptorMemberID, &sw.RequesterID, models.ActivitySwapAccepted, "swap", swapID, fmt.Sprintf("Accepted swap for chore '%s'", chore.Title))
	return nil
}

func (p *PostgresStore) RejectSwap(swapID string) error {
	ctx := context.Background()
	r, err := p.q.GetSwap(ctx, swapID)
	if err != nil {
		return ErrNotFound
	}

	sw := p.swapFromDB(r)
	if sw.Status != models.SwapStatusPending {
		return errors.New("SWAP_NOT_PENDING")
	}

	now := time.Now().UTC()
	_, err = p.q.UpdateSwapStatus(ctx, db.UpdateSwapStatusParams{
		ID:         swapID,
		Status:     models.SwapStatusRejected,
		ResolvedAt: timeToPg(now),
	})
	return err
}

func (p *PostgresStore) CancelSwap(swapID string) error {
	ctx := context.Background()
	r, err := p.q.GetSwap(ctx, swapID)
	if err != nil {
		return ErrNotFound
	}

	sw := p.swapFromDB(r)
	if sw.Status != models.SwapStatusPending {
		return errors.New("SWAP_NOT_PENDING")
	}

	now := time.Now().UTC()
	_, err = p.q.UpdateSwapStatus(ctx, db.UpdateSwapStatusParams{
		ID:         swapID,
		Status:     models.SwapStatusCancelled,
		ResolvedAt: timeToPg(now),
	})
	return err
}

// ----------------- REWARDS & REDEMPTIONS -----------------

func (p *PostgresStore) GetRewards(householdID string) []models.RewardItem {
	ctx := context.Background()
	rows, err := p.q.GetRewardsByHousehold(ctx, householdID)
	if err != nil {
		return []models.RewardItem{}
	}
	result := make([]models.RewardItem, len(rows))
	for i, r := range rows {
		result[i] = p.rewardFromDB(r)
	}
	return result
}

func (p *PostgresStore) GetReward(rewardID string) (*models.RewardItem, error) {
	ctx := context.Background()
	r, err := p.q.GetReward(ctx, rewardID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	item := p.rewardFromDB(r)
	return &item, nil
}

func (p *PostgresStore) CreateReward(householdID string, req models.RewardItemCreateRequest) (*models.RewardItem, error) {
	ctx := context.Background()
	if _, err := p.GetHousehold(householdID); err != nil {
		return nil, ErrNotFound
	}

	id := generateID("rew")
	now := time.Now().UTC()
	r, err := p.q.CreateReward(ctx, db.CreateRewardParams{
		ID:          id,
		HouseholdID: householdID,
		Title:       req.Title,
		Description: req.Description,
		PointsCost:  int32(req.PointsCost),
		Icon:        textToPg(req.Icon),
		IsActive:    true,
		CreatedAt:   timeToPg(now),
	})
	if err != nil {
		return nil, err
	}

	item := p.rewardFromDB(r)
	return &item, nil
}

func (p *PostgresStore) RedeemReward(rewardID, memberID string) (*models.RewardRedemption, error) {
	ctx := context.Background()
	rew, err := p.GetReward(rewardID)
	if err != nil {
		return nil, ErrNotFound
	}

	mem, err := p.GetMember(memberID)
	if err != nil {
		return nil, ErrNotFound
	}

	if mem.PointsBalance < rew.PointsCost {
		return nil, errors.New("INSUFFICIENT_POINTS")
	}

	// Deduct points
	_, _ = p.q.UpdateMemberStats(ctx, db.UpdateMemberStatsParams{
		ID:                memberID,
		PointsBalance:     int32(mem.PointsBalance - rew.PointsCost),
		TotalPointsEarned: int32(mem.TotalPointsEarned),
		Streak:            int32(mem.Streak),
	})

	id := generateID("red")
	now := time.Now().UTC()
	r, err := p.q.CreateRedemption(ctx, db.CreateRedemptionParams{
		ID:           id,
		HouseholdID:  rew.HouseholdID,
		MemberID:     memberID,
		RewardItemID: rewardID,
		PointsSpent:  int32(rew.PointsCost),
		Status:       models.RedemptionStatusRequested,
		RequestedAt:  timeToPg(now),
		FulfilledAt:  pgtype.Timestamptz{Valid: false},
	})
	if err != nil {
		return nil, err
	}

	red := p.redemptionFromDB(r)
	p.addActivity(rew.HouseholdID, memberID, nil, models.ActivityRewardRedeemed, "reward", rewardID, fmt.Sprintf("Redeemed '%s' (-%d pts)", rew.Title, rew.PointsCost))
	return &red, nil
}

func (p *PostgresStore) FulfillReward(redemptionID string) error {
	ctx := context.Background()
	r, err := p.q.GetRedemption(ctx, redemptionID)
	if err != nil {
		return ErrNotFound
	}

	red := p.redemptionFromDB(r)
	if red.Status != models.RedemptionStatusRequested {
		return errors.New("REDEMPTION_NOT_REQUESTED")
	}

	now := time.Now().UTC()
	_, err = p.q.UpdateRedemptionStatus(ctx, db.UpdateRedemptionStatusParams{
		ID:          redemptionID,
		Status:      models.RedemptionStatusFulfilled,
		FulfilledAt: timeToPg(now),
	})
	if err != nil {
		return err
	}

	p.addActivity(red.HouseholdID, red.MemberID, nil, models.ActivityRewardFulfilled, "redemption", redemptionID, "Reward fulfilled")
	return nil
}

func (p *PostgresStore) GetRedemptions(householdID string) []models.RewardRedemptionItem {
	ctx := context.Background()
	rows, err := p.q.GetRedemptionsByHousehold(ctx, householdID)
	if err != nil {
		return []models.RewardRedemptionItem{}
	}

	items := make([]models.RewardRedemptionItem, 0)
	for _, r := range rows {
		red := p.redemptionFromDB(r)
		rew, err := p.GetReward(red.RewardItemID)
		if err != nil {
			continue
		}
		mem, err := p.GetMember(red.MemberID)
		if err != nil {
			continue
		}

		items = append(items, models.RewardRedemptionItem{
			Redemption: red,
			Reward:     *rew,
			Member:     *mem,
		})
	}
	return items
}

// ----------------- ACTIVITIES & COMMENTS -----------------

func (p *PostgresStore) GetActivitiesFiltered(householdID string, entityType, eventType string, limit, offset int) []models.ActivityLog {
	ctx := context.Background()
	if limit <= 0 {
		limit = 50
	}
	rows, err := p.q.GetActivitiesByHousehold(ctx, db.GetActivitiesByHouseholdParams{
		HouseholdID: householdID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return []models.ActivityLog{}
	}

	result := make([]models.ActivityLog, 0)
	for _, r := range rows {
		a := p.activityFromDB(r)
		if entityType != "" && a.EntityType != entityType {
			continue
		}
		if eventType != "" && a.EventType != eventType {
			continue
		}
		result = append(result, a)
	}
	return result
}

func (p *PostgresStore) GetActivities(householdID string) []models.ActivityLog {
	return p.GetActivitiesFiltered(householdID, "", "", 50, 0)
}

func (p *PostgresStore) SendNudge(choreID, senderMemberID string) (string, error) {
	c, err := p.GetChore(choreID)
	if err != nil {
		return "", ErrNotFound
	}
	if c.CurrentAssigneeID == nil {
		return "", errors.New("CHORE_HAS_NO_ASSIGNEE")
	}

	msg := fmt.Sprintf("Friendly reminder about chore '%s'", c.Title)
	p.addActivity(c.HouseholdID, senderMemberID, c.CurrentAssigneeID, models.ActivityChoreNudge, "chore", choreID, msg)
	return msg, nil
}

func (p *PostgresStore) GetComments(choreID string) []models.ChoreCommentWithMember {
	ctx := context.Background()
	rows, err := p.q.GetCommentsByChore(ctx, choreID)
	if err != nil {
		return []models.ChoreCommentWithMember{}
	}

	result := make([]models.ChoreCommentWithMember, 0)
	for _, r := range rows {
		var memberPtr *models.Member
		if mem, err := p.GetMember(r.MemberID); err == nil {
			memberPtr = mem
		}

		result = append(result, models.ChoreCommentWithMember{
			Comment: models.ChoreComment{
				ID:        r.ID,
				ChoreID:   r.ChoreID,
				MemberID:  r.MemberID,
				Message:   r.Message,
				CreatedAt: r.CreatedAt.Time.Format(time.RFC3339),
			},
			Member: memberPtr,
		})
	}
	return result
}

func (p *PostgresStore) AddComment(choreID, memberID, message string) (*models.ChoreComment, error) {
	ctx := context.Background()
	if _, err := p.GetChore(choreID); err != nil {
		return nil, ErrNotFound
	}
	if _, err := p.GetMember(memberID); err != nil {
		return nil, ErrNotFound
	}

	id := generateID("comm")
	now := time.Now().UTC()
	r, err := p.q.CreateComment(ctx, db.CreateCommentParams{
		ID:        id,
		ChoreID:   choreID,
		MemberID:  memberID,
		Message:   strings.TrimSpace(message),
		CreatedAt: timeToPg(now),
	})
	if err != nil {
		return nil, err
	}

	comment := models.ChoreComment{
		ID:        r.ID,
		ChoreID:   r.ChoreID,
		MemberID:  r.MemberID,
		Message:   r.Message,
		CreatedAt: r.CreatedAt.Time.Format(time.RFC3339),
	}
	return &comment, nil
}

// ----------------- AUTH & MULTI-TENANCY -----------------

func (p *PostgresStore) VerifyAdminPIN(householdID, pin string) (bool, error) {
	h, err := p.GetHousehold(householdID)
	if err != nil {
		return false, ErrNotFound
	}
	expectedPIN := h.Settings.AdminPIN
	if expectedPIN == "" {
		expectedPIN = "1234"
	}
	return pin == expectedPIN, nil
}

func (p *PostgresStore) GetMemberByEmail(email string) (*models.Member, *models.Household, error) {
	ctx := context.Background()
	r, err := p.q.GetMemberByEmail(ctx, textToPg(&email))
	if err != nil {
		return nil, nil, ErrNotFound
	}
	mem := p.memberFromDB(r)
	hh, err := p.GetHousehold(mem.HouseholdID)
	if err != nil {
		return nil, nil, err
	}
	return &mem, hh, nil
}

func (p *PostgresStore) CreateMagicLink(email string) (string, *models.Member, *models.Household, error) {
	ctx := context.Background()
	mem, hh, err := p.GetMemberByEmail(email)
	if err != nil {
		return "", nil, nil, err
	}

	token := fmt.Sprintf("MAGIC-%s-%d", strings.ToUpper(strings.ReplaceAll(mem.Name, " ", "")), time.Now().Unix())
	now := time.Now().UTC()
	expiresAt := now.Add(15 * time.Minute)

	_, err = p.pool.Exec(ctx, `
		INSERT INTO magic_links (token, member_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (token) DO UPDATE
		SET expires_at = EXCLUDED.expires_at, created_at = EXCLUDED.created_at
	`, token, mem.ID, expiresAt, now)
	if err != nil {
		// Fallback in case expires_at column does not exist yet
		_, err = p.pool.Exec(ctx, `
			INSERT INTO magic_links (token, member_id, created_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (token) DO UPDATE
			SET created_at = EXCLUDED.created_at
		`, token, mem.ID, now)
		if err != nil {
			return "", nil, nil, err
		}
	}

	return token, mem, hh, nil
}

type verifiedMagicLinkCacheEntry struct {
	member    *models.Member
	household *models.Household
	cachedAt  time.Time
}

var (
	recentVerifiedTokensMu sync.RWMutex
	recentVerifiedTokens   = make(map[string]verifiedMagicLinkCacheEntry)
)

func (p *PostgresStore) VerifyMagicLink(token string) (*models.Member, *models.Household, error) {
	ctx := context.Background()
	cleanToken := strings.TrimSpace(token)
	upperToken := strings.ToUpper(cleanToken)

	// 0. Idempotent short-lived cache check (30 seconds):
	// Protects against React StrictMode double invocation, component remounts, or browser double-clicks
	recentVerifiedTokensMu.RLock()
	if entry, ok := recentVerifiedTokens[cleanToken]; ok && time.Since(entry.cachedAt) < 30*time.Second {
		recentVerifiedTokensMu.RUnlock()
		return entry.member, entry.household, nil
	}
	if entry, ok := recentVerifiedTokens[upperToken]; ok && time.Since(entry.cachedAt) < 30*time.Second {
		recentVerifiedTokensMu.RUnlock()
		return entry.member, entry.household, nil
	}
	recentVerifiedTokensMu.RUnlock()

	var actualToken string
	var memberID string
	var expiresAt time.Time
	var createdAt time.Time

	// Query token, member_id, and timestamps
	err := p.pool.QueryRow(ctx, `
		SELECT token, member_id, expires_at, created_at 
		FROM magic_links 
		WHERE token = $1 OR token = $2 
		LIMIT 1
	`, cleanToken, upperToken).Scan(&actualToken, &memberID, &expiresAt, &createdAt)

	if err != nil {
		// Fallback query if expires_at column was not returned
		err = p.pool.QueryRow(ctx, `
			SELECT token, member_id, created_at 
			FROM magic_links 
			WHERE token = $1 OR token = $2 
			LIMIT 1
		`, cleanToken, upperToken).Scan(&actualToken, &memberID, &createdAt)
		if err != nil {
			return nil, nil, errors.New("INVALID_OR_EXPIRED_TOKEN")
		}
		expiresAt = createdAt.Add(15 * time.Minute)
	}

	now := time.Now().UTC()
	isDemoToken := actualToken == "MAGIC-SARAH" || actualToken == "MAGIC-DAVID" || actualToken == "MAGIC-ALEX"

	// 1. Time-Based Expiration check (15 minutes validity for generated tokens)
	if !isDemoToken && (now.After(expiresAt) || now.Sub(createdAt) > 15*time.Minute) {
		// Purge expired token from DB
		_, _ = p.pool.Exec(ctx, `DELETE FROM magic_links WHERE token = $1`, actualToken)
		return nil, nil, errors.New("INVALID_OR_EXPIRED_TOKEN")
	}

	// 2. Single-Use Invalidation: Delete token immediately upon login so it cannot be reused (preserve static demo tokens)
	if !isDemoToken {
		_, err = p.pool.Exec(ctx, `DELETE FROM magic_links WHERE token = $1`, actualToken)
		if err != nil {
			log.Printf("[PostgresStore] Warning: failed to delete used magic link token %s: %v", actualToken, err)
		}
	}

	mem, err := p.GetMember(memberID)
	if err != nil {
		return nil, nil, ErrNotFound
	}

	hh, err := p.GetHousehold(mem.HouseholdID)
	if err != nil {
		return nil, nil, ErrNotFound
	}

	// Record in short-lived memory cache so concurrent requests within 30s succeed
	recentVerifiedTokensMu.Lock()
	recentVerifiedTokens[cleanToken] = verifiedMagicLinkCacheEntry{
		member:    mem,
		household: hh,
		cachedAt:  time.Now(),
	}
	recentVerifiedTokens[upperToken] = verifiedMagicLinkCacheEntry{
		member:    mem,
		household: hh,
		cachedAt:  time.Now(),
	}
	// Prune older than 2 minutes
	for k, v := range recentVerifiedTokens {
		if time.Since(v.cachedAt) > 2*time.Minute {
			delete(recentVerifiedTokens, k)
		}
	}
	recentVerifiedTokensMu.Unlock()

	return mem, hh, nil
}

func (p *PostgresStore) GetDemoMember(memberID string) (*models.Member, *models.Household, error) {
	mem, err := p.GetMember(memberID)
	if err != nil {
		return nil, nil, ErrNotFound
	}

	hh, err := p.GetHousehold(mem.HouseholdID)
	if err != nil {
		return nil, nil, ErrNotFound
	}

	return mem, hh, nil
}

// ----------------- INTERNAL HELPERS -----------------

func (p *PostgresStore) addActivity(householdID, actorID string, recipientID *string, eventType, entityType, entityID, message string) {
	ctx := context.Background()
	id := generateID("act")
	now := time.Now().UTC()

	_, _ = p.q.CreateActivityLog(ctx, db.CreateActivityLogParams{
		ID:          id,
		HouseholdID: householdID,
		RecipientID: textToPg(recipientID),
		ActorID:     actorID,
		EventType:   eventType,
		EntityType:  entityType,
		EntityID:    entityID,
		Message:     message,
		Read:        false,
		CreatedAt:   timeToPg(now),
	})
}

func (p *PostgresStore) advanceRecurrence(chore models.Chore) {
	if chore.RecurrenceType == models.ChoreRecurrenceNone || chore.RecurrenceType == "" {
		return
	}

	ctx := context.Background()
	now := time.Now().UTC()
	baseDate := now
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

	rotJSON, _ := json.Marshal(chore.RotationMemberIDs)
	var recJSON []byte
	if chore.RecurrenceRule != nil {
		recJSON, _ = json.Marshal(chore.RecurrenceRule)
	}

	nextChoreID := generateID("c")
	_, _ = p.q.CreateChore(ctx, db.CreateChoreParams{
		ID:                   nextChoreID,
		HouseholdID:          chore.HouseholdID,
		Title:                chore.Title,
		Description:          textToPg(chore.Description),
		Category:             chore.Category,
		EffortPoints:         int32(chore.EffortPoints),
		AssignmentType:       chore.AssignmentType,
		CurrentAssigneeID:    textToPg(nextAssigneeID),
		RotationMemberIds:    rotJSON,
		CurrentRotationIndex: intToPg(&nextIndex),
		RecurrenceType:       chore.RecurrenceType,
		RecurrenceRule:       recJSON,
		DueDate:              timePtrToPg(&nextDueDate),
		RequiresApproval:     chore.RequiresApproval,
		RequiresProof:        chore.RequiresProof,
		Status:               status,
		CreatedBy:            chore.CreatedBy,
		CreatedAt:            timeToPg(now),
		UpdatedAt:            timeToPg(now),
	})

	targetDesc := "Placed in Open Pool"
	if nextAssigneeID != nil {
		if m, err := p.GetMember(*nextAssigneeID); err == nil {
			targetDesc = fmt.Sprintf("Assigned to %s", m.Name)
		}
	}

	p.addActivity(chore.HouseholdID, chore.CreatedBy, nextAssigneeID, models.ActivityChoreCreated, "chore", nextChoreID,
		fmt.Sprintf("Scheduled next instance for \"%s\". %s", chore.Title, targetDesc))
}

// VerifyPassword checks if the provided password matches the member's credentials in Postgres.
func (p *PostgresStore) VerifyPassword(memberID, password string) bool {
	ctx := context.Background()
	row := p.pool.QueryRow(ctx, "SELECT password_hash FROM members WHERE id = $1", memberID)
	var hash string
	if err := row.Scan(&hash); err != nil {
		return false
	}

	if hash == "" {
		return password == "password123"
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// RegisterMember registers a new member in PostgreSQL and either creates a new household or attaches to an existing one.
func (p *PostgresStore) RegisterMember(req models.RegisterRequest) (*models.Member, *models.Household, error) {
	ctx := context.Background()
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if already registered
	var existingID string
	err := p.pool.QueryRow(ctx, "SELECT id FROM members WHERE LOWER(email) = $1", cleanEmail).Scan(&existingID)
	if err == nil && existingID != "" {
		return nil, nil, fmt.Errorf("email '%s' is already registered", cleanEmail)
	}

	var household models.Household
	role := req.Role
	if role == "" {
		role = models.MemberRoleAdmin
	}

	if strings.TrimSpace(req.InviteCode) != "" {
		h, err := p.JoinHousehold(req.InviteCode)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid invite code '%s'", req.InviteCode)
		}
		household = *h
		if req.Role == "" {
			role = models.MemberRoleMember
		}
	} else {
		hName := strings.TrimSpace(req.HouseholdName)
		if hName == "" {
			hName = fmt.Sprintf("%s's Home", req.Name)
		}
		createdH := p.CreateHousehold(models.HouseholdCreateRequest{
			Name: hName,
		})
		household = createdH
		role = models.MemberRoleAdmin
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	memberID := generateID("m")
	now := time.Now().UTC()
	avatar := "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=120&auto=format&fit=crop&q=80"

	// Insert into members table including password_hash
	_, err = p.pool.Exec(ctx, `
		INSERT INTO members (id, household_id, name, email, password_hash, avatar_url, role, points_balance, total_points_earned, streak, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0, 0, 0, $8, $9)
	`, memberID, household.ID, req.Name, cleanEmail, string(hash), avatar, role, now, now)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to insert member: %w", err)
	}

	createdMember := models.Member{
		ID:                memberID,
		HouseholdID:       household.ID,
		Name:              req.Name,
		Email:             &cleanEmail,
		PasswordHash:      string(hash),
		AvatarURL:         &avatar,
		Role:              role,
		PointsBalance:     0,
		TotalPointsEarned: 0,
		Streak:            0,
		CreatedAt:         now.Format(time.RFC3339),
		UpdatedAt:         now.Format(time.RFC3339),
	}

	return &createdMember, &household, nil
}

func (p *PostgresStore) CreatePasswordResetToken(email string) (string, *models.Member, error) {
	ctx := context.Background()
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

	mem, _, err := p.GetMemberByEmail(cleanEmail)
	if err != nil {
		return "", nil, ErrNotFound
	}

	token := fmt.Sprintf("RST-%d", time.Now().UnixNano()%100000000)
	now := time.Now().UTC()
	expiresAt := now.Add(1 * time.Hour)

	_, err = p.pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (token, member_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (token) DO UPDATE
		SET expires_at = EXCLUDED.expires_at, created_at = EXCLUDED.created_at
	`, token, mem.ID, expiresAt, now)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create password reset token: %w", err)
	}

	return token, mem, nil
}

func (p *PostgresStore) ResetPasswordWithToken(token, newPassword string) (*models.Member, *models.Household, error) {
	ctx := context.Background()

	var memberID string
	var expiresAt time.Time

	err := p.pool.QueryRow(ctx, `
		SELECT member_id, expires_at FROM password_reset_tokens WHERE token = $1
	`, token).Scan(&memberID, &expiresAt)
	if err != nil {
		return nil, nil, ErrForbidden
	}

	if time.Now().UTC().After(expiresAt) {
		return nil, nil, ErrForbidden
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	_, err = p.pool.Exec(ctx, `
		UPDATE members SET password_hash = $1, updated_at = $2 WHERE id = $3
	`, string(hash), now, memberID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update password: %w", err)
	}

	_, _ = p.pool.Exec(ctx, `DELETE FROM password_reset_tokens WHERE token = $1`, token)

	mem, err := p.GetMember(memberID)
	if err != nil {
		return nil, nil, ErrNotFound
	}

	hh, err := p.GetHousehold(mem.HouseholdID)
	if err != nil {
		return nil, nil, ErrNotFound
	}

	return mem, hh, nil
}

// UpdatePassword updates a member's password hash in PostgreSQL.
func (p *PostgresStore) UpdatePassword(memberID, password string) error {
	ctx := context.Background()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	_, err = p.pool.Exec(ctx, `UPDATE members SET password_hash = $1, updated_at = $2 WHERE id = $3`, string(hash), now, memberID)
	return err
}


