package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
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
	filters := dto.ParseScheduleFilters(r)

	// Convert DTO filters to repository filters
	repoFilter := &repositories.ScheduleFilter{
		IsActive: filters.IsActive,
		Search:   filters.Search,
		SortBy:   filters.SortBy,
		Order:    filters.Order,
	}
	if filters.WorkflowID != nil {
		if wfID, err := uuid.Parse(*filters.WorkflowID); err == nil {
			repoFilter.WorkflowID = &wfID
		}
	}

	schedules, total, err := h.scheduleSvc.GetByWorkspaceWithFilters(r.Context(), wsCtx.WorkspaceID, repoFilter, pg.Opts)
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
			ScheduleResponse: buildScheduleResponse(&sched),
			Actions:          actions,
		})
	}

	// Build links with filter query string preservation
	basePath := "/api/v1/workspaces/" + wsID + "/schedules"
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

	dto.Created(w, buildScheduleResponse(schedule))
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
		ScheduleResponse: buildScheduleResponse(schedule),
		Actions:          actions,
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

	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	wsID := wsCtx.WorkspaceID.String()
	schedID := schedule.ID.String()
	basePath := "/api/v1/workspaces/" + wsID + "/schedules/" + schedID

	dto.NewResponse(buildScheduleResponse(schedule)).
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

// buildScheduleResponse creates a ScheduleResponse from a Schedule model
func buildScheduleResponse(s *models.Schedule) dto.ScheduleResponse {
	var nextRunAt, lastRunAt *int64
	if s.NextRunAt != nil {
		ts := s.NextRunAt.Unix()
		nextRunAt = &ts
	}
	if s.LastRunAt != nil {
		ts := s.LastRunAt.Unix()
		lastRunAt = &ts
	}

	var lastExecutionID *string
	if s.LastExecutionID != nil {
		id := s.LastExecutionID.String()
		lastExecutionID = &id
	}

	return dto.ScheduleResponse{
		ID:              s.ID.String(),
		WorkflowID:      s.WorkflowID.String(),
		WorkspaceID:     s.WorkspaceID.String(),
		CreatedBy:       s.CreatedBy.String(),
		Name:            s.Name,
		Description:     s.Description,
		CronExpression:  s.CronExpression,
		Timezone:        s.Timezone,
		IsActive:        s.IsActive,
		InputData:       s.InputData,
		NextRunAt:       nextRunAt,
		LastRunAt:       lastRunAt,
		LastExecutionID: lastExecutionID,
		RunCount:        s.RunCount,
		CreatedAt:       s.CreatedAt.Unix(),
		UpdatedAt:       s.UpdatedAt.Unix(),
	}
}
