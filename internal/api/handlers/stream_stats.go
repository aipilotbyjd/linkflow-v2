package handlers

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/pkg/streams"
)

// StreamStatsHandler handles webhook stream monitoring endpoints
type StreamStatsHandler struct {
	webhookStream *streams.WebhookStream
}

// NewStreamStatsHandler creates a new stream stats handler
func NewStreamStatsHandler(webhookStream *streams.WebhookStream) *StreamStatsHandler {
	return &StreamStatsHandler{webhookStream: webhookStream}
}

// GetStats returns webhook stream statistics
// GET /api/v1/admin/streams/webhooks/stats
func (h *StreamStatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if h.webhookStream == nil {
		dto.ErrorResponse(w, http.StatusServiceUnavailable, "webhook streaming not enabled")
		return
	}

	stats, err := h.webhookStream.GetStats(r.Context())
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to get stats: "+err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"enabled": true,
		"stats":   stats,
	})
}

// ReplayDLQ replays messages from the dead letter queue
// POST /api/v1/admin/streams/webhooks/replay?count=10
func (h *StreamStatsHandler) ReplayDLQ(w http.ResponseWriter, r *http.Request) {
	if h.webhookStream == nil {
		dto.ErrorResponse(w, http.StatusServiceUnavailable, "webhook streaming not enabled")
		return
	}

	count := int64(10) // default
	if c := r.URL.Query().Get("count"); c != "" {
		if parsed, err := strconv.ParseInt(c, 10, 64); err == nil {
			count = parsed
		}
	}

	replayed, err := h.webhookStream.ReplayFromDLQ(r.Context(), count)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to replay: "+err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"replayed": replayed,
		"message":  "Messages replayed from dead letter queue",
	})
}

// Trim removes old messages from the stream
// POST /api/v1/admin/streams/webhooks/trim?maxlen=50000
func (h *StreamStatsHandler) Trim(w http.ResponseWriter, r *http.Request) {
	if h.webhookStream == nil {
		dto.ErrorResponse(w, http.StatusServiceUnavailable, "webhook streaming not enabled")
		return
	}

	maxLen := int64(50000) // default
	if m := r.URL.Query().Get("maxlen"); m != "" {
		if parsed, err := strconv.ParseInt(m, 10, 64); err == nil {
			maxLen = parsed
		}
	}

	trimmed, err := h.webhookStream.Trim(r.Context(), maxLen)
	if err != nil {
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to trim: "+err.Error())
		return
	}

	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"trimmed": trimmed,
		"maxLen":  maxLen,
	})
}
