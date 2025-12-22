package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type WorkflowHandler struct {
	workflowSvc *services.WorkflowService
	billingSvc  *services.BillingService
	queueClient *queue.Client
}

func NewWorkflowHandler(
	workflowSvc *services.WorkflowService,
	billingSvc *services.BillingService,
	queueClient *queue.Client,
) *WorkflowHandler {
	return &WorkflowHandler{
		workflowSvc: workflowSvc,
		billingSvc:  billingSvc,
		queueClient: queueClient,
	}
}

func (h *WorkflowHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	opts := repositories.NewListOptions(page, perPage)

	workflows, total, err := h.workflowSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID, opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list workflows")
		return
	}

	response := []dto.WorkflowResponse{}
	for _, wf := range workflows {
		var lastExecutedAt *int64
		if wf.LastExecutedAt != nil {
			ts := wf.LastExecutedAt.Unix()
			lastExecutedAt = &ts
		}

		response = append(response, dto.WorkflowResponse{
			ID:             wf.ID.String(),
			Name:           wf.Name,
			Description:    wf.Description,
			Status:         wf.Status,
			Version:        wf.Version,
			Tags:           wf.Tags,
			ExecutionCount: wf.ExecutionCount,
			LastExecutedAt: lastExecutedAt,
			CreatedAt:      wf.CreatedAt.Unix(),
			UpdatedAt:      wf.UpdatedAt.Unix(),
		})
	}

	totalPages := int(total) / opts.Limit
	if int(total)%opts.Limit > 0 {
		totalPages++
	}

	dto.JSONWithMeta(w, http.StatusOK, response, &dto.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *WorkflowHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	var req dto.CreateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	workflow, err := h.workflowSvc.Create(r.Context(), services.CreateWorkflowInput{
		WorkspaceID: wsCtx.WorkspaceID,
		CreatedBy:   claims.UserID,
		Name:        req.Name,
		Description: req.Description,
		Nodes:       req.Nodes,
		Connections: req.Connections,
		Settings:    req.Settings,
		Tags:        req.Tags,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create workflow")
		return
	}

	dto.Created(w, dto.WorkflowResponse{
		ID:          workflow.ID.String(),
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      workflow.Status,
		Version:     workflow.Version,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Settings:    workflow.Settings,
		Tags:        workflow.Tags,
		CreatedAt:   workflow.CreatedAt.Unix(),
		UpdatedAt:   workflow.UpdatedAt.Unix(),
	})
}

func (h *WorkflowHandler) Get(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	workflow, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
		return
	}

	var lastExecutedAt *int64
	if workflow.LastExecutedAt != nil {
		ts := workflow.LastExecutedAt.Unix()
		lastExecutedAt = &ts
	}

	dto.JSON(w, http.StatusOK, dto.WorkflowResponse{
		ID:             workflow.ID.String(),
		Name:           workflow.Name,
		Description:    workflow.Description,
		Status:         workflow.Status,
		Version:        workflow.Version,
		Nodes:          workflow.Nodes,
		Connections:    workflow.Connections,
		Settings:       workflow.Settings,
		Tags:           workflow.Tags,
		ExecutionCount: workflow.ExecutionCount,
		LastExecutedAt: lastExecutedAt,
		CreatedAt:      workflow.CreatedAt.Unix(),
		UpdatedAt:      workflow.UpdatedAt.Unix(),
	})
}

func (h *WorkflowHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	var req dto.UpdateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	workflow, err := h.workflowSvc.Update(r.Context(), workflowID, services.UpdateWorkflowInput{
		Name:        req.Name,
		Description: req.Description,
		Nodes:       req.Nodes,
		Connections: req.Connections,
		Settings:    req.Settings,
		Tags:        req.Tags,
	}, claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update workflow")
		return
	}

	dto.JSON(w, http.StatusOK, dto.WorkflowResponse{
		ID:          workflow.ID.String(),
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      workflow.Status,
		Version:     workflow.Version,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Settings:    workflow.Settings,
		Tags:        workflow.Tags,
		CreatedAt:   workflow.CreatedAt.Unix(),
		UpdatedAt:   workflow.UpdatedAt.Unix(),
	})
}

