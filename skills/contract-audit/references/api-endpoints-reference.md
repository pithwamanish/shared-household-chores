# ChoreSync OpenAPI 3.1 Contract Reference

This document summarizes the frozen API contract paths (`contracts/openapi.yaml`) for ChoreSync.

---

## 1. Authentication Endpoints (`/api/v1/auth`)

| Path | Method | Description | Request Body / Parameters | Response Status |
|:---|:---|:---|:---|:---|
| `/api/v1/auth/login` | POST | Authenticates member with email and password | `LoginRequest` (email, password) | 200 OK (`AuthResponse`), 401 Unauthorized |
| `/api/v1/auth/register` | POST | Registers new household or joins with invite code | `RegisterRequest` (email, password, name, household_name, invite_code) | 201 Created (`AuthResponse`), 400 Bad Request |
| `/api/v1/auth/magic-link` | POST | Dispatches 15-minute single-use magic login link | `MagicLinkRequest` (email) | 200 OK (`MessageResponse`) |
| `/api/v1/auth/verify-magic-link` | GET | Validates single-use magic link token | Query `token` | 200 OK (`AuthResponse`), 401 Unauthorized |
| `/api/v1/auth/forgot-password` | POST | Triggers password reset email token | `ForgotPasswordRequest` (email) | 200 OK (`MessageResponse`) |
| `/api/v1/auth/reset-password` | POST | Updates password using reset token | `ResetPasswordRequest` (token, new_password) | 200 OK (`MessageResponse`), 400 Bad Request |

---

## 2. Households & Members (`/api/v1/households`, `/api/v1/members`)

| Path | Method | Description | Response Status |
|:---|:---|:---|:---|
| `/api/v1/households` | GET | Returns current household profile & invite code | 200 OK (`Household`) |
| `/api/v1/households/{id}` | GET, PUT | Retrieves or updates household settings & mode | 200 OK (`Household`) |
| `/api/v1/members` | GET, POST | Lists members or creates new household member | 200 OK (`MemberList`), 201 Created |
| `/api/v1/members/{id}` | GET, PUT, DELETE | Member profile, role update, or deactivation | 200 OK (`Member`), 204 No Content |

---

## 3. Chores & Lifecycle (`/api/v1/chores`)

| Path | Method | Description | Response Status |
|:---|:---|:---|:---|
| `/api/v1/chores` | GET, POST | Lists active chores or creates chore instance | 200 OK (`ChoreList`), 201 Created (`Chore`) |
| `/api/v1/chores/{id}` | GET, PUT, DELETE | Chore details, updates, or archiving | 200 OK (`Chore`), 204 No Content |
| `/api/v1/chores/{id}/complete` | POST | Marks chore done; queues for approval if required | 200 OK (`ChoreCompletionResponse`) |
| `/api/v1/chores/{id}/approve` | POST | Admin/Parent signs off; credits gamified points | 200 OK (`ChoreApprovalResponse`) |
| `/api/v1/chores/{id}/nudge` | POST | Queues async reminder message to Floci SQS | 202 Accepted (`NudgeResponse`) |

---

## 4. Swaps, Rewards & Media (`/api/v1/swaps`, `/api/v1/rewards`, `/api/v1/uploads`)

| Path | Method | Description | Response Status |
|:---|:---|:---|:---|
| `/api/v1/swaps` | GET, POST | Lists swap requests or proposes chore swap | 200 OK (`SwapList`), 201 Created (`Swap`) |
| `/api/v1/swaps/{id}/accept` | POST | Peer accepts swap; transfers responsibility | 200 OK (`Swap`) |
| `/api/v1/swaps/{id}/reject` | POST | Peer rejects swap proposal | 200 OK (`Swap`) |
| `/api/v1/rewards` | GET, POST | Catalogs redeemable rewards or creates new reward | 200 OK (`RewardList`), 201 Created |
| `/api/v1/rewards/{id}/redeem` | POST | Deducts member points and logs redemption | 200 OK (`RedemptionResponse`) |
| `/api/v1/uploads/photo` | POST | Uploads completion proof to Floci S3 bucket | 201 Created (`UploadResponse`) |
| `/api/v1/activity` | GET | Household audit log and event feed | 200 OK (`ActivityList`) |
| `/healthz` | GET | Liveness and readiness probe | 200 OK (`HealthResponse`) |

---

## 5. Universal JSON Error Schema

All 4xx and 5xx responses conform to:
```json
{
  "error": "RESOURCE_NOT_FOUND",
  "message": "Chore with ID c-456 was not found in household h-101",
  "code": 404,
  "details": {}
}
```
