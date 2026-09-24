package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/choresync/backend/internal/email"
	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/server"
	"github.com/choresync/backend/internal/store"
)

func newTestRouter(s store.Store) (http.Handler, email.Service) {
	em := email.NewMockService(email.Config{})
	return server.NewRouter(s, em, nil, nil), em
}

func TestAuthDemoLoginAndJWTIssuance(t *testing.T) {
	s := store.NewStore()
	handler, _ := newTestRouter(s)

	// 1. Valid demo login for Sarah (Admin)
	body, _ := json.Marshal(map[string]string{"member_id": "m-sarah"})
	req := httptest.NewRequest("POST", "/api/v1/auth/demo-login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on demo login, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.AuthTokenResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode AuthTokenResponse: %v", err)
	}

	if resp.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
	if resp.TokenType != "Bearer" {
		t.Fatalf("expected token type Bearer, got %s", resp.TokenType)
	}
	if resp.Member == nil || resp.Member.ID != "m-sarah" {
		t.Fatalf("expected member ID m-sarah, got %+v", resp.Member)
	}
	if resp.Household == nil || resp.Household.ID != "h-roommates" {
		t.Fatalf("expected household ID h-roommates, got %+v", resp.Household)
	}

	// 2. Non-existent demo member -> 404
	badBody, _ := json.Marshal(map[string]string{"member_id": "m-non-existent"})
	reqBad := httptest.NewRequest("POST", "/api/v1/auth/demo-login", bytes.NewReader(badBody))
	reqBad.Header.Set("Content-Type", "application/json")
	wBad := httptest.NewRecorder()
	handler.ServeHTTP(wBad, reqBad)

	if wBad.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found for bad demo member, got %d", wBad.Code)
	}
}

func TestMagicLinkFlow(t *testing.T) {
	s := store.NewStore()
	handler, em := newTestRouter(s)

	// 1. Request magic link
	reqBody, _ := json.Marshal(map[string]string{"email": "sarah@example.com"})
	req := httptest.NewRequest("POST", "/api/v1/auth/magic-link", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on magic link request, got %d: %s", w.Code, w.Body.String())
	}

	var mlResp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&mlResp)
	token := mlResp["token"]
	if token == "" {
		t.Fatal("expected non-empty magic link token")
	}

	// Verify email was dispatched
	devEmails := em.GetRecentDevEmails()
	if len(devEmails) == 0 || devEmails[0].Type != "magic_link" {
		t.Fatalf("expected magic_link email to be dispatched, got %+v", devEmails)
	}

	// 2. Verify magic link token
	verifyBody, _ := json.Marshal(map[string]string{"token": token})
	reqVerify := httptest.NewRequest("POST", "/api/v1/auth/verify", bytes.NewReader(verifyBody))
	reqVerify.Header.Set("Content-Type", "application/json")
	wVerify := httptest.NewRecorder()
	handler.ServeHTTP(wVerify, reqVerify)

	if wVerify.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on verify, got %d: %s", wVerify.Code, wVerify.Body.String())
	}

	var authResp models.AuthTokenResponse
	_ = json.NewDecoder(wVerify.Body).Decode(&authResp)
	if authResp.AccessToken == "" {
		t.Fatal("expected non-empty access token after verification")
	}
	if authResp.Member == nil || authResp.Member.ID != "m-sarah" {
		t.Fatalf("expected Sarah Chen as member, got %+v", authResp.Member)
	}

	// 3. Single-Use Invalidation: Reusing the exact same token immediately must fail with 401 Unauthorized
	reqReuse := httptest.NewRequest("POST", "/api/v1/auth/verify", bytes.NewReader(verifyBody))
	reqReuse.Header.Set("Content-Type", "application/json")
	wReuse := httptest.NewRecorder()
	handler.ServeHTTP(wReuse, reqReuse)

	if wReuse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized when reusing single-use magic link token, got %d: %s", wReuse.Code, wReuse.Body.String())
	}

	// 4. Bad token -> 401 Unauthorized
	badVerify, _ := json.Marshal(map[string]string{"token": "INVALID-TOKEN"})
	reqBad := httptest.NewRequest("POST", "/api/v1/auth/verify", bytes.NewReader(badVerify))
	reqBad.Header.Set("Content-Type", "application/json")
	wBad := httptest.NewRecorder()
	handler.ServeHTTP(wBad, reqBad)

	if wBad.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for invalid token, got %d", wBad.Code)
	}
}

