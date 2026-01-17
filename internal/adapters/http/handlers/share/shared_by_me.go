package share

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// SharedByMeHandler handles get shares created by user
type SharedByMeHandler struct {
	repo ShareRepository
}

// NewSharedByMeHandler creates a new handler
func NewSharedByMeHandler(repo ShareRepository) *SharedByMeHandler {
	return &SharedByMeHandler{repo: repo}
}

// Handle handles the shared by me request
func (h *SharedByMeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shares := []WorkflowShare{
		{
			ID:              uuid.New().String(),
			WorkflowID:      uuid.New().String(),
			WorkflowName:    "My Workflow",
			SharedBy:        userID.String(),
			SharedByName:    "You",
			SharedWith:      uuid.New().String(),
			SharedWithName:  "John Doe",
			SharedWithEmail: "john@example.com",
			Permission:      "view",
			Status:          "accepted",
			CreatedAt:       time.Now().AddDate(0, 0, -7),
			AcceptedAt:      func() *time.Time { t := time.Now().AddDate(0, 0, -6); return &t }(),
		},
	}

	common.Success(w, map[string]interface{}{
		"shares": shares,
		"total":  len(shares),
	})
}
