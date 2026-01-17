package template

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// UseTemplateRequest represents use template request
type UseTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	FolderID    string `json:"folderId,omitempty"`
}

// UseTemplateResponse represents use template response
type UseTemplateResponse struct {
	WorkflowID string `json:"workflowId"`
	Name       string `json:"name"`
}

// UseHandler handles use template request
type UseHandler struct {
	repo TemplateRepository
}

// NewUseHandler creates a new handler
func NewUseHandler(repo TemplateRepository) *UseHandler {
	return &UseHandler{repo: repo}
}

// Handle handles the use template request
func (h *UseHandler) Handle(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req UseTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		req.Name = "Workflow from template"
	}

	workflowID := uuid.New().String()
	_ = templateID
	_ = workspaceID

	common.Success(w, UseTemplateResponse{
		WorkflowID: workflowID,
		Name:       req.Name,
	})
}
