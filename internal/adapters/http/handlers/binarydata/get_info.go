package binarydata

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// GetInfoHandler handles get binary data info request
type GetInfoHandler struct {
	storage StorageService
}

// NewGetInfoHandler creates a new handler
func NewGetInfoHandler(storage StorageService) *GetInfoHandler {
	return &GetInfoHandler{storage: storage}
}

// Handle handles the get info request
func (h *GetInfoHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
