package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

// =============================================================================
// AUDIT LOG HANDLER
// =============================================================================

type AuditLogHandler struct {
	auditSvc *services.AuditLogService
}

func NewAuditLogHandler(auditSvc *services.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{auditSvc: auditSvc}
}

func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	opts := repositories.NewListOptions(page, perPage)

	logs, total, err := h.auditSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID, opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list audit logs")
		return
	}

	dto.JSONWithMeta(w, http.StatusOK, logs, &dto.Meta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

func (h *AuditLogHandler) Search(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	action := r.URL.Query().Get("action")
	resourceType := r.URL.Query().Get("resource_type")

	var userID *uuid.UUID
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			userID = &id
		}
	}

	var start, end *time.Time
	if s := r.URL.Query().Get("start"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			start = &t
		}
	}
	if e := r.URL.Query().Get("end"); e != "" {
		if t, err := time.Parse(time.RFC3339, e); err == nil {
			end = &t
		}
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	opts := repositories.NewListOptions(page, perPage)

	logs, total, err := h.auditSvc.Search(r.Context(), wsCtx.WorkspaceID, action, resourceType, userID, start, end, opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to search audit logs")
		return
	}

	dto.JSONWithMeta(w, http.StatusOK, logs, &dto.Meta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// =============================================================================
// ALERT HANDLER
// =============================================================================

type AlertHandler struct {
	alertSvc *services.AlertService
}

func NewAlertHandler(alertSvc *services.AlertService) *AlertHandler {
	return &AlertHandler{alertSvc: alertSvc}
}

func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	alerts, err := h.alertSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	dto.OK(w, alerts)
}

func (h *AlertHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	var req struct {
		Name         string                 `json:"name"`
		WorkflowID   *string                `json:"workflow_id,omitempty"`
		Type         string                 `json:"type"`
		Trigger      string                 `json:"trigger"`
		Config       map[string]interface{} `json:"config"`
		Conditions   map[string]interface{} `json:"conditions,omitempty"`
		CooldownMins int                    `json:"cooldown_mins"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	var workflowID *uuid.UUID
	if req.WorkflowID != nil {
		id, err := uuid.Parse(*req.WorkflowID)
		if err != nil {
			dto.BadRequest(w, "invalid workflow_id")
			return
		}
		workflowID = &id
	}

	alert, err := h.alertSvc.Create(r.Context(), services.CreateAlertInput{
		WorkspaceID:  wsCtx.WorkspaceID,
		WorkflowID:   workflowID,
		CreatedBy:    claims.UserID,
		Name:         req.Name,
		Type:         req.Type,
		Trigger:      req.Trigger,
		Config:       req.Config,
		Conditions:   req.Conditions,
		CooldownMins: req.CooldownMins,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create alert")
		return
	}

	dto.Created(w, alert)
}

func (h *AlertHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "alertId"))
	if err != nil {
		dto.BadRequest(w, "invalid alert ID")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if err := h.alertSvc.Update(r.Context(), id, updates); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update alert")
		return
	}

	dto.OK(w, map[string]string{"message": "alert updated"})
}

func (h *AlertHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "alertId"))
	if err != nil {
		dto.BadRequest(w, "invalid alert ID")
		return
	}

	if err := h.alertSvc.Delete(r.Context(), id); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete alert")
		return
	}

	dto.NoContent(w)
}

// =============================================================================
// WORKFLOW COMMENT HANDLER
// =============================================================================

type WorkflowCommentHandler struct {
	commentSvc *services.WorkflowCommentService
}

func NewWorkflowCommentHandler(commentSvc *services.WorkflowCommentService) *WorkflowCommentHandler {
	return &WorkflowCommentHandler{commentSvc: commentSvc}
}

func (h *WorkflowCommentHandler) List(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		dto.BadRequest(w, "invalid workflow ID")
		return
	}

	nodeID := r.URL.Query().Get("node_id")

	var comments interface{}
	if nodeID != "" {
		comments, err = h.commentSvc.GetByNode(r.Context(), workflowID, nodeID)
	} else {
		comments, err = h.commentSvc.GetByWorkflow(r.Context(), workflowID)
	}

	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list comments")
		return
	}

	dto.OK(w, comments)
}

func (h *WorkflowCommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		dto.BadRequest(w, "invalid workflow ID")
		return
	}

	var req struct {
		NodeID   *string `json:"node_id,omitempty"`
		ParentID *string `json:"parent_id,omitempty"`
		Content  string  `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			dto.BadRequest(w, "invalid parent_id")
			return
		}
		parentID = &id
	}

	comment, err := h.commentSvc.Create(r.Context(), services.CreateCommentInput{
		WorkflowID:  workflowID,
		WorkspaceID: wsCtx.WorkspaceID,
		NodeID:      req.NodeID,
		ParentID:    parentID,
		CreatedBy:   claims.UserID,
		Content:     req.Content,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create comment")
		return
	}

	dto.Created(w, comment)
}

func (h *WorkflowCommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "commentId"))
	if err != nil {
		dto.BadRequest(w, "invalid comment ID")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if err := h.commentSvc.Update(r.Context(), id, req.Content); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update comment")
		return
	}

	dto.OK(w, map[string]string{"message": "comment updated"})
}