func TestMultiTenancyIsolation(t *testing.T) {
	s := store.NewStore()
	handler, _ := newTestRouter(s)

	// Step 1: Mint token for roommate Sarah in h-roommates
	body, _ := json.Marshal(map[string]string{"member_id": "m-sarah"})
	req := httptest.NewRequest("POST", "/api/v1/auth/demo-login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var sarahAuth models.AuthTokenResponse
	_ = json.NewDecoder(w.Body).Decode(&sarahAuth)
	token := sarahAuth.AccessToken

	// Step 2: Sarah accesses her own household (h-roommates) -> 200 OK
	reqOwn := httptest.NewRequest("GET", "/api/v1/households/h-roommates/chores", nil)
	reqOwn.Header.Set("Authorization", "Bearer "+token)
	wOwn := httptest.NewRecorder()
	handler.ServeHTTP(wOwn, reqOwn)

	if wOwn.Code != http.StatusOK {
		t.Fatalf("expected 200 OK when accessing own household, got %d: %s", wOwn.Code, wOwn.Body.String())
	}

	// Step 3: Sarah attempts to access another household (h-family) -> 403 Forbidden!
	reqForeign := httptest.NewRequest("GET", "/api/v1/households/h-family/chores", nil)
	reqForeign.Header.Set("Authorization", "Bearer "+token)
	wForeign := httptest.NewRecorder()
	handler.ServeHTTP(wForeign, reqForeign)

	if wForeign.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for cross-tenant access, got %d: %s", wForeign.Code, wForeign.Body.String())
	}

	var errResp models.ErrorResponse
	_ = json.NewDecoder(wForeign.Body).Decode(&errResp)
	if errResp.Error != "FORBIDDEN" {
		t.Fatalf("expected error code FORBIDDEN, got %s", errResp.Error)
	}
}

func TestAdminPINVerification(t *testing.T) {
	s := store.NewStore()
	handler, _ := newTestRouter(s)

	// 1. Correct PIN (1234) -> 200 OK
	bodyCorrect, _ := json.Marshal(map[string]string{"pin": "1234"})
	req := httptest.NewRequest("POST", "/api/v1/households/h-family/verify-pin", bytes.NewReader(bodyCorrect))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for correct PIN, got %d: %s", w.Code, w.Body.String())
	}

	var pinResp map[string]any
	_ = json.NewDecoder(w.Body).Decode(&pinResp)
	if valid, ok := pinResp["valid"].(bool); !ok || !valid {
		t.Fatalf("expected valid=true, got %+v", pinResp)
	}

	// 2. Incorrect PIN (9999) -> 401 Unauthorized
	bodyWrong, _ := json.Marshal(map[string]string{"pin": "9999"})
	reqWrong := httptest.NewRequest("POST", "/api/v1/households/h-family/verify-pin", bytes.NewReader(bodyWrong))
	reqWrong.Header.Set("Content-Type", "application/json")
	wWrong := httptest.NewRecorder()
	handler.ServeHTTP(wWrong, reqWrong)

	if wWrong.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for wrong PIN, got %d: %s", wWrong.Code, wWrong.Body.String())
	}
}

