package tests

import (
	"os"
	"testing"

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
