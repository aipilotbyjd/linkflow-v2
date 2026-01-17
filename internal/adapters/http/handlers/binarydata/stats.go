package binarydata

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// StatsHandler handles get binary data stats request
type StatsHandler struct {
	storage StorageService
}

// NewStatsHandler creates a new handler
func NewStatsHandler(storage StorageService) *StatsHandler {
	return &StatsHandler{storage: storage}
}

// Handle handles the get stats request
func (h *StatsHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
