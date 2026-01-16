package execution

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Repository defines the interface for execution persistence
type Repository interface {
	Create(ctx context.Context, execution *Execution) error
	Update(ctx context.Context, execution *Execution) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Execution, error)
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID, opts *ListOptions) ([]Execution, int64, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]Execution, int64, error)
	FindByBatchID(ctx context.Context, batchID uuid.UUID) ([]Execution, error)
	FindRunning(ctx context.Context) ([]Execution, error)
	FindStale(ctx context.Context, timeout time.Duration) ([]Execution, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	CountByWorkflowID(ctx context.Context, workflowID uuid.UUID) (int64, error)
	CountByStatus(ctx context.Context, workspaceID uuid.UUID, status Status) (int64, error)
	DeleteOlderThan(ctx context.Context, workspaceID uuid.UUID, before time.Time) (int64, error)
	BulkDelete(ctx context.Context, ids []uuid.UUID) (int64, error)
}

// NodeExecutionRepository defines the interface for node execution persistence
type NodeExecutionRepository interface {
	Create(ctx context.Context, nodeExec *NodeExecution) error
	Update(ctx context.Context, nodeExec *NodeExecution) error
	FindByID(ctx context.Context, id uuid.UUID) (*NodeExecution, error)
	FindByExecutionID(ctx context.Context, executionID uuid.UUID) ([]NodeExecution, error)
	FindByExecutionAndNodeID(ctx context.Context, executionID uuid.UUID, nodeID string) (*NodeExecution, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status NodeStatus) error
	DeleteByExecutionID(ctx context.Context, executionID uuid.UUID) error
}

// LogRepository defines the interface for execution log persistence
type LogRepository interface {
	Create(ctx context.Context, log *Log) error
	CreateBatch(ctx context.Context, logs []Log) error
	FindByExecutionID(ctx context.Context, executionID uuid.UUID, opts *types.ListOptions) ([]Log, int64, error)
	FindByNodeID(ctx context.Context, executionID uuid.UUID, nodeID string) ([]Log, error)
	DeleteByExecutionID(ctx context.Context, executionID uuid.UUID) error
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

// ListOptions for execution queries
type ListOptions struct {
	*types.ListOptions
	Status      *Status
	TriggerType *string
	WorkflowID  *uuid.UUID
	TriggeredBy *uuid.UUID
	DateRange   *types.DateRange
	Search      string
}

// NewListOptions creates default list options
func NewListOptions(page, perPage int) *ListOptions {
	return &ListOptions{
		ListOptions: types.NewListOptions(page, perPage),
	}
}

// Stats represents execution statistics
type Stats struct {
	Total       int64         `json:"total"`
	Completed   int64         `json:"completed"`
	Failed      int64         `json:"failed"`
	Cancelled   int64         `json:"cancelled"`
	Running     int64         `json:"running"`
	Queued      int64         `json:"queued"`
	AvgDuration time.Duration `json:"avg_duration_ms"`
	SuccessRate float64       `json:"success_rate"`
}

// StatsRepository provides execution statistics
type StatsRepository interface {
	GetStats(ctx context.Context, workspaceID uuid.UUID, from, to time.Time) (*Stats, error)
	GetWorkflowStats(ctx context.Context, workflowID uuid.UUID, from, to time.Time) (*Stats, error)
	GetDailyStats(ctx context.Context, workspaceID uuid.UUID, days int) ([]DailyStat, error)
}

// DailyStat represents daily execution statistics
type DailyStat struct {
	Date      time.Time `json:"date"`
	Total     int64     `json:"total"`
	Completed int64     `json:"completed"`
	Failed    int64     `json:"failed"`
}
