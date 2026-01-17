package template

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

type Template struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Author      string    `json:"author"`
	Version     string    `json:"version"`
	Downloads   int       `json:"downloads"`
	Rating      float64   `json:"rating"`
	RatingCount int       `json:"ratingCount"`
	Featured    bool      `json:"featured"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TemplateCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Count       int    `json:"count"`
	Icon        string `json:"icon,omitempty"`
}

type TemplateRepository interface {
	FindAll(limit, offset int) ([]Template, int64, error)
	FindByID(id string) (*Template, error)
	FindFeatured(limit int) ([]Template, error)
	FindByCategory(category string, limit, offset int) ([]Template, int64, error)
	Search(query string, limit, offset int) ([]Template, int64, error)
	GetCategories() ([]TemplateCategory, error)
}

type Handler struct {
	repo TemplateRepository
}

func NewHandler(repo TemplateRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	templates := getStaticTemplates()
	total := int64(len(templates))

	if offset < len(templates) {
		end := offset + limit
		if end > len(templates) {
			end = len(templates)
		}
		templates = templates[offset:end]
	} else {
		templates = []Template{}
	}

	common.Success(w, map[string]interface{}{
		"templates": templates,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *Handler) GetFeatured(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 6
	}

	templates := getStaticTemplates()
	featured := []Template{}
	for _, t := range templates {
		if t.Featured && len(featured) < limit {
			featured = append(featured, t)
		}
	}

	common.Success(w, map[string]interface{}{
		"templates": featured,
	})
}

func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories := []TemplateCategory{
		{ID: "automation", Name: "Automation", Description: "Automate repetitive tasks", Count: 25, Icon: "robot"},
		{ID: "integration", Name: "Integration", Description: "Connect different services", Count: 42, Icon: "link"},
		{ID: "data", Name: "Data Processing", Description: "Transform and analyze data", Count: 18, Icon: "database"},
		{ID: "communication", Name: "Communication", Description: "Email, Slack, and messaging", Count: 15, Icon: "message"},
		{ID: "crm", Name: "CRM & Sales", Description: "Sales and customer management", Count: 12, Icon: "users"},
		{ID: "devops", Name: "DevOps", Description: "CI/CD and infrastructure", Count: 8, Icon: "code"},
	}

	common.Success(w, map[string]interface{}{
		"categories": categories,
	})
}

func (h *Handler) GetByCategory(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	templates := getStaticTemplates()
	filtered := []Template{}
	for _, t := range templates {
		if t.Category == category {
			filtered = append(filtered, t)
		}
	}

	total := int64(len(filtered))
	if offset < len(filtered) {
		end := offset + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[offset:end]
	} else {
		filtered = []Template{}
	}

	common.Success(w, map[string]interface{}{
		"templates": filtered,
		"category":  category,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	templates := getStaticTemplates()

	common.Success(w, map[string]interface{}{
		"templates": templates,
		"query":     query,
		"total":     len(templates),
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")

	templates := getStaticTemplates()
	for _, t := range templates {
		if t.ID == templateID {
			common.Success(w, t)
			return
		}
	}

	common.Error(w, http.StatusNotFound, "TEMPLATE_NOT_FOUND", "Template not found")
}

type UseTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	FolderID    string `json:"folderId,omitempty"`
}

type UseTemplateResponse struct {
	WorkflowID string `json:"workflowId"`
	Name       string `json:"name"`
}

func (h *Handler) Use(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req UseTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Name == "" {
		req.Name = "Workflow from template"
	}

	workflowID := uuid.New().String()
	_ = templateID
	_ = workspaceID

	common.Success(w, UseTemplateResponse{
		WorkflowID: workflowID,
		Name:       req.Name,
	})
}

func getStaticTemplates() []Template {
	return []Template{
		{
			ID:          "tpl-1",
			Name:        "Slack to Email Notifier",
			Description: "Forward important Slack messages to email",
			Category:    "communication",
			Tags:        []string{"slack", "email", "notifications"},
			Author:      "LinkFlow",
			Version:     "1.0.0",
			Downloads:   1250,
			Rating:      4.8,
			RatingCount: 45,
			Featured:    true,
			CreatedAt:   time.Now().AddDate(0, -3, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -7),
		},
		{
			ID:          "tpl-2",
			Name:        "GitHub to Slack Reporter",
			Description: "Get Slack notifications for GitHub events",
			Category:    "devops",
			Tags:        []string{"github", "slack", "ci-cd"},
			Author:      "LinkFlow",
			Version:     "1.2.0",
			Downloads:   980,
			Rating:      4.6,
			RatingCount: 32,
			Featured:    true,
			CreatedAt:   time.Now().AddDate(0, -6, 0),
			UpdatedAt:   time.Now().AddDate(0, -1, 0),
		},
		{
			ID:          "tpl-3",
			Name:        "CRM Lead Sync",
			Description: "Sync leads between Salesforce and HubSpot",
			Category:    "crm",
			Tags:        []string{"salesforce", "hubspot", "leads"},
			Author:      "LinkFlow",
			Version:     "2.0.0",
			Downloads:   750,
			Rating:      4.5,
			RatingCount: 28,
			Featured:    true,
			CreatedAt:   time.Now().AddDate(0, -4, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -14),
		},
		{
			ID:          "tpl-4",
			Name:        "Data Backup to S3",
			Description: "Automatically backup data to AWS S3",
			Category:    "data",
			Tags:        []string{"aws", "s3", "backup"},
			Author:      "LinkFlow",
			Version:     "1.1.0",
			Downloads:   620,
			Rating:      4.7,
			RatingCount: 21,
			Featured:    false,
			CreatedAt:   time.Now().AddDate(0, -2, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -3),
		},
	}
}
