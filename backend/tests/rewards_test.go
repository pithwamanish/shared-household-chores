package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/choresync/backend/internal/models"
)

func TestRewardsListAndCreation(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. List rewards for h-roommates
	w := executeRequest(handler, "GET", "/api/v1/households/h-roommates/rewards", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK listing rewards, got %d: %s", w.Code, w.Body.String())
	}

	var rewards []models.RewardItem
	if err := json.Unmarshal(w.Body.Bytes(), &rewards); err != nil {
		t.Fatalf("failed to parse rewards: %v", err)
	}
	if len(rewards) == 0 {
		t.Fatalf("expected at least 1 seed reward in h-roommates, got 0")
	}

	// 2. Create a new reward
	icon := "Coffee"
	newRewardReq := models.RewardItemCreateRequest{
		Title:       "Specialty Pour-Over Coffee on Sunday",
		Description: "Freshly ground Ethiopian single-origin pour-over prepared by roommate",
		PointsCost:  40,
		Icon:        &icon,
	}
	wCreate := executeRequest(handler, "POST", "/api/v1/households/h-roommates/rewards", newRewardReq)
	if wCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for new reward, got %d: %s", wCreate.Code, wCreate.Body.String())
	}

	var created models.RewardItem
	if err := json.Unmarshal(wCreate.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode created reward: %v", err)
	}
	if created.Title != newRewardReq.Title {
		t.Errorf("expected title '%s', got '%s'", newRewardReq.Title, created.Title)
	}
	if created.PointsCost != 40 {
		t.Errorf("expected points cost 40, got %d", created.PointsCost)
	}
}

func TestRewardRedemptionSufficientAndInsufficientBalance(t *testing.T) {
	handler, appStore := setupTestRouter()

	// Leo in h-family has 180 points in seed data
	memberBefore, err := appStore.GetMember("m-leo-kid")
	if err != nil {
		t.Fatalf("failed to get member: %v", err)
	}
	initialBalance := memberBefore.PointsBalance

	// rew-fam-icecream costs 60 points
	reward, err := appStore.GetReward("rew-fam-icecream")
	if err != nil {
		t.Fatalf("failed to get reward: %v", err)
	}

	// 1. Successful redemption with sufficient points
	redeemReq := models.RewardRedeemRequest{
		MemberID: "m-leo-kid",
	}
	w := executeRequest(handler, "POST", "/api/v1/rewards/rew-fam-icecream/redeem", redeemReq)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created redeeming reward, got %d: %s", w.Code, w.Body.String())
	}

	var redemption models.RewardRedemption
	if err := json.Unmarshal(w.Body.Bytes(), &redemption); err != nil {
		t.Fatalf("failed to parse redemption: %v", err)
	}
	if redemption.Status != models.RedemptionStatusRequested {
		t.Errorf("expected status 'requested', got '%s'", redemption.Status)
	}
	if redemption.PointsSpent != reward.PointsCost {
		t.Errorf("expected points spent %d, got %d", reward.PointsCost, redemption.PointsSpent)
	}

	// Verify exact balance deduction
	memberMid, _ := appStore.GetMember("m-leo-kid")
	expectedBalance := initialBalance - reward.PointsCost
	if memberMid.PointsBalance != expectedBalance {
		t.Errorf("expected balance %d after redemption, got %d", expectedBalance, memberMid.PointsBalance)
	}

	// 2. Insufficient points check:
	// Create an expensive reward costing 500 points
	expensiveReq := models.RewardItemCreateRequest{
		Title:       "Trip to Amusement Park",
		Description: "Full day pass to theme park",
		PointsCost:  500,
	}
	wExp := executeRequest(handler, "POST", "/api/v1/households/h-family/rewards", expensiveReq)
	if wExp.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created creating expensive reward, got %d", wExp.Code)
	}
	var expensiveReward models.RewardItem
	_ = json.Unmarshal(wExp.Body.Bytes(), &expensiveReward)

	// Attempt redemption with insufficient points
	wFail := executeRequest(handler, "POST", "/api/v1/rewards/"+expensiveReward.ID+"/redeem", redeemReq)
	if wFail.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for insufficient points, got %d: %s", wFail.Code, wFail.Body.String())
	}

	// Verify balance remains unchanged after failed redemption attempt
	memberAfter, _ := appStore.GetMember("m-leo-kid")
	if memberAfter.PointsBalance != expectedBalance {
		t.Errorf("expected balance to remain %d, got %d", expectedBalance, memberAfter.PointsBalance)
	}
}

