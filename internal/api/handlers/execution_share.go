package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type ExecutionShareHandler struct {
	shareSvc *services.ExecutionShareService
	execSvc  *services.ExecutionService
}

func NewExecutionShareHandler(shareSvc *services.ExecutionShareService, execSvc *services.ExecutionService) *ExecutionShareHandler {
	return &ExecutionShareHandler{shareSvc: shareSvc, execSvc: execSvc}
}

func (h *ExecutionShareHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	execID, ok := middleware.ParseUUID(w, r, "executionId")
	if !ok {
		return
	}

	var req struct {
		ExpiresIn     *int    `json:"expires_in,omitempty"`
		Password      *string `json:"password,omitempty"`
		MaxViews      *int    `json:"max_views,omitempty"`
		AllowDownload bool    `json:"allow_download"`
		IncludeLogs   bool    `json:"include_logs"`
		IncludeData   bool    `json:"include_data"`
	}
	req.IncludeLogs = true

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

	password := dto.QueryString(r, "password")
	var pwd *string
	if password != "" {
		pwd = &password
	}

	if err := h.shareSvc.ValidateAccess(r.Context(), share, pwd); err != nil {
		if err == services.ErrInvalidPassword {
			dto.Unauthorized(w, "invalid password")
			return
		}
		dto.Forbidden(w, err.Error())
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
	id, ok := middleware.ParseUUID(w, r, "shareId")
	if !ok {
		return
	}

	if err := h.shareSvc.Delete(r.Context(), id); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete share link")
		return
	}

	dto.NoContent(w)
}
