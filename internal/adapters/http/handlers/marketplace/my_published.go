package marketplace

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// MyPublishedHandler handles get my published templates request
type MyPublishedHandler struct{}

// NewMyPublishedHandler creates a new handler
func NewMyPublishedHandler() *MyPublishedHandler {
	return &MyPublishedHandler{}
}

// Handle handles the get my published templates request
func (h *MyPublishedHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
