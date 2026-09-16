package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/store"
)

type ChoreHandler struct {
	Store store.Store
}

func NewChoreHandler(s store.Store) *ChoreHandler {
	return &ChoreHandler{Store: s}
}

// GetChores queries chores for a household with optional query filtering.
func (h *ChoreHandler) GetChores(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if _, err := h.Store.GetHousehold(householdID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	q := r.URL.Query()
	status := q.Get("status")
	category := q.Get("category")
	search := q.Get("search")
	view := q.Get("view")
	assigneeID := q.Get("assignee_id")

	chores := h.Store.GetChores(householdID, status, category, search, view, assigneeID)
	respondJSON(w, http.StatusOK, chores)
}

// CreateChore creates a new chore in the specified household.
func (h *ChoreHandler) CreateChore(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if _, err := h.Store.GetHousehold(householdID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	var req models.ChoreCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	req.HouseholdID = householdID
	if strings.TrimSpace(req.Title) == "" {
		respondError(w, "BAD_REQUEST", "Chore title is required", http.StatusBadRequest, nil)
		return
	}
	if req.EffortPoints < 0 {
		respondError(w, "BAD_REQUEST", "Effort points must be non-negative", http.StatusBadRequest, nil)
		return
	}

	chore, err := h.Store.CreateChore(req)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Household with ID '%s' not found", householdID))
		return
	}

	respondJSON(w, http.StatusCreated, chore)
}

// GetChore retrieves a single chore by ID.
func (h *ChoreHandler) GetChore(w http.ResponseWriter, r *http.Request) {
	choreID := chi.URLParam(r, "chore_id")
	chore, err := h.Store.GetChore(choreID)
	if err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Chore with ID '%s' not found", choreID), http.StatusNotFound, nil)
		return
	}

	respondJSON(w, http.StatusOK, chore)
}

// UpdateChore modifies fields of an existing chore.
func (h *ChoreHandler) UpdateChore(w http.ResponseWriter, r *http.Request) {
	choreID := chi.URLParam(r, "chore_id")

	var req models.ChoreUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	chore, err := h.Store.UpdateChore(choreID, req)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Chore with ID '%s' not found", choreID))
		return
	}

	respondJSON(w, http.StatusOK, chore)
}

// DeleteChore removes a chore from the board.
func (h *ChoreHandler) DeleteChore(w http.ResponseWriter, r *http.Request) {
	choreID := chi.URLParam(r, "chore_id")
	if err := h.Store.DeleteChore(choreID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Chore with ID '%s' not found", choreID), http.StatusNotFound, nil)
		return
	}

	respondJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Chore deleted successfully",
	})
}

// ClaimChore assigns an open pool chore to a member.
func (h *ChoreHandler) ClaimChore(w http.ResponseWriter, r *http.Request) {
	choreID := chi.URLParam(r, "chore_id")

	var req models.ChoreClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.MemberID) == "" {
		respondError(w, "BAD_REQUEST", "member_id is required", http.StatusBadRequest, nil)
		return
	}

	chore, err := h.Store.ClaimChore(choreID, req.MemberID)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Chore with ID '%s' or Member not found", choreID))
		return
	}

	respondJSON(w, http.StatusOK, chore)
}

// CompleteChore submits chore completion.
func (h *ChoreHandler) CompleteChore(w http.ResponseWriter, r *http.Request) {
	choreID := chi.URLParam(r, "chore_id")

	var req models.ChoreCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.MemberID) == "" {
		respondError(w, "BAD_REQUEST", "member_id is required", http.StatusBadRequest, nil)
		return
	}

	chore, comp, pointsAwarded, err := h.Store.CompleteChore(choreID, req.MemberID, req.ProofNotes, req.ProofPhotoURL)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Chore with ID '%s' or Member not found", choreID))
		return
	}

	resp := models.ChoreCompleteResponse{
		Chore:         *chore,
		Completion:    comp,
		PointsAwarded: pointsAwarded,
	}

	respondJSON(w, http.StatusOK, resp)
}

// RotateChore advances a round-robin chore to the next member in sequence.
func (h *ChoreHandler) RotateChore(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	choreID := chi.URLParam(r, "chore_id")
	if choreID == "" {
		choreID = chi.URLParam(r, "choreId")
	}

	chore, err := h.Store.RotateChore(householdID, choreID)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Chore with ID '%s' not found", choreID))
		return
	}

	respondJSON(w, http.StatusOK, chore)
}

