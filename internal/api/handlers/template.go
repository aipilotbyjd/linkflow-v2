package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type TemplateHandler struct {
	templateSvc *services.TemplateService
}

func NewTemplateHandler(templateSvc *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateSvc: templateSvc}
}

// List returns public templates
func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	opts := parseListOptions(r)
	templates, total, err := h.templateSvc.GetPublicTemplates(r.Context(), opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"total":     total,
		"page":      opts.Offset/opts.Limit + 1,
		"limit":     opts.Limit,
	})
}

// GetFeatured returns featured templates
func (h *TemplateHandler) GetFeatured(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	templates, err := h.templateSvc.GetFeatured(r.Context(), limit)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
	})
}

// GetCategories returns all template categories
func (h *TemplateHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.templateSvc.GetCategories(r.Context())
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

// GetByCategory returns templates in a category
func (h *TemplateHandler) GetByCategory(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	opts := parseListOptions(r)

	templates, total, err := h.templateSvc.GetByCategory(r.Context(), category, opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"total":     total,
		"category":  category,
	})
}

// Search searches templates
func (h *TemplateHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "search query required")
		return
	}

	opts := parseListOptions(r)
	templates, total, err := h.templateSvc.Search(r.Context(), query, opts)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"total":     total,
		"query":     query,
	})
}

// Get returns a template by ID
func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	templateIDStr := chi.URLParam(r, "templateID")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid template ID")
		return
	}

	template, err := h.templateSvc.GetByID(r.Context(), templateID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusNotFound, "template not found")
		return
	}

	dto.JSON(w, http.StatusOK, template)
}

// UseTemplateRequest represents a request to use a template
type UseTemplateRequest struct {
	Name      string      `json:"name"`
	Variables models.JSON `json:"variables,omitempty"`
}

// UseTemplate creates a workflow from a template
func (h *TemplateHandler) UseTemplate(w http.ResponseWriter, r *http.Request) {
	templateIDStr := chi.URLParam(r, "templateID")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid template ID")
		return
	}

	var req UseTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		dto.ErrorResponse(w, http.StatusBadRequest, "workflow name required")
		return
	}

	userCtx := middleware.GetUserFromContext(r.Context())
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if userCtx == nil || wsCtx == nil {
		dto.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	workflow, err := h.templateSvc.CreateWorkflowFromTemplate(r.Context(), services.CreateFromTemplateInput{
		TemplateID:  templateID,
		WorkspaceID: wsCtx.WorkspaceID,
		UserID:      userCtx.UserID,
		Name:        req.Name,
		Variables:   req.Variables,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "Workflow created from template",
		"workflow": workflow,
	})
}

func parseListOptions(r *http.Request) *repositories.ListOptions {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	return &repositories.ListOptions{
		Offset: (page - 1) * limit,
		Limit:  limit,
	}
}
