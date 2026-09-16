package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/choresync/backend/internal/models"
)

func TestSwapProposeAndAccept(t *testing.T) {
	handler, appStore := setupTestRouter()

	// 1. Liam proposes swap for c-recycling-rr (assigned to m-liam)
	choreBefore, err := appStore.GetChore("c-recycling-rr")
	if err != nil {
		t.Fatalf("failed to fetch chore: %v", err)
	}
	if choreBefore.CurrentAssigneeID == nil || *choreBefore.CurrentAssigneeID != "m-liam" {
		t.Fatalf("expected chore initially assigned to m-liam, got %v", choreBefore.CurrentAssigneeID)
	}

	reason := "Studying for midterm exams this evening"
	targetMember := "m-maya"
	proposeReq := models.ChoreSwapCreateRequest{
		RequesterID:    "m-liam",
		TargetMemberID: &targetMember,
		Reason:         &reason,
	}

	w := executeRequest(handler, "POST", "/api/v1/chores/c-recycling-rr/swap", proposeReq)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created proposing swap, got %d: %s", w.Code, w.Body.String())
	}

	var createdSwap models.ChoreSwapRequest
	if err := json.Unmarshal(w.Body.Bytes(), &createdSwap); err != nil {
		t.Fatalf("failed to unmarshal swap: %v", err)
	}

	if createdSwap.Status != models.SwapStatusPending {
		t.Errorf("expected swap status pending, got %s", createdSwap.Status)
	}
	if createdSwap.RequesterID != "m-liam" {
		t.Errorf("expected requester m-liam, got %s", createdSwap.RequesterID)
	}

	// 2. Maya accepts the swap proposal
	acceptReq := models.SwapAcceptRequest{
		AcceptorMemberID: "m-maya",
	}
	w = executeRequest(handler, "POST", "/api/v1/swaps/"+createdSwap.ID+"/accept", acceptReq)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK accepting swap, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Verify chore assignee is now Maya
	choreAfter, err := appStore.GetChore("c-recycling-rr")
	if err != nil {
		t.Fatalf("failed to fetch chore after swap: %v", err)
	}
	if choreAfter.CurrentAssigneeID == nil || *choreAfter.CurrentAssigneeID != "m-maya" {
		t.Errorf("expected chore reassigned to m-maya, got %v", choreAfter.CurrentAssigneeID)
	}

	// 4. Verify cannot accept already accepted swap
	w = executeRequest(handler, "POST", "/api/v1/swaps/"+createdSwap.ID+"/accept", acceptReq)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request accepting already accepted swap, got %d", w.Code)
	}
}

func TestSwapRejectAndCancelWorkflow(t *testing.T) {
	handler, _ := setupTestRouter()

	// Propose swap for c-kitchen-deep (assigned to m-sarah)
	reason := "Out of town this weekend"
	targetMember := "m-liam"
	proposeReq := models.ChoreSwapCreateRequest{
		RequesterID:    "m-sarah",
		TargetMemberID: &targetMember,
		Reason:         &reason,
	}

	w := executeRequest(handler, "POST", "/api/v1/chores/c-kitchen-deep/swap", proposeReq)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created proposing swap, got %d: %s", w.Code, w.Body.String())
	}

	var swap models.ChoreSwapRequest
	_ = json.Unmarshal(w.Body.Bytes(), &swap)

	// Reject swap
	w = executeRequest(handler, "POST", "/api/v1/swaps/"+swap.ID+"/reject", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK rejecting swap, got %d: %s", w.Code, w.Body.String())
	}

	// Cancel swap testing with second proposal
	w2 := executeRequest(handler, "POST", "/api/v1/chores/c-kitchen-deep/swap", proposeReq)
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created proposing second swap, got %d: %s", w2.Code, w2.Body.String())
	}
	var swap2 models.ChoreSwapRequest
	_ = json.Unmarshal(w2.Body.Bytes(), &swap2)

	// Requester cancels swap
	wCancel := executeRequest(handler, "DELETE", "/api/v1/swaps/"+swap2.ID, nil)
	if wCancel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK cancelling swap, got %d: %s", wCancel.Code, wCancel.Body.String())
	}
}

func TestSwapGuardrailsAndValidations(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. Proposing swap for chore not owned by requester
	notOwnerReq := models.ChoreSwapCreateRequest{
		RequesterID: "m-maya", // chore c-kitchen-deep is assigned to m-sarah
	}
	w := executeRequest(handler, "POST", "/api/v1/chores/c-kitchen-deep/swap", notOwnerReq)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request when non-owner requests swap, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Proposing swap in household where allow_swaps is disabled
	// Disable swaps in h-roommates
	allowSwaps := false
	wPatch := executeRequest(handler, "PATCH", "/api/v1/households/h-roommates/settings", models.HouseholdSettingsUpdate{
		AllowSwaps: &allowSwaps,
	})
	if wPatch.Code != http.StatusOK {
		t.Fatalf("expected 200 OK disabling swaps, got %d", wPatch.Code)
	}

	validReq := models.ChoreSwapCreateRequest{
		RequesterID: "m-sarah",
	}
	wDisabled := executeRequest(handler, "POST", "/api/v1/chores/c-kitchen-deep/swap", validReq)
	if wDisabled.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request when swaps disabled, got %d: %s", wDisabled.Code, wDisabled.Body.String())
	}

	// 3. Proposing swap for completed chore
	// Re-enable swaps
	allowSwapsTrue := true
	_ = executeRequest(handler, "PATCH", "/api/v1/households/h-roommates/settings", models.HouseholdSettingsUpdate{
		AllowSwaps: &allowSwapsTrue,
	})
	// Complete chore first
	_ = executeRequest(handler, "POST", "/api/v1/chores/c-kitchen-deep/complete", models.ChoreCompleteRequest{
		MemberID: "m-sarah",
	})
	// Attempt swap
	wCompleted := executeRequest(handler, "POST", "/api/v1/chores/c-kitchen-deep/swap", validReq)
	if wCompleted.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request swapping completed chore, got %d: %s", wCompleted.Code, wCompleted.Body.String())
	}
}
