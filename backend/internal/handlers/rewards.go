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

type RewardHandler struct {
	Store store.Store
}

func NewRewardHandler(s store.Store) *RewardHandler {
	return &RewardHandler{Store: s}
}

// GetRewards lists active rewards for a household.
func (h *RewardHandler) GetRewards(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if _, err := h.Store.GetHousehold(householdID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	rewards := h.Store.GetRewards(householdID)
	respondJSON(w, http.StatusOK, rewards)
}

// CreateReward creates a new reward in the catalog.
func (h *RewardHandler) CreateReward(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if _, err := h.Store.GetHousehold(householdID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	var req models.RewardItemCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		respondError(w, "BAD_REQUEST", "Reward title is required", http.StatusBadRequest, nil)
		return
	}
	if req.PointsCost <= 0 {
		respondError(w, "BAD_REQUEST", "points_cost must be greater than zero", http.StatusBadRequest, nil)
		return
	}

	reward, err := h.Store.CreateReward(householdID, req)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Household with ID '%s' not found", householdID))
		return
	}

	respondJSON(w, http.StatusCreated, reward)
}

// RedeemReward allows a member to spend points on a reward.
func (h *RewardHandler) RedeemReward(w http.ResponseWriter, r *http.Request) {
	rewardID := chi.URLParam(r, "reward_id")

	var req models.RewardRedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.MemberID) == "" {
		respondError(w, "BAD_REQUEST", "member_id is required", http.StatusBadRequest, nil)
		return
	}

	redemption, err := h.Store.RedeemReward(rewardID, req.MemberID)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Reward with ID '%s' or Member not found", rewardID))
		return
	}

	respondJSON(w, http.StatusCreated, redemption)
}

// FulfillReward marks a redemption as delivered.
func (h *RewardHandler) FulfillReward(w http.ResponseWriter, r *http.Request) {
	redemptionID := chi.URLParam(r, "redemption_id")

	err := h.Store.FulfillReward(redemptionID)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Redemption with ID '%s' not found", redemptionID))
		return
	}

	respondJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Reward redemption fulfilled successfully",
	})
}

// GetRedemptions retrieves all reward redemptions for a household.
func (h *RewardHandler) GetRedemptions(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if _, err := h.Store.GetHousehold(householdID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	redemptions := h.Store.GetRedemptions(householdID)
	respondJSON(w, http.StatusOK, redemptions)
}
