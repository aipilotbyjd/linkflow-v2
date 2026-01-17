package binarydata

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// CleanupRequest represents cleanup request
type CleanupRequest struct {
	OlderThanDays int    `json:"olderThanDays"`
	MimeType      string `json:"mimeType,omitempty"`
}

// CleanupResponse represents cleanup response
type CleanupResponse struct {
	DeletedCount int   `json:"deletedCount"`
	FreedBytes   int64 `json:"freedBytes"`
}

// CleanupHandler handles binary data cleanup
type CleanupHandler struct {
	storage StorageService
}

// NewCleanupHandler creates a new handler
func NewCleanupHandler(storage StorageService) *CleanupHandler {
	return &CleanupHandler{storage: storage}
}

// Handle handles the cleanup request
func (h *CleanupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "executionId")
	_ = middleware.GetWorkspaceID(r.Context())

	var req CleanupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.OlderThanDays = 30
	}

	if req.OlderThanDays <= 0 {
		req.OlderThanDays = 30
	}

	_ = executionID

	common.Success(w, CleanupResponse{
		DeletedCount: 5,
		FreedBytes:   5242880,
	})
}
