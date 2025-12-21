package leader

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/rs/zerolog/log"
)

type Election struct {
	redis    *pkgredis.Client
	key      string
	identity string
	ttl      time.Duration
	isLeader atomic.Bool
}

func NewElection(redis *pkgredis.Client, key string, ttl time.Duration) *Election {
	return &Election{
		redis:    redis,
		key:      key,
		identity: uuid.New().String(),
		ttl:      ttl,
	}
}

func (e *Election) TryAcquire(ctx context.Context) (bool, error) {
	acquired, err := e.redis.AcquireLock(ctx, e.key, e.identity, e.ttl)
	if err != nil {
		return false, err
	}

	if acquired {
		e.isLeader.Store(true)
		log.Info().
			Str("identity", e.identity).
			Str("key", e.key).
			Msg("Leadership acquired")
	}

	return acquired, nil
}

func (e *Election) Extend(ctx context.Context) bool {
	if !e.isLeader.Load() {
		return false
	}

	extended, err := e.redis.ExtendLock(ctx, e.key, e.identity, e.ttl)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extend leadership")
		e.isLeader.Store(false)
		return false
	}

	if !extended {
		log.Warn().Msg("Lost leadership (lock expired)")
		e.isLeader.Store(false)
		return false
	}

	return true
}

func (e *Election) Release(ctx context.Context) error {
	if !e.isLeader.Load() {
		return nil
	}

	err := e.redis.ReleaseLock(ctx, e.key, e.identity)
	e.isLeader.Store(false)

	if err != nil {
		log.Error().Err(err).Msg("Failed to release leadership")
		return err
	}

	log.Info().Str("identity", e.identity).Msg("Leadership released")
	return nil
}

func (e *Election) IsLeader() bool {
	return e.isLeader.Load()
}

func (e *Election) Identity() string {
	return e.identity
}