func (h *WorkflowHandler) Delete(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	if err := h.workflowSvc.Delete(r.Context(), workflowID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete workflow")
		return
	}

	dto.NoContent(w)
}

func (h *WorkflowHandler) Execute(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	// Check execution limit
	allowed, err := h.billingSvc.CheckExecutionLimit(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to check limits")
		return
	}
	if !allowed {
		dto.ErrorResponse(w, http.StatusForbidden, "execution limit reached")
		return
	}

	var req dto.ExecuteWorkflowRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Queue execution
	task, err := h.queueClient.EnqueueWorkflowExecution(r.Context(), queue.WorkflowExecutionPayload{
		WorkflowID:  workflowID,
		WorkspaceID: wsCtx.WorkspaceID,
		TriggeredBy: &claims.UserID,
		TriggerType: models.TriggerManual,
		InputData:   req.InputData,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to queue execution")
		return
	}

	dto.Accepted(w, map[string]string{
		"task_id": task.ID,
		"status":  "queued",
	})
}

func (h *WorkflowHandler) Clone(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	var req dto.CloneWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workflow, err := h.workflowSvc.Clone(r.Context(), workflowID, claims.UserID, req.Name)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to clone workflow")
		return
	}

	dto.Created(w, dto.WorkflowResponse{
		ID:          workflow.ID.String(),
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      workflow.Status,
		Version:     workflow.Version,
		CreatedAt:   workflow.CreatedAt.Unix(),
		UpdatedAt:   workflow.UpdatedAt.Unix(),
	})
}

func (h *WorkflowHandler) Activate(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	if err := h.workflowSvc.Activate(r.Context(), workflowID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to activate workflow")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]string{"status": "active"})
}

func (h *WorkflowHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	if err := h.workflowSvc.Deactivate(r.Context(), workflowID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to deactivate workflow")
		return
	}

	dto.JSON(w, http.StatusOK, map[string]string{"status": "inactive"})
}

func (h *WorkflowHandler) GetVersions(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	versions, err := h.workflowSvc.GetVersions(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get versions")
		return
	}

	response := []dto.WorkflowVersionResponse{}
	for _, v := range versions {
		response = append(response, dto.WorkflowVersionResponse{
			ID:            v.ID.String(),
			Version:       v.Version,
			Nodes:         v.Nodes,
			Connections:   v.Connections,
			Settings:      v.Settings,
			ChangeMessage: v.ChangeMessage,
			CreatedAt:     v.CreatedAt.Unix(),
		})
	}

	dto.JSON(w, http.StatusOK, response)
}

func (h *WorkflowHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	versionStr := chi.URLParam(r, "version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid version")
		return
	}

	v, err := h.workflowSvc.GetVersion(r.Context(), workflowID, version)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "version not found")
		return
	}

	dto.JSON(w, http.StatusOK, dto.WorkflowVersionResponse{
		ID:            v.ID.String(),
		Version:       v.Version,
		Nodes:         v.Nodes,
		Connections:   v.Connections,
		Settings:      v.Settings,
		ChangeMessage: v.ChangeMessage,
		CreatedAt:     v.CreatedAt.Unix(),
	})
}

// Export exports a workflow as JSON
func (h *WorkflowHandler) Export(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	workflow, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
		return
	}

	exportData := map[string]interface{}{
		"version":     "1.0",
		"exportedAt":  time.Now().Unix(),
		"workflow": map[string]interface{}{
			"name":        workflow.Name,
			"description": workflow.Description,
			"nodes":       workflow.Nodes,
			"connections": workflow.Connections,
			"settings":    workflow.Settings,
			"tags":        workflow.Tags,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.json\"", workflow.Name))
	_ = json.NewEncoder(w).Encode(exportData)
}

