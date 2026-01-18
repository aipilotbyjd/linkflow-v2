package template

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	templateDomain "github.com/linkflow-ai/linkflow/internal/core/domain/template"
	workflowDomain "github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type UseHandler struct {
	templateRepo templateDomain.Repository
	workflowRepo workflowDomain.Repository
}

func NewUseHandler(templateRepo templateDomain.Repository, workflowRepo workflowDomain.Repository) *UseHandler {
	return &UseHandler{
		templateRepo: templateRepo,
		workflowRepo: workflowRepo,
	}
}

func (h *UseHandler) Handle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "templateId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.BadRequest(w, "invalid template ID")
		return
	}

	workspaceID := middleware.GetWorkspaceID(r.Context())
	claims := middleware.GetUserFromContext(r.Context())

	// Get template
	tmpl, err := h.templateRepo.FindByID(r.Context(), id)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if tmpl == nil {
		common.NotFound(w, "template")
		return
	}

	// Create workflow from template
	name := tmpl.Name + " (from template)"
	workflow := workflowDomain.NewWorkflow(workspaceID, claims.UserID, name)
	if tmpl.Description != nil {
		workflow.Description = tmpl.Description
	}
	workflow.Nodes = tmpl.Nodes
	workflow.Connections = tmpl.Connections

	if err := h.workflowRepo.Create(r.Context(), workflow); err != nil {
		common.HandleError(w, err)
		return
	}

	// Increment template usage
	_ = h.templateRepo.IncrementUsage(r.Context(), id)

	common.Created(w, map[string]interface{}{
		"workflow_id": workflow.ID.String(),
		"message":     "Workflow created from template",
	})
}
