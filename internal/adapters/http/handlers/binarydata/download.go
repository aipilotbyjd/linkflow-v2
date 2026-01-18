package binarydata

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/binarydata"
	"gorm.io/gorm"
)

// DownloadHandler handles file download request
type DownloadHandler struct {
	repo    binarydata.Repository
	storage binarydata.StorageService
}

// NewDownloadHandler creates a new handler
func NewDownloadHandler(repo binarydata.Repository, storage binarydata.StorageService) *DownloadHandler {
	return &DownloadHandler{repo: repo, storage: storage}
}

// Handle handles the file download request
func (h *DownloadHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	reader, err := h.storage.Download(r.Context(), data.StoragePath)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", data.MimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+data.FileName+"\"")
	if _, err := io.Copy(w, reader); err != nil {
		// Can't return error after headers sent
		return
	}
}
