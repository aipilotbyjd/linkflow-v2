package marketplace

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

// UseTemplateRequest represents use marketplace template request
type UseTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	FolderID    string `json:"folderId,omitempty"`
}

// UseHandler handles use marketplace template request
type UseHandler struct{}

// NewUseHandler creates a new handler
func NewUseHandler() *UseHandler {
	return &UseHandler{}
}

// Handle handles the use template request
func (h *UseHandler) Handle(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req UseTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	if req.Name == "" {
		req.Name = "Workflow from marketplace"
	}

	_ = templateID
	_ = workspaceID

	common.Success(w, map[string]interface{}{
		"workflowId": uuid.New().String(),
		"name":       req.Name,
	})
}
