package share

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// AcceptHandler handles accept share request
type AcceptHandler struct {
	repo ShareRepository
}

// NewAcceptHandler creates a new handler
func NewAcceptHandler(repo ShareRepository) *AcceptHandler {
	return &AcceptHandler{repo: repo}
}

// Handle handles the accept share request
func (h *AcceptHandler) Handle(w http.ResponseWriter, r *http.Request) {
	shareID := chi.URLParam(r, "shareId")

	share := WorkflowShare{
		ID:           shareID,
		WorkflowID:   uuid.New().String(),
		WorkflowName: "Accepted Workflow",
		Status:       "accepted",
		AcceptedAt:   func() *time.Time { t := time.Now(); return &t }(),
	}

	common.Success(w, share)
}
