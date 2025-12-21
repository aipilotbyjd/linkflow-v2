package poller

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/linkflow-ai/linkflow/internal/scheduler/cron"
	"github.com/linkflow-ai/linkflow/internal/scheduler/dispatcher"
	"github.com/linkflow-ai/linkflow/internal/scheduler/store"
	"github.com/rs/zerolog/log"
)

type Poller struct {
	store       store.ScheduleStore
	dispatcher  *dispatcher.Dispatcher
	calculator  *cron.Calculator
	backpressure *dispatcher.BackpressureMonitor

	batchSize    int
	pollInterval time.Duration

	// Metrics
	pollCount    atomic.Int64
	lastPollAt   atomic.Value // time.Time
	lastPollDur  atomic.Int64 // milliseconds
}

func NewPoller(
	scheduleStore store.ScheduleStore,
	disp *dispatcher.Dispatcher,
	calc *cron.Calculator,
	batchSize int,
	pollInterval time.Duration,
) *Poller {
	p := &Poller{
		store:        scheduleStore,
		dispatcher:   disp,
		calculator:   calc,
		batchSize:    batchSize,
		pollInterval: pollInterval,
	}
	p.lastPollAt.Store(time.Time{})
	return p
}

func (p *Poller) SetBackpressure(bp *dispatcher.BackpressureMonitor) {
	p.backpressure = bp
}

func (p *Poller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	// Check backpressure
	if p.backpressure != nil && p.backpressure.ShouldPause() {
		log.Debug().Msg("Skipping poll due to backpressure")
		return
	}

	start := time.Now()
	p.pollCount.Add(1)

	// Fetch due schedules
	schedules, err := p.store.GetDue(ctx, p.batchSize)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch due schedules")
		return
	}

	if len(schedules) == 0 {
		p.recordPoll(start)
		return
	}

	// Dispatch each schedule
	dispatched := 0
	skipped := 0

	for _, schedule := range schedules {
		result := p.dispatcher.Dispatch(ctx, schedule)

		if result.Success {
			dispatched++
			// Update next run time
			p.updateNextRun(ctx, schedule)
		} else if result.Skipped {
			skipped++
		}
	}

	p.recordPoll(start)

	if dispatched > 0 || skipped > 0 {
		log.Info().
			Int("dispatched", dispatched).
			Int("skipped", skipped).
			Int("total", len(schedules)).
			Dur("duration", time.Since(start)).
			Msg("Poll completed")
	}
}

func (p *Poller) updateNextRun(ctx context.Context, schedule *store.Schedule) {
	nextRun, err := p.calculator.NextRun(schedule.ID, schedule.CronExpression, schedule.Timezone)
	if err != nil {
		log.Error().
			Err(err).
			Str("schedule_id", schedule.ID.String()).
			Str("cron", schedule.CronExpression).
			Msg("Failed to calculate next run")
		return
	}

	if err := p.store.RecordRun(ctx, schedule.ID, nextRun); err != nil {
		log.Error().
			Err(err).
			Str("schedule_id", schedule.ID.String()).
			Msg("Failed to record run")
	}
}

func (p *Poller) recordPoll(start time.Time) {
	p.lastPollAt.Store(time.Now())
	p.lastPollDur.Store(time.Since(start).Milliseconds())
}

func (p *Poller) PollOnce(ctx context.Context) {
	p.poll(ctx)
}

type Stats struct {
	PollCount      int64
	LastPollAt     time.Time
	LastPollDurMs  int64
}

func (p *Poller) Stats() Stats {
	lastPoll := p.lastPollAt.Load().(time.Time)
	return Stats{
		PollCount:     p.pollCount.Load(),
		LastPollAt:    lastPoll,
		LastPollDurMs: p.lastPollDur.Load(),
	}
}
