package share

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// PendingHandler handles get pending shares
type PendingHandler struct {
	repo ShareRepository
}

// NewPendingHandler creates a new handler
func NewPendingHandler(repo ShareRepository) *PendingHandler {
	return &PendingHandler{repo: repo}
}

// Handle handles the pending shares request
func (h *PendingHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shares := []WorkflowShare{
		{
			ID:           uuid.New().String(),
			WorkflowID:   uuid.New().String(),
			WorkflowName: "Pending Workflow",
			SharedBy:     uuid.New().String(),
			SharedByName: "Bob Wilson",
			SharedWith:   userID.String(),
			Permission:   "view",
			Status:       "pending",
			CreatedAt:    time.Now().AddDate(0, 0, -1),
		},
	}

	common.Success(w, map[string]interface{}{
		"shares": shares,
		"total":  len(shares),
	})
}
