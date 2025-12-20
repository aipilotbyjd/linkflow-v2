package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/rs/zerolog/log"
)

// BatchProcessor handles high-volume schedule processing
type BatchProcessor struct {
	scheduleSvc *services.ScheduleService
	queueClient *queue.Client
	batchSize   int
	workers     int
	mu          sync.Mutex
}

func NewBatchProcessor(scheduleSvc *services.ScheduleService, queueClient *queue.Client) *BatchProcessor {
	return &BatchProcessor{
		scheduleSvc: scheduleSvc,
		queueClient: queueClient,
		batchSize:   100,
		workers:     5,
	}
}

func (p *BatchProcessor) ProcessDueSchedules(ctx context.Context) error {
	// Get all due schedules in batches
	offset := 0
	totalProcessed := 0

	for {
		schedules, err := p.scheduleSvc.GetDueBatch(ctx, p.batchSize, offset)
		if err != nil {
			return err
		}

		if len(schedules) == 0 {
			break
		}

		// Process batch with worker pool
		processed := p.processBatch(ctx, schedules)
		totalProcessed += processed
		offset += p.batchSize

		// Safety limit
		if offset > 10000 {
			log.Warn().Msg("Batch processor hit safety limit")
			break
		}
	}

	if totalProcessed > 0 {
		log.Info().Int("processed", totalProcessed).Msg("Batch processed due schedules")
	}

	return nil
}

func (p *BatchProcessor) processBatch(ctx context.Context, schedules []models.Schedule) int {
	if len(schedules) == 0 {
		return 0
	}

	// Create worker pool
	jobs := make(chan models.Schedule, len(schedules))
	results := make(chan bool, len(schedules))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for schedule := range jobs {
				success := p.processSchedule(ctx, schedule)
				results <- success
			}
		}()
	}

	// Send jobs
	for _, schedule := range schedules {
		jobs <- schedule
	}
	close(jobs)

	// Wait and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	processed := 0
	for success := range results {
		if success {
			processed++
		}
	}

	return processed
}