func TestRewardRedemptionZeroBalanceAndCrossHouseholdSecurity(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. Create a brand new member with 0 points
	newMemberReq := models.MemberCreateRequest{
		Name: "New Roommate",
		Role: "member",
	}
	wMember := executeRequest(handler, "POST", "/api/v1/households/h-roommates/members", newMemberReq)
	if wMember.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created creating member, got %d", wMember.Code)
	}
	var newMember models.Member
	_ = json.Unmarshal(wMember.Body.Bytes(), &newMember)

	if newMember.PointsBalance != 0 {
		t.Fatalf("expected new member to have 0 points, got %d", newMember.PointsBalance)
	}

	// Attempt redemption with 0 balance -> must fail with 400 Bad Request
	redeemReq := models.RewardRedeemRequest{
		MemberID: newMember.ID,
	}
	wZero := executeRequest(handler, "POST", "/api/v1/rewards/rew-room-coffee/redeem", redeemReq)
	if wZero.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for zero balance redemption, got %d: %s", wZero.Code, wZero.Body.String())
	}

	// 2. Cross-household redemption attempt
	// Member from h-roommates attempts to redeem reward from h-family (rew-fam-icecream)
	wCross := executeRequest(handler, "POST", "/api/v1/rewards/rew-fam-icecream/redeem", models.RewardRedeemRequest{
		MemberID: "m-sarah", // in h-roommates
	})
	if wCross.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for cross-household reward redemption, got %d: %s", wCross.Code, wCross.Body.String())
	}
}

func TestRewardFulfillAndRedemptionsList(t *testing.T) {
	handler, _ := setupTestRouter()

	// 1. Leo redeems reward
	redeemReq := models.RewardRedeemRequest{
		MemberID: "m-leo-kid",
	}
	w := executeRequest(handler, "POST", "/api/v1/rewards/rew-fam-icecream/redeem", redeemReq)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created redeeming reward, got %d", w.Code)
	}
	var redemption models.RewardRedemption
	_ = json.Unmarshal(w.Body.Bytes(), &redemption)

	// 2. Fulfill the redemption
	wFulfill := executeRequest(handler, "POST", "/api/v1/redemptions/"+redemption.ID+"/fulfill", nil)
	if wFulfill.Code != http.StatusOK {
		t.Fatalf("expected 200 OK fulfilling reward, got %d: %s", wFulfill.Code, wFulfill.Body.String())
	}

	// 3. Attempting to fulfill again should fail
	wDoubleFulfill := executeRequest(handler, "POST", "/api/v1/redemptions/"+redemption.ID+"/fulfill", nil)
	if wDoubleFulfill.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on duplicate fulfillment, got %d", wDoubleFulfill.Code)
	}

	// 4. List redemptions for h-family
	wList := executeRequest(handler, "GET", "/api/v1/households/h-family/redemptions", nil)
	if wList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK listing redemptions, got %d", wList.Code)
	}
	var redemptionsList []models.RewardRedemptionItem
	if err := json.Unmarshal(wList.Body.Bytes(), &redemptionsList); err != nil {
		t.Fatalf("failed to decode redemptions list: %v", err)
	}

	found := false
	for _, item := range redemptionsList {
		if item.Redemption.ID == redemption.ID {
			found = true
			if item.Redemption.Status != models.RedemptionStatusFulfilled {
				t.Errorf("expected status 'fulfilled', got '%s'", item.Redemption.Status)
			}
			if item.Redemption.FulfilledAt == nil {
				t.Errorf("expected fulfilled_at timestamp to be set")
			}
			if item.Member.ID != "m-leo-kid" {
				t.Errorf("expected member m-leo-kid, got %s", item.Member.ID)
			}
		}
	}
	if !found {
		t.Errorf("redemption %s not found in household redemptions list", redemption.ID)
	}
}
