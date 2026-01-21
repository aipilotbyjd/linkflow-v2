package topup

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// UpdateSettingsHandler handles updating auto top-up settings
type UpdateSettingsHandler struct {
	repo AutoTopUpRepository
}

func NewUpdateSettingsHandler(repo AutoTopUpRepository) *UpdateSettingsHandler {
	return &UpdateSettingsHandler{repo: repo}
}

func (h *UpdateSettingsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	var req UpdateAutoTopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	// Get or create settings
	settings, err := h.repo.FindByWorkspaceID(ctx, workspaceID)
	isNew := false
	if err != nil {
		settings = billing.NewAutoTopUp(workspaceID)
		isNew = true
	}

	// Update fields
	if req.Enabled != nil {
		settings.Enabled = *req.Enabled
	}
	if req.TriggerThreshold != nil {
		if *req.TriggerThreshold < 5 || *req.TriggerThreshold > 50 {
			common.BadRequest(w, "Trigger threshold must be between 5 and 50")
			return
		}
		settings.TriggerThreshold = *req.TriggerThreshold
	}
	if req.PurchaseType != nil {
		settings.PurchaseType = billing.TopUpType(*req.PurchaseType)
	}
	if req.CreditAmount != nil {
		settings.CreditAmount = *req.CreditAmount
	}
	if req.AICreditsAmount != nil {
		settings.AICreditsAmount = *req.AICreditsAmount
	}
	if req.MaxPurchasesPerMonth != nil {
		settings.MaxPurchasesPerMonth = *req.MaxPurchasesPerMonth
	}
	if req.MaxSpendPerMonth != nil {
		settings.MaxSpendPerMonth = *req.MaxSpendPerMonth
	}

	settings.UpdatedAt = time.Now()

	// Save
	if isNew {
		if err := h.repo.Create(ctx, settings); err != nil {
			common.HandleError(w, err)
			return
		}
	} else {
		if err := h.repo.Update(ctx, settings); err != nil {
			common.HandleError(w, err)
			return
		}
	}

	common.Success(w, ToAutoTopUpResponse(settings))
}
