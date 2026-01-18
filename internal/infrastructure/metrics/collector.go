package metrics

import (
	"runtime"
	"time"
)

// Collector implements MetricsCollector for system metrics
type Collector struct {
	startTime time.Time
	version   string
}

// NewCollector creates a new metrics collector
func NewCollector(version string) *Collector {
	return &Collector{
		startTime: time.Now(),
		version:   version,
	}
}

// CollectMetrics collects system metrics
func (c *Collector) CollectMetrics() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"uptime":           time.Since(c.startTime).String(),
		"version":          c.version,
		"goVersion":        runtime.Version(),
		"numGoroutines":    runtime.NumGoroutine(),
		"numCpu":           runtime.NumCPU(),
		"memoryAlloc":      memStats.Alloc,
		"memoryTotalAlloc": memStats.TotalAlloc,
		"memorySys":        memStats.Sys,
		"numGc":            memStats.NumGC,
	}
}
