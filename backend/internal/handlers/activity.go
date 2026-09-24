package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/choresync/backend/internal/cloud"
	"github.com/choresync/backend/internal/email"
	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/store"
)

type ActivityHandler struct {
	Store store.Store
	Email email.Service
	Queue cloud.QueueService
}

func NewActivityHandler(s store.Store, em email.Service, q cloud.QueueService) *ActivityHandler {
	return &ActivityHandler{Store: s, Email: em, Queue: q}
}

// GetActivities retrieves audit logs for a household with optional query filtering and pagination.
func (h *ActivityHandler) GetActivities(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if _, err := h.Store.GetHousehold(householdID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	query := r.URL.Query()
	entityType := strings.TrimSpace(query.Get("entity_type"))
	eventType := strings.TrimSpace(query.Get("event_type"))

	limit := 0
	if l := query.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	offset := 0
	if o := query.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	activities := h.Store.GetActivitiesFiltered(householdID, entityType, eventType, limit, offset)
	respondJSON(w, http.StatusOK, activities)
}

// SendNudge sends a friendly reminder for a chore.
func (h *ActivityHandler) SendNudge(w http.ResponseWriter, r *http.Request) {
	choreID := chi.URLParam(r, "chore_id")

	var req models.ChoreNudgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.SenderMemberID) == "" {
		respondError(w, "BAD_REQUEST", "sender_member_id is required", http.StatusBadRequest, nil)
		return
	}

	msg, err := h.Store.SendNudge(choreID, req.SenderMemberID)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Chore with ID '%s' or sender not found", choreID))
		return
	}

	emailDispatched := false
	sqsQueued := false

	chore, err := h.Store.GetChore(choreID)
	if err == nil && chore.CurrentAssigneeID != nil {
		assignee, err := h.Store.GetMember(*chore.CurrentAssigneeID)
		sender, _ := h.Store.GetMember(req.SenderMemberID)
		household, _ := h.Store.GetHousehold(chore.HouseholdID)

		senderName := "A roommate"
		if sender != nil {
			senderName = sender.Name
		}
		hhName := "our household"
		if household != nil {
			hhName = household.Name
		}

		if err == nil && assignee != nil && assignee.Email != nil && *assignee.Email != "" {
			dueStr := "today"
			if chore.DueDate != nil {
				dueStr = *chore.DueDate
			}

			// If cloud SQS queue is enabled, enqueue asynchronous reminder job
			if h.Queue != nil && h.Queue.IsEnabled() {
				job := cloud.ReminderJob{
					ChoreID:       chore.ID,
					ChoreTitle:    chore.Title,
					DueDate:       dueStr,
					AssigneeEmail: *assignee.Email,
					AssigneeName:  assignee.Name,
					SenderName:    senderName,
					HouseholdName: hhName,
				}
				_, qErr := h.Queue.PublishReminder(r.Context(), job)
				if qErr == nil {
					sqsQueued = true
					emailDispatched = true
				}
			}

			// Fallback: synchronous dispatch if SQS was not used
			if !sqsQueued && h.Email != nil {
				_ = h.Email.SendChoreReminder(*assignee.Email, assignee.Name, senderName, chore.Title, dueStr, hhName)
				emailDispatched = true
			}
		}
	}

	respondJSON(w, http.StatusOK, models.ChoreNudgeResponse{
		Success:         true,
		Message:         msg,
		EmailDispatched: emailDispatched,
		SQSQueued:       sqsQueued,
	})
}

// GetComments retrieves comments posted on a chore.
func (h *ActivityHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	choreID := chi.URLParam(r, "chore_id")
	if _, err := h.Store.GetChore(choreID); err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Chore with ID '%s' not found", choreID), http.StatusNotFound, nil)
		return
	}

	comments := h.Store.GetComments(choreID)
	respondJSON(w, http.StatusOK, comments)
}

// AddComment posts a new comment on a chore.
func (h *ActivityHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	choreID := chi.URLParam(r, "chore_id")

	var req models.ChoreCommentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.MemberID) == "" {
		respondError(w, "BAD_REQUEST", "member_id is required", http.StatusBadRequest, nil)
		return
	}
	if strings.TrimSpace(req.Message) == "" {
		respondError(w, "BAD_REQUEST", "Message cannot be empty", http.StatusBadRequest, nil)
		return
	}

	comment, err := h.Store.AddComment(choreID, req.MemberID, req.Message)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Chore with ID '%s' or Member not found", choreID))
		return
	}

	respondJSON(w, http.StatusCreated, comment)
}
