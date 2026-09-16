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

type MemberHandler struct {
	Store store.Store
}

func NewMemberHandler(s store.Store) *MemberHandler {
	return &MemberHandler{Store: s}
}

// GetMembers returns all members for a household.
func (h *MemberHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if _, err := h.Store.GetHousehold(householdID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	members := h.Store.GetMembers(householdID)
	respondJSON(w, http.StatusOK, members)
}

// AddMember adds or invites a member to a household.
func (h *MemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if _, err := h.Store.GetHousehold(householdID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	var req models.MemberCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		respondError(w, "BAD_REQUEST", "Member name is required", http.StatusBadRequest, nil)
		return
	}
	if req.Role != models.MemberRoleAdmin && req.Role != models.MemberRoleMember && req.Role != models.MemberRoleChild {
		respondError(w, "BAD_REQUEST", fmt.Sprintf("Invalid role: '%s'. Must be admin, member, or child", req.Role), http.StatusBadRequest, nil)
		return
	}

	member, err := h.Store.AddMember(householdID, req)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Household with ID '%s' not found", householdID))
		return
	}

	respondJSON(w, http.StatusCreated, member)
}
