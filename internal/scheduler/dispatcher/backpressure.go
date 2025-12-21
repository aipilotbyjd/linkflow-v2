package dispatcher

import (
	"context"
	"sync/atomic"
	"time"

	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/rs/zerolog/log"
)

type BackpressureMonitor struct {
	redis          *pkgredis.Client
	queueKey       string
	maxDepth       int64
	checkInterval  time.Duration
	currentDepth   atomic.Int64
	isPaused       atomic.Bool
}

func NewBackpressureMonitor(redis *pkgredis.Client, queueKey string, maxDepth int64) *BackpressureMonitor {
	return &BackpressureMonitor{
		redis:         redis,
		queueKey:      queueKey,
		maxDepth:      maxDepth,
		checkInterval: 5 * time.Second,
	}
}

func (m *BackpressureMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.check(ctx)
		}
	}
}

func (m *BackpressureMonitor) check(ctx context.Context) {
	depth, err := m.redis.LLen(ctx, m.queueKey).Result()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to check queue depth")
		return
	}

	m.currentDepth.Store(depth)

	wasPaused := m.isPaused.Load()

	if depth >= m.maxDepth {
		if !wasPaused {
			m.isPaused.Store(true)
			log.Warn().
				Int64("depth", depth).
				Int64("max", m.maxDepth).
				Msg("Backpressure: pausing dispatch")
		}
	} else if depth < m.maxDepth/2 {
		if wasPaused {
			m.isPaused.Store(false)
			log.Info().
				Int64("depth", depth).
				Msg("Backpressure: resuming dispatch")
		}
	}
}

func (m *BackpressureMonitor) ShouldPause() bool {
	return m.isPaused.Load()
}

func (m *BackpressureMonitor) QueueDepth() int64 {
	return m.currentDepth.Load()
}
