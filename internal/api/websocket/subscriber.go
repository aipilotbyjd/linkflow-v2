package websocket

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Subscriber subscribes to Redis pubsub and broadcasts to WebSocket clients
type Subscriber struct {
	redisClient *redis.Client
	hub         *Hub
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewSubscriber(redisClient *redis.Client, hub *Hub) *Subscriber {
	ctx, cancel := context.WithCancel(context.Background())
	return &Subscriber{
		redisClient: redisClient,
		hub:         hub,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *Subscriber) Start() {
	s.wg.Add(1)
	go s.subscribeToEvents()
}

func (s *Subscriber) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *Subscriber) subscribeToEvents() {
	defer s.wg.Done()

	// Subscribe to all workspace channels using pattern
	// Worker publishes to "workspace:{id}" channel
	pubsub := s.redisClient.PSubscribe(s.ctx, "workspace:*")
	defer pubsub.Close()

	ch := pubsub.Channel()

	log.Info().Msg("WebSocket subscriber started")

	for {
		select {
		case <-s.ctx.Done():
			log.Info().Msg("WebSocket subscriber stopped")
			return
		case msg := <-ch:
			s.handleMessage(msg)
		}
	}
}

// WorkerEvent is the event structure from the worker
type WorkerEvent struct {
	Type        string                 `json:"type"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	WorkflowID  uuid.UUID              `json:"workflow_id,omitempty"`
	ExecutionID uuid.UUID              `json:"execution_id,omitempty"`
	NodeID      string                 `json:"node_id,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   string                 `json:"timestamp"`
}

func (s *Subscriber) handleMessage(msg *redis.Message) {
	// Parse workspace ID from channel name
	var workspaceID uuid.UUID
	_, err := parseChannelWorkspaceID(msg.Channel, &workspaceID)
	if err != nil {
		log.Error().Err(err).Str("channel", msg.Channel).Msg("Failed to parse channel")
		return
	}

	// Parse the worker event
	var workerEvent WorkerEvent
	if err := json.Unmarshal([]byte(msg.Payload), &workerEvent); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal event")
		return
	}

	// Convert to WebSocket event format
	data := workerEvent.Data
	if data == nil {
		data = make(map[string]interface{})
	}
	data["workflow_id"] = workerEvent.WorkflowID.String()
	data["execution_id"] = workerEvent.ExecutionID.String()
	if workerEvent.NodeID != "" {
		data["node_id"] = workerEvent.NodeID
	}

	wsEvent := Event{
		Type: EventType(workerEvent.Type),
		Data: data,
	}

	log.Debug().
		Str("type", workerEvent.Type).
		Str("workspace_id", workspaceID.String()).
		Str("execution_id", workerEvent.ExecutionID.String()).
		Msg("Broadcasting WebSocket event")

	// Broadcast to all clients in the workspace
	s.hub.BroadcastToWorkspace(workspaceID, wsEvent)
}

func parseChannelWorkspaceID(channel string, workspaceID *uuid.UUID) (bool, error) {
	// Channel format: workspace:{workspaceId}
	prefix := "workspace:"
	
	if len(channel) <= len(prefix) || channel[:len(prefix)] != prefix {
		return false, nil
	}
	
	wsIDStr := channel[len(prefix):]
	
	id, err := uuid.Parse(wsIDStr)
	if err != nil {
		return false, err
	}

	*workspaceID = id
	return true, nil
}

// ExecutionLogStreamer streams execution logs in real-time
type ExecutionLogStreamer struct {
	redisClient *redis.Client
	hub         *Hub
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewExecutionLogStreamer(redisClient *redis.Client, hub *Hub) *ExecutionLogStreamer {
	ctx, cancel := context.WithCancel(context.Background())
	return &ExecutionLogStreamer{
		redisClient: redisClient,
		hub:         hub,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *ExecutionLogStreamer) Start() {
	s.wg.Add(1)
	go s.subscribeToLogs()
}

func (s *ExecutionLogStreamer) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *ExecutionLogStreamer) subscribeToLogs() {
	defer s.wg.Done()

	pubsub := s.redisClient.PSubscribe(s.ctx, "execution:*:logs")
	defer pubsub.Close()

	ch := pubsub.Channel()

	log.Info().Msg("Execution log streamer started")

	for {
		select {
		case <-s.ctx.Done():
			log.Info().Msg("Execution log streamer stopped")
			return
		case msg := <-ch:
			s.handleLogMessage(msg)
		}
	}
}

func (s *ExecutionLogStreamer) handleLogMessage(msg *redis.Message) {
	// Parse execution ID from channel
	// Channel format: execution:{executionId}:logs
	var execIDStr string
	parseExecutionChannel(msg.Channel, &execIDStr)
	
	execID, err := uuid.Parse(execIDStr)
	if err != nil {
		return
	}

	// Parse log entry
	var logEntry ExecutionLogEntry
	if err := json.Unmarshal([]byte(msg.Payload), &logEntry); err != nil {
		return
	}

	// Create event
	event := Event{
		Type: "log",
		Data: map[string]interface{}{
			"execution_id": execID.String(),
			"node_id":      logEntry.NodeID,
			"level":        logEntry.Level,
			"message":      logEntry.Message,
			"data":         logEntry.Data,
			"timestamp":    logEntry.Timestamp,
		},
	}

	// Broadcast to clients watching this execution
	s.hub.BroadcastToExecution(execID, event)
}

func parseExecutionChannel(channel string, execID *string) {
	prefix := "execution:"
	suffix := ":logs"
	
	if len(channel) > len(prefix)+len(suffix) && channel[:len(prefix)] == prefix {
		rest := channel[len(prefix):]
		if idx := len(rest) - len(suffix); idx > 0 && rest[idx:] == suffix {
			*execID = rest[:idx]
		}
	}
}

type ExecutionLogEntry struct {
	NodeID    string                 `json:"node_id"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// LogPublisher publishes execution logs to Redis
type LogPublisher struct {
	redisClient *redis.Client
}

func NewLogPublisher(redisClient *redis.Client) *LogPublisher {
	return &LogPublisher{redisClient: redisClient}
}

func (p *LogPublisher) PublishLog(ctx context.Context, executionID uuid.UUID, nodeID, level, message string, data map[string]interface{}) error {
	channel := getExecutionLogChannel(executionID)
	
	entry := ExecutionLogEntry{
		NodeID:    nodeID,
		Level:     level,
		Message:   message,
		Data:      data,
		Timestamp: getTimestamp(),
	}
	
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	
	return p.redisClient.Publish(ctx, channel, payload).Err()
}

func getExecutionLogChannel(executionID uuid.UUID) string {
	return "execution:" + executionID.String() + ":logs"
}

func getTimestamp() string {
	return json.Number(string(rune(json.Number("").String()[0]))).String()
}
