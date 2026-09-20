package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/store"
	"github.com/choresync/backend/internal/telemetry"
)

type SwapHandler struct {
	Store store.Store
}

func NewSwapHandler(s store.Store) *SwapHandler {
	return &SwapHandler{Store: s}
}

// GetSwaps retrieves all swap requests for a household.
func (h *SwapHandler) GetSwaps(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if _, err := h.Store.GetHousehold(householdID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	swaps := h.Store.GetSwaps(householdID)
	respondJSON(w, http.StatusOK, swaps)
}

// RequestChoreSwap proposes a swap via chore ID route /api/chores/{chore_id}/swap.
func (h *SwapHandler) RequestChoreSwap(w http.ResponseWriter, r *http.Request) {
	choreID := chi.URLParam(r, "chore_id")
	chore, err := h.Store.GetChore(choreID)
	if err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Chore with ID '%s' not found", choreID), http.StatusNotFound, nil)
		return
	}

	var req models.ChoreSwapCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.RequesterID) == "" {
		respondError(w, "BAD_REQUEST", "requester_id is required", http.StatusBadRequest, nil)
		return
	}

	swap, err := h.Store.CreateSwap(chore.HouseholdID, choreID, req.RequesterID, req.TargetMemberID, req.Reason)
	if err != nil {
		handleStoreError(w, err, "Failed to create swap")
		return
	}

	respondJSON(w, http.StatusCreated, swap)
}

// CreateSwap proposes a swap via /api/swaps.
func (h *SwapHandler) CreateSwap(w http.ResponseWriter, r *http.Request) {
	var req models.SwapCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.ChoreID) == "" || strings.TrimSpace(req.RequesterID) == "" {
		respondError(w, "BAD_REQUEST", "chore_id and requester_id are required", http.StatusBadRequest, nil)
		return
	}

	chore, err := h.Store.GetChore(req.ChoreID)
	if err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Chore with ID '%s' not found", req.ChoreID), http.StatusNotFound, nil)
		return
	}

	swap, err := h.Store.CreateSwap(chore.HouseholdID, req.ChoreID, req.RequesterID, req.TargetMemberID, req.Reason)
	if err != nil {
		handleStoreError(w, err, "Failed to create swap")
		return
	}

	telemetry.RecordChoreSwap(r.Context(), "proposed")
	telemetry.Info(r.Context(), "Chore swap proposed", "swap_id", swap.ID, "chore_id", req.ChoreID, "requester_id", req.RequesterID)
	span := trace.SpanFromContext(r.Context())
	if span != nil {
		span.SetAttributes(
			attribute.String("swap.id", swap.ID),
			attribute.String("chore.id", req.ChoreID),
			attribute.String("swap.status", "proposed"),
		)
	}

	respondJSON(w, http.StatusCreated, swap)
}

// AcceptSwap transfers chore assignee to the accepting member.
func (h *SwapHandler) AcceptSwap(w http.ResponseWriter, r *http.Request) {
	swapID := chi.URLParam(r, "swap_id")

	var req models.SwapAcceptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.AcceptorMemberID) == "" {
		respondError(w, "BAD_REQUEST", "acceptor_member_id is required", http.StatusBadRequest, nil)
		return
	}

	err := h.Store.AcceptSwap(swapID, req.AcceptorMemberID)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Swap with ID '%s' not found or invalid", swapID))
		return
	}

	telemetry.RecordChoreSwap(r.Context(), "accepted")
	telemetry.Info(r.Context(), "Chore swap accepted", "swap_id", swapID, "acceptor_id", req.AcceptorMemberID)
	span := trace.SpanFromContext(r.Context())
	if span != nil {
		span.SetAttributes(
			attribute.String("swap.id", swapID),
			attribute.String("swap.status", "accepted"),
		)
	}

	respondJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Swap accepted and chore reassigned",
	})
}

// RejectSwap rejects a proposed swap.
func (h *SwapHandler) RejectSwap(w http.ResponseWriter, r *http.Request) {
	swapID := chi.URLParam(r, "swap_id")
	err := h.Store.RejectSwap(swapID)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Swap with ID '%s' not found", swapID))
		return
	}

	telemetry.RecordChoreSwap(r.Context(), "rejected")
	telemetry.Info(r.Context(), "Chore swap rejected", "swap_id", swapID)
	span := trace.SpanFromContext(r.Context())
	if span != nil {
		span.SetAttributes(
			attribute.String("swap.id", swapID),
			attribute.String("swap.status", "rejected"),
		)
	}

	respondJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Swap rejected successfully",
	})
}

// CancelSwap cancels a swap proposed by the requester.
func (h *SwapHandler) CancelSwap(w http.ResponseWriter, r *http.Request) {
	swapID := chi.URLParam(r, "swap_id")
	err := h.Store.CancelSwap(swapID)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Swap with ID '%s' not found", swapID))
		return
	}

	telemetry.RecordChoreSwap(r.Context(), "cancelled")
	telemetry.Info(r.Context(), "Chore swap cancelled", "swap_id", swapID)
	span := trace.SpanFromContext(r.Context())
	if span != nil {
		span.SetAttributes(
			attribute.String("swap.id", swapID),
			attribute.String("swap.status", "cancelled"),
		)
	}

	respondJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Swap cancelled successfully",
	})
}
