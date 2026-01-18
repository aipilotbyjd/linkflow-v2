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

// GetInfoHandler handles get binary data info request
type GetInfoHandler struct {
	repo binarydata.Repository
}

// NewGetInfoHandler creates a new handler
func NewGetInfoHandler(repo binarydata.Repository) *GetInfoHandler {
	return &GetInfoHandler{repo: repo}
}

// Handle handles the get binary data info request
func (h *GetInfoHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	common.Success(w, ToBinaryDataResponse(*data))
}
