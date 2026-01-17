package marketplace

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// RatingStatsHandler handles get rating stats request
type RatingStatsHandler struct{}

// NewRatingStatsHandler creates a new handler
func NewRatingStatsHandler() *RatingStatsHandler {
	return &RatingStatsHandler{}
}

// Handle handles the get rating stats request
func (h *RatingStatsHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
