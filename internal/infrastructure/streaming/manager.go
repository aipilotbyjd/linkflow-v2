package streaming

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// StreamStats represents stream statistics
type StreamStats struct {
	Name          string `json:"name"`
	Length        int64  `json:"length"`
	Pending       int64  `json:"pending"`
	Consumers     int    `json:"consumers"`
	LastDelivered string `json:"lastDelivered"`
}

// Manager implements StreamManager for Redis streams
type Manager struct {
	client *redis.Client
}

// NewManager creates a new stream manager
func NewManager(client *redis.Client) *Manager {
	return &Manager{client: client}
}

// GetStats gets statistics for a stream
func (m *Manager) GetStats(streamName string) (*StreamStats, error) {
	ctx := context.Background()

	// Get stream length
	length, err := m.client.XLen(ctx, streamName).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get stream length: %w", err)
	}

	// Get stream info
	info, err := m.client.XInfoStream(ctx, streamName).Result()
	if err != nil {
		// Stream may not exist
		return &StreamStats{
			Name:   streamName,
			Length: length,
		}, nil
	}

	// Get consumer groups
	groups, _ := m.client.XInfoGroups(ctx, streamName).Result()

	var pending int64
	for _, g := range groups {
		pending += g.Pending
	}

	return &StreamStats{
		Name:          streamName,
		Length:        length,
		Pending:       pending,
		Consumers:     len(groups),
		LastDelivered: info.LastGeneratedID,
	}, nil
}

// ReplayDLQ replays messages from dead letter queue
func (m *Manager) ReplayDLQ(streamName string, count int) (int, error) {
	ctx := context.Background()
	dlqName := streamName + ":dlq"

	// Read from DLQ
	messages, err := m.client.XRange(ctx, dlqName, "-", "+").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to read DLQ: %w", err)
	}

	if len(messages) == 0 {
		return 0, nil
	}

	// Limit to count
	if count > 0 && len(messages) > count {
		messages = messages[:count]
	}

	// Re-add to main stream
	replayed := 0
	for _, msg := range messages {
		_, err := m.client.XAdd(ctx, &redis.XAddArgs{
			Stream: streamName,
			Values: msg.Values,
		}).Result()
		if err != nil {
			continue
		}

		// Remove from DLQ
		m.client.XDel(ctx, dlqName, msg.ID)
		replayed++
	}

	return replayed, nil
}

// TrimStream trims a stream to max length
func (m *Manager) TrimStream(streamName string, maxLen int64) (int64, error) {
	ctx := context.Background()

	// Get current length
	beforeLen, err := m.client.XLen(ctx, streamName).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get stream length: %w", err)
	}

	// Trim stream
	err = m.client.XTrimMaxLen(ctx, streamName, maxLen).Err()
	if err != nil {
		return 0, fmt.Errorf("failed to trim stream: %w", err)
	}

	// Get new length
	afterLen, err := m.client.XLen(ctx, streamName).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get new stream length: %w", err)
	}

	return beforeLen - afterLen, nil
}
