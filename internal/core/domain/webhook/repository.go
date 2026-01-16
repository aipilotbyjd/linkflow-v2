package webhook

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Repository defines the interface for webhook endpoint persistence
type Repository interface {
	Create(ctx context.Context, endpoint *Endpoint) error
	Update(ctx context.Context, endpoint *Endpoint) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Endpoint, error)
	FindByPath(ctx context.Context, path string) (*Endpoint, error)
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Endpoint, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *types.ListOptions) ([]Endpoint, int64, error)
	FindByWorkflowAndNode(ctx context.Context, workflowID uuid.UUID, nodeID string) (*Endpoint, error)
	ExistsByPath(ctx context.Context, path string) (bool, error)
	SetActive(ctx context.Context, id uuid.UUID, isActive bool) error
	RecordCall(ctx context.Context, id uuid.UUID) error
	CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	DeleteByWorkflowID(ctx context.Context, workflowID uuid.UUID) error
}

// EventRepository defines the interface for webhook event persistence
type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	FindByID(ctx context.Context, id uuid.UUID) (*Event, error)
	FindByEndpointID(ctx context.Context, endpointID uuid.UUID, opts *types.ListOptions) ([]Event, int64, error)
	DeleteByEndpointID(ctx context.Context, endpointID uuid.UUID) error
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
	CountByEndpointID(ctx context.Context, endpointID uuid.UUID) (int64, error)
}

// SignatureVerifier defines the interface for webhook signature verification
type SignatureVerifier interface {
	Sign(payload []byte, secret string) string
	Verify(payload []byte, signature, secret string) bool
}
