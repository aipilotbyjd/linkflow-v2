package cron

import (
	"sync"
	"time"

	"github.com/google/uuid"
	cronlib "github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Calculator struct {
	parser   *Parser
	cache    map[uuid.UUID]*cacheEntry
	cacheTTL time.Duration
	mu       sync.RWMutex
}

type cacheEntry struct {
	schedule   cronlib.Schedule
	expression string
	cachedAt   time.Time
}

func NewCalculator() *Calculator {
	return &Calculator{
		parser:   NewParser(),
		cache:    make(map[uuid.UUID]*cacheEntry),
		cacheTTL: 10 * time.Minute,
	}
}

func (c *Calculator) NextRun(scheduleID uuid.UUID, expression, timezone string) (time.Time, error) {
	// Load timezone
	loc := time.UTC
	if timezone != "" {
		var err error
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			log.Warn().Str("timezone", timezone).Msg("Invalid timezone, using UTC")
			loc = time.UTC
		}
	}

	// Get or parse schedule
	schedule, err := c.getSchedule(scheduleID, expression)
	if err != nil {
		return time.Time{}, err
	}

	// Calculate next run
	now := time.Now().In(loc)
	return schedule.Next(now), nil
}

func (c *Calculator) NextNRuns(expression, timezone string, n int) ([]time.Time, error) {
	loc := time.UTC
	if timezone != "" {
		var err error
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			return nil, err
		}
	}

	schedule, err := c.parser.Parse(expression)
	if err != nil {
		return nil, err
	}

	runs := make([]time.Time, n)
	current := time.Now().In(loc)

	for i := 0; i < n; i++ {
		current = schedule.Next(current)
		runs[i] = current
	}

	return runs, nil
}

func (c *Calculator) getSchedule(id uuid.UUID, expression string) (cronlib.Schedule, error) {
	c.mu.RLock()
	entry, exists := c.cache[id]
	c.mu.RUnlock()

	if exists && entry.expression == expression && time.Since(entry.cachedAt) < c.cacheTTL {
		return entry.schedule, nil
	}

	schedule, err := c.parser.Parse(expression)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[id] = &cacheEntry{
		schedule:   schedule,
		expression: expression,
		cachedAt:   time.Now(),
	}
	c.mu.Unlock()

	return schedule, nil
}

func (c *Calculator) Invalidate(id uuid.UUID) {
	c.mu.Lock()
	delete(c.cache, id)
	c.mu.Unlock()
}

func (c *Calculator) ClearCache() {
	c.mu.Lock()
	c.cache = make(map[uuid.UUID]*cacheEntry)
	c.mu.Unlock()
}

func (c *Calculator) Validate(expression string) error {
	return c.parser.Validate(expression)
}
