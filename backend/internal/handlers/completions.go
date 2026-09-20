package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/choresync/backend/internal/auth"
	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/store"
	"github.com/choresync/backend/internal/telemetry"
)

// CompletionHandler handles chore completion verification, approval, and rejection.
type CompletionHandler struct {
	Store store.Store
}

// NewCompletionHandler instantiates a new CompletionHandler.
func NewCompletionHandler(s store.Store) *CompletionHandler {
	return &CompletionHandler{Store: s}
}

// GetPendingApprovals returns all completions awaiting verification for a household.
func (h *CompletionHandler) GetPendingApprovals(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if householdID == "" {
		householdID = chi.URLParam(r, "householdId")
	}

	if _, err := h.Store.GetHousehold(householdID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	approvals := h.Store.GetPendingApprovals(householdID)
	respondJSON(w, http.StatusOK, approvals)
}

// ApproveChore allows an admin/parent to approve a chore completion and award effort points.
func (h *CompletionHandler) ApproveChore(w http.ResponseWriter, r *http.Request) {
	completionID := chi.URLParam(r, "completion_id")
	if completionID == "" {
		completionID = chi.URLParam(r, "completionId")
	}

	var req models.ApproveCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	caller, hasCaller := auth.GetCaller(r.Context())
	adminID := strings.TrimSpace(req.AdminMemberID)
	if adminID == "" && hasCaller {
		adminID = caller.Sub
	}

	if adminID == "" {
		respondError(w, "BAD_REQUEST", "admin_member_id is required", http.StatusBadRequest, nil)
		return
	}

	if hasCaller && caller.Role != models.MemberRoleAdmin {
		respondError(w, "FORBIDDEN", fmt.Sprintf("Admin privileges required: caller '%s' has role '%s'", caller.Sub, caller.Role), http.StatusForbidden, nil)
		return
	}

	err := h.Store.ApproveChore(completionID, adminID)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Completion with ID '%s' not found", completionID))
		return
	}

	telemetry.RecordChoreApproval(r.Context(), "approved", 10)
	telemetry.Info(r.Context(), "Chore approved", "completion_id", completionID, "admin_id", adminID)
	span := trace.SpanFromContext(r.Context())
	if span != nil {
		span.SetAttributes(
			attribute.String("completion.id", completionID),
			attribute.String("admin.id", adminID),
			attribute.String("approval.decision", "approved"),
		)
	}

	respondJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Chore completion approved successfully",
	})
}

// RejectChore allows an admin/parent to reject a completion with feedback, returning chore for rework.
func (h *CompletionHandler) RejectChore(w http.ResponseWriter, r *http.Request) {
	completionID := chi.URLParam(r, "completion_id")
	if completionID == "" {
		completionID = chi.URLParam(r, "completionId")
	}

	var req models.RejectCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	caller, hasCaller := auth.GetCaller(r.Context())
	adminID := strings.TrimSpace(req.AdminMemberID)
	if adminID == "" && hasCaller {
		adminID = caller.Sub
	}

	if adminID == "" {
		respondError(w, "BAD_REQUEST", "admin_member_id is required", http.StatusBadRequest, nil)
		return
	}

	if hasCaller && caller.Role != models.MemberRoleAdmin {
		respondError(w, "FORBIDDEN", fmt.Sprintf("Admin privileges required: caller '%s' has role '%s'", caller.Sub, caller.Role), http.StatusForbidden, nil)
		return
	}

	if strings.TrimSpace(req.Reason) == "" {
		respondError(w, "BAD_REQUEST", "Reason is required when rejecting a chore", http.StatusBadRequest, nil)
		return
	}

	err := h.Store.RejectChore(completionID, adminID, req.Reason)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Completion with ID '%s' not found", completionID))
		return
	}

	telemetry.RecordChoreApproval(r.Context(), "rejected", 0)
	telemetry.Info(r.Context(), "Chore rejected", "completion_id", completionID, "admin_id", adminID, "reason", req.Reason)
	span := trace.SpanFromContext(r.Context())
	if span != nil {
		span.SetAttributes(
			attribute.String("completion.id", completionID),
			attribute.String("admin.id", adminID),
			attribute.String("approval.decision", "rejected"),
			attribute.String("rejection.reason", req.Reason),
		)
	}

	respondJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Chore completion rejected and returned for rework",
	})
}
