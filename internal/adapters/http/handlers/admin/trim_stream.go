package admin

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

// TrimStreamRequest represents the trim stream request
type TrimStreamRequest struct {
	StreamName string `json:"streamName" validate:"required"`
	MaxLen     int64  `json:"maxLen"`
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

	if req.MaxLen <= 0 {
		req.MaxLen = 10000 // Default
	}

	trimmed, err := h.streams.TrimStream(req.StreamName, req.MaxLen)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"trimmed": trimmed,
		"message": "Stream trimmed successfully",
	})
}
