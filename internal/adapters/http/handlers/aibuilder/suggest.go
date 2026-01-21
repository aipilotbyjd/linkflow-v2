package aibuilder

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	appbuilder "github.com/linkflow-ai/linkflow/internal/core/application/aibuilder"
)

// SuggestHandler handles workflow improvement suggestions
type SuggestHandler struct {
	service *appbuilder.Service
}

// NewSuggestHandler creates a new suggest handler
func NewSuggestHandler(service *appbuilder.Service) *SuggestHandler {
	return &SuggestHandler{service: service}
}

// Handle handles the suggest improvements request
func (h *SuggestHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req SuggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if req.WorkflowJSON == "" {
		common.BadRequest(w, "workflow_json is required")
		return
	}

	suggestions, err := h.service.SuggestImprovements(ctx, req.WorkflowJSON)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, SuggestResponse{Suggestions: suggestions})
}
