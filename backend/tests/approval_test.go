package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/choresync/backend/internal/models"
)

func TestCompleteChoreRequiringApprovalWorkflow(t *testing.T) {
	handler, appStore := setupTestRouter()

	// Initial member points for Leo
	memberBefore, err := appStore.GetMember("m-leo-kid")
	if err != nil {
		t.Fatalf("failed to fetch member: %v", err)
	}
	initialPoints := memberBefore.PointsBalance
	initialStreak := memberBefore.Streak

	// c-fam-dog-walk is in h-family, RequiresApproval: true, EffortPoints: 20
	notes := "Walked Buster in the park for 25 minutes and filled his bowl with fresh water."
	photoURL := "https://images.unsplash.com/photo-1543466835-00a7907e9de1"

	// 1. Leo completes chore requiring approval
	completeReq := models.ChoreCompleteRequest{
		MemberID:      "m-leo-kid",
		ProofNotes:    &notes,
		ProofPhotoURL: &photoURL,
	}
	w := executeRequest(handler, "POST", "/api/v1/chores/c-fam-dog-walk/complete", completeReq)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK submitting completion, got %d: %s", w.Code, w.Body.String())
	}

	var completeResp models.ChoreCompleteResponse
	if err := json.Unmarshal(w.Body.Bytes(), &completeResp); err != nil {
		t.Fatalf("failed to decode completion response: %v", err)
	}

	// Verify pointsAwarded is 0 and status is pending_approval
	if completeResp.PointsAwarded != 0 {
		t.Errorf("expected 0 points awarded before approval, got %d", completeResp.PointsAwarded)
	}
	if completeResp.Chore.Status != models.ChoreStatusPendingApproval {
		t.Errorf("expected chore status pending_approval, got %s", completeResp.Chore.Status)
	}
	if completeResp.Completion == nil || completeResp.Completion.Status != models.CompletionStatusPendingApproval {
		t.Errorf("expected completion status pending_approval, got %v", completeResp.Completion)
	}

	// Verify member points and streak are NOT changed yet
	memberMid, _ := appStore.GetMember("m-leo-kid")
	if memberMid.PointsBalance != initialPoints {
		t.Errorf("expected member points to remain %d, got %d", initialPoints, memberMid.PointsBalance)
	}
	if memberMid.Streak != initialStreak {
		t.Errorf("expected member streak to remain %d, got %d", initialStreak, memberMid.Streak)
	}

	// Verify it appears in pending approvals queue
	w = executeRequest(handler, "GET", "/api/v1/households/h-family/approvals", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from approvals endpoint, got %d", w.Code)
	}
	var approvals []models.PendingApprovalItem
	if err := json.Unmarshal(w.Body.Bytes(), &approvals); err != nil {
		t.Fatalf("failed to decode approvals list: %v", err)
	}

	found := false
	for _, item := range approvals {
		if item.Chore.ID == "c-fam-dog-walk" {
			found = true
			if item.Completion.ProofNotes == nil || *item.Completion.ProofNotes != notes {
				t.Errorf("expected proof notes %s, got %v", notes, item.Completion.ProofNotes)
			}
			if item.Completion.ProofPhotoURL == nil || *item.Completion.ProofPhotoURL != photoURL {
				t.Errorf("expected proof photo %s, got %v", photoURL, item.Completion.ProofPhotoURL)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected c-fam-dog-walk to be in pending approvals queue")
	}
}

func TestNonAdminCannotApproveOrReject(t *testing.T) {
	handler, _ := setupTestRouter()

	// comp-seed-leo-1 is a seed completion awaiting approval in h-family
	compID := "comp-seed-leo-1"

	// 1. Non-admin child attempts to approve -> 403 Forbidden
	w := executeRequest(handler, "POST", "/api/v1/completions/"+compID+"/approve", models.ApproveCompletionRequest{
		AdminMemberID: "m-emma-kid", // Child, not Admin
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for child approving, got %d: %s", w.Code, w.Body.String())
	}
	var errResp models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error JSON: %v", err)
	}
	if errResp.Error != "FORBIDDEN" {
		t.Errorf("expected error code FORBIDDEN, got %s", errResp.Error)
	}

	// 2. Non-admin child attempts to reject -> 403 Forbidden
	w = executeRequest(handler, "POST", "/api/v1/completions/"+compID+"/reject", models.RejectCompletionRequest{
		AdminMemberID: "m-emma-kid",
		Reason:        "Not good enough",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for child rejecting, got %d", w.Code)
	}

	// 3. Admin from different household attempts to approve -> 403 Forbidden
	w = executeRequest(handler, "POST", "/api/v1/completions/"+compID+"/approve", models.ApproveCompletionRequest{
		AdminMemberID: "m-sarah", // Admin in h-roommates, NOT h-family
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for cross-household admin, got %d", w.Code)
	}
}

func TestAdminApproveAwardsPointsAndRecurrence(t *testing.T) {
	handler, appStore := setupTestRouter()

	// comp-seed-leo-1: chore c-fam-bedroom-leo (15 points), submitter m-leo-kid (75 points)
	compID := "comp-seed-leo-1"
	memberBefore, _ := appStore.GetMember("m-leo-kid")
	initialPoints := memberBefore.PointsBalance
	initialEarned := memberBefore.TotalPointsEarned
	initialStreak := memberBefore.Streak

	// Admin Sarah (Mom) approves
	w := executeRequest(handler, "POST", "/api/v1/completions/"+compID+"/approve", models.ApproveCompletionRequest{
		AdminMemberID: "m-sarah-fam",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on admin approval, got %d: %s", w.Code, w.Body.String())
	}

	var succResp models.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &succResp); err != nil {
		t.Fatalf("failed to decode success response: %v", err)
	}
	if !succResp.Success {
		t.Errorf("expected success true, got %v", succResp.Success)
	}

	// Verify points credited
	memberAfter, _ := appStore.GetMember("m-leo-kid")
	if memberAfter.PointsBalance != initialPoints+15 {
		t.Errorf("expected points balance %d, got %d", initialPoints+15, memberAfter.PointsBalance)
	}
	if memberAfter.TotalPointsEarned != initialEarned+15 {
		t.Errorf("expected total points %d, got %d", initialEarned+15, memberAfter.TotalPointsEarned)
	}
	if memberAfter.Streak != initialStreak+1 {
		t.Errorf("expected streak %d, got %d", initialStreak+1, memberAfter.Streak)
	}

	// Verify chore status is completed
	chore, _ := appStore.GetChore("c-fam-bedroom-leo")
	if chore.Status != models.ChoreStatusCompleted {
		t.Errorf("expected chore status completed, got %s", chore.Status)
	}

	// Verify removed from pending approvals queue
	w = executeRequest(handler, "GET", "/api/v1/households/h-family/approvals", nil)
	var approvals []models.PendingApprovalItem
	_ = json.Unmarshal(w.Body.Bytes(), &approvals)
	for _, a := range approvals {
		if a.Completion.ID == compID {
			t.Errorf("expected completion %s to be removed from pending approvals", compID)
		}
	}

	// Verify duplicate approval is rejected with 400 Bad Request
	w = executeRequest(handler, "POST", "/api/v1/completions/"+compID+"/approve", models.ApproveCompletionRequest{
		AdminMemberID: "m-sarah-fam",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on duplicate approval, got %d", w.Code)
	}
}

func TestAdminRejectRevertsChoreAndRequiresFeedback(t *testing.T) {
	handler, appStore := setupTestRouter()

	compID := "comp-seed-leo-1"
	memberBefore, _ := appStore.GetMember("m-leo-kid")
	initialPoints := memberBefore.PointsBalance

	// 1. Rejection with empty reason -> 400 Bad Request
	w := executeRequest(handler, "POST", "/api/v1/completions/"+compID+"/reject", models.RejectCompletionRequest{
		AdminMemberID: "m-david-fam",
		Reason:        "",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for empty reason, got %d", w.Code)
	}

	// 2. Admin rejects with feedback
	reason := "Please put the vacuum back in the hallway closet and wipe the desk surface."
	w = executeRequest(handler, "POST", "/api/v1/completions/"+compID+"/reject", models.RejectCompletionRequest{
		AdminMemberID: "m-david-fam",
		Reason:        reason,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on rejection, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Verify points are NOT credited
	memberAfter, _ := appStore.GetMember("m-leo-kid")
	if memberAfter.PointsBalance != initialPoints {
		t.Errorf("expected points balance to remain %d, got %d", initialPoints, memberAfter.PointsBalance)
	}

	// 4. Verify chore reverted to assigned status
	chore, _ := appStore.GetChore("c-fam-bedroom-leo")
	if chore.Status != models.ChoreStatusAssigned {
		t.Errorf("expected chore to revert to assigned, got %s", chore.Status)
	}

	// 5. Verify removed from pending approvals
	w = executeRequest(handler, "GET", "/api/v1/households/h-family/approvals", nil)
	var approvals []models.PendingApprovalItem
	_ = json.Unmarshal(w.Body.Bytes(), &approvals)
	for _, a := range approvals {
		if a.Completion.ID == compID {
			t.Errorf("expected completion %s to be removed from pending approvals", compID)
		}
	}

	// 6. Cannot approve a rejected completion
	w = executeRequest(handler, "POST", "/api/v1/completions/"+compID+"/approve", models.ApproveCompletionRequest{
		AdminMemberID: "m-david-fam",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request when approving already rejected completion, got %d", w.Code)
	}
}

func TestHouseholdScopedApprovalEndpoints(t *testing.T) {
	handler, appStore := setupTestRouter()

	compID := "comp-seed-leo-1"

	// Household-scoped approve route: POST /api/v1/households/h-family/completions/{completion_id}/approve
	w := executeRequest(handler, "POST", "/api/v1/households/h-family/completions/"+compID+"/approve", models.ApproveCompletionRequest{
		AdminMemberID: "m-david-fam",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from household-scoped route, got %d: %s", w.Code, w.Body.String())
	}

	chore, _ := appStore.GetChore("c-fam-bedroom-leo")
	if chore.Status != models.ChoreStatusCompleted {
		t.Errorf("expected chore status completed, got %s", chore.Status)
	}
}

func TestCompleteAlreadyPendingOrCompletedErrors(t *testing.T) {
	handler, _ := setupTestRouter()

	// c-fam-bedroom-leo is already in pending_approval initially in seed data
	w := executeRequest(handler, "POST", "/api/v1/chores/c-fam-bedroom-leo/complete", models.ChoreCompleteRequest{
		MemberID: "m-leo-kid",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request when submitting completion for chore already in pending_approval, got %d", w.Code)
	}
}
