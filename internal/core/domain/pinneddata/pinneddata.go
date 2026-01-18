package pinneddata

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PinnedData represents pinned/test data for a workflow node
type PinnedData struct {
	ID         uuid.UUID       `json:"id"`
	WorkflowID uuid.UUID       `json:"workflowId"`
	NodeID     string          `json:"nodeId"`
	Data       json.RawMessage `json:"data"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// NewPinnedData creates new pinned data
func NewPinnedData(workflowID uuid.UUID, nodeID string, data json.RawMessage) *PinnedData {
	now := time.Now()
	return &PinnedData{
		ID:         uuid.New(),
		WorkflowID: workflowID,
		NodeID:     nodeID,
		Data:       data,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
