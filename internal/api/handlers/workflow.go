package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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
	scheduleSvc *services.ScheduleService
	queueClient *queue.Client
}

// NewWorkflowHandler creates a new WorkflowHandler for workflow CRUD operations.
func NewWorkflowHandler(
	workflowSvc *services.WorkflowService,
	billingSvc *services.BillingService,
	scheduleSvc *services.ScheduleService,
	queueClient *queue.Client,
) *WorkflowHandler {
	return &WorkflowHandler{
		workflowSvc: workflowSvc,
		billingSvc:  billingSvc,
		scheduleSvc: scheduleSvc,
		queueClient: queueClient,
	}
}

// List returns paginated workflows for a workspace.
func (h *WorkflowHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	pg := dto.ParsePagination(r)
	filters := dto.ParseWorkflowFilters(r)

	// Convert DTO filters to repository filters
	repoFilter := filtersToWorkflowRepoFilter(filters)

	workflows, total, err := h.workflowSvc.GetByWorkspaceWithFilters(r.Context(), wsCtx.WorkspaceID, repoFilter, pg.Opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list workflows")
		return
	}

	// Parse includes (e.g., ?include=schedules)
	includes := dto.ParseIncludes(r)
	includeBuilder := dto.NewIncludeBuilder()

	// Build workflow responses with actions
	type WorkflowWithActions struct {
		dto.WorkflowResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}

	response := []WorkflowWithActions{}
	for _, wf := range workflows {
		var lastExecutedAt *int64
		if wf.LastExecutedAt != nil {
			ts := wf.LastExecutedAt.Unix()
			lastExecutedAt = &ts
		}

		wsID := wsCtx.WorkspaceID.String()
		wfID := wf.ID.String()

		// Build actions based on workflow status
		var actions []dto.Action
		actions = append(actions, dto.ExecuteAction(wsID, wfID))
		if wf.Status == "active" {
			actions = append(actions, dto.DeactivateAction(wsID, wfID))
		} else {
			actions = append(actions, dto.ActivateAction(wsID, wfID))
		}
		actions = append(actions, dto.DeleteAction("/api/v1/workspaces/"+wsID+"/workflows/"+wfID))

		response = append(response, WorkflowWithActions{
			WorkflowResponse: dto.WorkflowResponse{
				ID:             wfID,
				Name:           wf.Name,
				Description:    wf.Description,
				Status:         wf.Status,
				Version:        wf.Version,
				Tags:           wf.Tags,
				ExecutionCount: wf.ExecutionCount,
				LastExecutedAt: lastExecutedAt,
				CreatedAt:      wf.CreatedAt.Unix(),
				UpdatedAt:      wf.UpdatedAt.Unix(),
			},
			Actions: actions,
		})
	}

	// Include schedules if requested
	if dto.HasInclude(includes, "schedules") {
		schedules, _, _ := h.scheduleSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID, nil)
		if len(schedules) > 0 {
			scheduleMap := make(map[string]interface{})
			for _, s := range schedules {
				scheduleMap[s.WorkflowID.String()] = map[string]interface{}{
					"id":              s.ID.String(),
					"cron_expression": s.CronExpression,
					"is_active":       s.IsActive,
					"next_run_at":     s.NextRunAt,
				}
			}
			includeBuilder.Add("schedules", scheduleMap)
		}
	}

	// Build links with filter query string preservation
	basePath := "/api/v1/workspaces/" + wsCtx.WorkspaceID.String() + "/workflows"
	filterQS := filters.ToQueryString()
	links := &dto.Links{
		Self: fmt.Sprintf("%s?page=%d&per_page=%d%s", basePath, pg.Page, pg.PerPage, filterQS),
	}
	meta := pg.NewMeta(total)
	if pg.Page < meta.TotalPages {
		links.Next = fmt.Sprintf("%s?page=%d&per_page=%d%s", basePath, pg.Page+1, pg.PerPage, filterQS)
	}
	if pg.Page > 1 {
		links.Prev = fmt.Sprintf("%s?page=%d&per_page=%d%s", basePath, pg.Page-1, pg.PerPage, filterQS)
	}
	links.First = fmt.Sprintf("%s?page=1&per_page=%d%s", basePath, pg.PerPage, filterQS)
	if meta.TotalPages > 0 {
		links.Last = fmt.Sprintf("%s?page=%d&per_page=%d%s", basePath, meta.TotalPages, pg.PerPage, filterQS)
	}

	// Apply field selection if requested (e.g., ?fields=id,name,status)
	var data interface{} = response
	data = dto.SelectFields(r, data)

	// Send enhanced response
	dto.NewResponse(data).
		WithIncluded(includeBuilder.Build()).
		WithLinks(links).
		WithMeta(meta).
		Send(w)
}

