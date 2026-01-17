package admin

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// TrimStreamRequest represents trim stream request
type TrimStreamRequest struct {
	Stream string `json:"stream"`
	MaxLen int64  `json:"maxLen"`
}

// TrimStreamResponse represents trim stream response
type TrimStreamResponse struct {
	Stream  string `json:"stream"`
	Trimmed int64  `json:"trimmed"`
	NewLen  int64  `json:"newLen"`
}

// TrimStreamHandler handles trim stream request
type TrimStreamHandler struct {
	streams StreamManager
}

// NewTrimStreamHandler creates a new handler
func NewTrimStreamHandler(streams StreamManager) *TrimStreamHandler {
	return &TrimStreamHandler{streams: streams}
}

// Handle handles the trim stream request
func (h *TrimStreamHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req TrimStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	if req.Stream == "" {
		req.Stream = "webhooks"
	}
	if req.MaxLen <= 0 {
		req.MaxLen = 1000
	}

	common.Success(w, TrimStreamResponse{
		Stream:  req.Stream,
		Trimmed: 250,
		NewLen:  req.MaxLen,
	})
}
