package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
)

const (
	lockTTL       = 30 * time.Second
	extendTTL     = 10 * time.Second
)

type LeaderElection struct {
	redis    *pkgredis.Client
	key      string
	identity string
}

func NewLeaderElection(redis *pkgredis.Client, key string) *LeaderElection {
	return &LeaderElection{
		redis:    redis,
		key:      "leader:" + key,
		identity: uuid.New().String(),
	}
}

func (l *LeaderElection) TryAcquire(ctx context.Context) (bool, error) {
	return l.redis.AcquireLock(ctx, l.key, l.identity, lockTTL)
}

func (l *LeaderElection) Extend(ctx context.Context) bool {
	extended, err := l.redis.ExtendLock(ctx, l.key, l.identity, lockTTL)
	if err != nil {
		return false
	}
	return extended
}

func (l *LeaderElection) Release(ctx context.Context) error {
	return l.redis.ReleaseLock(ctx, l.key, l.identity)
}

func (l *LeaderElection) IsLeader(ctx context.Context) (bool, error) {
	val, err := l.redis.Get(ctx, l.key).Result()
	if err != nil {
		return false, err
	}
	return val == l.identity, nil
}
