package marketplace

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// SyncHandler handles sync marketplace template request
type SyncHandler struct{}

// NewSyncHandler creates a new handler
func NewSyncHandler() *SyncHandler {
	return &SyncHandler{}
}

// Handle handles the sync request
func (h *SyncHandler) Handle(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")

	common.Success(w, map[string]interface{}{
		"templateId": templateID,
		"version":    "1.1.0",
		"synced":     true,
		"syncedAt":   time.Now(),
	})
}
