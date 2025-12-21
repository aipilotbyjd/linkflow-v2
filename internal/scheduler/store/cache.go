package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
)

type CachedStore struct {
	store ScheduleStore
	redis *pkgredis.Client
	ttl   time.Duration
}

func NewCachedStore(store ScheduleStore, redis *pkgredis.Client) *CachedStore {
	return &CachedStore{
		store: store,
		redis: redis,
		ttl:   5 * time.Minute,
	}
}

func (s *CachedStore) GetDue(ctx context.Context, limit int) ([]*Schedule, error) {
	// Due schedules change frequently, don't cache
	return s.store.GetDue(ctx, limit)
}

func (s *CachedStore) GetDueByPriority(ctx context.Context, priority string, limit int) ([]*Schedule, error) {
	return s.store.GetDueByPriority(ctx, priority, limit)
}

func (s *CachedStore) GetDueByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*Schedule, error) {
	return s.store.GetDueByWorkspace(ctx, workspaceID, limit)
}

func (s *CachedStore) UpdateNextRun(ctx context.Context, id uuid.UUID, nextRun time.Time) error {
	err := s.store.UpdateNextRun(ctx, id, nextRun)
	if err == nil {
		s.invalidate(ctx, id)
	}
	return err
}

func (s *CachedStore) RecordRun(ctx context.Context, id uuid.UUID, nextRun time.Time) error {
	err := s.store.RecordRun(ctx, id, nextRun)
	if err == nil {
		s.invalidate(ctx, id)
	}
	return err
}

func (s *CachedStore) GetStale(ctx context.Context, threshold time.Duration) ([]*Schedule, error) {
	return s.store.GetStale(ctx, threshold)
}

func (s *CachedStore) GetByID(ctx context.Context, id uuid.UUID) (*Schedule, error) {
	key := s.cacheKey(id)

	// Try cache first
	data, err := s.redis.Get(ctx, key).Bytes()
	if err == nil {
		var schedule Schedule
		if json.Unmarshal(data, &schedule) == nil {
			return &schedule, nil
		}
	}

	// Fetch from store
	schedule, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache it
	if data, err := json.Marshal(schedule); err == nil {
		s.redis.Set(ctx, key, data, s.ttl)
	}

	return schedule, nil
}

func (s *CachedStore) cacheKey(id uuid.UUID) string {
	return fmt.Sprintf("scheduler:schedule:%s", id.String())
}

func (s *CachedStore) invalidate(ctx context.Context, id uuid.UUID) {
	s.redis.Del(ctx, s.cacheKey(id))
}
