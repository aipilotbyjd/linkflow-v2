package marketplace

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// GetMyRatingHandler handles get my rating request
type GetMyRatingHandler struct{}

// NewGetMyRatingHandler creates a new handler
func NewGetMyRatingHandler() *GetMyRatingHandler {
	return &GetMyRatingHandler{}
}

// Handle handles the get my rating request
func (h *GetMyRatingHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