func (p *BatchProcessor) processSchedule(ctx context.Context, schedule models.Schedule) bool {
	// Queue workflow execution
	_, err := p.queueClient.EnqueueWorkflowExecution(ctx, queue.WorkflowExecutionPayload{
		WorkflowID:  schedule.WorkflowID,
		WorkspaceID: schedule.WorkspaceID,
		TriggerType: models.TriggerSchedule,
		InputData:   schedule.InputData,
		TriggerData: models.JSON{
			"scheduleId": schedule.ID.String(),
			"scheduledAt": time.Now().Format(time.RFC3339),
		},
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("schedule_id", schedule.ID.String()).
			Msg("Failed to queue scheduled workflow")
		return false
	}

	// Update next run time
	if err := p.scheduleSvc.RecordRun(ctx, schedule.ID, uuid.Nil); err != nil {
		log.Error().
			Err(err).
			Str("schedule_id", schedule.ID.String()).
			Msg("Failed to record schedule run")
	}

	return true
}

// PriorityScheduler handles priority-based scheduling
type PriorityScheduler struct {
	scheduleSvc *services.ScheduleService
	queueClient *queue.Client
}

func NewPriorityScheduler(scheduleSvc *services.ScheduleService, queueClient *queue.Client) *PriorityScheduler {
	return &PriorityScheduler{
		scheduleSvc: scheduleSvc,
		queueClient: queueClient,
	}
}

func (s *PriorityScheduler) ProcessWithPriority(ctx context.Context) error {
	// Get schedules grouped by priority
	highPriority, err := s.scheduleSvc.GetDueByPriority(ctx, "high")
	if err != nil {
		return err
	}

	normalPriority, err := s.scheduleSvc.GetDueByPriority(ctx, "normal")
	if err != nil {
		return err
	}

	lowPriority, err := s.scheduleSvc.GetDueByPriority(ctx, "low")
	if err != nil {
		return err
	}

	// Process high priority first with critical queue
	for _, schedule := range highPriority {
		s.queueClient.EnqueuePriorityWorkflowExecution(ctx, queue.WorkflowExecutionPayload{
			WorkflowID:  schedule.WorkflowID,
			WorkspaceID: schedule.WorkspaceID,
			TriggerType: models.TriggerSchedule,
			InputData:   schedule.InputData,
		})
		s.scheduleSvc.RecordRun(ctx, schedule.ID, uuid.Nil)
	}

	// Normal priority
	for _, schedule := range normalPriority {
		s.queueClient.EnqueueWorkflowExecution(ctx, queue.WorkflowExecutionPayload{
			WorkflowID:  schedule.WorkflowID,
			WorkspaceID: schedule.WorkspaceID,
			TriggerType: models.TriggerSchedule,
			InputData:   schedule.InputData,
		})
		s.scheduleSvc.RecordRun(ctx, schedule.ID, uuid.Nil)
	}

	// Low priority with delay
	for _, schedule := range lowPriority {
		s.queueClient.EnqueueDelayedWorkflowExecution(ctx, queue.WorkflowExecutionPayload{
			WorkflowID:  schedule.WorkflowID,
			WorkspaceID: schedule.WorkspaceID,
			TriggerType: models.TriggerSchedule,
			InputData:   schedule.InputData,
		}, 30*time.Second)
		s.scheduleSvc.RecordRun(ctx, schedule.ID, uuid.Nil)
	}

	total := len(highPriority) + len(normalPriority) + len(lowPriority)
	if total > 0 {
		log.Info().
			Int("high", len(highPriority)).
			Int("normal", len(normalPriority)).
			Int("low", len(lowPriority)).
			Msg("Processed prioritized schedules")
	}

	return nil
}

// RateLimitedScheduler prevents overwhelming downstream systems
type RateLimitedScheduler struct {
	scheduleSvc   *services.ScheduleService
	queueClient   *queue.Client
	maxPerMinute  int
	windowSize    time.Duration
	currentCount  int
	windowStart   time.Time
	mu            sync.Mutex
}

func NewRateLimitedScheduler(scheduleSvc *services.ScheduleService, queueClient *queue.Client, maxPerMinute int) *RateLimitedScheduler {
	return &RateLimitedScheduler{
		scheduleSvc:  scheduleSvc,
		queueClient:  queueClient,
		maxPerMinute: maxPerMinute,
		windowSize:   time.Minute,
		windowStart:  time.Now(),
	}
}

func (s *RateLimitedScheduler) CanProcess() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if now.Sub(s.windowStart) > s.windowSize {
		s.windowStart = now
		s.currentCount = 0
	}

	return s.currentCount < s.maxPerMinute
}

func (s *RateLimitedScheduler) IncrementCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentCount++
}

func (s *RateLimitedScheduler) Process(ctx context.Context) error {
	schedules, err := s.scheduleSvc.GetDue(ctx)
	if err != nil {
		return err
	}

	processed := 0
	deferred := 0

	for _, schedule := range schedules {
		if !s.CanProcess() {
			// Defer remaining to next minute
			deferred++
			continue
		}

		_, err := s.queueClient.EnqueueWorkflowExecution(ctx, queue.WorkflowExecutionPayload{
			WorkflowID:  schedule.WorkflowID,
			WorkspaceID: schedule.WorkspaceID,
			TriggerType: models.TriggerSchedule,
			InputData:   schedule.InputData,
		})

		if err != nil {
			log.Error().Err(err).Str("schedule_id", schedule.ID.String()).Msg("Failed to queue")
			continue
		}

		s.IncrementCount()
		s.scheduleSvc.RecordRun(ctx, schedule.ID, uuid.Nil)
		processed++
	}

	if processed > 0 || deferred > 0 {
		log.Info().
			Int("processed", processed).
			Int("deferred", deferred).
			Msg("Rate-limited schedule processing")
	}

	return nil
}
