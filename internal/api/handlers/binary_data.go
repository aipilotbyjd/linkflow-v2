package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
)

type BinaryDataHandler struct {
	binarySvc *services.BinaryDataService
}

func NewBinaryDataHandler(binarySvc *services.BinaryDataService) *BinaryDataHandler {
	return &BinaryDataHandler{binarySvc: binarySvc}
}

func (h *BinaryDataHandler) Upload(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	executionID, err := uuid.Parse(chi.URLParam(r, "executionID"))
	if err != nil {
		dto.BadRequest(w, "invalid execution ID")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		dto.BadRequest(w, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		dto.BadRequest(w, "file is required")
		return
	}
	defer file.Close()

	nodeID := r.FormValue("node_id")
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	var expiresAt *time.Time
	if expiresStr := r.FormValue("expires_in_hours"); expiresStr != "" {
		hours, err := strconv.Atoi(expiresStr)
		if err == nil && hours > 0 {
			t := time.Now().Add(time.Duration(hours) * time.Hour)
			expiresAt = &t
		}
	}

	var metadata models.JSON
	if metaStr := r.FormValue("metadata"); metaStr != "" {
		metadata = models.JSON{"custom": metaStr}
	}

	data, err := h.binarySvc.Store(r.Context(), services.StoreBinaryInput{
		ExecutionID: executionID,
		WorkspaceID: wsCtx.WorkspaceID,
		NodeID:      nodeID,
		FileName:    header.Filename,
		MimeType:    mimeType,
		Data:        file,
		Size:        header.Size,
		Metadata:    metadata,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.Created(w, data)
}

func (h *BinaryDataHandler) Download(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	binaryID, err := uuid.Parse(chi.URLParam(r, "binaryID"))
	if err != nil {
		dto.BadRequest(w, "invalid binary ID")
		return
	}

	data, reader, err := h.binarySvc.GetContent(r.Context(), binaryID, wsCtx.WorkspaceID)
	if err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "binary data not found")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", data.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", data.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(data.Size, 10))
	w.Header().Set("X-Checksum-SHA256", data.Checksum)

	io.Copy(w, reader)
}

func (h *BinaryDataHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	binaryID, err := uuid.Parse(chi.URLParam(r, "binaryID"))
	if err != nil {
		dto.BadRequest(w, "invalid binary ID")
		return
	}

	data, err := h.binarySvc.Get(r.Context(), binaryID, wsCtx.WorkspaceID)
	if err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "binary data not found")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, data)
}

func (h *BinaryDataHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	binaryID, err := uuid.Parse(chi.URLParam(r, "binaryID"))
	if err != nil {
		dto.BadRequest(w, "invalid binary ID")
		return
	}

	if err := h.binarySvc.Delete(r.Context(), binaryID, wsCtx.WorkspaceID); err != nil {
		if err == services.ErrNotFound {
			dto.NotFound(w, "binary data not found")
			return
		}
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.NoContent(w)
}

func (h *BinaryDataHandler) ListByExecution(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	executionID, err := uuid.Parse(chi.URLParam(r, "executionID"))
	if err != nil {
		dto.BadRequest(w, "invalid execution ID")
		return
	}

	nodeID := r.URL.Query().Get("node_id")

	var data []models.BinaryData
	if nodeID != "" {
		data, err = h.binarySvc.ListByNode(r.Context(), executionID, nodeID, wsCtx.WorkspaceID)
	} else {
		data, err = h.binarySvc.ListByExecution(r.Context(), executionID, wsCtx.WorkspaceID)
	}

	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"files": data,
		"count": len(data),
	})
}

func (h *BinaryDataHandler) GetStorageStats(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	stats, err := h.binarySvc.GetStorageStats(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, stats)
}

func (h *BinaryDataHandler) CleanupByExecution(w http.ResponseWriter, r *http.Request) {
	_, wsCtx := middleware.MustUserAndWorkspace(w, r)
	if wsCtx == nil {
		return
	}

	executionID, err := uuid.Parse(chi.URLParam(r, "executionID"))
	if err != nil {
		dto.BadRequest(w, "invalid execution ID")
		return
	}

	if err := h.binarySvc.CleanupByExecution(r.Context(), executionID, wsCtx.WorkspaceID); err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]string{"message": "cleanup complete"})
}
