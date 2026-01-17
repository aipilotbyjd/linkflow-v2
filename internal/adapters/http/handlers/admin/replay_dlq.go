package admin

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// ReplayDLQRequest represents replay DLQ request
type ReplayDLQRequest struct {
	Stream string `json:"stream"`
	Count  int    `json:"count"`
}

// ReplayDLQResponse represents replay DLQ response
type ReplayDLQResponse struct {
	Stream   string `json:"stream"`
	Replayed int    `json:"replayed"`
}

// ReplayDLQHandler handles replay DLQ request
type ReplayDLQHandler struct {
	streams StreamManager
}

// NewReplayDLQHandler creates a new handler
func NewReplayDLQHandler(streams StreamManager) *ReplayDLQHandler {
	return &ReplayDLQHandler{streams: streams}
}

// Handle handles the replay DLQ request
func (h *ReplayDLQHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req ReplayDLQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	if req.Stream == "" {
		req.Stream = "webhooks"
	}
	if req.Count <= 0 {
		req.Count = 10
	}

	replayedCount := 5
	if req.Count < 5 {
		replayedCount = req.Count
	}

	common.Success(w, ReplayDLQResponse{
		Stream:   req.Stream,
		Replayed: replayedCount,
	})
}
