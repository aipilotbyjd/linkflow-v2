package binarydata

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// UploadResponse represents upload response
type UploadResponse struct {
	ID          string    `json:"id"`
	ExecutionID string    `json:"executionId"`
	NodeID      string    `json:"nodeId"`
	FileName    string    `json:"fileName"`
	MimeType    string    `json:"mimeType"`
	Size        int64     `json:"size"`
	StoragePath string    `json:"storagePath"`
	CreatedAt   time.Time `json:"createdAt"`
}

// UploadHandler handles binary data upload
type UploadHandler struct {
	storage StorageService
}

// NewUploadHandler creates a new handler
func NewUploadHandler(storage StorageService) *UploadHandler {
	return &UploadHandler{storage: storage}
}

// Handle handles the upload request
func (h *UploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "executionId")
	_ = middleware.GetWorkspaceID(r.Context())

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		common.BadRequest(w, "could not parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		common.BadRequest(w, "file is required")
		return
	}
	defer file.Close()

	nodeID := r.FormValue("nodeId")
	if nodeID == "" {
		nodeID = "unknown"
	}

	binaryID := uuid.New().String()

	storagePath := "uploads/" + binaryID + "/" + header.Filename
	mimeType := header.Header.Get("Content-Type")

	// Store the file
	if h.storage != nil {
		if err := h.storage.Store(r.Context(), storagePath, file, mimeType); err != nil {
			common.HandleError(w, err)
			return
		}
	}

	// Create metadata record
	binaryData := &BinaryData{
		ID:          binaryID,
		ExecutionID: executionID,
		NodeID:      nodeID,
		FileName:    header.Filename,
		MimeType:    mimeType,
		Size:        header.Size,
		StoragePath: storagePath,
		CreatedAt:   time.Now(),
	}

	// TODO: Save metadata to repository

	common.Created(w, UploadResponse{
		ID:          binaryData.ID,
		ExecutionID: binaryData.ExecutionID,
		NodeID:      binaryData.NodeID,
		FileName:    binaryData.FileName,
		MimeType:    binaryData.MimeType,
		Size:        binaryData.Size,
		StoragePath: binaryData.StoragePath,
		CreatedAt:   binaryData.CreatedAt,
	})
}