func (h *WorkflowHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
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

	// Validate workflow structure (nodes, connections, graph)
	if validationResult, err := validator.ParseAndValidateWorkflow(req.Nodes, req.Connections); err != nil {
		dto.BadRequest(w, "failed to parse workflow structure: "+err.Error())
		return
	} else if !validationResult.Valid {
		// Convert to DTO format
		errors := make([]dto.WorkflowValidationError, len(validationResult.Errors))
		for i, e := range validationResult.Errors {
			errors[i] = dto.WorkflowValidationError{
				Field:   e.Field,
				NodeID:  e.NodeID,
				Code:    e.Code,
				Message: e.Message,
			}
		}
		dto.WorkflowValidationErrorResponse(w, errors)
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
		Color:       req.Color,
		Icon:        req.Icon,
		Category:    req.Category,
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
	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	workflow, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
		return
	}

	// SECURITY: Validate workspace ownership to prevent cross-tenant access
	if !ValidateWorkspaceOwnership(w, r, workflow) {
		return
	}

	var lastExecutedAt *int64
	if workflow.LastExecutedAt != nil {
		ts := workflow.LastExecutedAt.Unix()
		lastExecutedAt = &ts
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflow.ID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID

	actions := []dto.Action{
		{Name: "edit", Method: "PUT", Href: basePath, Label: "Edit Workflow"},
		{Name: "versions", Method: "GET", Href: basePath + "/versions", Label: "View Versions"},
	}
	if workflow.Status == "active" {
		actions = append(actions, dto.Action{Name: "execute", Method: "POST", Href: basePath + "/execute", Label: "Execute"})
		actions = append(actions, dto.Action{Name: "deactivate", Method: "POST", Href: basePath + "/deactivate", Label: "Deactivate"})
	} else {
		actions = append(actions, dto.Action{Name: "activate", Method: "POST", Href: basePath + "/activate", Label: "Activate"})
	}
	actions = append(actions, dto.Action{Name: "delete", Method: "DELETE", Href: basePath, Label: "Delete"})

	response := struct {
		dto.WorkflowResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}{
		WorkflowResponse: dto.WorkflowResponse{
			ID:             wfID,
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
		},
		Actions: actions,
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

func (h *WorkflowHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before modification
	existing, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
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

	// Validate workflow structure if nodes or connections are being updated
	nodesToValidate := req.Nodes
	connectionsToValidate := req.Connections
	if nodesToValidate == nil {
		nodesToValidate = existing.Nodes
	}
	if connectionsToValidate == nil {
		connectionsToValidate = existing.Connections
	}
	if req.Nodes != nil || req.Connections != nil {
		if validationResult, err := validator.ParseAndValidateWorkflow(nodesToValidate, connectionsToValidate); err != nil {
			dto.BadRequest(w, "failed to parse workflow structure: "+err.Error())
			return
		} else if !validationResult.Valid {
			errors := make([]dto.WorkflowValidationError, len(validationResult.Errors))
			for i, e := range validationResult.Errors {
				errors[i] = dto.WorkflowValidationError{
					Field:   e.Field,
					NodeID:  e.NodeID,
					Code:    e.Code,
					Message: e.Message,
				}
			}
			dto.WorkflowValidationErrorResponse(w, errors)
			return
		}
	}

	workflow, err := h.workflowSvc.Update(r.Context(), workflowID, services.UpdateWorkflowInput{
		Name:        req.Name,
		Description: req.Description,
		Nodes:       req.Nodes,
		Connections: req.Connections,
		Settings:    req.Settings,
		Tags:        req.Tags,
		Color:       req.Color,
		Icon:        req.Icon,
		Category:    req.Category,
		IsFavorite:  req.IsFavorite,
	}, claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update workflow")
		return
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflow.ID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID

	response := dto.WorkflowResponse{
		ID:          wfID,
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
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

func (h *WorkflowHandler) Delete(w http.ResponseWriter, r *http.Request) {
	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before deletion
	existing, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	if err := h.workflowSvc.Delete(r.Context(), workflowID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete workflow")
		return
	}

	dto.NoContent(w)
}

func (h *WorkflowHandler) Execute(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before execution
	existing, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.NotFound(w, "Workflow")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	// Business rule: Check if workflow can be executed
	hasNodes := len(existing.Nodes) > 0
	if err := validator.CanExecuteWorkflow(existing.Status, hasNodes); err != nil {
		dto.BadRequest(w, err.Error())
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
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before cloning
	existing, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
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
	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before activation
	existing, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.NotFound(w, "Workflow")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	// Business rule: Check if workflow can be activated
	hasTrigger := hasTriggerNode(existing.Nodes)
	if err := validator.CanActivateWorkflow(existing.Status, hasTrigger); err != nil {
		dto.BadRequest(w, err.Error())
		return
	}

	if err := h.workflowSvc.Activate(r.Context(), workflowID); err != nil {
		dto.InternalServerError(w, "failed to activate workflow")
		return
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID

	dto.NewResponse(map[string]string{"status": "active"}).
		WithLinks(&dto.Links{Self: basePath}).
		WithActions(
			dto.Action{Name: "execute", Method: "POST", Href: basePath + "/execute", Label: "Execute"},
			dto.Action{Name: "deactivate", Method: "POST", Href: basePath + "/deactivate", Label: "Deactivate"},
		).
		Send(w)
}

func (h *WorkflowHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before deactivation
	existing, err := h.workflowSvc.GetByID(r.Context(), workflowID)
	if err != nil {
		dto.NotFound(w, "Workflow")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	// Business rule: Check if workflow can be deactivated
	hasActiveSchedules := false
	if h.scheduleSvc != nil {
		schedules, err := h.scheduleSvc.GetByWorkflow(r.Context(), workflowID)
		if err == nil {
			for _, s := range schedules {
				if s.IsActive {
					hasActiveSchedules = true
					break
				}
			}
		}
	}
	if err := validator.CanDeactivateWorkflow(existing.Status, hasActiveSchedules); err != nil {
		dto.BadRequest(w, err.Error())
		return
	}

	if err := h.workflowSvc.Deactivate(r.Context(), workflowID); err != nil {
		dto.InternalServerError(w, "failed to deactivate workflow")
		return
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID

	dto.NewResponse(map[string]string{"status": "inactive"}).
		WithLinks(&dto.Links{Self: basePath}).
		WithActions(
			dto.Action{Name: "activate", Method: "POST", Href: basePath + "/activate", Label: "Activate"},
		).
		Send(w)
}

// hasTriggerNode checks if workflow has a trigger node
func hasTriggerNode(nodes models.JSONArray) bool {
	for _, node := range nodes {
		if nodeMap, ok := node.(map[string]interface{}); ok {
			if nodeType, ok := nodeMap["type"].(string); ok {
				if strings.HasPrefix(nodeType, "trigger.") {
					return true
				}
			}
		}
	}
	return false
}

func (h *WorkflowHandler) GetVersions(w http.ResponseWriter, r *http.Request) {
	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
		return
	}

	versions, err := h.workflowSvc.GetVersions(r.Context(), workflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get versions")
		return
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID + "/versions"

	type VersionWithActions struct {
		dto.WorkflowVersionResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}

	response := []VersionWithActions{}
	for _, v := range versions {
		vNum := strconv.Itoa(v.Version)
		actions := []dto.Action{
			{Name: "view", Method: "GET", Href: basePath + "/" + vNum, Label: "View Version"},
			{Name: "diff", Method: "GET", Href: basePath + "/" + vNum + "/diff", Label: "Compare with Current"},
			{Name: "rollback", Method: "POST", Href: basePath + "/" + vNum + "/rollback", Label: "Rollback"},
		}

		response = append(response, VersionWithActions{
			WorkflowVersionResponse: dto.WorkflowVersionResponse{
				ID:            v.ID.String(),
				Version:       v.Version,
				Nodes:         v.Nodes,
				Connections:   v.Connections,
				Settings:      v.Settings,
				ChangeMessage: v.ChangeMessage,
				CreatedAt:     v.CreatedAt.Unix(),
			},
			Actions: actions,
		})
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		WithMeta(&dto.Meta{Total: int64(len(response)), Page: 1, PerPage: len(response), TotalPages: 1}).
		Send(w)
}

func (h *WorkflowHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
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

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflowID.String()
	vNum := strconv.Itoa(version)
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID + "/versions/" + vNum

	response := struct {
		dto.WorkflowVersionResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}{
		WorkflowVersionResponse: dto.WorkflowVersionResponse{
			ID:            v.ID.String(),
			Version:       v.Version,
			Nodes:         v.Nodes,
			Connections:   v.Connections,
			Settings:      v.Settings,
			ChangeMessage: v.ChangeMessage,
			CreatedAt:     v.CreatedAt.Unix(),
		},
		Actions: []dto.Action{
			{Name: "diff", Method: "GET", Href: basePath + "/diff", Label: "Compare with Current"},
			{Name: "rollback", Method: "POST", Href: basePath + "/rollback", Label: "Rollback to This Version"},
		},
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

// Export exports a workflow as JSON
func (h *WorkflowHandler) Export(w http.ResponseWriter, r *http.Request) {
	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
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
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
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
	claims := middleware.MustUser(w, r)
	if claims == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
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

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	wfID := workflow.ID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/workflows/" + wfID

	response := dto.WorkflowResponse{
		ID:          wfID,
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
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

// Duplicate creates a copy of a workflow with optional variable substitution
func (h *WorkflowHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	workflowID, ok := middleware.ParseUUID(w, r, "workflowID")
	if !ok {
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

// filtersToWorkflowRepoFilter converts DTO filters to repository filters
func filtersToWorkflowRepoFilter(f *dto.WorkflowFilters) *repositories.WorkflowFilter {
	if f == nil {
		return nil
	}
	return &repositories.WorkflowFilter{
		Status:        f.Status,
		Search:        f.Search,
		Tags:          f.Tags,
		CreatedAfter:  f.CreatedAfter,
		CreatedBefore: f.CreatedBefore,
		UpdatedAfter:  f.UpdatedAfter,
		UpdatedBefore: f.UpdatedBefore,
		SortBy:        f.SortBy,
		Order:         f.Order,
	}
}
