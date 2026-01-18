package binarydata

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/binarydata"
	"gorm.io/gorm"
)

// DeleteHandler handles file delete request
type DeleteHandler struct {
	repo    binarydata.Repository
	storage binarydata.StorageService
}

// NewDeleteHandler creates a new handler
func NewDeleteHandler(repo binarydata.Repository, storage binarydata.StorageService) *DeleteHandler {
	return &DeleteHandler{repo: repo, storage: storage}
}

// Handle handles the file delete request
func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "fileId")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		common.BadRequest(w, "Invalid file ID")
		return
	}

	data, err := h.repo.FindByID(r.Context(), fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.NotFound(w, "File")
			return
		}
		common.HandleError(w, err)
		return
	}

	// Delete from storage
	if err := h.storage.Delete(r.Context(), data.StoragePath); err != nil {
		common.HandleError(w, err)
		return
	}

	// Delete metadata
	if err := h.repo.Delete(r.Context(), fileID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"message": "File deleted successfully",
	})
}
