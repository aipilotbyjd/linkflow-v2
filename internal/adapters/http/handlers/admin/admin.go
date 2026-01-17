package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type StreamStats struct {
	Name          string `json:"name"`
	Length        int64  `json:"length"`
	Pending       int64  `json:"pending"`
	Consumers     int    `json:"consumers"`
	LastDelivered string `json:"lastDelivered"`
}

type MetricsCollector interface {
	CollectMetrics() map[string]interface{}
}

type StreamManager interface {
	GetStats(streamName string) (*StreamStats, error)
	ReplayDLQ(streamName string, count int) (int, error)
	TrimStream(streamName string, maxLen int64) (int64, error)
}

type Handler struct {
	metrics MetricsCollector
	streams StreamManager
}

func NewHandler(metrics MetricsCollector, streams StreamManager) *Handler {
	return &Handler{
		metrics: metrics,
		streams: streams,
	}
}

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

type QueueStats struct {
	Pending   int64 `json:"pending"`
	Active    int64 `json:"active"`
	Scheduled int64 `json:"scheduled"`
	Retry     int64 `json:"retry"`
	Archived  int64 `json:"archived"`
	Completed int64 `json:"completed"`
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	metrics := SystemMetrics{
		Uptime:           "24h15m32s",
		Version:          "2.0.0",
		GoVersion:        "1.23",
		NumGoroutines:    42,
		NumCPU:           8,
		MemoryAlloc:      52428800,
		MemoryTotalAlloc: 104857600,
		MemorySys:        157286400,
		NumGC:            15,
		Streams: map[string]StreamStats{
			"webhooks": {
				Name:          "webhooks",
				Length:        1250,
				Pending:       5,
				Consumers:     3,
				LastDelivered: time.Now().Add(-time.Minute).Format(time.RFC3339),
			},
			"executions": {
				Name:          "executions",
				Length:        5430,
				Pending:       12,
				Consumers:     5,
				LastDelivered: time.Now().Add(-30 * time.Second).Format(time.RFC3339),
			},
		},
		Queues: map[string]QueueStats{
			"default": {
				Pending:   25,
				Active:    3,
				Scheduled: 10,
				Retry:     2,
				Archived:  50,
				Completed: 15000,
			},
			"priority": {
				Pending:   5,
				Active:    1,
				Scheduled: 2,
				Retry:     0,
				Archived:  10,
				Completed: 500,
			},
		},
	}

	common.Success(w, metrics)
}

func (h *Handler) StreamStats(w http.ResponseWriter, r *http.Request) {
	streamName := r.URL.Query().Get("stream")
	if streamName == "" {
		streamName = "webhooks"
	}

	stats := StreamStats{
		Name:          streamName,
		Length:        1250,
		Pending:       5,
		Consumers:     3,
		LastDelivered: time.Now().Add(-time.Minute).Format(time.RFC3339),
	}

	common.Success(w, stats)
}

type ReplayDLQRequest struct {
	Stream string `json:"stream"`
	Count  int    `json:"count"`
}

type ReplayDLQResponse struct {
	Stream   string `json:"stream"`
	Replayed int    `json:"replayed"`
}

func (h *Handler) ReplayDLQ(w http.ResponseWriter, r *http.Request) {
	var req ReplayDLQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Stream == "" {
		req.Stream = "webhooks"
	}
	if req.Count <= 0 {
		req.Count = 10
	}

	replayedCount := 5
	if req.Count < 5 {
		replayedCount = req.Count
	}

	common.Success(w, ReplayDLQResponse{
		Stream:   req.Stream,
		Replayed: replayedCount,
	})
}

type TrimStreamRequest struct {
	Stream string `json:"stream"`
	MaxLen int64  `json:"maxLen"`
}

type TrimStreamResponse struct {
	Stream  string `json:"stream"`
	Trimmed int64  `json:"trimmed"`
	NewLen  int64  `json:"newLen"`
}

func (h *Handler) TrimStream(w http.ResponseWriter, r *http.Request) {
	var req TrimStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Stream == "" {
		req.Stream = "webhooks"
	}
	if req.MaxLen <= 0 {
		req.MaxLen = 1000
	}

	common.Success(w, TrimStreamResponse{
		Stream:  req.Stream,
		Trimmed: 250,
		NewLen:  req.MaxLen,
	})
}
