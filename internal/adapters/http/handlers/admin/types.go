package admin

// StreamStats represents stream statistics
type StreamStats struct {
	Name          string `json:"name"`
	Length        int64  `json:"length"`
	Pending       int64  `json:"pending"`
	Consumers     int    `json:"consumers"`
	LastDelivered string `json:"lastDelivered"`
}

// QueueStats represents queue statistics
type QueueStats struct {
	Pending   int64 `json:"pending"`
	Active    int64 `json:"active"`
	Scheduled int64 `json:"scheduled"`
	Retry     int64 `json:"retry"`
	Archived  int64 `json:"archived"`
	Completed int64 `json:"completed"`
}

// SystemMetrics represents system metrics
type SystemMetrics struct {
	Uptime           string                 `json:"uptime"`
	Version          string                 `json:"version"`
	GoVersion        string                 `json:"goVersion"`
	NumGoroutines    int                    `json:"numGoroutines"`
	NumCPU           int                    `json:"numCpu"`
	MemoryAlloc      uint64                 `json:"memoryAlloc"`
	MemoryTotalAlloc uint64                 `json:"memoryTotalAlloc"`
	MemorySys        uint64                 `json:"memorySys"`
	NumGC            uint32                 `json:"numGc"`
	Streams          map[string]StreamStats `json:"streams"`
	Queues           map[string]QueueStats  `json:"queues"`
}

// MetricsCollector defines the metrics collector interface
type MetricsCollector interface {
	CollectMetrics() map[string]interface{}
}

// StreamManager defines the stream manager interface
type StreamManager interface {
	GetStats(streamName string) (*StreamStats, error)
	ReplayDLQ(streamName string, count int) (int, error)
	TrimStream(streamName string, maxLen int64) (int64, error)
}
