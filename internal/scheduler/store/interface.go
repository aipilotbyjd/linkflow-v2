package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	WorkspaceID    uuid.UUID
	Name           string
	CronExpression string
	Timezone       string
	Priority       string
	InputData      map[string]interface{}
	NextRunAt      *time.Time
	LastRunAt      *time.Time
	RunCount       int
	IsActive       bool
}

type ScheduleStore interface {
	// GetDue fetches schedules that are due for execution
	GetDue(ctx context.Context, limit int) ([]*Schedule, error)

	// GetDueByPriority fetches due schedules filtered by priority
	GetDueByPriority(ctx context.Context, priority string, limit int) ([]*Schedule, error)

	// GetDueByWorkspace fetches due schedules for a specific workspace
	GetDueByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*Schedule, error)

	// UpdateNextRun updates the next run time for a schedule
	UpdateNextRun(ctx context.Context, id uuid.UUID, nextRun time.Time) error

	// RecordRun records that a schedule was executed
	RecordRun(ctx context.Context, id uuid.UUID, nextRun time.Time) error

	// GetStale fetches schedules that appear stuck
	GetStale(ctx context.Context, threshold time.Duration) ([]*Schedule, error)

	// GetByID fetches a single schedule
	GetByID(ctx context.Context, id uuid.UUID) (*Schedule, error)
}
