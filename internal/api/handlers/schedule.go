package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

type ScheduleHandler struct {
	scheduleSvc *services.ScheduleService
}

// NewScheduleHandler creates a new ScheduleHandler for workflow scheduling.
func NewScheduleHandler(scheduleSvc *services.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleSvc: scheduleSvc}
}

// List returns all schedules for a workspace.
func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	pg := dto.ParsePagination(r)
	schedules, total, err := h.scheduleSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID, pg.Opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list schedules")
		return
	}

	// Build response with actions
	type ScheduleWithActions struct {
		dto.ScheduleResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}

	response := []ScheduleWithActions{}
	wsID := wsCtx.WorkspaceID.String()

	for _, sched := range schedules {
		var nextRunAt, lastRunAt *int64
		if sched.NextRunAt != nil {
			ts := sched.NextRunAt.Unix()
			nextRunAt = &ts
		}
		if sched.LastRunAt != nil {
			ts := sched.LastRunAt.Unix()
			lastRunAt = &ts
		}

		schedID := sched.ID.String()
		basePath := "/api/v1/workspaces/" + wsID + "/schedules/" + schedID

		// Build actions based on schedule status
		var actions []dto.Action
		if sched.IsActive {
			actions = append(actions, dto.Action{Name: "pause", Method: "POST", Href: basePath + "/pause", Label: "Pause Schedule"})
		} else {
			actions = append(actions, dto.Action{Name: "resume", Method: "POST", Href: basePath + "/resume", Label: "Resume Schedule"})
		}
		actions = append(actions, dto.Action{Name: "trigger", Method: "POST", Href: basePath + "/trigger", Label: "Trigger Now"})
		actions = append(actions, dto.Action{Name: "edit", Method: "PUT", Href: basePath, Label: "Edit Schedule"})
		actions = append(actions, dto.Action{Name: "delete", Method: "DELETE", Href: basePath, Label: "Delete"})

		response = append(response, ScheduleWithActions{
			ScheduleResponse: dto.ScheduleResponse{
				ID:             schedID,
				WorkflowID:     sched.WorkflowID.String(),
				Name:           sched.Name,
				Description:    sched.Description,
				CronExpression: sched.CronExpression,
				Timezone:       sched.Timezone,
				IsActive:       sched.IsActive,
				InputData:      sched.InputData,
				NextRunAt:      nextRunAt,
				LastRunAt:      lastRunAt,
				RunCount:       sched.RunCount,
				CreatedAt:      sched.CreatedAt.Unix(),
			},
			Actions: actions,
		})
	}

	// Build links
	basePath := "/api/v1/workspaces/" + wsID + "/schedules"
	links := &dto.Links{
		Self: fmt.Sprintf("%s?page=%d&per_page=%d", basePath, pg.Page, pg.PerPage),
	}
	meta := pg.NewMeta(total)
	if pg.Page < meta.TotalPages {
		links.Next = fmt.Sprintf("%s?page=%d&per_page=%d", basePath, pg.Page+1, pg.PerPage)
	}
	if pg.Page > 1 {
		links.Prev = fmt.Sprintf("%s?page=%d&per_page=%d", basePath, pg.Page-1, pg.PerPage)
	}
	links.First = fmt.Sprintf("%s?page=1&per_page=%d", basePath, pg.PerPage)
	if meta.TotalPages > 0 {
		links.Last = fmt.Sprintf("%s?page=%d&per_page=%d", basePath, meta.TotalPages, pg.PerPage)
	}

	// Apply field selection
	data := dto.SelectFields(r, response)

	dto.NewResponse(data).
		WithLinks(links).
		WithMeta(meta).
		Send(w)
}

func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if claims == nil {
		return
	}

	var req dto.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	workflowID, err := uuid.Parse(req.WorkflowID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid workflow ID")
		return
	}

	schedule, err := h.scheduleSvc.Create(r.Context(), services.CreateScheduleInput{
		WorkflowID:     workflowID,
		WorkspaceID:    wsCtx.WorkspaceID,
		CreatedBy:      claims.UserID,
		Name:           req.Name,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		Timezone:       req.Timezone,
		InputData:      req.InputData,
	})
	if err != nil {
		if err == services.ErrInvalidCron {
			dto.ErrorResponse(w, http.StatusBadRequest, "invalid cron expression")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create schedule")
		return
	}

	var nextRunAt *int64
	if schedule.NextRunAt != nil {
		ts := schedule.NextRunAt.Unix()
		nextRunAt = &ts
	}

	dto.Created(w, dto.ScheduleResponse{
		ID:             schedule.ID.String(),
		WorkflowID:     schedule.WorkflowID.String(),
		Name:           schedule.Name,
		Description:    schedule.Description,
		CronExpression: schedule.CronExpression,
		Timezone:       schedule.Timezone,
		IsActive:       schedule.IsActive,
		InputData:      schedule.InputData,
		NextRunAt:      nextRunAt,
		RunCount:       schedule.RunCount,
		CreatedAt:      schedule.CreatedAt.Unix(),
	})
}