func (h *WorkflowCommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "commentId"))
	if err != nil {
		dto.BadRequest(w, "invalid comment ID")
		return
	}

	if err := h.commentSvc.Delete(r.Context(), id); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete comment")
		return
	}

	dto.NoContent(w)
}

func (h *WorkflowCommentHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "commentId"))
	if err != nil {
		dto.BadRequest(w, "invalid comment ID")
		return
	}

	if err := h.commentSvc.Resolve(r.Context(), id, claims.UserID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to resolve comment")
		return
	}

	dto.OK(w, map[string]string{"message": "comment resolved"})
}

// =============================================================================
// EXECUTION SHARE HANDLER
// =============================================================================

type ExecutionShareHandler struct {
	shareSvc *services.ExecutionShareService
	execSvc  *services.ExecutionService
}

func NewExecutionShareHandler(shareSvc *services.ExecutionShareService, execSvc *services.ExecutionService) *ExecutionShareHandler {
	return &ExecutionShareHandler{shareSvc: shareSvc, execSvc: execSvc}
}

func (h *ExecutionShareHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	execID, err := uuid.Parse(chi.URLParam(r, "executionId"))
	if err != nil {
		dto.BadRequest(w, "invalid execution ID")
		return
	}

	var req struct {
		ExpiresIn     *int    `json:"expires_in,omitempty"` // hours
		Password      *string `json:"password,omitempty"`
		MaxViews      *int    `json:"max_views,omitempty"`
		AllowDownload bool    `json:"allow_download"`
		IncludeLogs   bool    `json:"include_logs"`
		IncludeData   bool    `json:"include_data"`
	}
	req.IncludeLogs = true // default

	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	var expiresAt *time.Time
	if req.ExpiresIn != nil {
		t := time.Now().Add(time.Duration(*req.ExpiresIn) * time.Hour)
		expiresAt = &t
	}

	share, err := h.shareSvc.Create(r.Context(), services.CreateShareInput{
		ExecutionID:   execID,
		WorkspaceID:   wsCtx.WorkspaceID,
		CreatedBy:     claims.UserID,
		ExpiresAt:     expiresAt,
		Password:      req.Password,
		MaxViews:      req.MaxViews,
		AllowDownload: req.AllowDownload,
		IncludeLogs:   req.IncludeLogs,
		IncludeData:   req.IncludeData,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create share link")
		return
	}

	dto.Created(w, map[string]interface{}{
		"id":    share.ID,
		"token": share.Token,
		"url":   "/shared/executions/" + share.Token,
	})
}

func (h *ExecutionShareHandler) GetShared(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	share, err := h.shareSvc.GetByToken(r.Context(), token)
	if err != nil {
		dto.NotFound(w, "share link not found")
		return
	}

	password := r.URL.Query().Get("password")
	var pwd *string
	if password != "" {
		pwd = &password
	}

	if err := h.shareSvc.ValidateAccess(r.Context(), share, pwd); err != nil {
		if err == services.ErrInvalidPassword {
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid password")
			return
		}
		dto.ErrorResponse(w, http.StatusForbidden, err.Error())
		return
	}

	h.shareSvc.IncrementViews(r.Context(), share.ID)

	execution, err := h.execSvc.GetByID(r.Context(), share.ExecutionID)
	if err != nil {
		dto.NotFound(w, "execution not found")
		return
	}

	response := map[string]interface{}{
		"id":               execution.ID,
		"status":           execution.Status,
		"trigger_type":     execution.TriggerType,
		"workflow_version": execution.WorkflowVersion,
		"nodes_total":      execution.NodesTotal,
		"nodes_completed":  execution.NodesCompleted,
		"queued_at":        execution.QueuedAt,
		"started_at":       execution.StartedAt,
		"completed_at":     execution.CompletedAt,
	}

	if share.IncludeData {
		response["input_data"] = execution.InputData
		response["output_data"] = execution.OutputData
	}

	if share.IncludeLogs {
		nodeExecs, _ := h.execSvc.GetNodeExecutions(r.Context(), execution.ID)
		response["node_executions"] = nodeExecs
	}

	dto.OK(w, response)
}

func (h *ExecutionShareHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "shareId"))
	if err != nil {
		dto.BadRequest(w, "invalid share ID")
		return
	}

	if err := h.shareSvc.Delete(r.Context(), id); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete share link")
		return
	}

	dto.NoContent(w)
}