// Import imports a workflow from JSON
func (h *WorkflowHandler) Import(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	var importData struct {
		Version  string `json:"version"`
		Workflow struct {
			Name        string              `json:"name"`
			Description *string             `json:"description"`
			Nodes       models.JSONArray    `json:"nodes"`
			Connections models.JSONArray    `json:"connections"`
			Settings    models.JSON         `json:"settings"`
			Tags        []string            `json:"tags"`
		} `json:"workflow"`
	}

	if err := json.NewDecoder(r.Body).Decode(&importData); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid import data")
		return
	}

	if importData.Workflow.Name == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "workflow name is required")
		return
	}

	workflow, err := h.workflowSvc.Create(r.Context(), services.CreateWorkflowInput{
		WorkspaceID: wsCtx.WorkspaceID,
		CreatedBy:   claims.UserID,
		Name:        importData.Workflow.Name + " (Imported)",
		Description: importData.Workflow.Description,
		Nodes:       importData.Workflow.Nodes,
		Connections: importData.Workflow.Connections,
		Settings:    importData.Workflow.Settings,
		Tags:        importData.Workflow.Tags,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to import workflow")
		return
	}

	dto.Created(w, dto.WorkflowResponse{
		ID:          workflow.ID.String(),
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      workflow.Status,
		Version:     workflow.Version,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Settings:    workflow.Settings,
		Tags:        workflow.Tags,
		CreatedAt:   workflow.CreatedAt.Unix(),
		UpdatedAt:   workflow.UpdatedAt.Unix(),
	})
}

// RollbackVersion restores workflow to a previous version
func (h *WorkflowHandler) RollbackVersion(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	versionStr := chi.URLParam(r, "version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid version")
		return
	}

	workflow, err := h.workflowSvc.RestoreVersion(r.Context(), workflowID, version, claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to rollback version")
		return
	}

	dto.JSON(w, http.StatusOK, dto.WorkflowResponse{
		ID:          workflow.ID.String(),
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      workflow.Status,
		Version:     workflow.Version,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Settings:    workflow.Settings,
		Tags:        workflow.Tags,
		CreatedAt:   workflow.CreatedAt.Unix(),
		UpdatedAt:   workflow.UpdatedAt.Unix(),
	})
}

// Duplicate creates a copy of a workflow with optional variable substitution
func (h *WorkflowHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	workflowIDStr := chi.URLParam(r, "workflowID")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	var req struct {
		Name      string            `json:"name"`
		Variables map[string]string `json:"variables"`
	}
	// Body is optional for duplicate - silently ignore decode errors
	_ = json.NewDecoder(r.Body).Decode(&req)

	original, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
		return
	}

	name := req.Name
	if name == "" {
		name = original.Name + " (Copy)"
	}

	// Apply variable substitution to nodes if variables provided
	nodes := original.Nodes
	connections := original.Connections
	if len(req.Variables) > 0 {
		nodesJSON, _ := json.Marshal(nodes)
		nodesStr := string(nodesJSON)
		for key, value := range req.Variables {
			nodesStr = strings.ReplaceAll(nodesStr, "{{"+key+"}}", value)
		}
		_ = json.Unmarshal([]byte(nodesStr), &nodes)
	}

	workflow, err := h.workflowSvc.Create(r.Context(), services.CreateWorkflowInput{
		WorkspaceID: wsCtx.WorkspaceID,
		CreatedBy:   claims.UserID,
		Name:        name,
		Description: original.Description,
		Nodes:       nodes,
		Connections: connections,
		Settings:    original.Settings,
		Tags:        original.Tags,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to duplicate workflow")
		return
	}

	dto.Created(w, dto.WorkflowResponse{
		ID:          workflow.ID.String(),
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      workflow.Status,
		Version:     workflow.Version,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Settings:    workflow.Settings,
		Tags:        workflow.Tags,
		CreatedAt:   workflow.CreatedAt.Unix(),
		UpdatedAt:   workflow.UpdatedAt.Unix(),
	})
}
