package aibuilder

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	appbuilder "github.com/linkflow-ai/linkflow/internal/core/application/aibuilder"
	"github.com/linkflow-ai/linkflow/internal/core/domain/aibuilder"
)

// GenerateHandler handles workflow generation from natural language
type GenerateHandler struct {
	service *appbuilder.Service
}

// NewGenerateHandler creates a new generate handler
func NewGenerateHandler(service *appbuilder.Service) *GenerateHandler {
	return &GenerateHandler{service: service}
}

// Handle handles the generate workflow request
func (h *GenerateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	workspaceID := middleware.GetWorkspaceID(ctx)

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if req.Prompt == "" || len(req.Prompt) < 10 {
		common.BadRequest(w, "Prompt must be at least 10 characters")
		return
	}

	genCtx := &aibuilder.Context{
		PreferredTrigger:   req.PreferredTrigger,
		AvailableCredTypes: req.AvailableCreds,
		Constraints:        req.Constraints,
	}

	result, err := h.service.GenerateWorkflow(ctx, workspaceID, userID, req.Prompt, genCtx)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	response := GenerateResponse{
		Name:        result.Name,
		Description: result.Description,
		Nodes:       result.Nodes,
		Connections: result.Connections,
		Settings:    result.Settings,
		Explanation: result.Explanation,
		Suggestions: result.Suggestions,
	}

	common.Success(w, response)
}
