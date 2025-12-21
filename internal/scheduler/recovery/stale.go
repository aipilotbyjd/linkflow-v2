package recovery

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/scheduler/cron"
	"github.com/linkflow-ai/linkflow/internal/scheduler/store"
	"github.com/rs/zerolog/log"
)

type StaleRecovery struct {
	store      store.ScheduleStore
	calculator *cron.Calculator
	threshold  time.Duration
	interval   time.Duration
}

func NewStaleRecovery(
	scheduleStore store.ScheduleStore,
	calculator *cron.Calculator,
	threshold time.Duration,
) *StaleRecovery {
	return &StaleRecovery{
		store:      scheduleStore,
		calculator: calculator,
		threshold:  threshold,
		interval:   5 * time.Minute,
	}
}

func (r *StaleRecovery) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	// Run once on start
	r.recover(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.recover(ctx)
		}
	}
}

func (r *StaleRecovery) recover(ctx context.Context) {
	stale, err := r.store.GetStale(ctx, r.threshold)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch stale schedules")
		return
	}

	if len(stale) == 0 {
		return
	}

	recovered := 0
	for _, schedule := range stale {
		// Calculate new next_run_at
		nextRun, err := r.calculator.NextRun(schedule.ID, schedule.CronExpression, schedule.Timezone)
		if err != nil {
			log.Error().
				Err(err).
				Str("schedule_id", schedule.ID.String()).
				Msg("Failed to calculate next run for stale schedule")
			continue
		}

		// Update the schedule
		if err := r.store.UpdateNextRun(ctx, schedule.ID, nextRun); err != nil {
			log.Error().
				Err(err).
				Str("schedule_id", schedule.ID.String()).
				Msg("Failed to update stale schedule")
			continue
		}

		recovered++
		log.Warn().
			Str("schedule_id", schedule.ID.String()).
			Time("old_next_run", *schedule.NextRunAt).
			Time("new_next_run", nextRun).
			Msg("Recovered stale schedule")
	}

	if recovered > 0 {
		log.Info().Int("count", recovered).Msg("Recovered stale schedules")
	}
}

func (r *StaleRecovery) RecoverOnce(ctx context.Context) {
	r.recover(ctx)
}
