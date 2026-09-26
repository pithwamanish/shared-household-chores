package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/choresync/backend/internal/auth"
	"github.com/choresync/backend/internal/cloud"
	"github.com/choresync/backend/internal/email"
	"github.com/choresync/backend/internal/handlers"
	"github.com/choresync/backend/internal/store"
	"github.com/choresync/backend/internal/telemetry"
)

// tenantGuardMiddleware verifies that if caller is authenticated with a JWT,
// they cannot access another household's data.
func tenantGuardMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		caller, hasCaller := auth.GetCaller(r.Context())
		if hasCaller {
			householdID := chi.URLParam(r, "household_id")
			if householdID != "" && householdID != caller.HouseholdID {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":   "FORBIDDEN",
					"message": fmt.Sprintf("Forbidden: cross-tenant access denied. Token household '%s' cannot access '%s'", caller.HouseholdID, householdID),
					"code":    http.StatusForbidden,
					"details": map[string]any{},
				})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// NewRouter configures and returns the Chi HTTP router with all ChoreSync routes.
func NewRouter(s store.Store, em email.Service, storage cloud.StorageService, queue cloud.QueueService) http.Handler {
	if storage == nil {
		storage = cloud.NewMockStorageService()
	}

	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	// OpenTelemetry distributed tracing middleware (extracts W3C traceparent, excludes /healthz)
	r.Use(telemetry.Middleware())
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Cross-Origin Resource Sharing (CORS) configuration for Vite/frontend dev
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "traceparent", "tracestate", "baggage"},
		ExposedHeaders:   []string{"Link", "traceparent", "tracestate"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// JWT Authentication Middleware (attaches Claims to context if valid Bearer token provided)
	r.Use(auth.Middleware(auth.GetDefaultSecret()))

	// Handlers
	householdH := handlers.NewHouseholdHandler(s)
	memberH := handlers.NewMemberHandler(s)
	choreH := handlers.NewChoreHandler(s)
	approvalH := handlers.NewApprovalHandler(s)
	swapH := handlers.NewSwapHandler(s)
	rewardH := handlers.NewRewardHandler(s)
	activityH := handlers.NewActivityHandler(s, em, queue)
	uploadH := handlers.NewUploadHandler(storage)
	systemH := handlers.NewSystemHandler(s)
	authH := handlers.NewAuthHandler(s, em)

	// Health Check
	r.Get("/healthz", systemH.HealthCheck)

	// API Routes (Contract-First OpenAPI 3.1 supporting both /api and /api/v1)
	mountAPIRoutes := func(api chi.Router) {
		// System
		api.Post("/reset", systemH.ResetDemoData)

		// Auth, Magic Links, Passwords & Registration
		api.Post("/auth/login", authH.Login)
		api.Post("/auth/register", authH.Register)
		api.Post("/auth/magic-link", authH.RequestMagicLink)
		api.Post("/auth/verify", authH.VerifyMagicLink)
		api.Post("/auth/demo-login", authH.DemoLogin)
		api.Post("/auth/forgot-password", authH.RequestPasswordReset)
		api.Post("/auth/reset-password", authH.ResetPassword)
		api.Post("/auth/supabase-login", authH.SupabaseLogin)

		// Dev email inspection (for testing and local verification)
		api.Get("/dev/emails", authH.GetDevEmails)
		api.Delete("/dev/emails", authH.ClearDevEmails)

		// Cloud Object Storage (S3 / Floci) - Chore photo proof uploads
		api.Post("/uploads/photo", uploadH.UploadPhoto)
		api.Get("/uploads/photo/*", uploadH.GetPhoto)
		api.Head("/uploads/photo/*", uploadH.GetPhoto)

		// Dev failure and latency simulation endpoints for alert & incident verification
		api.Get("/dev/simulate-error", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":   "INTERNAL_SERVER_ERROR",
				"message": "Simulated synthetic 500 error for alert testing",
				"code":    500,
				"details": map[string]any{},
			})
		})
		api.Post("/dev/simulate-error", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":   "INTERNAL_SERVER_ERROR",
				"message": "Simulated synthetic 500 error for alert testing",
				"code":    500,
				"details": map[string]any{},
			})
		})
		api.Get("/dev/simulate-latency", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1200 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":  "delayed",
				"latency": "1.2s",
			})
		})

		// Household Discovery & Join
		api.Get("/households", householdH.GetHouseholds)
		api.Post("/households", householdH.CreateHousehold)
		api.Post("/households/join", householdH.JoinHousehold)

		// Tenant-guarded Household routes
		api.With(tenantGuardMiddleware).Group(func(tg chi.Router) {
			tg.Get("/households/{household_id}", householdH.GetHousehold)
			tg.Patch("/households/{household_id}/settings", householdH.UpdateHouseholdSettings)
			tg.Post("/households/{household_id}/verify-pin", authH.VerifyPIN)

			// Household Members
			tg.Get("/households/{household_id}/members", memberH.GetMembers)
			tg.Post("/households/{household_id}/members", memberH.AddMember)

			// Household Chores
			tg.Get("/households/{household_id}/chores", choreH.GetChores)
			tg.Post("/households/{household_id}/chores", choreH.CreateChore)
			tg.Post("/households/{household_id}/chores/{chore_id}/rotate", choreH.RotateChore)

			// Approvals & Completions
			tg.Get("/households/{household_id}/approvals", approvalH.GetPendingApprovals)
			tg.Post("/households/{household_id}/completions/{completion_id}/approve", approvalH.ApproveChore)
			tg.Post("/households/{household_id}/completions/{completion_id}/reject", approvalH.RejectChore)

			// Swaps
			tg.Get("/households/{household_id}/swaps", swapH.GetSwaps)

			// Rewards & Redemptions
			tg.Get("/households/{household_id}/rewards", rewardH.GetRewards)
			tg.Post("/households/{household_id}/rewards", rewardH.CreateReward)
			tg.Get("/households/{household_id}/redemptions", rewardH.GetRedemptions)

			// Activities
			tg.Get("/households/{household_id}/activity", activityH.GetActivities)
		})

		// Direct Chore Operations
		api.Get("/chores/{chore_id}", choreH.GetChore)
		api.Patch("/chores/{chore_id}", choreH.UpdateChore)
		api.Delete("/chores/{chore_id}", choreH.DeleteChore)
		api.Post("/chores/{chore_id}/claim", choreH.ClaimChore)
		api.Post("/chores/{chore_id}/complete", choreH.CompleteChore)
		api.Post("/chores/{chore_id}/rotate", choreH.RotateChore)
		api.Post("/chores/{chore_id}/swap", swapH.RequestChoreSwap)
		api.Post("/chores/{chore_id}/nudge", activityH.SendNudge)
		api.Get("/chores/{chore_id}/comments", activityH.GetComments)
		api.Post("/chores/{chore_id}/comments", activityH.AddComment)

		// Direct Completions & Approvals
		api.Post("/completions/{completion_id}/approve", approvalH.ApproveChore)
		api.Post("/completions/{completion_id}/reject", approvalH.RejectChore)

		// Direct Swaps & Rewards
		api.Post("/swaps", swapH.CreateSwap)
		api.Post("/swaps/{swap_id}/accept", swapH.AcceptSwap)
		api.Post("/swaps/{swap_id}/reject", swapH.RejectSwap)
		api.Delete("/swaps/{swap_id}", swapH.CancelSwap)
		api.Post("/rewards/{reward_id}/redeem", rewardH.RedeemReward)
		api.Post("/redemptions/{redemption_id}/fulfill", rewardH.FulfillReward)
	}

	r.Route("/api", mountAPIRoutes)
	r.Route("/api/v1", mountAPIRoutes)

	return r
}
