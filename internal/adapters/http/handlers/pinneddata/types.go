package pinneddata

import (
	"encoding/json"
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/pinneddata"
)

// PinnedDataResponse represents pinned data in API response
type PinnedDataResponse struct {
	ID         string          `json:"id"`
	WorkflowID string          `json:"workflowId"`
	NodeID     string          `json:"nodeId"`
	Data       json.RawMessage `json:"data"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// SetPinnedDataRequest represents the request to set pinned data
type SetPinnedDataRequest struct {
	Data json.RawMessage `json:"data" validate:"required"`
}

// ToPinnedDataResponse converts domain to response
func ToPinnedDataResponse(p pinneddata.PinnedData) PinnedDataResponse {
	return PinnedDataResponse{
		ID:         p.ID.String(),
		WorkflowID: p.WorkflowID.String(),
		NodeID:     p.NodeID,
		Data:       p.Data,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

// ToPinnedDataResponseList converts domain list to response list
func ToPinnedDataResponseList(list []pinneddata.PinnedData) []PinnedDataResponse {
	result := make([]PinnedDataResponse, len(list))
	for i, p := range list {
		result[i] = ToPinnedDataResponse(p)
	}
	return result
}
