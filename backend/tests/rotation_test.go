package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/choresync/backend/internal/models"
)

func TestRoundRobinRotationSuccess(t *testing.T) {
	handler, _ := setupTestRouter()

	// c-recycling-rr starts at index 1 (m-liam), members: ["m-sarah", "m-liam", "m-maya", "m-noah"]
	// 1. First rotation -> index 2 (m-maya)
	w := executeRequest(handler, "POST", "/api/v1/households/h-roommates/chores/c-recycling-rr/rotate", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var chore models.Chore
	if err := json.Unmarshal(w.Body.Bytes(), &chore); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if chore.CurrentRotationIndex == nil || *chore.CurrentRotationIndex != 2 {
		t.Errorf("expected rotation index 2, got %v", chore.CurrentRotationIndex)
	}
	if chore.CurrentAssigneeID == nil || *chore.CurrentAssigneeID != "m-maya" {
		t.Errorf("expected assignee m-maya, got %v", chore.CurrentAssigneeID)
	}
	if chore.Status != models.ChoreStatusAssigned {
		t.Errorf("expected status assigned, got %s", chore.Status)
	}

	// 2. Second rotation -> index 3 (m-noah)
	w = executeRequest(handler, "POST", "/api/v1/households/h-roommates/chores/c-recycling-rr/rotate", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	_ = json.Unmarshal(w.Body.Bytes(), &chore)
	if chore.CurrentRotationIndex == nil || *chore.CurrentRotationIndex != 3 {
		t.Errorf("expected rotation index 3, got %v", chore.CurrentRotationIndex)
	}
	if chore.CurrentAssigneeID == nil || *chore.CurrentAssigneeID != "m-noah" {
		t.Errorf("expected assignee m-noah, got %v", chore.CurrentAssigneeID)
	}

	// 3. Third rotation -> wraps around to index 0 (m-sarah)
	w = executeRequest(handler, "POST", "/api/v1/households/h-roommates/chores/c-recycling-rr/rotate", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	_ = json.Unmarshal(w.Body.Bytes(), &chore)
	if chore.CurrentRotationIndex == nil || *chore.CurrentRotationIndex != 0 {
		t.Errorf("expected rotation index 0 (wrap-around), got %v", chore.CurrentRotationIndex)
	}
	if chore.CurrentAssigneeID == nil || *chore.CurrentAssigneeID != "m-sarah" {
		t.Errorf("expected assignee m-sarah, got %v", chore.CurrentAssigneeID)
	}

	// 4. Fourth rotation -> index 1 (m-liam)
	w = executeRequest(handler, "POST", "/api/v1/households/h-roommates/chores/c-recycling-rr/rotate", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	_ = json.Unmarshal(w.Body.Bytes(), &chore)
	if chore.CurrentRotationIndex == nil || *chore.CurrentRotationIndex != 1 {
		t.Errorf("expected rotation index 1, got %v", chore.CurrentRotationIndex)
	}
	if chore.CurrentAssigneeID == nil || *chore.CurrentAssigneeID != "m-liam" {
		t.Errorf("expected assignee m-liam, got %v", chore.CurrentAssigneeID)
	}
}

func TestRoundRobinRotationNonRoundRobinError(t *testing.T) {
	handler, _ := setupTestRouter()

	// c-kitchen-deep is direct assignment, not round-robin
	w := executeRequest(handler, "POST", "/api/v1/households/h-roommates/chores/c-kitchen-deep/rotate", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for non-round-robin chore, got %d: %s", w.Code, w.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error JSON: %v", err)
	}
	if errResp.Error != "BAD_REQUEST" {
		t.Errorf("expected error code BAD_REQUEST, got %s", errResp.Error)
	}
}

func TestRoundRobinRotationNotFound(t *testing.T) {
	handler, _ := setupTestRouter()

	w := executeRequest(handler, "POST", "/api/v1/households/h-roommates/chores/c-non-existent/rotate", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d: %s", w.Code, w.Body.String())
	}

	var errResp models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error JSON: %v", err)
	}
	if errResp.Error != "RESOURCE_NOT_FOUND" {
		t.Errorf("expected error code RESOURCE_NOT_FOUND, got %s", errResp.Error)
	}
}

func TestRoundRobinRotationDirectRoute(t *testing.T) {
	handler, _ := setupTestRouter()

	// Test direct route /api/chores/{id}/rotate and /api/v1/chores/{id}/rotate
	w := executeRequest(handler, "POST", "/api/chores/c-recycling-rr/rotate", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	w = executeRequest(handler, "POST", "/api/v1/chores/c-recycling-rr/rotate", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRecurrenceSpawnsNextChoreWithRotatedAssignee(t *testing.T) {
	handler, _ := setupTestRouter()

	// c-recycling-rr has weekly recurrence, round-robin, assigned to Liam (index 1)
	completePayload := models.ChoreCompleteRequest{
		MemberID: "m-liam",
	}
	w := executeRequest(handler, "POST", "/api/v1/chores/c-recycling-rr/complete", completePayload)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var completeResp models.ChoreCompleteResponse
	if err := json.Unmarshal(w.Body.Bytes(), &completeResp); err != nil {
		t.Fatalf("failed to decode complete response: %v", err)
	}
	if completeResp.Chore.Status != models.ChoreStatusCompleted {
		t.Errorf("expected completed chore to have status completed, got %s", completeResp.Chore.Status)
	}

	// Verify all chores for household: a new instance of "Recycling & Garbage to Curb" should exist
	w = executeRequest(handler, "GET", "/api/v1/households/h-roommates/chores", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var chores []models.Chore
	if err := json.Unmarshal(w.Body.Bytes(), &chores); err != nil {
		t.Fatalf("failed to decode chores: %v", err)
	}

	var spawnedChore *models.Chore
	for _, c := range chores {
		if c.ID != "c-recycling-rr" && c.Title == "Recycling & Garbage to Curb" {
			sc := c
			spawnedChore = &sc
			break
		}
	}

	if spawnedChore == nil {
		t.Fatalf("expected a newly spawned recurring chore instance, none found")
	}

	if spawnedChore.Status != models.ChoreStatusAssigned {
		t.Errorf("expected spawned chore status assigned, got %s", spawnedChore.Status)
	}
	// Next assignee should be rotated to index 2 (m-maya)
	if spawnedChore.CurrentRotationIndex == nil || *spawnedChore.CurrentRotationIndex != 2 {
		t.Errorf("expected spawned chore rotation index 2, got %v", spawnedChore.CurrentRotationIndex)
	}
	if spawnedChore.CurrentAssigneeID == nil || *spawnedChore.CurrentAssigneeID != "m-maya" {
		t.Errorf("expected spawned chore assignee m-maya, got %v", spawnedChore.CurrentAssigneeID)
	}
	if spawnedChore.DueDate == nil {
		t.Errorf("expected spawned chore to have a due date")
	} else {
		_, err := time.Parse(time.RFC3339, *spawnedChore.DueDate)
		if err != nil {
			t.Errorf("invalid due date format: %v", err)
		}
	}
}