func TestAdminApprovalWithJWTBearerToken(t *testing.T) {
	s := store.NewStore()
	handler, _ := newTestRouter(s)

	// 1. Mint admin token for Dad (m-david-fam)
	body, _ := json.Marshal(map[string]string{"member_id": "m-david-fam"})
	req := httptest.NewRequest("POST", "/api/v1/auth/demo-login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var davidAuth models.AuthTokenResponse
	_ = json.NewDecoder(w.Body).Decode(&davidAuth)

	// 2. Dad approves chore without specifying admin_member_id in body (derived from Bearer JWT)
	reqApprove := httptest.NewRequest("POST", "/api/v1/households/h-family/completions/comp-seed-leo-1/approve", bytes.NewReader([]byte("{}")))
	reqApprove.Header.Set("Authorization", "Bearer "+davidAuth.AccessToken)
	reqApprove.Header.Set("Content-Type", "application/json")
	wApprove := httptest.NewRecorder()
	handler.ServeHTTP(wApprove, reqApprove)

	if wApprove.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on JWT-derived admin approval, got %d: %s", wApprove.Code, wApprove.Body.String())
	}

	// 3. Child Emma tries to approve -> 403 Forbidden
	bodyChild, _ := json.Marshal(map[string]string{"member_id": "m-emma-kid"})
	reqC := httptest.NewRequest("POST", "/api/v1/auth/demo-login", bytes.NewReader(bodyChild))
	reqC.Header.Set("Content-Type", "application/json")
	wC := httptest.NewRecorder()
	handler.ServeHTTP(wC, reqC)

	var emmaAuth models.AuthTokenResponse
	_ = json.NewDecoder(wC.Body).Decode(&emmaAuth)

	reqApproveChild := httptest.NewRequest("POST", "/api/v1/households/h-family/completions/comp-seed-leo-1/approve", bytes.NewReader([]byte("{}")))
	reqApproveChild.Header.Set("Authorization", "Bearer "+emmaAuth.AccessToken)
	reqApproveChild.Header.Set("Content-Type", "application/json")
	wApproveChild := httptest.NewRecorder()
	handler.ServeHTTP(wApproveChild, reqApproveChild)

	if wApproveChild.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for child JWT approval attempt, got %d: %s", wApproveChild.Code, wApproveChild.Body.String())
	}
}

func TestPasswordLoginAndRegister(t *testing.T) {
	s := store.NewStore()
	handler, _ := newTestRouter(s)

	// 1. Password login for seed admin Sarah
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "sarah@example.com",
		"password": "password123",
	})
	reqL := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	reqL.Header.Set("Content-Type", "application/json")
	wL := httptest.NewRecorder()
	handler.ServeHTTP(wL, reqL)

	if wL.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on password login, got %d: %s", wL.Code, wL.Body.String())
	}

	var authResp models.AuthTokenResponse
	_ = json.NewDecoder(wL.Body).Decode(&authResp)
	if authResp.AccessToken == "" || authResp.Member.Name != "Sarah Chen" {
		t.Fatalf("unexpected auth response: %+v", authResp)
	}

	// 2. Bad password -> 401 Unauthorized
	badPassBody, _ := json.Marshal(map[string]string{
		"email":    "sarah@example.com",
		"password": "wrongpassword",
	})
	reqBad := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(badPassBody))
	reqBad.Header.Set("Content-Type", "application/json")
	wBad := httptest.NewRecorder()
	handler.ServeHTTP(wBad, reqBad)

	if wBad.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for wrong password, got %d", wBad.Code)
	}

	// 3. Register a new user with new household
	regBody, _ := json.Marshal(map[string]string{
		"name":           "Lucas Vance",
		"email":          "lucas@example.com",
		"password":       "securepass456",
		"household_name": "Lucas & Friends",
	})
	reqR := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regBody))
	reqR.Header.Set("Content-Type", "application/json")
	wR := httptest.NewRecorder()
	handler.ServeHTTP(wR, reqR)

	if wR.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on user registration, got %d: %s", wR.Code, wR.Body.String())
	}

	var regAuth models.AuthTokenResponse
	_ = json.NewDecoder(wR.Body).Decode(&regAuth)
	if regAuth.AccessToken == "" || regAuth.Member.Name != "Lucas Vance" {
		t.Fatalf("unexpected registration response: %+v", regAuth)
	}
	if regAuth.Household == nil || regAuth.Household.Name != "Lucas & Friends" {
		t.Fatalf("expected new household 'Lucas & Friends', got %+v", regAuth.Household)
	}

	// 4. Log in as Lucas with new password
	lucasLoginBody, _ := json.Marshal(map[string]string{
		"email":    "lucas@example.com",
		"password": "securepass456",
	})
	reqLL := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(lucasLoginBody))
	reqLL.Header.Set("Content-Type", "application/json")
	wLL := httptest.NewRecorder()
	handler.ServeHTTP(wLL, reqLL)

	if wLL.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on Lucas login, got %d: %s", wLL.Code, wLL.Body.String())
	}

	// 5. Register with existing email -> 409 Conflict
	dupBody, _ := json.Marshal(map[string]string{
		"name":     "Lucas Duplicate",
		"email":    "lucas@example.com",
		"password": "anypassword",
	})
	reqDup := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(dupBody))
	reqDup.Header.Set("Content-Type", "application/json")
	wDup := httptest.NewRecorder()
	handler.ServeHTTP(wDup, reqDup)

	if wDup.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate email, got %d: %s", wDup.Code, wDup.Body.String())
	}
}