// =============================================================================
// ENVIRONMENT VARIABLE HANDLER
// =============================================================================

type EnvironmentVariableHandler struct {
	envSvc *services.EnvironmentVariableService
}

func NewEnvironmentVariableHandler(envSvc *services.EnvironmentVariableService) *EnvironmentVariableHandler {
	return &EnvironmentVariableHandler{envSvc: envSvc}
}

func (h *EnvironmentVariableHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	env := r.URL.Query().Get("environment")
	var envPtr *string
	if env != "" {
		envPtr = &env
	}

	vars, err := h.envSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID, envPtr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list environment variables")
		return
	}

	// Mask secret values
	response := make([]map[string]interface{}, len(vars))
	for i, v := range vars {
		response[i] = map[string]interface{}{
			"id":          v.ID,
			"name":        v.Name,
			"is_secret":   v.IsSecret,
			"environment": v.Environment,
			"description": v.Description,
			"created_at":  v.CreatedAt,
			"updated_at":  v.UpdatedAt,
		}
		if !v.IsSecret {
			response[i]["value"] = "[REDACTED]" // Always hide values in list
		}
	}

	dto.OK(w, response)
}

func (h *EnvironmentVariableHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Value       string  `json:"value"`
		IsSecret    bool    `json:"is_secret"`
		Environment *string `json:"environment,omitempty"`
		Description *string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	envVar, err := h.envSvc.Create(r.Context(), services.CreateEnvVarInput{
		WorkspaceID: wsCtx.WorkspaceID,
		CreatedBy:   claims.UserID,
		Name:        req.Name,
		Value:       req.Value,
		IsSecret:    req.IsSecret,
		Environment: req.Environment,
		Description: req.Description,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create environment variable")
		return
	}

	dto.Created(w, map[string]interface{}{
		"id":          envVar.ID,
		"name":        envVar.Name,
		"is_secret":   envVar.IsSecret,
		"environment": envVar.Environment,
	})
}

func (h *EnvironmentVariableHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "varId"))
	if err != nil {
		dto.BadRequest(w, "invalid variable ID")
		return
	}

	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if err := h.envSvc.Update(r.Context(), id, req.Value); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update environment variable")
		return
	}

	dto.OK(w, map[string]string{"message": "environment variable updated"})
}

func (h *EnvironmentVariableHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "varId"))
	if err != nil {
		dto.BadRequest(w, "invalid variable ID")
		return
	}

	if err := h.envSvc.Delete(r.Context(), id); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete environment variable")
		return
	}

	dto.NoContent(w)
}

