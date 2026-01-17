package pinneddata

import (
	"encoding/json"
	"time"
)

// PinnedData represents pinned data for a node
type PinnedData struct {
	ID         string          `json:"id"`
	WorkflowID string          `json:"workflowId"`
	NodeID     string          `json:"nodeId"`
	Data       json.RawMessage `json:"data"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// PinnedDataRepository defines the pinned data repository interface
type PinnedDataRepository interface {
	GetByWorkflow(workflowID string) ([]PinnedData, error)
	GetByNode(workflowID, nodeID string) (*PinnedData, error)
	Set(workflowID, nodeID string, data json.RawMessage) (*PinnedData, error)
	Delete(workflowID, nodeID string) error
}
