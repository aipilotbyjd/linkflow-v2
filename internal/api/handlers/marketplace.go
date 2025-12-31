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
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type MarketplaceHandler struct {
	marketplaceSvc *services.MarketplaceService
	ratingSvc      *services.TemplateRatingService
}

func NewMarketplaceHandler(marketplaceSvc *services.MarketplaceService, ratingSvc *services.TemplateRatingService) *MarketplaceHandler {
	return &MarketplaceHandler{
		marketplaceSvc: marketplaceSvc,
		ratingSvc:      ratingSvc,
	}
}

type PublishTemplateRequest struct {
	WorkflowID  string   `json:"workflow_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Icon        string   `json:"icon"`
	IsPublic    bool     `json:"is_public"`
}

func (h *MarketplaceHandler) Publish(w http.ResponseWriter, r *http.Request) {
	userCtx, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if userCtx == nil {
		return
	}

	var req PublishTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	workflowID, err := uuid.Parse(req.WorkflowID)
	if err != nil {
		dto.BadRequest(w, "invalid workflow_id")
		return
	}

	if req.Name == "" {
		dto.BadRequest(w, "name is required")
		return
	}
	if req.Category == "" {
		dto.BadRequest(w, "category is required")
		return
	}

	template, err := h.marketplaceSvc.Publish(r.Context(), services.PublishTemplateInput{
		WorkflowID:  workflowID,
		WorkspaceID: wsCtx.WorkspaceID,
		PublishedBy: userCtx.UserID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Icon:        req.Icon,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.Created(w, template)
}

func (h *MarketplaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	userCtx, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if userCtx == nil {
		return
	}

	templateID, err := uuid.Parse(chi.URLParam(r, "templateID"))
	if err != nil {
		dto.BadRequest(w, "invalid template ID")
		return
	}

	var req PublishTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	template, err := h.marketplaceSvc.Update(r.Context(), templateID, wsCtx.WorkspaceID, services.PublishTemplateInput{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Icon:        req.Icon,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "template not found")
			return
		}
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, template)
}

func (h *MarketplaceHandler) Sync(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	templateID, err := uuid.Parse(chi.URLParam(r, "templateID"))
	if err != nil {
		dto.BadRequest(w, "invalid template ID")
		return
	}

	template, err := h.marketplaceSvc.SyncFromWorkflow(r.Context(), templateID, wsCtx.WorkspaceID)
	if err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "template not found")
			return
		}
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, template)
}

func (h *MarketplaceHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	templateID, err := uuid.Parse(chi.URLParam(r, "templateID"))
	if err != nil {
		dto.BadRequest(w, "invalid template ID")
		return
	}

	if err := h.marketplaceSvc.Unpublish(r.Context(), templateID, wsCtx.WorkspaceID); err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "template not found")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.NoContent(w)
}

func (h *MarketplaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	templateID, err := uuid.Parse(chi.URLParam(r, "templateID"))
	if err != nil {
		dto.BadRequest(w, "invalid template ID")
		return
	}

	template, err := h.marketplaceSvc.Get(r.Context(), templateID)
	if err != nil {
		dto.NotFound(w, "template not found")
		return
	}

	dto.JSON(w, http.StatusOK, template)
}

func (h *MarketplaceHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	templates, total, err := h.marketplaceSvc.ListPublic(r.Context(), category, limit, offset)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *MarketplaceHandler) ListFeatured(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	templates, err := h.marketplaceSvc.ListFeatured(r.Context(), limit)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

func (h *MarketplaceHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		dto.BadRequest(w, "search query required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	templates, total, err := h.marketplaceSvc.Search(r.Context(), query, limit, offset)
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

func (h *MarketplaceHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.marketplaceSvc.GetCategories(r.Context())
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

func (h *MarketplaceHandler) ListMyPublished(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	templates, err := h.marketplaceSvc.ListByWorkspace(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

type UseTemplateMarketplaceRequest struct {
	Name      string      `json:"name"`
	Variables models.JSON `json:"variables,omitempty"`
}

func (h *MarketplaceHandler) UseTemplate(w http.ResponseWriter, r *http.Request) {
	userCtx, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if userCtx == nil {
		return
	}

	templateID, err := uuid.Parse(chi.URLParam(r, "templateID"))
	if err != nil {
		dto.BadRequest(w, "invalid template ID")
		return
	}

	var req UseTemplateMarketplaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		dto.BadRequest(w, "workflow name required")
		return
	}

	workflow, err := h.marketplaceSvc.UseTemplate(r.Context(), services.UseMarketplaceTemplateInput{
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

	dto.Created(w, workflow)
}

// Rating endpoints

type RateTemplateRequest struct {
	Rating int     `json:"rating"`
	Review *string `json:"review,omitempty"`
}

func (h *MarketplaceHandler) Rate(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.MustUser(w, r)
	if userCtx == nil {
		return
	}

	templateID, err := uuid.Parse(chi.URLParam(r, "templateID"))
	if err != nil {
		dto.BadRequest(w, "invalid template ID")
		return
	}

	var req RateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.BadRequest(w, "invalid request body")
		return
	}

	rating, err := h.ratingSvc.Rate(r.Context(), services.RateTemplateInput{
		TemplateID: templateID,
		UserID:     userCtx.UserID,
		Rating:     req.Rating,
		Review:     req.Review,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, rating)
}

func (h *MarketplaceHandler) DeleteRating(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.MustUser(w, r)
	if userCtx == nil {
		return
	}

	templateID, err := uuid.Parse(chi.URLParam(r, "templateID"))
	if err != nil {
		dto.BadRequest(w, "invalid template ID")
		return
	}

	if err := h.ratingSvc.Delete(r.Context(), templateID, userCtx.UserID); err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "rating not found")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.NoContent(w)
}

func (h *MarketplaceHandler) GetMyRating(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.MustUser(w, r)
	if userCtx == nil {
		return
	}

	templateID, err := uuid.Parse(chi.URLParam(r, "templateID"))
	if err != nil {
		dto.BadRequest(w, "invalid template ID")
		return
	}

	rating, err := h.ratingSvc.GetUserRating(r.Context(), templateID, userCtx.UserID)
	if err != nil {
		dto.NotFound(w, "no rating found")
		return
	}

	dto.JSON(w, http.StatusOK, rating)
}

func (h *MarketplaceHandler) ListRatings(w http.ResponseWriter, r *http.Request) {
	templateID, err := uuid.Parse(chi.URLParam(r, "templateID"))
	if err != nil {
		dto.BadRequest(w, "invalid template ID")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	ratings, total, err := h.ratingSvc.ListByTemplate(r.Context(), templateID, limit, offset)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"ratings": ratings,
		"total":   total,
	})
}

func (h *MarketplaceHandler) GetRatingStats(w http.ResponseWriter, r *http.Request) {
	templateID, err := uuid.Parse(chi.URLParam(r, "templateID"))
	if err != nil {
		dto.BadRequest(w, "invalid template ID")
		return
	}

	stats, err := h.ratingSvc.GetRatingStats(r.Context(), templateID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, stats)
}
