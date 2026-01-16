package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type LeaderElection struct {
	redis      *redis.Client
	instanceID string
	leaseKey   string
	leaseTTL   time.Duration
	isLeader   bool
	mu         sync.RWMutex
	stopCh     chan struct{}
}

func NewLeaderElection(redisClient *redis.Client, leaseTTL time.Duration) *LeaderElection {
	return &LeaderElection{
		redis:      redisClient,
		instanceID: uuid.New().String(),
		leaseKey:   "scheduler:leader",
		leaseTTL:   leaseTTL,
		stopCh:     make(chan struct{}),
	}
}

func (l *LeaderElection) Run(ctx context.Context) {
	ticker := time.NewTicker(l.leaseTTL / 3)
	defer ticker.Stop()

	// Try to acquire leadership immediately
	l.tryAcquire(ctx)

	for {
		select {
		case <-ctx.Done():
			l.release(ctx)
			return
		case <-l.stopCh:
			l.release(ctx)
			return
		case <-ticker.C:
			if l.IsLeader() {
				l.renewLease(ctx)
			} else {
				l.tryAcquire(ctx)
			}
		}
	}
}

func (l *LeaderElection) tryAcquire(ctx context.Context) {
	// Try to set the key only if it doesn't exist
	ok, err := l.redis.SetNX(ctx, l.leaseKey, l.instanceID, l.leaseTTL).Result()
	if err != nil {
		return
	}

	l.mu.Lock()
	l.isLeader = ok
	l.mu.Unlock()
}

func (l *LeaderElection) renewLease(ctx context.Context) {
	// Get current value
	val, err := l.redis.Get(ctx, l.leaseKey).Result()
	if err != nil {
		l.mu.Lock()
		l.isLeader = false
		l.mu.Unlock()
		return
	}

	// Only renew if we're still the leader
	if val != l.instanceID {
		l.mu.Lock()
		l.isLeader = false
		l.mu.Unlock()
		return
	}

	// Extend the lease
	_, err = l.redis.Expire(ctx, l.leaseKey, l.leaseTTL).Result()
	if err != nil {
		l.mu.Lock()
		l.isLeader = false
		l.mu.Unlock()
	}
}

func (l *LeaderElection) release(ctx context.Context) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.isLeader {
		return
	}

	// Only delete if we're the owner
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	l.redis.Eval(ctx, script, []string{l.leaseKey}, l.instanceID)
	l.isLeader = false
}

func (l *LeaderElection) IsLeader() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.isLeader
}

func (l *LeaderElection) Stop() {
	close(l.stopCh)
}

func (l *LeaderElection) InstanceID() string {
	return l.instanceID
}
