package marketplace

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// PublishRequest represents publish to marketplace request
type PublishRequest struct {
	WorkflowID  string   `json:"workflowId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Price       float64  `json:"price"`
}

// PublishHandler handles publish to marketplace request
type PublishHandler struct{}

// NewPublishHandler creates a new handler
func NewPublishHandler() *PublishHandler {
	return &PublishHandler{}
}

// Handle handles the publish request
func (h *PublishHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	if req.WorkflowID == "" || req.Name == "" {
		common.BadRequest(w, "workflow ID and name are required")
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

	common.Created(w, template)
}
