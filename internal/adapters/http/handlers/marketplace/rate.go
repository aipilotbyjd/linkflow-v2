package marketplace

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// RateRequest represents rate template request
type RateRequest struct {
	Score   int    `json:"score"`
	Comment string `json:"comment,omitempty"`
}

// RateHandler handles rate template request
type RateHandler struct{}

// NewRateHandler creates a new handler
func NewRateHandler() *RateHandler {
	return &RateHandler{}
}

// Handle handles the rate request
func (h *RateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	userID := middleware.GetUserID(r.Context())

	var req RateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	if req.Score < 1 || req.Score > 5 {
		common.BadRequest(w, "score must be between 1 and 5")
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

	common.Created(w, rating)
}
