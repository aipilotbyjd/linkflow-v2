package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type AlertHandler struct {
	alertSvc *services.AlertService
}

func NewAlertHandler(alertSvc *services.AlertService) *AlertHandler {
	return &AlertHandler{alertSvc: alertSvc}
}

func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	alerts, err := h.alertSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	// Build response with actions
	type AlertWithActions struct {
		ID           string      `json:"id"`
		Name         string      `json:"name"`
		Type         string      `json:"type"`
		Trigger      string      `json:"trigger"`
		IsActive     bool        `json:"is_active"`
		CooldownMins int         `json:"cooldown_mins"`
		FireCount    int         `json:"fire_count"`
		LastFiredAt  *int64      `json:"last_fired_at,omitempty"`
		CreatedAt    int64       `json:"created_at"`
		Actions      []dto.Action `json:"actions,omitempty"`
	}

	wsID := wsCtx.WorkspaceID.String()
	response := []AlertWithActions{}

	for _, alert := range alerts {
		alertID := alert.ID.String()
		basePath := "/api/v1/workspaces/" + wsID + "/alerts/" + alertID

		var lastFiredAt *int64
		if alert.LastFiredAt != nil {
			ts := alert.LastFiredAt.Unix()
			lastFiredAt = &ts
		}

		var actions []dto.Action
		if alert.IsActive {
			actions = append(actions, dto.Action{Name: "disable", Method: "POST", Href: basePath + "/disable", Label: "Disable Alert"})
		} else {
			actions = append(actions, dto.Action{Name: "enable", Method: "POST", Href: basePath + "/enable", Label: "Enable Alert"})
		}
		actions = append(actions, dto.Action{Name: "edit", Method: "PUT", Href: basePath, Label: "Edit Alert"})
		actions = append(actions, dto.Action{Name: "test", Method: "POST", Href: basePath + "/test", Label: "Test Alert"})
		actions = append(actions, dto.Action{Name: "delete", Method: "DELETE", Href: basePath, Label: "Delete"})

		response = append(response, AlertWithActions{
			ID:           alertID,
			Name:         alert.Name,
			Type:         alert.Type,
			Trigger:      alert.Trigger,
			IsActive:     alert.IsActive,
			CooldownMins: alert.CooldownMins,
			FireCount:    alert.FireCount,
			LastFiredAt:  lastFiredAt,
			CreatedAt:    alert.CreatedAt.Unix(),
			Actions:      actions,
		})
	}

	// Apply field selection
	data := dto.SelectFields(r, response)

	dto.NewResponse(data).
		WithLinks(&dto.Links{Self: "/api/v1/workspaces/" + wsID + "/alerts"}).
		Send(w)
}

func (h *AlertHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
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
	id, ok := middleware.ParseUUID(w, r, "alertID")
	if !ok {
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
	id, ok := middleware.ParseUUID(w, r, "alertID")
	if !ok {
		return
	}

	if err := h.alertSvc.Delete(r.Context(), id); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete alert")
		return
	}

	dto.NoContent(w)
}
