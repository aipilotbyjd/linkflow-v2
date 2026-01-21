package byok

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// BYOKRepository interface for BYOK data access
type BYOKRepository interface {
	Create(ctx interface{}, config *billing.BYOKConfig) error
	Update(ctx interface{}, config *billing.BYOKConfig) error
	Delete(ctx interface{}, id interface{}) error
	FindByID(ctx interface{}, id interface{}) (*billing.BYOKConfig, error)
	FindByWorkspaceID(ctx interface{}, workspaceID interface{}) ([]*billing.BYOKConfig, error)
	FindByWorkspaceAndProvider(ctx interface{}, workspaceID interface{}, provider billing.AIProvider) (*billing.BYOKConfig, error)
}

// ListHandler handles listing BYOK configurations
type ListHandler struct {
	repo BYOKRepository
}

func NewListHandler(repo BYOKRepository) *ListHandler {
	return &ListHandler{repo: repo}
}

func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	configs, err := h.repo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	responses := make([]BYOKResponse, len(configs))
	for i, config := range configs {
		responses[i] = ToBYOKResponse(config)
	}

	common.Success(w, map[string]interface{}{
		"configurations": responses,
		"total":          len(responses),
	})
}