func (h *ScheduleHandler) Get(w http.ResponseWriter, r *http.Request) {
	scheduleID, ok := middleware.ParseUUID(w, r, "scheduleID")
	if !ok {
		return
	}

	schedule, err := h.scheduleSvc.GetByID(r.Context(), scheduleID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "schedule not found")
		return
	}

	if !ValidateWorkspaceOwnership(w, r, schedule) {
		return
	}

	var nextRunAt, lastRunAt *int64
	if schedule.NextRunAt != nil {
		ts := schedule.NextRunAt.Unix()
		nextRunAt = &ts
	}
	if schedule.LastRunAt != nil {
		ts := schedule.LastRunAt.Unix()
		lastRunAt = &ts
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	schedID := schedule.ID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/schedules/" + schedID

	actions := []dto.Action{
		{Name: "edit", Method: "PUT", Href: basePath, Label: "Edit Schedule"},
		{Name: "trigger", Method: "POST", Href: basePath + "/trigger", Label: "Trigger Now"},
	}
	if schedule.IsActive {
		actions = append(actions, dto.Action{Name: "pause", Method: "POST", Href: basePath + "/pause", Label: "Pause"})
	} else {
		actions = append(actions, dto.Action{Name: "resume", Method: "POST", Href: basePath + "/resume", Label: "Resume"})
	}
	actions = append(actions, dto.Action{Name: "delete", Method: "DELETE", Href: basePath, Label: "Delete"})

	response := struct {
		dto.ScheduleResponse
		Actions []dto.Action `json:"actions,omitempty"`
	}{
		ScheduleResponse: dto.ScheduleResponse{
			ID:             schedID,
			WorkflowID:     schedule.WorkflowID.String(),
			Name:           schedule.Name,
			Description:    schedule.Description,
			CronExpression: schedule.CronExpression,
			Timezone:       schedule.Timezone,
			IsActive:       schedule.IsActive,
			InputData:      schedule.InputData,
			NextRunAt:      nextRunAt,
			LastRunAt:      lastRunAt,
			RunCount:       schedule.RunCount,
			CreatedAt:      schedule.CreatedAt.Unix(),
		},
		Actions: actions,
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	scheduleID, ok := middleware.ParseUUID(w, r, "scheduleID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before modification
	existing, err := h.scheduleSvc.GetByID(r.Context(), scheduleID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "schedule not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	var req dto.UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		dto.ValidationErrorResponse(w, err)
		return
	}

	schedule, err := h.scheduleSvc.Update(r.Context(), scheduleID, services.UpdateScheduleInput{
		Name:           req.Name,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		Timezone:       req.Timezone,
		InputData:      req.InputData,
	})
	if err != nil {
		if err == services.ErrInvalidCron {
			dto.ErrorResponse(w, http.StatusBadRequest, "invalid cron expression")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to update schedule")
		return
	}

	var nextRunAt, lastRunAt *int64
	if schedule.NextRunAt != nil {
		ts := schedule.NextRunAt.Unix()
		nextRunAt = &ts
	}
	if schedule.LastRunAt != nil {
		ts := schedule.LastRunAt.Unix()
		lastRunAt = &ts
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	schedID := schedule.ID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/schedules/" + schedID

	response := dto.ScheduleResponse{
		ID:             schedID,
		WorkflowID:     schedule.WorkflowID.String(),
		Name:           schedule.Name,
		Description:    schedule.Description,
		CronExpression: schedule.CronExpression,
		Timezone:       schedule.Timezone,
		IsActive:       schedule.IsActive,
		InputData:      schedule.InputData,
		NextRunAt:      nextRunAt,
		LastRunAt:      lastRunAt,
		RunCount:       schedule.RunCount,
		CreatedAt:      schedule.CreatedAt.Unix(),
	}

	dto.NewResponse(response).
		WithLinks(&dto.Links{Self: basePath}).
		Send(w)
}

func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	scheduleID, ok := middleware.ParseUUID(w, r, "scheduleID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before deletion
	existing, err := h.scheduleSvc.GetByID(r.Context(), scheduleID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "schedule not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	if err := h.scheduleSvc.Delete(r.Context(), scheduleID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to delete schedule")
		return
	}

	dto.NoContent(w)
}

func (h *ScheduleHandler) Pause(w http.ResponseWriter, r *http.Request) {
	scheduleID, ok := middleware.ParseUUID(w, r, "scheduleID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before pause
	existing, err := h.scheduleSvc.GetByID(r.Context(), scheduleID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "schedule not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	if err := h.scheduleSvc.Pause(r.Context(), scheduleID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to pause schedule")
		return
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	schedID := scheduleID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/schedules/" + schedID

	dto.NewResponse(map[string]bool{"is_active": false}).
		WithLinks(&dto.Links{Self: basePath}).
		WithActions(dto.Action{Name: "resume", Method: "POST", Href: basePath + "/resume", Label: "Resume"}).
		Send(w)
}

func (h *ScheduleHandler) Resume(w http.ResponseWriter, r *http.Request) {
	scheduleID, ok := middleware.ParseUUID(w, r, "scheduleID")
	if !ok {
		return
	}

	// SECURITY: Validate ownership before resume
	existing, err := h.scheduleSvc.GetByID(r.Context(), scheduleID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "schedule not found")
		return
	}
	if !ValidateWorkspaceOwnership(w, r, existing) {
		return
	}

	if err := h.scheduleSvc.Resume(r.Context(), scheduleID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to resume schedule")
		return
	}

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	schedID := scheduleID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/schedules/" + schedID

	dto.NewResponse(map[string]bool{"is_active": true}).
		WithLinks(&dto.Links{Self: basePath}).
		WithActions(dto.Action{Name: "pause", Method: "POST", Href: basePath + "/pause", Label: "Pause"}).
		Send(w)
}
