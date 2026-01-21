package middleware

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	billingapp "github.com/linkflow-ai/linkflow/internal/core/application/billing"
)

// UsageLimitMiddleware enforces usage limits before allowing requests
type UsageLimitMiddleware struct {
	usageService *billingapp.UsageService
}

func NewUsageLimitMiddleware(usageService *billingapp.UsageService) *UsageLimitMiddleware {
	return &UsageLimitMiddleware{usageService: usageService}
}

// CheckOperations ensures workspace has available operations
func (m *UsageLimitMiddleware) CheckOperations(operationsNeeded int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			workspaceID := GetWorkspaceID(ctx)

			if err := m.usageService.CheckOperationsAvailable(ctx, workspaceID, operationsNeeded); err != nil {
				if err == billingapp.ErrOperationsExceeded {
					w.Header().Set("X-RateLimit-Exceeded", "operations")
					common.Error(w, http.StatusPaymentRequired, "OPERATIONS_EXCEEDED", 
						"You have exceeded your operations limit. Please upgrade your plan or wait for the next billing cycle.")
					return
				}
				if err == billingapp.ErrNoActiveSubscription {
					common.Error(w, http.StatusPaymentRequired, "NO_SUBSCRIPTION", 
						"No active subscription found. Please subscribe to a plan.")
					return
				}
				common.HandleError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CheckAICredits ensures workspace has available AI credits
func (m *UsageLimitMiddleware) CheckAICredits(creditsNeeded int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			workspaceID := GetWorkspaceID(ctx)

			if err := m.usageService.CheckAICreditsAvailable(ctx, workspaceID, creditsNeeded); err != nil {
				if err == billingapp.ErrAICreditsExceeded {
					w.Header().Set("X-RateLimit-Exceeded", "ai_credits")
					common.Error(w, http.StatusPaymentRequired, "AI_CREDITS_EXCEEDED", 
						"You have exceeded your AI credits limit. Please upgrade your plan or purchase additional credits.")
					return
				}
				common.HandleError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireFeature ensures workspace has access to a specific feature
func (m *UsageLimitMiddleware) RequireFeature(feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			workspaceID := GetWorkspaceID(ctx)

			if err := m.usageService.CheckFeatureAccess(ctx, workspaceID, feature); err != nil {
				if err == billingapp.ErrFeatureNotAvailable {
					common.Error(w, http.StatusForbidden, "FEATURE_NOT_AVAILABLE", 
						"This feature is not available on your current plan. Please upgrade to access it.")
					return
				}
				common.HandleError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CheckScenarioLimit ensures workspace can create/activate more scenarios
func (m *UsageLimitMiddleware) CheckScenarioLimit(currentActive int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			workspaceID := GetWorkspaceID(ctx)

			if err := m.usageService.CheckActiveScenarios(ctx, workspaceID, currentActive); err != nil {
				if err == billingapp.ErrScenariosExceeded {
					common.Error(w, http.StatusForbidden, "SCENARIOS_EXCEEDED", 
						"You have reached your active scenarios limit. Please deactivate a scenario or upgrade your plan.")
					return
				}
				common.HandleError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
