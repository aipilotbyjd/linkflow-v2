package binarydata

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

type BinaryData struct {
	ID          string    `json:"id"`
	ExecutionID string    `json:"executionId"`
	NodeID      string    `json:"nodeId"`
	FileName    string    `json:"fileName"`
	MimeType    string    `json:"mimeType"`
	Size        int64     `json:"size"`
	StoragePath string    `json:"storagePath,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

type BinaryDataStats struct {
	TotalFiles     int   `json:"totalFiles"`
	TotalSizeBytes int64 `json:"totalSizeBytes"`
	ByMimeType     map[string]MimeTypeStats `json:"byMimeType"`
}

type MimeTypeStats struct {
	Count int   `json:"count"`
	Size  int64 `json:"size"`
}

type StorageService interface {
	Upload(data io.Reader, fileName, mimeType string) (string, error)
	Download(path string) (io.ReadCloser, error)
	Delete(path string) error
	GetInfo(path string) (*BinaryData, error)
}

type Handler struct {
	storage StorageService
}

func NewHandler(storage StorageService) *Handler {
	return &Handler{storage: storage}
}

type UploadResponse struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "executionId")
	_ = middleware.GetWorkspaceID(r.Context())

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Could not parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		common.Error(w, http.StatusBadRequest, "MISSING_FILE", "File is required")
		return
	}
	defer file.Close()

	nodeID := r.FormValue("nodeId")
	if nodeID == "" {
		nodeID = "unknown"
	}

	binaryID := uuid.New().String()

	binaryData := BinaryData{
		ID:          binaryID,
		ExecutionID: executionID,
		NodeID:      nodeID,
		FileName:    header.Filename,
		MimeType:    header.Header.Get("Content-Type"),
		Size:        header.Size,
		StoragePath: "uploads/" + binaryID + "/" + header.Filename,
		CreatedAt:   time.Now(),
	}

	common.JSON(w, http.StatusCreated, UploadResponse{
		ID:       binaryData.ID,
		FileName: binaryData.FileName,
		MimeType: binaryData.MimeType,
		Size:     binaryData.Size,
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
	binaryID := chi.URLParam(r, "binaryId")

	binaryData := BinaryData{
		ID:          binaryID,
		ExecutionID: uuid.New().String(),
		NodeID:      "node-1",
		FileName:    "output.json",
		MimeType:    "application/json",
		Size:        1024,
		StoragePath: "uploads/" + binaryID + "/output.json",
		CreatedAt:   time.Now().Add(-time.Hour),
	}

	common.Success(w, binaryData)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	binaryID := chi.URLParam(r, "binaryId")

	fileName := "output.json"
	mimeType := "application/json"
	data := []byte(`{"status": "sample data", "binaryId": "` + binaryID + `"}`)

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	binaryID := chi.URLParam(r, "binaryId")
	_ = binaryID

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetWorkspaceID(r.Context())

	stats := BinaryDataStats{
		TotalFiles:     156,
		TotalSizeBytes: 52428800,
		ByMimeType: map[string]MimeTypeStats{
			"application/json": {Count: 85, Size: 10485760},
			"image/png":        {Count: 42, Size: 31457280},
			"text/csv":         {Count: 29, Size: 10485760},
		},
	}

	common.Success(w, stats)
}

type CleanupRequest struct {
	OlderThanDays int    `json:"olderThanDays"`
	MimeType      string `json:"mimeType,omitempty"`
}

type CleanupResponse struct {
	DeletedCount int   `json:"deletedCount"`
	FreedBytes   int64 `json:"freedBytes"`
}

func (h *Handler) Cleanup(w http.ResponseWriter, r *http.Request) {
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
