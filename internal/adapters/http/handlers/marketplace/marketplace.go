package marketplace

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

type MarketplaceTemplate struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	LongDescription string   `json:"longDescription,omitempty"`
	Category       string    `json:"category"`
	Tags           []string  `json:"tags"`
	Author         Author    `json:"author"`
	Version        string    `json:"version"`
	Downloads      int       `json:"downloads"`
	Rating         float64   `json:"rating"`
	RatingCount    int       `json:"ratingCount"`
	Featured       bool      `json:"featured"`
	Verified       bool      `json:"verified"`
	Price          float64   `json:"price"`
	Currency       string    `json:"currency"`
	Screenshots    []string  `json:"screenshots,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Author struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar,omitempty"`
	Verified bool   `json:"verified"`
}

type Rating struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Score     int       `json:"score"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type RatingStats struct {
	Average     float64         `json:"average"`
	Total       int             `json:"total"`
	Distribution map[int]int    `json:"distribution"`
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Browse(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	category := r.URL.Query().Get("category")

	if limit <= 0 {
		limit = 20
	}

	templates := getMarketplaceTemplates()
	if category != "" {
		filtered := []MarketplaceTemplate{}
		for _, t := range templates {
			if t.Category == category {
				filtered = append(filtered, t)
			}
		}
		templates = filtered
	}

	common.Success(w, map[string]interface{}{
		"templates": templates,
		"total":     len(templates),
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *Handler) Featured(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 6
	}

	templates := getMarketplaceTemplates()
	featured := []MarketplaceTemplate{}
	for _, t := range templates {
		if t.Featured && len(featured) < limit {
			featured = append(featured, t)
		}
	}

	common.Success(w, map[string]interface{}{
		"templates": featured,
	})
}

func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	categories := []map[string]interface{}{
		{"id": "automation", "name": "Automation", "count": 45, "icon": "robot"},
		{"id": "integration", "name": "Integration", "count": 78, "icon": "link"},
		{"id": "data", "name": "Data Processing", "count": 32, "icon": "database"},
		{"id": "communication", "name": "Communication", "count": 56, "icon": "message"},
		{"id": "crm", "name": "CRM & Sales", "count": 29, "icon": "users"},
		{"id": "devops", "name": "DevOps", "count": 41, "icon": "code"},
		{"id": "ai", "name": "AI & ML", "count": 23, "icon": "brain"},
	}

	common.Success(w, map[string]interface{}{
		"categories": categories,
	})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	templates := getMarketplaceTemplates()

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

	templates := getMarketplaceTemplates()
	for _, t := range templates {
		if t.ID == templateID {
			t.LongDescription = "This is a detailed description of the template with usage instructions and examples."
			t.Screenshots = []string{"/screenshots/1.png", "/screenshots/2.png"}
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

func (h *Handler) Use(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req UseTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Name == "" {
		req.Name = "Workflow from marketplace"
	}

	_ = templateID
	_ = workspaceID

	common.Success(w, map[string]interface{}{
		"workflowId": uuid.New().String(),
		"name":       req.Name,
	})
}

type PublishRequest struct {
	WorkflowID  string   `json:"workflowId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Price       float64  `json:"price"`
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.WorkflowID == "" || req.Name == "" {
		common.Error(w, http.StatusBadRequest, "MISSING_FIELDS", "Workflow ID and name are required")
		return
	}

	template := MarketplaceTemplate{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Author: Author{
			ID:       userID.String(),
			Name:     "Current User",
			Verified: false,
		},
		Version:     "1.0.0",
		Downloads:   0,
		Rating:      0,
		RatingCount: 0,
		Featured:    false,
		Verified:    false,
		Price:       req.Price,
		Currency:    "USD",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	common.JSON(w, http.StatusCreated, template)
}

