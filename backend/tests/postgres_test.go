package tests

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/choresync/backend/internal/cloud"
	"github.com/choresync/backend/internal/email"
	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/store"
)

func TestPostgresStoreIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping PostgreSQL integration test: DATABASE_URL not set")
	}

	pgStore, err := store.NewPostgresStore(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgStore.Close()

	// 1. Verify seed fixtures populated
	households := pgStore.GetHouseholds()
	if len(households) < 3 {
		t.Fatalf("Expected at least 3 seeded households in Postgres, got %d", len(households))
	}

	// 2. Verify Member retrieval
	member, err := pgStore.GetMember("m-sarah")
	if err != nil {
		t.Fatalf("Failed to get Sarah Chen from Postgres: %v", err)
	}
	if member.Name != "Sarah Chen" {
		t.Errorf("Expected member name 'Sarah Chen', got '%s'", member.Name)
	}

	// 3. Create, Claim, and Complete a chore
	choreReq := models.ChoreCreateRequest{
		HouseholdID:  "h-roommates",
		Title:        "Wipe Kitchen Counters (Postgres Test)",
		Category:     models.ChoreCategoryCleaning,
		EffortPoints: 15,
		CreatedBy:    "m-sarah",
	}
	createdChore, err := pgStore.CreateChore(choreReq)
	if err != nil {
		t.Fatalf("Failed to create chore in Postgres: %v", err)
	}

	claimedChore, err := pgStore.ClaimChore(createdChore.ID, "m-liam")
	if err != nil {
		t.Fatalf("Failed to claim chore in Postgres: %v", err)
	}
	if claimedChore.Status != models.ChoreStatusInProgress {
		t.Errorf("Expected status in_progress, got '%s'", claimedChore.Status)
	}

	notes := "Done sparkling clean"
	completedChore, completion, pts, err := pgStore.CompleteChore(claimedChore.ID, "m-liam", &notes, nil)
	if err != nil {
		t.Fatalf("Failed to complete chore in Postgres: %v", err)
	}
	if pts != 15 {
		t.Errorf("Expected 15 points awarded, got %d", pts)
	}
	if completedChore.Status != models.ChoreStatusCompleted {
		t.Errorf("Expected chore status completed, got '%s'", completedChore.Status)
	}
	if completion == nil || completion.PointsAwarded != 15 {
		t.Errorf("Expected completion record with 15 pts, got %+v", completion)
	}

	// 4. Verify Activity Logs written to Postgres
	activities := pgStore.GetActivities("h-roommates")
	if len(activities) == 0 {
		t.Errorf("Expected activity logs recorded in Postgres, got 0")
	}

	// 5. Test Reset() reverts to initial seed fixtures
	pgStore.Reset()
	_, err = pgStore.GetChore(createdChore.ID)
	if err != store.ErrNotFound {
		t.Errorf("Expected deleted test chore to return ErrNotFound after Reset(), got %v", err)
	}
}

func TestPostgresQueueIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping PostgreSQL integration test: DATABASE_URL not set")
	}

	pgStore, err := store.NewPostgresStore(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgStore.Close()

	queue := cloud.NewPostgresQueueService(pgStore.Pool())
	if !queue.IsEnabled() {
		t.Fatal("expected PostgresQueueService to be enabled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockEmail := email.NewMockService(email.Config{})
	queue.StartReminderWorker(ctx, mockEmail)

	job := cloud.ReminderJob{
		ChoreID:       "c-dishes",
		ChoreTitle:    "Clean Dishes (Postgres Queue Test)",
		DueDate:       "tonight",
		AssigneeEmail: "test-neon-queue@example.com",
		AssigneeName:  "Test Roommate",
		SenderName:    "Sarah Chen",
		HouseholdName: "Oakwood Flatmates",
	}

	msgID, err := queue.PublishReminder(ctx, job)
	if err != nil {
		t.Fatalf("Failed to publish reminder to Neon Queue: %v", err)
	}
	if !strings.HasPrefix(msgID, "neon-job-") {
		t.Errorf("expected msgID to start with 'neon-job-', got %s", msgID)
	}

	// Wait for background worker to consume job and invoke email
	deadline := time.Now().Add(5 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		emails := mockEmail.GetRecentDevEmails()
		for _, em := range emails {
			if em.To == "test-neon-queue@example.com" {
				found = true
				break
			}
		}
		if found {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}

	if !found {
		t.Fatalf("Neon Queue background worker did not process job within deadline")
	}
}
