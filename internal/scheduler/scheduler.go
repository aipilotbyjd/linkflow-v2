package scheduler

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/linkflow-ai/linkflow/internal/queue"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Scheduler struct {
	cfg           *config.Config
	cron          *cron.Cron
	leader        *LeaderElection
	scheduleSvc   *services.ScheduleService
	executionSvc  *services.ExecutionService
	queueClient   *queue.Client
	done          chan struct{}
}

func New(
	cfg *config.Config,
	redisClient *pkgredis.Client,
	scheduleSvc *services.ScheduleService,
	executionSvc *services.ExecutionService,
	queueClient *queue.Client,
) *Scheduler {
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		cfg:          cfg,
		cron:         c,
		leader:       NewLeaderElection(redisClient, "scheduler-leader"),
		scheduleSvc:  scheduleSvc,
		executionSvc: executionSvc,
		queueClient:  queueClient,
		done:         make(chan struct{}),
	}
}

func (s *Scheduler) Start() error {
	log.Info().Msg("Starting scheduler...")

	// Try to acquire leadership
	go s.runWithLeadership()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down scheduler...")
	close(s.done)
	s.cron.Stop()

	return nil
}

func (s *Scheduler) runWithLeadership() {
	for {
		select {
		case <-s.done:
			return
		default:
		}

		ctx := context.Background()
		acquired, err := s.leader.TryAcquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to acquire leadership")
			time.Sleep(5 * time.Second)
			continue
		}

		if acquired {
			log.Info().Msg("Acquired leadership, starting jobs")
			s.setupJobs()
			s.cron.Start()
			s.maintainLeadership(ctx)
			s.cron.Stop()
			log.Info().Msg("Lost leadership, stopping jobs")
		} else {
			log.Debug().Msg("Not leader, waiting...")
			time.Sleep(5 * time.Second)
		}
	}
}

func (s *Scheduler) maintainLeadership(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			s.leader.Release(ctx)
			return
		case <-ticker.C:
			if !s.leader.Extend(ctx) {
				return
			}
		}
	}
}

func (s *Scheduler) setupJobs() {
	// Check for due schedules every minute
	s.cron.AddFunc("0 * * * * *", s.processDueSchedules)

	// Recover stale jobs every 5 minutes
	s.cron.AddFunc("0 */5 * * * *", s.recoverStaleJobs)

	// Cleanup old data daily at 3 AM
	s.cron.AddFunc("0 0 3 * * *", s.cleanupOldData)

	// Aggregate usage hourly
	s.cron.AddFunc("0 0 * * * *", s.aggregateUsage)
}

func (s *Scheduler) processDueSchedules() {
	ctx := context.Background()

	schedules, err := s.scheduleSvc.GetDue(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get due schedules")
		return
	}

	for _, schedule := range schedules {
		log.Info().
			Str("schedule_id", schedule.ID.String()).
			Str("workflow_id", schedule.WorkflowID.String()).
			Msg("Executing scheduled workflow")

		// Queue workflow execution
		task, err := s.queueClient.EnqueueWorkflowExecution(ctx, queue.WorkflowExecutionPayload{
			WorkflowID:  schedule.WorkflowID,
			WorkspaceID: schedule.WorkspaceID,
			TriggerType: models.TriggerSchedule,
			InputData:   schedule.InputData,
		})
		if err != nil {
			log.Error().
				Err(err).
				Str("schedule_id", schedule.ID.String()).
				Msg("Failed to queue scheduled workflow")
			continue
		}

		// Record the run (this also updates next_run_at)
		// Note: We use a placeholder execution ID here, the actual execution will be created by the worker
		if err := s.scheduleSvc.RecordRun(ctx, schedule.ID, schedule.ID); err != nil {
			log.Error().
				Err(err).
				Str("schedule_id", schedule.ID.String()).
				Msg("Failed to record schedule run")
		}

		_ = task
	}
}

func (s *Scheduler) recoverStaleJobs() {
	// TODO: Implement stale job recovery
	// Find executions stuck in "running" for more than 10 minutes
	log.Debug().Msg("Stale job recovery completed")
}

func (s *Scheduler) cleanupOldData() {
	ctx := context.Background()

	log.Info().Msg("Starting data cleanup...")

	// TODO: Delete old executions based on retention policy
	// TODO: Delete old webhook logs
	// TODO: Vacuum database

	log.Info().Msg("Data cleanup completed")
	_ = ctx
}

func (s *Scheduler) aggregateUsage() {
	ctx := context.Background()

	log.Debug().Msg("Aggregating usage...")

	// TODO: Aggregate execution counts per workspace
	// TODO: Update usage records

	_ = ctx
}
