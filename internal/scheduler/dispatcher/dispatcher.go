package dispatcher

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/linkflow-ai/linkflow/internal/scheduler/store"
	"github.com/rs/zerolog/log"
)

type Dispatcher struct {
	queue         *queue.Client
	globalLimiter RateLimiter
	wsLimiter     RateLimiter

	// Metrics
	dispatched atomic.Int64
	skipped    atomic.Int64
	failed     atomic.Int64
}

func NewDispatcher(queueClient *queue.Client, globalLimiter, wsLimiter RateLimiter) *Dispatcher {
	return &Dispatcher{
		queue:         queueClient,
		globalLimiter: globalLimiter,
		wsLimiter:     wsLimiter,
	}
}

type DispatchResult struct {
	ScheduleID string
	Success    bool
	Skipped    bool
	Error      error
}

func (d *Dispatcher) Dispatch(ctx context.Context, schedule *store.Schedule) *DispatchResult {
	result := &DispatchResult{ScheduleID: schedule.ID.String()}

	// Check global rate limit
	if !d.globalLimiter.Allow(ctx, "global") {
		result.Skipped = true
		d.skipped.Add(1)
		return result
	}

	// Check workspace rate limit
	wsKey := fmt.Sprintf("workspace:%s", schedule.WorkspaceID)
	if !d.wsLimiter.Allow(ctx, wsKey) {
		result.Skipped = true
		d.skipped.Add(1)
		return result
	}

	// Enqueue to worker
	payload := queue.WorkflowExecutionPayload{
		WorkflowID:  schedule.WorkflowID,
		WorkspaceID: schedule.WorkspaceID,
		TriggerType: models.TriggerSchedule,
		InputData:   schedule.InputData,
		TriggerData: models.JSON{
			"schedule_id":   schedule.ID.String(),
			"schedule_name": schedule.Name,
			"scheduled_at":  time.Now().Format(time.RFC3339),
		},
	}

	_, err := d.queue.EnqueueWorkflowExecution(ctx, payload)
	if err != nil {
		result.Error = err
		d.failed.Add(1)
		log.Error().
			Err(err).
			Str("schedule_id", schedule.ID.String()).
			Str("workflow_id", schedule.WorkflowID.String()).
			Msg("Failed to enqueue schedule")
		return result
	}

	result.Success = true
	d.dispatched.Add(1)

	log.Debug().
		Str("schedule_id", schedule.ID.String()).
		Str("workflow_id", schedule.WorkflowID.String()).
		Msg("Schedule dispatched")

	return result
}

func (d *Dispatcher) DispatchBatch(ctx context.Context, schedules []*store.Schedule) []*DispatchResult {
	results := make([]*DispatchResult, len(schedules))

	for i, schedule := range schedules {
		results[i] = d.Dispatch(ctx, schedule)
	}

	return results
}

type Stats struct {
	Dispatched int64
	Skipped    int64
	Failed     int64
}

func (d *Dispatcher) Stats() Stats {
	return Stats{
		Dispatched: d.dispatched.Load(),
		Skipped:    d.skipped.Load(),
		Failed:     d.failed.Load(),
	}
}

func (d *Dispatcher) ResetStats() {
	d.dispatched.Store(0)
	d.skipped.Store(0)
	d.failed.Store(0)
}
