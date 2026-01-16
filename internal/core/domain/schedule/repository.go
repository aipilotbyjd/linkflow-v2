package schedule

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Repository defines the interface for schedule persistence
type Repository interface {
	Create(ctx context.Context, schedule *Schedule) error
	Update(ctx context.Context, schedule *Schedule) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Schedule, error)
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]Schedule, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *types.ListOptions) ([]Schedule, int64, error)
	FindDue(ctx context.Context, before time.Time) ([]Schedule, error)
	FindDueSchedules(ctx context.Context, from, to time.Time) ([]*Schedule, error)
	FindActive(ctx context.Context) ([]Schedule, error)
	UpdateNextRunAt(ctx context.Context, id uuid.UUID, nextRunAt time.Time) error
	RecordRun(ctx context.Context, id uuid.UUID, executionID uuid.UUID, nextRunAt time.Time) error
	SetActive(ctx context.Context, id uuid.UUID, isActive bool) error
	CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	CountActiveByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	DeleteByWorkflowID(ctx context.Context, workflowID uuid.UUID) error
}

// CronParser defines the interface for cron expression parsing
type CronParser interface {
	Parse(expression string) error
	Next(from time.Time) time.Time
	NextN(from time.Time, n int) []time.Time
}
