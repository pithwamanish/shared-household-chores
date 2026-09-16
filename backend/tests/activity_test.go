package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/choresync/backend/internal/models"
)

func TestActivityListAndReverseChronologicalOrder(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. List activity for h-roommates
	w := executeRequest(handler, "GET", "/api/v1/households/h-roommates/activity", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK listing activities, got %d: %s", w.Code, w.Body.String())
	}

	var activities []models.ActivityLog
	if err := json.Unmarshal(w.Body.Bytes(), &activities); err != nil {
		t.Fatalf("failed to decode activities: %v", err)
	}

	if len(activities) < 2 {
		t.Fatalf("expected at least 2 activities in seed data, got %d", len(activities))
	}

	// Verify reverse chronological ordering (newest first)
	for i := 0; i < len(activities)-1; i++ {
		t1, err1 := time.Parse(time.RFC3339, activities[i].CreatedAt)
		t2, err2 := time.Parse(time.RFC3339, activities[i+1].CreatedAt)
		if err1 == nil && err2 == nil {
			if t1.Before(t2) {
				t.Errorf("activity at index %d (%s) is older than index %d (%s)", i, activities[i].CreatedAt, i+1, activities[i+1].CreatedAt)
			}
		}
	}
}

func TestActivityLifecycleEmission(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. Create a chore -> should emit chore_created
	newChore := models.ChoreCreateRequest{
		HouseholdID:    "h-roommates",
		Title:          "Deep Clean Oven",
		Category:       "kitchen",
		EffortPoints:   30,
		AssignmentType: models.ChoreAssignmentOpenPool,
		CreatedBy:      "m-sarah",
	}
	wChore := executeRequest(handler, "POST", "/api/v1/households/h-roommates/chores", newChore)
	if wChore.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created creating chore, got %d", wChore.Code)
	}

	// 2. Fetch activity log to check latest event
	wAct := executeRequest(handler, "GET", "/api/v1/households/h-roommates/activity", nil)
	var activities []models.ActivityLog
	_ = json.Unmarshal(wAct.Body.Bytes(), &activities)

	if len(activities) == 0 {
		t.Fatalf("expected activities, got 0")
	}
	latest := activities[0]
	if latest.EventType != models.ActivityChoreCreated {
		t.Errorf("expected latest event chore_created, got %s", latest.EventType)
	}
	if latest.ActorID != "m-sarah" {
		t.Errorf("expected actor m-sarah, got %s", latest.ActorID)
	}
}

func TestActivityFilteringAndPagination(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. Filter by entity_type=chore
	wChore := executeRequest(handler, "GET", "/api/v1/households/h-roommates/activity?entity_type=chore", nil)
	if wChore.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wChore.Code)
	}
	var choreActs []models.ActivityLog
	_ = json.Unmarshal(wChore.Body.Bytes(), &choreActs)
	for _, act := range choreActs {
		if act.EntityType != "chore" {
			t.Errorf("expected entity_type 'chore', got '%s'", act.EntityType)
		}
	}

	// 2. Filter by entity_type=reward
	wReward := executeRequest(handler, "GET", "/api/v1/households/h-roommates/activity?entity_type=reward", nil)
	if wReward.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wReward.Code)
	}
	var rewardActs []models.ActivityLog
	_ = json.Unmarshal(wReward.Body.Bytes(), &rewardActs)
	for _, act := range rewardActs {
		if act.EntityType != "reward" {
			t.Errorf("expected entity_type 'reward', got '%s'", act.EntityType)
		}
	}

	// 3. Test limit parameter
	wLimit := executeRequest(handler, "GET", "/api/v1/households/h-roommates/activity?limit=2", nil)
	if wLimit.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wLimit.Code)
	}
	var limitedActs []models.ActivityLog
	_ = json.Unmarshal(wLimit.Body.Bytes(), &limitedActs)
	if len(limitedActs) > 2 {
		t.Errorf("expected at most 2 items with limit=2, got %d", len(limitedActs))
	}

	// 4. Test offset parameter
	wAll := executeRequest(handler, "GET", "/api/v1/households/h-roommates/activity", nil)
	var allActs []models.ActivityLog
	_ = json.Unmarshal(wAll.Body.Bytes(), &allActs)

	if len(allActs) >= 2 {
		wOffset := executeRequest(handler, "GET", "/api/v1/households/h-roommates/activity?offset=1", nil)
		var offsetActs []models.ActivityLog
		_ = json.Unmarshal(wOffset.Body.Bytes(), &offsetActs)
		if len(offsetActs) != len(allActs)-1 {
			t.Errorf("expected %d items with offset=1, got %d", len(allActs)-1, len(offsetActs))
		}
		if len(offsetActs) > 0 && offsetActs[0].ID != allActs[1].ID {
			t.Errorf("expected offset item to match second item of full list")
		}
	}
}

func TestSocialNudgeAndComments(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. Send nudge for c-vacuum-living (assigned to m-sarah)
	nudgeReq := models.ChoreNudgeRequest{
		SenderMemberID: "m-liam",
	}
	wNudge := executeRequest(handler, "POST", "/api/v1/chores/c-vacuum-living/nudge", nudgeReq)
	if wNudge.Code != http.StatusOK {
		t.Fatalf("expected 200 OK sending nudge, got %d: %s", wNudge.Code, wNudge.Body.String())
	}
	var nudgeResp models.ChoreNudgeResponse
	if err := json.Unmarshal(wNudge.Body.Bytes(), &nudgeResp); err != nil {
		t.Fatalf("failed to decode nudge response: %v", err)
	}
	if !nudgeResp.Success {
		t.Errorf("expected nudge success true")
	}

	// Verify nudge appeared in activity feed
	wAct := executeRequest(handler, "GET", "/api/v1/households/h-roommates/activity", nil)
	var activities []models.ActivityLog
	_ = json.Unmarshal(wAct.Body.Bytes(), &activities)
	if len(activities) == 0 || activities[0].EventType != models.ActivityChoreNudge {
		t.Errorf("expected chore_nudge in activity feed, got %v", activities[0].EventType)
	}

	// 2. Add comment to chore
	commentReq := models.ChoreCommentCreateRequest{
		MemberID: "m-liam",
		Message:  "I bought more vacuum bags, they are in the hall closet!",
	}
	wComm := executeRequest(handler, "POST", "/api/v1/chores/c-vacuum-living/comments", commentReq)
	if wComm.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created posting comment, got %d: %s", wComm.Code, wComm.Body.String())
	}

	// 3. List comments for chore
	wGetComm := executeRequest(handler, "GET", "/api/v1/chores/c-vacuum-living/comments", nil)
	if wGetComm.Code != http.StatusOK {
		t.Fatalf("expected 200 OK getting comments, got %d", wGetComm.Code)
	}
	var comments []models.ChoreCommentWithMember
	if err := json.Unmarshal(wGetComm.Body.Bytes(), &comments); err != nil {
		t.Fatalf("failed to decode comments: %v", err)
	}
	if len(comments) == 0 {
		t.Fatalf("expected at least 1 comment, got 0")
	}
	if comments[len(comments)-1].Comment.Message != commentReq.Message {
		t.Errorf("expected comment message '%s', got '%s'", commentReq.Message, comments[len(comments)-1].Comment.Message)
	}
}
