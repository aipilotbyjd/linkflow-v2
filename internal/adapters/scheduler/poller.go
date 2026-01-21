package scheduler

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
)

type Poller struct {
	scheduleRepo schedule.Repository
	pollInterval time.Duration
}

func NewPoller(scheduleRepo schedule.Repository, pollInterval time.Duration) *Poller {
	return &Poller{
		scheduleRepo: scheduleRepo,
		pollInterval: pollInterval,
	}
}

// Poll returns schedules that are due for execution
func (p *Poller) Poll(ctx context.Context) ([]*schedule.Schedule, error) {
	now := time.Now()
	window := now.Add(p.pollInterval)

	// Find schedules where next_run_at is between now and now + poll_interval
	schedules, err := p.scheduleRepo.FindDueSchedules(ctx, now, window)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

// UpdateNextRun calculates and updates the next run time for a schedule
func (p *Poller) UpdateNextRun(ctx context.Context, sched *schedule.Schedule) error {
	if _, err := sched.CalculateNextRun(); err != nil {
		return err
	}
	return p.scheduleRepo.Update(ctx, sched)
}

// MarkExecuted marks a schedule as having been executed
func (p *Poller) MarkExecuted(ctx context.Context, sched *schedule.Schedule) error {
	sched.MarkExecuted()
	return p.scheduleRepo.Update(ctx, sched)
}
