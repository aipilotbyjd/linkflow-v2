package admin

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type StreamStatsHandler struct{}

func NewStreamStatsHandler() *StreamStatsHandler {
	return &StreamStatsHandler{}
}

type StreamStats struct {
	WebhookStream WebhookStreamStats `json:"webhook_stream"`
	EventStream   EventStreamStats   `json:"event_stream"`
}

type WebhookStreamStats struct {
	Length    int64 `json:"length"`
	Consumers int   `json:"consumers"`
	Pending   int64 `json:"pending"`
}

type EventStreamStats struct {
	Length    int64 `json:"length"`
	Consumers int   `json:"consumers"`
}

func (h *StreamStatsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement actual stream stats from Redis
	stats := StreamStats{
		WebhookStream: WebhookStreamStats{
			Length:    0,
			Consumers: 0,
			Pending:   0,
		},
		EventStream: EventStreamStats{
			Length:    0,
			Consumers: 0,
		},
	}

	common.Success(w, stats)
}
