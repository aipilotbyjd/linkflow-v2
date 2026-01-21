package aibuilder

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	appbuilder "github.com/linkflow-ai/linkflow/internal/core/application/aibuilder"
)

// ExplainHandler handles workflow explanation requests
type ExplainHandler struct {
	service *appbuilder.Service
}

// NewExplainHandler creates a new explain handler
func NewExplainHandler(service *appbuilder.Service) *ExplainHandler {
	return &ExplainHandler{service: service}
}

// Handle handles the explain workflow request
func (h *ExplainHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ExplainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if req.WorkflowJSON == "" {
		common.BadRequest(w, "workflow_json is required")
		return
	}

	explanation, err := h.service.ExplainWorkflow(ctx, req.WorkflowJSON)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ExplainResponse{Explanation: explanation})
}
