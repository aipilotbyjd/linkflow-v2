package marketplace

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// ListRatingsHandler handles list ratings request
type ListRatingsHandler struct{}

// NewListRatingsHandler creates a new handler
func NewListRatingsHandler() *ListRatingsHandler {
	return &ListRatingsHandler{}
}

// Handle handles the list ratings request
func (h *ListRatingsHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
