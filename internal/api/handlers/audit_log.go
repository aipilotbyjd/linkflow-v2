package handlers

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type AuditLogHandler struct {
	auditSvc *services.AuditLogService
}

func NewAuditLogHandler(auditSvc *services.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{auditSvc: auditSvc}
}

func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	pg := dto.ParsePagination(r)
	logs, total, err := h.auditSvc.GetByWorkspace(r.Context(), wsCtx.WorkspaceID, pg.Opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to list audit logs")
		return
	}

	dto.JSONWithMeta(w, http.StatusOK, logs, pg.NewMeta(total))
}

func (h *AuditLogHandler) Search(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.MustWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	action := dto.QueryString(r, "action")
	resourceType := dto.QueryString(r, "resource_type")
	userID, ok := middleware.ParseUUIDQuery(w, r, "user_id")
	if !ok {
		return
	}

	var start, end *time.Time
	if s := dto.QueryString(r, "start"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			start = &t
		}
	}
	if e := dto.QueryString(r, "end"); e != "" {
		if t, err := time.Parse(time.RFC3339, e); err == nil {
			end = &t
		}
	}

	pg := dto.ParsePagination(r)
	logs, total, err := h.auditSvc.Search(r.Context(), wsCtx.WorkspaceID, action, resourceType, userID, start, end, pg.Opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to search audit logs")
		return
	}

	dto.JSONWithMeta(w, http.StatusOK, logs, pg.NewMeta(total))
}
