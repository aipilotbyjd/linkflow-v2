package topup

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// AutoTopUpRepository interface
type AutoTopUpRepository interface {
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (*billing.AutoTopUp, error)
	Create(ctx context.Context, topup *billing.AutoTopUp) error
	Update(ctx context.Context, topup *billing.AutoTopUp) error
}

// GetSettingsHandler handles getting auto top-up settings
type GetSettingsHandler struct {
	repo AutoTopUpRepository
}

func NewGetSettingsHandler(repo AutoTopUpRepository) *GetSettingsHandler {
	return &GetSettingsHandler{repo: repo}
}

func (h *GetSettingsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	settings, err := h.repo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		// Return default settings if not configured
		settings = billing.NewAutoTopUp(workspaceID)
	}

	common.Success(w, ToAutoTopUpResponse(settings))
}
