package pinneddata

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// Repository defines the pinned data repository interface
type Repository interface {
	GetByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]PinnedData, error)
	GetByNode(ctx context.Context, workflowID uuid.UUID, nodeID string) (*PinnedData, error)
	Set(ctx context.Context, workflowID uuid.UUID, nodeID string, data json.RawMessage) (*PinnedData, error)
	Delete(ctx context.Context, workflowID uuid.UUID, nodeID string) error
}
