package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// CancellationManager manages workflow execution cancellation
type CancellationManager struct {
	redis   *redis.Client
	active  sync.Map // executionID -> cancelFunc
	channel string
}

// CancellationMessage represents a cancellation request
type CancellationMessage struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	Reason      string    `json:"reason"`
	RequestedBy string    `json:"requested_by"`
	RequestedAt time.Time `json:"requested_at"`
}

// NewCancellationManager creates a new cancellation manager
func NewCancellationManager(redis *redis.Client) *CancellationManager {
	return &CancellationManager{
		redis:   redis,
		channel: "workflow:cancel",
	}
}

// Register registers an execution for cancellation tracking
func (cm *CancellationManager) Register(executionID uuid.UUID, cancel context.CancelFunc) {
	cm.active.Store(executionID.String(), cancel)
}

// Unregister removes an execution from cancellation tracking
func (cm *CancellationManager) Unregister(executionID uuid.UUID) {
	cm.active.Delete(executionID.String())
}

// Cancel cancels an execution
func (cm *CancellationManager) Cancel(ctx context.Context, executionID uuid.UUID, reason, requestedBy string) error {
	// Try local cancellation first
	if cancel, ok := cm.active.Load(executionID.String()); ok {
		if cancelFunc, ok := cancel.(context.CancelFunc); ok {
			cancelFunc()
			cm.active.Delete(executionID.String())
			log.Info().
				Str("execution_id", executionID.String()).
				Str("reason", reason).
				Msg("Execution cancelled locally")
			return nil
		}
	}

	// Publish to Redis for distributed cancellation
	msg := CancellationMessage{
		ExecutionID: executionID,
		Reason:      reason,
		RequestedBy: requestedBy,
		RequestedAt: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal cancellation message: %w", err)
	}

	if err := cm.redis.Publish(ctx, cm.channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish cancellation: %w", err)
	}

	log.Info().
		Str("execution_id", executionID.String()).
		Str("reason", reason).
		Msg("Cancellation request published")

	return nil
}

// Listen starts listening for cancellation requests
func (cm *CancellationManager) Listen(ctx context.Context) {
	pubsub := cm.redis.Subscribe(ctx, cm.channel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	log.Info().Msg("Cancellation manager listening for requests")

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			var cancellation CancellationMessage
			if err := json.Unmarshal([]byte(msg.Payload), &cancellation); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal cancellation message")
				continue
			}

			cm.handleCancellation(cancellation)
		}
	}
}

func (cm *CancellationManager) handleCancellation(msg CancellationMessage) {
	if cancel, ok := cm.active.Load(msg.ExecutionID.String()); ok {
		if cancelFunc, ok := cancel.(context.CancelFunc); ok {
			cancelFunc()
			cm.active.Delete(msg.ExecutionID.String())
			log.Info().
				Str("execution_id", msg.ExecutionID.String()).
				Str("reason", msg.Reason).
				Msg("Execution cancelled via pubsub")
		}
	}
}

// IsActive checks if an execution is currently active
func (cm *CancellationManager) IsActive(executionID uuid.UUID) bool {
	_, ok := cm.active.Load(executionID.String())
	return ok
}

// ActiveCount returns the number of active executions
func (cm *CancellationManager) ActiveCount() int {
	count := 0
	cm.active.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// GetActiveExecutions returns all active execution IDs
func (cm *CancellationManager) GetActiveExecutions() []uuid.UUID {
	var ids []uuid.UUID
	cm.active.Range(func(key, value interface{}) bool {
		if id, err := uuid.Parse(key.(string)); err == nil {
			ids = append(ids, id)
		}
		return true
	})
	return ids
}

// CancellationToken is a token that can be checked for cancellation
type CancellationToken struct {
	ctx    context.Context
	cancel context.CancelFunc
	reason string
	mu     sync.RWMutex
}

// NewCancellationToken creates a new cancellation token
func NewCancellationToken(parent context.Context) *CancellationToken {
	ctx, cancel := context.WithCancel(parent)
	return &CancellationToken{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Cancel cancels the token with a reason
func (ct *CancellationToken) Cancel(reason string) {
	ct.mu.Lock()
	ct.reason = reason
	ct.mu.Unlock()
	ct.cancel()
}

// IsCancelled checks if the token is cancelled
func (ct *CancellationToken) IsCancelled() bool {
	select {
	case <-ct.ctx.Done():
		return true
	default:
		return false
	}
}

// Reason returns the cancellation reason
func (ct *CancellationToken) Reason() string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.reason
}

// Context returns the underlying context
func (ct *CancellationToken) Context() context.Context {
	return ct.ctx
}

// Done returns a channel that's closed when cancelled
func (ct *CancellationToken) Done() <-chan struct{} {
	return ct.ctx.Done()
}

// ProgressTracker tracks execution progress
type ProgressTracker struct {
	redis       *redis.Client
	executionID uuid.UUID
	totalNodes  int
	completed   int
	mu          sync.Mutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(redis *redis.Client, executionID uuid.UUID, totalNodes int) *ProgressTracker {
	return &ProgressTracker{
		redis:       redis,
		executionID: executionID,
		totalNodes:  totalNodes,
	}
}

// Update updates progress
func (pt *ProgressTracker) Update(ctx context.Context, completed int, currentNode string) error {
	pt.mu.Lock()
	pt.completed = completed
	pt.mu.Unlock()

	progress := map[string]interface{}{
		"execution_id": pt.executionID.String(),
		"total_nodes":  pt.totalNodes,
		"completed":    completed,
		"current_node": currentNode,
		"percentage":   (completed * 100) / pt.totalNodes,
		"updated_at":   time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(progress)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("execution:progress:%s", pt.executionID)
	return pt.redis.Set(ctx, key, data, 1*time.Hour).Err()
}

// GetProgress gets current progress
func (pt *ProgressTracker) GetProgress(ctx context.Context) (map[string]interface{}, error) {
	key := fmt.Sprintf("execution:progress:%s", pt.executionID)
	data, err := pt.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var progress map[string]interface{}
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, err
	}

	return progress, nil
}

// GetProgressByID gets progress for any execution
func GetProgressByID(ctx context.Context, redis *redis.Client, executionID uuid.UUID) (map[string]interface{}, error) {
	key := fmt.Sprintf("execution:progress:%s", executionID)
	data, err := redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var progress map[string]interface{}
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, err
	}

	return progress, nil
}
