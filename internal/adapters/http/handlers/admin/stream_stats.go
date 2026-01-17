package admin

import (
	"net/http"
	"time"

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
	streamName := r.URL.Query().Get("stream")
	if streamName == "" {
		streamName = "webhooks"
	}

	stats := StreamStats{
		Name:          streamName,
		Length:        1250,
		Pending:       5,
		Consumers:     3,
		LastDelivered: time.Now().Add(-time.Minute).Format(time.RFC3339),
	}

	common.Success(w, stats)
}