func TestUpdateAdminPINAndVerification(t *testing.T) {
	s := store.NewStore()
	handler, _ := newTestRouter(s)

	// 1. Initial PIN (1234) -> 200 OK
	verifyBody1, _ := json.Marshal(map[string]string{"pin": "1234"})
	reqV1 := httptest.NewRequest("POST", "/api/v1/households/h-family/verify-pin", bytes.NewReader(verifyBody1))
	reqV1.Header.Set("Content-Type", "application/json")
	wV1 := httptest.NewRecorder()
	handler.ServeHTTP(wV1, reqV1)
	if wV1.Code != http.StatusOK {
		t.Fatalf("expected initial PIN 1234 to pass, got %d: %s", wV1.Code, wV1.Body.String())
	}

	// 2. Reject non-4-digit PIN updates
	badPinBody, _ := json.Marshal(map[string]string{"admin_pin": "abc"})
	reqBad := httptest.NewRequest("PATCH", "/api/v1/households/h-family/settings", bytes.NewReader(badPinBody))
	reqBad.Header.Set("Content-Type", "application/json")
	wBad := httptest.NewRecorder()
	handler.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for non-4-digit PIN, got %d: %s", wBad.Code, wBad.Body.String())
	}

	// 3. Update Admin PIN to "8844"
	updateBody, _ := json.Marshal(map[string]string{"admin_pin": "8844"})
	reqUp := httptest.NewRequest("PATCH", "/api/v1/households/h-family/settings", bytes.NewReader(updateBody))
	reqUp.Header.Set("Content-Type", "application/json")
	wUp := httptest.NewRecorder()
	handler.ServeHTTP(wUp, reqUp)
	if wUp.Code != http.StatusOK {
		t.Fatalf("expected 200 OK updating PIN to 8844, got %d: %s", wUp.Code, wUp.Body.String())
	}

	// 4. Old PIN (1234) now fails -> 401 Unauthorized
	verifyBodyOld, _ := json.Marshal(map[string]string{"pin": "1234"})
	reqVOld := httptest.NewRequest("POST", "/api/v1/households/h-family/verify-pin", bytes.NewReader(verifyBodyOld))
	reqVOld.Header.Set("Content-Type", "application/json")
	wVOld := httptest.NewRecorder()
	handler.ServeHTTP(wVOld, reqVOld)
	if wVOld.Code != http.StatusUnauthorized {
		t.Fatalf("expected old PIN 1234 to fail with 401 Unauthorized, got %d: %s", wVOld.Code, wVOld.Body.String())
	}

	// 5. New PIN (8844) succeeds -> 200 OK
	verifyBodyNew, _ := json.Marshal(map[string]string{"pin": "8844"})
	reqVNew := httptest.NewRequest("POST", "/api/v1/households/h-family/verify-pin", bytes.NewReader(verifyBodyNew))
	reqVNew.Header.Set("Content-Type", "application/json")
	wVNew := httptest.NewRecorder()
	handler.ServeHTTP(wVNew, reqVNew)
	if wVNew.Code != http.StatusOK {
		t.Fatalf("expected new PIN 8844 to pass with 200 OK, got %d: %s", wVNew.Code, wVNew.Body.String())
	}
}