func (h *Handler) MyPublished(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	templates := []MarketplaceTemplate{
		{
			ID:          uuid.New().String(),
			Name:        "My Published Template",
			Description: "A template I published",
			Category:    "automation",
			Author: Author{
				ID:   userID.String(),
				Name: "You",
			},
			Version:     "1.0.0",
			Downloads:   15,
			Rating:      4.5,
			RatingCount: 3,
			CreatedAt:   time.Now().AddDate(0, -1, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -7),
		},
	}

	common.Success(w, map[string]interface{}{
		"templates": templates,
		"total":     len(templates),
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	template := MarketplaceTemplate{
		ID:          templateID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		Price:       req.Price,
		UpdatedAt:   time.Now(),
	}

	common.Success(w, template)
}

func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")

	common.Success(w, map[string]interface{}{
		"templateId": templateID,
		"version":    "1.1.0",
		"synced":     true,
		"syncedAt":   time.Now(),
	})
}

func (h *Handler) Unpublish(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	_ = templateID

	w.WriteHeader(http.StatusNoContent)
}

type RateRequest struct {
	Score   int    `json:"score"`
	Comment string `json:"comment,omitempty"`
}

func (h *Handler) Rate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	userID := middleware.GetUserID(r.Context())

	var req RateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Score < 1 || req.Score > 5 {
		common.Error(w, http.StatusBadRequest, "INVALID_SCORE", "Score must be between 1 and 5")
		return
	}

	rating := Rating{
		ID:        uuid.New().String(),
		UserID:    userID.String(),
		UserName:  "Current User",
		Score:     req.Score,
		Comment:   req.Comment,
		CreatedAt: time.Now(),
	}

	_ = templateID

	common.JSON(w, http.StatusCreated, rating)
}

func (h *Handler) GetMyRating(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	userID := middleware.GetUserID(r.Context())

	rating := Rating{
		ID:        uuid.New().String(),
		UserID:    userID.String(),
		UserName:  "Current User",
		Score:     4,
		Comment:   "Great template!",
		CreatedAt: time.Now().AddDate(0, 0, -7),
	}

	_ = templateID

	common.Success(w, rating)
}

func (h *Handler) ListRatings(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	ratings := []Rating{
		{
			ID:        uuid.New().String(),
			UserID:    uuid.New().String(),
			UserName:  "John Doe",
			Score:     5,
			Comment:   "Excellent template, saved me hours of work!",
			CreatedAt: time.Now().AddDate(0, 0, -3),
		},
		{
			ID:        uuid.New().String(),
			UserID:    uuid.New().String(),
			UserName:  "Jane Smith",
			Score:     4,
			Comment:   "Works well, could use better documentation.",
			CreatedAt: time.Now().AddDate(0, 0, -7),
		},
	}

	_ = templateID

	common.Success(w, map[string]interface{}{
		"ratings": ratings,
		"total":   len(ratings),
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *Handler) RatingStats(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	_ = templateID

	stats := RatingStats{
		Average: 4.5,
		Total:   25,
		Distribution: map[int]int{
			5: 15,
			4: 7,
			3: 2,
			2: 1,
			1: 0,
		},
	}

	common.Success(w, stats)
}

func (h *Handler) DeleteRating(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	_ = templateID

	w.WriteHeader(http.StatusNoContent)
}

func getMarketplaceTemplates() []MarketplaceTemplate {
	return []MarketplaceTemplate{
		{
			ID:          "mkt-1",
			Name:        "Advanced Slack Integration",
			Description: "Full-featured Slack integration with channels, DMs, and reactions",
			Category:    "communication",
			Tags:        []string{"slack", "messaging", "notifications"},
			Author:      Author{ID: "author-1", Name: "LinkFlow Team", Verified: true},
			Version:     "2.1.0",
			Downloads:   3500,
			Rating:      4.8,
			RatingCount: 127,
			Featured:    true,
			Verified:    true,
			Price:       0,
			Currency:    "USD",
			CreatedAt:   time.Now().AddDate(0, -6, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -14),
		},
		{
			ID:          "mkt-2",
			Name:        "AI Document Processor",
			Description: "Process documents with GPT-4 and extract structured data",
			Category:    "ai",
			Tags:        []string{"ai", "gpt", "documents", "extraction"},
			Author:      Author{ID: "author-2", Name: "AI Labs", Verified: true},
			Version:     "1.5.0",
			Downloads:   2100,
			Rating:      4.7,
			RatingCount: 89,
			Featured:    true,
			Verified:    true,
			Price:       9.99,
			Currency:    "USD",
			CreatedAt:   time.Now().AddDate(0, -4, 0),
			UpdatedAt:   time.Now().AddDate(0, 0, -7),
		},
		{
			ID:          "mkt-3",
			Name:        "E-commerce Order Flow",
			Description: "Complete order processing workflow for Shopify and WooCommerce",
			Category:    "integration",
			Tags:        []string{"ecommerce", "shopify", "woocommerce", "orders"},
			Author:      Author{ID: "author-3", Name: "Commerce Pro", Verified: false},
			Version:     "3.0.0",
			Downloads:   1800,
			Rating:      4.5,
			RatingCount: 63,
			Featured:    true,
			Verified:    false,
			Price:       19.99,
			Currency:    "USD",
			CreatedAt:   time.Now().AddDate(0, -8, 0),
			UpdatedAt:   time.Now().AddDate(0, -1, 0),
		},
	}
}
