package binarydata

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// ListHandler handles list binary data request
type ListHandler struct {
	storage StorageService
}

// NewListHandler creates a new handler
func NewListHandler(storage StorageService) *ListHandler {
	return &ListHandler{storage: storage}
}

// Handle handles the list request
func (h *ListHandler) Handle(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "executionId")

	files := []BinaryData{
		{
			ID:          uuid.New().String(),
			ExecutionID: executionID,
			NodeID:      "node-1",
			FileName:    "output.json",
			MimeType:    "application/json",
			Size:        1024,
			CreatedAt:   time.Now().Add(-time.Hour),
		},
		{
			ID:          uuid.New().String(),
			ExecutionID: executionID,
			NodeID:      "node-2",
			FileName:    "image.png",
			MimeType:    "image/png",
			Size:        52428,
			CreatedAt:   time.Now().Add(-30 * time.Minute),
		},
	}

	common.Success(w, map[string]interface{}{
		"files":       files,
		"executionId": executionID,
		"total":       len(files),
	})
}