// =============================================================================
// WORKFLOW IMPORT/EXPORT HANDLER
// =============================================================================

type WorkflowExportHandler struct {
	exportSvc *services.WorkflowExportService
}

func NewWorkflowExportHandler(exportSvc *services.WorkflowExportService) *WorkflowExportHandler {
	return &WorkflowExportHandler{exportSvc: exportSvc}
}

func (h *WorkflowExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		dto.BadRequest(w, "invalid workflow ID")
		return
	}

	includeCredentials := r.URL.Query().Get("include_credentials") == "true"

	data, err := h.exportSvc.Export(r.Context(), workflowID, claims.UserID, includeCredentials)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to export workflow")
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=workflow-"+workflowID.String()+".json")

	json.NewEncoder(w).Encode(data)
}

func (h *WorkflowExportHandler) Import(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if claims == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	var data services.WorkflowExportData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		dto.BadRequest(w, "invalid import data")
		return
	}

	workflow, err := h.exportSvc.Import(r.Context(), wsCtx.WorkspaceID, claims.UserID, &data)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to import workflow")
		return
	}

	dto.Created(w, map[string]interface{}{
		"id":      workflow.ID,
		"name":    workflow.Name,
		"message": "workflow imported successfully",
	})
}

// =============================================================================
// ANALYTICS HANDLER
// =============================================================================

type AnalyticsHandler struct {
	analyticsSvc *services.AnalyticsService
}

func NewAnalyticsHandler(analyticsSvc *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsSvc: analyticsSvc}
}

func (h *AnalyticsHandler) GetWorkspaceAnalytics(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "workspace context required")
		return
	}

	// Default to last 30 days
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	if s := r.URL.Query().Get("start"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start = t
		}
	}
	if e := r.URL.Query().Get("end"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end = t
		}
	}

	analytics, err := h.analyticsSvc.GetWorkspaceAnalytics(r.Context(), wsCtx.WorkspaceID, start, end)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get analytics")
		return
	}

	dto.OK(w, analytics)
}

func (h *AnalyticsHandler) GetWorkflowAnalytics(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
	if err != nil {
		dto.BadRequest(w, "invalid workflow ID")
		return
	}

	// Default to last 30 days
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	if s := r.URL.Query().Get("start"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start = t
		}
	}
	if e := r.URL.Query().Get("end"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end = t
		}
	}

	analytics, err := h.analyticsSvc.GetWorkflowAnalytics(r.Context(), workflowID, start, end)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get analytics")
		return
	}

	dto.OK(w, analytics)
}

// =============================================================================
// EXECUTION REPLAY HANDLER
// =============================================================================

type ExecutionReplayHandler struct {
	replaySvc *services.ExecutionReplayService
}

func NewExecutionReplayHandler(replaySvc *services.ExecutionReplayService) *ExecutionReplayHandler {
	return &ExecutionReplayHandler{replaySvc: replaySvc}
}

func (h *ExecutionReplayHandler) Replay(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	execID, err := uuid.Parse(chi.URLParam(r, "executionId"))
	if err != nil {
		dto.BadRequest(w, "invalid execution ID")
		return
	}

	execution, err := h.replaySvc.Replay(r.Context(), execID, &claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to replay execution")
		return
	}

	dto.Created(w, map[string]interface{}{
		"id":      execution.ID,
		"status":  execution.Status,
		"message": "execution replay started",
	})
}

func (h *ExecutionReplayHandler) ReplayFromNode(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		dto.ErrorResponse(w, http.StatusForbidden, "unauthorized")
		return
	}

	execID, err := uuid.Parse(chi.URLParam(r, "executionId"))
	if err != nil {
		dto.BadRequest(w, "invalid execution ID")
		return
	}

	var req struct {
		NodeID string `json:"node_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	execution, err := h.replaySvc.ReplayFromNode(r.Context(), execID, req.NodeID, &claims.UserID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to replay execution from node")
		return
	}

	dto.Created(w, map[string]interface{}{
		"id":        execution.ID,
		"status":    execution.Status,
		"from_node": req.NodeID,
		"message":   "partial execution replay started",
	})
}