func TestForgotPasswordAndResetFlow(t *testing.T) {
	s := store.NewStore()
	handler, em := newTestRouter(s)

	// 1. Request password reset for Sarah
	reqBody, _ := json.Marshal(map[string]string{"email": "sarah@example.com"})
	req := httptest.NewRequest("POST", "/api/v1/auth/forgot-password", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for forgot-password, got %d: %s", w.Code, w.Body.String())
	}

	var forgotResp models.ForgotPasswordResponse
	_ = json.NewDecoder(w.Body).Decode(&forgotResp)
	if forgotResp.Token == "" {
		t.Fatal("expected non-empty reset token")
	}

	// 2. Verify email recorded
	emails := em.GetRecentDevEmails()
	if len(emails) == 0 || emails[0].Type != "password_reset" {
		t.Fatalf("expected password_reset email in log, got %+v", emails)
	}

	// 3. Reset password with new value
	resetBody, _ := json.Marshal(map[string]string{
		"token":        forgotResp.Token,
		"new_password": "supersecretnewpass",
	})
	reqReset := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewReader(resetBody))
	reqReset.Header.Set("Content-Type", "application/json")
	wReset := httptest.NewRecorder()
	handler.ServeHTTP(wReset, reqReset)

	if wReset.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on reset-password, got %d: %s", wReset.Code, wReset.Body.String())
	}

	var resetAuth models.AuthTokenResponse
	_ = json.NewDecoder(wReset.Body).Decode(&resetAuth)
	if resetAuth.AccessToken == "" || resetAuth.Member.Name != "Sarah Chen" {
		t.Fatalf("unexpected resetAuth response: %+v", resetAuth)
	}

	// 4. Verify login succeeds with new password
	loginBodyNew, _ := json.Marshal(map[string]string{
		"email":    "sarah@example.com",
		"password": "supersecretnewpass",
	})
	reqLN := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBodyNew))
	reqLN.Header.Set("Content-Type", "application/json")
	wLN := httptest.NewRecorder()
	handler.ServeHTTP(wLN, reqLN)

	if wLN.Code != http.StatusOK {
		t.Fatalf("expected 200 OK logging in with new password, got %d", wLN.Code)
	}

	// 5. Verify old password now fails
	loginBodyOld, _ := json.Marshal(map[string]string{
		"email":    "sarah@example.com",
		"password": "password123",
	})
	reqLO := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBodyOld))
	reqLO.Header.Set("Content-Type", "application/json")
	wLO := httptest.NewRecorder()
	handler.ServeHTTP(wLO, reqLO)

	if wLO.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized logging in with old password, got %d", wLO.Code)
	}
}

func TestChoreReminderNudgeDispatchesEmail(t *testing.T) {
	s := store.NewStore()
	handler, em := newTestRouter(s)

	chores := s.GetChores("h-roommates", "", "", "", "", "")
	var targetChore models.Chore
	for _, c := range chores {
		if c.CurrentAssigneeID != nil {
			targetChore = c
			break
		}
	}
	if targetChore.ID == "" {
		t.Fatal("no assigned chore found in h-roommates")
	}

	nudgeBody, _ := json.Marshal(map[string]string{
		"sender_member_id": "m-sarah",
	})
	req := httptest.NewRequest("POST", "/api/v1/chores/"+targetChore.ID+"/nudge", bytes.NewReader(nudgeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on send nudge, got %d: %s", w.Code, w.Body.String())
	}

	var nudgeResp models.ChoreNudgeResponse
	_ = json.NewDecoder(w.Body).Decode(&nudgeResp)
	if !nudgeResp.Success || !nudgeResp.EmailDispatched {
		t.Fatalf("expected Success=true and EmailDispatched=true, got %+v", nudgeResp)
	}

	emails := em.GetRecentDevEmails()
	if len(emails) == 0 || emails[0].Type != "chore_reminder" {
		t.Fatalf("expected chore_reminder email recorded, got %+v", emails)
	}
}

func TestDevEmailsEndpoint(t *testing.T) {
	s := store.NewStore()
	handler, em := newTestRouter(s)

	_ = em.SendMagicLink("test@example.com", "Test", "http://localhost:3000/?token=123", "123")

	reqGet := httptest.NewRequest("GET", "/api/v1/dev/emails", nil)
	wGet := httptest.NewRecorder()
	handler.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from GET /dev/emails, got %d", wGet.Code)
	}

	var emails []models.DevEmail
	_ = json.NewDecoder(wGet.Body).Decode(&emails)
	if len(emails) != 1 || emails[0].To != "test@example.com" {
		t.Fatalf("unexpected emails array: %+v", emails)
	}

	reqDel := httptest.NewRequest("DELETE", "/api/v1/dev/emails", nil)
	wDel := httptest.NewRecorder()
	handler.ServeHTTP(wDel, reqDel)

	if wDel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from DELETE /dev/emails, got %d", wDel.Code)
	}

	wGet2 := httptest.NewRecorder()
	handler.ServeHTTP(wGet2, reqGet)
	var emails2 []models.DevEmail
	_ = json.NewDecoder(wGet2.Body).Decode(&emails2)
	if len(emails2) != 0 {
		t.Fatalf("expected empty emails array after delete, got %d", len(emails2))
	}
}
