package metrics

import (
	"sync/atomic"
	"time"
)

type Collector struct {
	// Counters
	pollsTotal       atomic.Int64
	dispatchedTotal  atomic.Int64
	skippedTotal     atomic.Int64
	failedTotal      atomic.Int64
	recoveredTotal   atomic.Int64

	// Gauges
	queueDepth       atomic.Int64
	activeSchedules  atomic.Int64

	// Timing
	lastPollDuration atomic.Int64 // milliseconds
	avgPollDuration  atomic.Int64 // milliseconds

	// State
	isLeader         atomic.Bool
	startedAt        time.Time
}

func NewCollector() *Collector {
	return &Collector{
		startedAt: time.Now(),
	}
}

func (c *Collector) IncPolls() {
	c.pollsTotal.Add(1)
}

func (c *Collector) IncDispatched(n int64) {
	c.dispatchedTotal.Add(n)
}

func (c *Collector) IncSkipped(n int64) {
	c.skippedTotal.Add(n)
}

func (c *Collector) IncFailed(n int64) {
	c.failedTotal.Add(n)
}

func (c *Collector) IncRecovered(n int64) {
	c.recoveredTotal.Add(n)
}

func (c *Collector) SetQueueDepth(depth int64) {
	c.queueDepth.Store(depth)
}

func (c *Collector) SetActiveSchedules(count int64) {
	c.activeSchedules.Store(count)
}

func (c *Collector) SetLeader(isLeader bool) {
	c.isLeader.Store(isLeader)
}

func (c *Collector) RecordPollDuration(d time.Duration) {
	ms := d.Milliseconds()
	c.lastPollDuration.Store(ms)

	// Simple moving average
	old := c.avgPollDuration.Load()
	if old == 0 {
		c.avgPollDuration.Store(ms)
	} else {
		c.avgPollDuration.Store((old + ms) / 2)
	}
}

type Snapshot struct {
	PollsTotal       int64         `json:"polls_total"`
	DispatchedTotal  int64         `json:"dispatched_total"`
	SkippedTotal     int64         `json:"skipped_total"`
	FailedTotal      int64         `json:"failed_total"`
	RecoveredTotal   int64         `json:"recovered_total"`
	QueueDepth       int64         `json:"queue_depth"`
	ActiveSchedules  int64         `json:"active_schedules"`
	LastPollDuration int64         `json:"last_poll_duration_ms"`
	AvgPollDuration  int64         `json:"avg_poll_duration_ms"`
	IsLeader         bool          `json:"is_leader"`
	Uptime           time.Duration `json:"uptime"`
}

func (c *Collector) Snapshot() *Snapshot {
	return &Snapshot{
		PollsTotal:       c.pollsTotal.Load(),
		DispatchedTotal:  c.dispatchedTotal.Load(),
		SkippedTotal:     c.skippedTotal.Load(),
		FailedTotal:      c.failedTotal.Load(),
		RecoveredTotal:   c.recoveredTotal.Load(),
		QueueDepth:       c.queueDepth.Load(),
		ActiveSchedules:  c.activeSchedules.Load(),
		LastPollDuration: c.lastPollDuration.Load(),
		AvgPollDuration:  c.avgPollDuration.Load(),
		IsLeader:         c.isLeader.Load(),
		Uptime:           time.Since(c.startedAt),
	}
}

func (c *Collector) Reset() {
	c.pollsTotal.Store(0)
	c.dispatchedTotal.Store(0)
	c.skippedTotal.Store(0)
	c.failedTotal.Store(0)
	c.recoveredTotal.Store(0)
}
