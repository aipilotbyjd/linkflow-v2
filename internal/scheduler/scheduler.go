package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/linkflow-ai/linkflow/internal/scheduler/cron"
	"github.com/linkflow-ai/linkflow/internal/scheduler/dispatcher"
	"github.com/linkflow-ai/linkflow/internal/scheduler/leader"
	"github.com/linkflow-ai/linkflow/internal/scheduler/metrics"
	"github.com/linkflow-ai/linkflow/internal/scheduler/poller"
	"github.com/linkflow-ai/linkflow/internal/scheduler/recovery"
	"github.com/linkflow-ai/linkflow/internal/scheduler/store"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Scheduler struct {
	config *Config

	// Components
	election     *leader.Election
	poller       *poller.Poller
	dispatcher   *dispatcher.Dispatcher
	staleRecov   *recovery.StaleRecovery
	cleanup      *recovery.Cleanup
	backpressure *dispatcher.BackpressureMonitor
	metrics      *metrics.Collector

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type Dependencies struct {
	DB    *gorm.DB
	Redis *pkgredis.Client
	Queue *queue.Client
}

func New(cfg *Config, deps *Dependencies) *Scheduler {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	cfg.Validate()

	ctx, cancel := context.WithCancel(context.Background())

	// Create store
	pgStore := store.NewPostgresStore(deps.DB)
	cachedStore := store.NewCachedStore(pgStore, deps.Redis)

	// Create leader election
	election := leader.NewElection(deps.Redis, cfg.LeaderKey, cfg.LeaderTTL)

	// Create cron calculator
	calculator := cron.NewCalculator()

	// Create rate limiters
	globalLimiter := dispatcher.NewSlidingWindowLimiter(
		deps.Redis, "scheduler:ratelimit:global", cfg.GlobalRateLimit, time.Minute,
	)
	wsLimiter := dispatcher.NewSlidingWindowLimiter(
		deps.Redis, "scheduler:ratelimit:workspace", cfg.WorkspaceLimit, time.Minute,
	)

	// Create dispatcher
	disp := dispatcher.NewDispatcher(deps.Queue, globalLimiter, wsLimiter)

	// Create poller
	poll := poller.NewPoller(cachedStore, disp, calculator, cfg.BatchSize, cfg.PollInterval)

	// Create backpressure monitor
	bp := dispatcher.NewBackpressureMonitor(deps.Redis, "asynq:queue:default", 10000)
	poll.SetBackpressure(bp)

	// Create recovery
	staleRecov := recovery.NewStaleRecovery(cachedStore, calculator, cfg.StaleThreshold)
	cleanup := recovery.NewCleanup(deps.DB, cfg.RetentionDays)

	// Create metrics
	metricsCollector := metrics.NewCollector()

	return &Scheduler{
		config:       cfg,
		election:     election,
		poller:       poll,
		dispatcher:   disp,
		staleRecov:   staleRecov,
		cleanup:      cleanup,
		backpressure: bp,
		metrics:      metricsCollector,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (s *Scheduler) Start() error {
	log.Info().
		Str("leader_key", s.config.LeaderKey).
		Dur("poll_interval", s.config.PollInterval).
		Int("batch_size", s.config.BatchSize).
		Msg("Starting scheduler")

	// Start leader election loop
	s.wg.Add(1)
	go s.leaderLoop()

	// Start backpressure monitor
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.backpressure.Start(s.ctx)
	}()

	return nil
}

func (s *Scheduler) Stop() error {
	log.Info().Msg("Stopping scheduler...")

	s.cancel()

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("Scheduler stopped gracefully")
	case <-time.After(s.config.ShutdownTimeout):
		log.Warn().Msg("Scheduler shutdown timed out")
	}

	// Release leadership
	s.election.Release(context.Background())

	return nil
}

func (s *Scheduler) leaderLoop() {
	defer s.wg.Done()

	extendTicker := time.NewTicker(s.config.LeaderTTL / 3)
	defer extendTicker.Stop()

	acquireTicker := time.NewTicker(5 * time.Second)
	defer acquireTicker.Stop()

	var pollerCancel context.CancelFunc
	var recoveryCancel context.CancelFunc
	var cleanupCancel context.CancelFunc

	stopWorkers := func() {
		if pollerCancel != nil {
			pollerCancel()
			pollerCancel = nil
		}
		if recoveryCancel != nil {
			recoveryCancel()
			recoveryCancel = nil
		}
		if cleanupCancel != nil {
			cleanupCancel()
			cleanupCancel = nil
		}
	}

	startWorkers := func() {
		var pollerCtx, recoveryCtx, cleanupCtx context.Context

		pollerCtx, pollerCancel = context.WithCancel(s.ctx)
		recoveryCtx, recoveryCancel = context.WithCancel(s.ctx)
		cleanupCtx, cleanupCancel = context.WithCancel(s.ctx)

		s.wg.Add(3)
		go func() {
			defer s.wg.Done()
			s.poller.Run(pollerCtx)
		}()
		go func() {
			defer s.wg.Done()
			s.staleRecov.Run(recoveryCtx)
		}()
		go func() {
			defer s.wg.Done()
			s.cleanup.Run(cleanupCtx)
		}()
	}

	for {
		select {
		case <-s.ctx.Done():
			stopWorkers()
			return

		case <-acquireTicker.C:
			if !s.election.IsLeader() {
				acquired, err := s.election.TryAcquire(s.ctx)
				if err != nil {
					log.Error().Err(err).Msg("Failed to acquire leadership")
					continue
				}
				if acquired {
					s.metrics.SetLeader(true)
					startWorkers()
				}
			}

		case <-extendTicker.C:
			if s.election.IsLeader() {
				if !s.election.Extend(s.ctx) {
					log.Warn().Msg("Lost leadership")
					s.metrics.SetLeader(false)
					stopWorkers()
				}
			}
		}
	}
}

func (s *Scheduler) IsLeader() bool {
	return s.election.IsLeader()
}

func (s *Scheduler) Metrics() *metrics.Collector {
	return s.metrics
}

func (s *Scheduler) Health() map[string]interface{} {
	snapshot := s.metrics.Snapshot()
	pollerStats := s.poller.Stats()
	dispatcherStats := s.dispatcher.Stats()

	return map[string]interface{}{
		"is_leader":        snapshot.IsLeader,
		"uptime_seconds":   int64(snapshot.Uptime.Seconds()),
		"polls_total":      pollerStats.PollCount,
		"last_poll_at":     pollerStats.LastPollAt,
		"dispatched_total": dispatcherStats.Dispatched,
		"skipped_total":    dispatcherStats.Skipped,
		"failed_total":     dispatcherStats.Failed,
		"queue_depth":      s.backpressure.QueueDepth(),
	}
}
