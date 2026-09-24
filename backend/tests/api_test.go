package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/choresync/backend/internal/email"
	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/server"
	"github.com/choresync/backend/internal/store"
)

func setupTestRouter() (http.Handler, store.Store) {
	appStore := store.NewStore()
	emailSvc := email.NewMockService(email.Config{})
	handler := server.NewRouter(appStore, emailSvc, nil, nil)
	return handler, appStore
}

func executeRequest(handler http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, target, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestHealthCheck(t *testing.T) {
	handler, _ := setupTestRouter()
	w := executeRequest(handler, "GET", "/healthz", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["status"] != "healthy" {
		t.Errorf("expected status healthy, got %s", res["status"])
	}
}

func TestHouseholdsAndMembers(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. List households
	w := executeRequest(handler, "GET", "/api/households", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var households []models.Household
	_ = json.Unmarshal(w.Body.Bytes(), &households)
	if len(households) != 3 {
		t.Fatalf("expected 3 seed households, got %d", len(households))
	}

	// 2. Get specific household
	w = executeRequest(handler, "GET", "/api/households/h-roommates", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var h models.Household
	_ = json.Unmarshal(w.Body.Bytes(), &h)
	if h.Name != "Apartment 4B" {
		t.Errorf("expected Apartment 4B, got %s", h.Name)
	}

	// 3. Join household
	w = executeRequest(handler, "POST", "/api/households/join", models.HouseholdJoinRequest{
		InviteCode: "APT4B-SHARE",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 4. Update household settings
	allowSwaps := false
	w = executeRequest(handler, "PATCH", "/api/households/h-roommates/settings", models.HouseholdSettingsUpdate{
		AllowSwaps: &allowSwaps,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	_ = json.Unmarshal(w.Body.Bytes(), &h)
	if h.Settings.AllowSwaps != false {
		t.Errorf("expected allow_swaps false, got %v", h.Settings.AllowSwaps)
	}

	// 5. List members
	w = executeRequest(handler, "GET", "/api/households/h-roommates/members", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var members []models.Member
	_ = json.Unmarshal(w.Body.Bytes(), &members)
	if len(members) != 4 {
		t.Fatalf("expected 4 roommates, got %d", len(members))
	}

	// 6. Add new member
	w = executeRequest(handler, "POST", "/api/households/h-roommates/members", models.MemberCreateRequest{
		Name: "Jordan Kim",
		Role: "member",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var newM models.Member
	_ = json.Unmarshal(w.Body.Bytes(), &newM)
	if newM.Name != "Jordan Kim" {
		t.Errorf("expected Jordan Kim, got %s", newM.Name)
	}
}

func TestChoresWorkflow(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. List chores for roommates
	w := executeRequest(handler, "GET", "/api/households/h-roommates/chores", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var chores []models.Chore
	_ = json.Unmarshal(w.Body.Bytes(), &chores)
	if len(chores) != 5 {
		t.Fatalf("expected 5 chores, got %d", len(chores))
	}

	// 2. Filter chores: view=open
	w = executeRequest(handler, "GET", "/api/households/h-roommates/chores?view=open", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var openChores []models.Chore
	_ = json.Unmarshal(w.Body.Bytes(), &openChores)
	if len(openChores) != 1 || openChores[0].ID != "c-balcony-sweep" {
		t.Fatalf("expected c-balcony-sweep in open pool, got %v", openChores)
	}

	// 3. Claim open chore
	w = executeRequest(handler, "POST", "/api/chores/c-balcony-sweep/claim", models.ChoreClaimRequest{
		MemberID: "m-sarah",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var claimed models.Chore
	_ = json.Unmarshal(w.Body.Bytes(), &claimed)
	if claimed.Status != "assigned" || *claimed.CurrentAssigneeID != "m-sarah" {
		t.Fatalf("chore not assigned properly to m-sarah: %v", claimed)
	}

	// 4. Complete chore (instant without approval gate)
	w = executeRequest(handler, "POST", "/api/chores/c-kitchen-deep/complete", models.ChoreCompleteRequest{
		MemberID: "m-sarah",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var compResp models.ChoreCompleteResponse
	_ = json.Unmarshal(w.Body.Bytes(), &compResp)
	if compResp.PointsAwarded != 25 {
		t.Errorf("expected 25 points awarded, got %d", compResp.PointsAwarded)
	}
	if compResp.Chore.Status != "completed" {
		t.Errorf("expected status completed, got %s", compResp.Chore.Status)
	}

	// Verify Sarah's points increased
	w = executeRequest(handler, "GET", "/api/households/h-roommates/members", nil)
	var members []models.Member
	_ = json.Unmarshal(w.Body.Bytes(), &members)
	for _, m := range members {
		if m.ID == "m-sarah" {
			if m.PointsBalance != 145+25 {
				t.Errorf("expected 170 points, got %d", m.PointsBalance)
			}
		}
	}
}

func TestApprovalGateWorkflow(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. Pending approval queue should have Leo's chore
	w := executeRequest(handler, "GET", "/api/households/h-family/approvals", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var approvals []models.PendingApprovalItem
	_ = json.Unmarshal(w.Body.Bytes(), &approvals)
	if len(approvals) != 1 {
		t.Fatalf("expected 1 pending approval, got %d", len(approvals))
	}
	compID := approvals[0].Completion.ID

	// 2. Non-admin member cannot approve (403 Forbidden)
	w = executeRequest(handler, "POST", "/api/completions/"+compID+"/approve", models.ApproveCompletionRequest{
		AdminMemberID: "m-emma-kid", // Emma is child role, not admin
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", w.Code)
	}

	// 3. Admin approves
	w = executeRequest(handler, "POST", "/api/completions/"+compID+"/approve", models.ApproveCompletionRequest{
		AdminMemberID: "m-sarah-fam", // Sarah is Admin
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Approval queue should now be empty
	w = executeRequest(handler, "GET", "/api/households/h-family/approvals", nil)
	_ = json.Unmarshal(w.Body.Bytes(), &approvals)
	if len(approvals) != 0 {
		t.Fatalf("expected 0 pending approvals after approval, got %d", len(approvals))
	}
}

func TestSwapsWorkflow(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. Get seed swap
	w := executeRequest(handler, "GET", "/api/households/h-roommates/swaps", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var swaps []models.SwapItem
	_ = json.Unmarshal(w.Body.Bytes(), &swaps)
	if len(swaps) != 1 {
		t.Fatalf("expected 1 swap, got %d", len(swaps))
	}

	// 2. Accept swap by Maya
	w = executeRequest(handler, "POST", "/api/swaps/swap-seed-1/accept", models.SwapAcceptRequest{
		AcceptorMemberID: "m-maya",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Verify chore assignee reassigned to Maya
	w = executeRequest(handler, "GET", "/api/chores/c-vacuum-living", nil)
	var chore models.Chore
	_ = json.Unmarshal(w.Body.Bytes(), &chore)
	if chore.CurrentAssigneeID == nil || *chore.CurrentAssigneeID != "m-maya" {
		t.Fatalf("expected chore reassigned to m-maya, got %v", chore.CurrentAssigneeID)
	}
}

func TestRewardsAndGamification(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. List rewards
	w := executeRequest(handler, "GET", "/api/households/h-roommates/rewards", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var rewards []models.RewardItem
	_ = json.Unmarshal(w.Body.Bytes(), &rewards)
	if len(rewards) != 3 {
		t.Fatalf("expected 3 rewards, got %d", len(rewards))
	}

	// 2. Redeem reward with insufficient points (Noah has 60 pts, coffee is 60 pts, dishes pass is 90 pts)
	w = executeRequest(handler, "POST", "/api/rewards/rew-room-dishes-pass/redeem", models.RewardRedeemRequest{
		MemberID: "m-noah",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for insufficient points, got %d", w.Code)
	}

	// 3. Noah redeems coffee (60 pts)
	w = executeRequest(handler, "POST", "/api/rewards/rew-room-coffee/redeem", models.RewardRedeemRequest{
		MemberID: "m-noah",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for valid redemption, got %d: %s", w.Code, w.Body.String())
	}
	var red models.RewardRedemption
	_ = json.Unmarshal(w.Body.Bytes(), &red)
	if red.PointsSpent != 60 {
		t.Errorf("expected 60 points spent, got %d", red.PointsSpent)
	}

	// 4. Fulfill redemption
	w = executeRequest(handler, "POST", "/api/redemptions/"+red.ID+"/fulfill", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on fulfill, got %d", w.Code)
	}
}

func TestSocialActivityAndReset(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. Send nudge
	w := executeRequest(handler, "POST", "/api/chores/c-recycling-rr/nudge", models.ChoreNudgeRequest{
		SenderMemberID: "m-sarah",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 2. Add comment
	w = executeRequest(handler, "POST", "/api/chores/c-recycling-rr/comments", models.ChoreCommentCreateRequest{
		MemberID: "m-sarah",
		Message:  "Reminder that glass recycling goes in the blue bin!",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// 3. Get comments
	w = executeRequest(handler, "GET", "/api/chores/c-recycling-rr/comments", nil)
	var comments []models.ChoreCommentWithMember
	_ = json.Unmarshal(w.Body.Bytes(), &comments)
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	// 4. Get activity
	w = executeRequest(handler, "GET", "/api/households/h-roommates/activity", nil)
	var acts []models.ActivityLog
	_ = json.Unmarshal(w.Body.Bytes(), &acts)
	if len(acts) < 2 {
		t.Fatalf("expected multiple activity logs, got %d", len(acts))
	}

	// 5. Reset demo
	w = executeRequest(handler, "POST", "/api/reset", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
