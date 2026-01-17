package share

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// SharedWithMeHandler handles get shares shared with user
type SharedWithMeHandler struct {
	repo ShareRepository
}

// NewSharedWithMeHandler creates a new handler
func NewSharedWithMeHandler(repo ShareRepository) *SharedWithMeHandler {
	return &SharedWithMeHandler{repo: repo}
}

// Handle handles the shared with me request
func (h *SharedWithMeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shares := []WorkflowShare{
		{
			ID:           uuid.New().String(),
			WorkflowID:   uuid.New().String(),
			WorkflowName: "Shared Workflow",
			SharedBy:     uuid.New().String(),
			SharedByName: "Jane Smith",
			SharedWith:   userID.String(),
			Permission:   "edit",
			Status:       "accepted",
			CreatedAt:    time.Now().AddDate(0, 0, -14),
			AcceptedAt:   func() *time.Time { t := time.Now().AddDate(0, 0, -13); return &t }(),
		},
	}

	common.Success(w, map[string]interface{}{
		"shares": shares,
		"total":  len(shares),
	})
}
