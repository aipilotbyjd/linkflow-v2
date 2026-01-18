package share

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the share repository interface
type Repository interface {
	Create(ctx context.Context, share *Share) error
	FindByID(ctx context.Context, id uuid.UUID) (*Share, error)
	Update(ctx context.Context, share *Share) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindSharedByUser(ctx context.Context, userID uuid.UUID) ([]Share, error)
	FindSharedWithUser(ctx context.Context, userID uuid.UUID) ([]Share, error)
	FindPendingForUser(ctx context.Context, userID uuid.UUID) ([]Share, error)
}
