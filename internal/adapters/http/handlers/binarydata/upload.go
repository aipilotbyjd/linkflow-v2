package binarydata

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/binarydata"
)

// UploadHandler handles file upload request
type UploadHandler struct {
	repo    binarydata.Repository
	storage binarydata.StorageService
}

// NewUploadHandler creates a new handler
func NewUploadHandler(repo binarydata.Repository, storage binarydata.StorageService) *UploadHandler {
	return &UploadHandler{repo: repo, storage: storage}
}

// Handle handles the file upload request
func (h *UploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	// Parse multipart form (32 MB max)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		common.BadRequest(w, "Failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		common.BadRequest(w, "No file provided")
		return
	}
	defer file.Close()

	// Upload to storage
	storagePath, err := h.storage.Upload(r.Context(), workspaceID, header.Filename, file, header.Size)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Create metadata
	data := binarydata.NewBinaryData(workspaceID, header.Filename, header.Header.Get("Content-Type"), header.Size, storagePath)
	if err := h.repo.Create(r.Context(), data); err != nil {
		// Clean up uploaded file
		_ = h.storage.Delete(r.Context(), storagePath)
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToBinaryDataResponse(*data))
}
