package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// StreamStatsHandler handles stream stats request
type StreamStatsHandler struct {
	streams StreamManager
}

// NewStreamStatsHandler creates a new handler
func NewStreamStatsHandler(streams StreamManager) *StreamStatsHandler {
	return &StreamStatsHandler{streams: streams}
}

// Handle handles the stream stats request
func (h *StreamStatsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	streamName := chi.URLParam(r, "streamName")
	if streamName == "" {
		streamName = r.URL.Query().Get("stream")
	}
	if streamName == "" {
		common.BadRequest(w, "Stream name is required")
		return
	}

	stats, err := h.streams.GetStats(streamName)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, stats)
}
