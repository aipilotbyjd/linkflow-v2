package scheduler

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	pollCount       int64
	pollErrors      int64
	pollDurationSum int64
	dispatchCount   int64
	dispatchErrors  int64
	schedulesFound  int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) RecordPoll() {
	atomic.AddInt64(&m.pollCount, 1)
}

func (m *Metrics) RecordPollError() {
	atomic.AddInt64(&m.pollErrors, 1)
}

func (m *Metrics) RecordPollDuration(d time.Duration) {
	atomic.AddInt64(&m.pollDurationSum, d.Milliseconds())
	atomic.AddInt64(&m.pollCount, 1)
}

func (m *Metrics) RecordDispatch() {
	atomic.AddInt64(&m.dispatchCount, 1)
}

func (m *Metrics) RecordDispatchError() {
	atomic.AddInt64(&m.dispatchErrors, 1)
}

func (m *Metrics) RecordSchedulesFound(count int) {
	atomic.AddInt64(&m.schedulesFound, int64(count))
}

func (m *Metrics) PollCount() int64 {
	return atomic.LoadInt64(&m.pollCount)
}

func (m *Metrics) PollErrors() int64 {
	return atomic.LoadInt64(&m.pollErrors)
}

func (m *Metrics) AveragePollDuration() time.Duration {
	count := atomic.LoadInt64(&m.pollCount)
	if count == 0 {
		return 0
	}
	sum := atomic.LoadInt64(&m.pollDurationSum)
	return time.Duration(sum/count) * time.Millisecond
}

func (m *Metrics) DispatchCount() int64 {
	return atomic.LoadInt64(&m.dispatchCount)
}

func (m *Metrics) DispatchErrors() int64 {
	return atomic.LoadInt64(&m.dispatchErrors)
}

func (m *Metrics) SchedulesFound() int64 {
	return atomic.LoadInt64(&m.schedulesFound)
}

func (m *Metrics) Snapshot() map[string]interface{} {
	return map[string]interface{}{
		"poll_count":            m.PollCount(),
		"poll_errors":           m.PollErrors(),
		"average_poll_duration": m.AveragePollDuration().String(),
		"dispatch_count":        m.DispatchCount(),
		"dispatch_errors":       m.DispatchErrors(),
		"schedules_found":       m.SchedulesFound(),
	}
}

func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.pollCount, 0)
	atomic.StoreInt64(&m.pollErrors, 0)
	atomic.StoreInt64(&m.pollDurationSum, 0)
	atomic.StoreInt64(&m.dispatchCount, 0)
	atomic.StoreInt64(&m.dispatchErrors, 0)
	atomic.StoreInt64(&m.schedulesFound, 0)
}
