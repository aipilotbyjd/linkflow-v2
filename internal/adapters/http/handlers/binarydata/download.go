package binarydata

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// DownloadHandler handles binary data download
type DownloadHandler struct {
	storage StorageService
}

// NewDownloadHandler creates a new handler
func NewDownloadHandler(storage StorageService) *DownloadHandler {
	return &DownloadHandler{storage: storage}
}

// Handle handles the download request
func (h *DownloadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	binaryID := chi.URLParam(r, "binaryId")

	fileName := "output.json"
	mimeType := "application/json"
	data := []byte(`{"status": "sample data", "binaryId": "` + binaryID + `"}`)

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}
