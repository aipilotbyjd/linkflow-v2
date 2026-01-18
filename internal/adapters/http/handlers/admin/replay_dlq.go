package admin

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

// ReplayDLQRequest represents the replay DLQ request
type ReplayDLQRequest struct {
	StreamName string `json:"streamName" validate:"required"`
	Count      int    `json:"count"`
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
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	if req.Count <= 0 {
		req.Count = 100 // Default
	}

	replayed, err := h.streams.ReplayDLQ(req.StreamName, req.Count)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"replayed": replayed,
		"message":  "DLQ replay completed successfully",
	})
}
