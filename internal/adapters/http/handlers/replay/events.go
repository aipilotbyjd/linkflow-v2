package replay

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// ListEventsHandler handles listing events for an execution
type ListEventsHandler struct{}

func NewListEventsHandler() *ListEventsHandler {
	return &ListEventsHandler{}
}

func (h *ListEventsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	execID, err := uuid.Parse(chi.URLParam(r, "executionId"))
	if err != nil {
		common.BadRequest(w, "Invalid execution ID")
		return
	}

	// In production, fetch from event store
	// For now, return sample events
	events := []EventLogResponse{
		{
			ID:          uuid.New().String(),
			ExecutionID: execID.String(),
			EventType:   "execution.started",
			Timestamp:   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			Data: map[string]interface{}{
				"trigger_type": "webhook",
				"input_size":   1024,
			},
		},
		{
			ID:          uuid.New().String(),
			ExecutionID: execID.String(),
			EventType:   "node.started",
			NodeID:      strPtr("node_1"),
			NodeType:    strPtr("trigger.webhook"),
			Timestamp:   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			Data:        map[string]interface{}{},
		},
		{
			ID:          uuid.New().String(),
			ExecutionID: execID.String(),
			EventType:   "node.completed",
			NodeID:      strPtr("node_1"),
			NodeType:    strPtr("trigger.webhook"),
			Timestamp:   time.Now().Add(-4*time.Minute - 55*time.Second).Format(time.RFC3339),
			Data: map[string]interface{}{
				"duration_ms": 50,
				"output_size": 512,
			},
		},
		{
			ID:          uuid.New().String(),
			ExecutionID: execID.String(),
			EventType:   "node.started",
			NodeID:      strPtr("node_2"),
			NodeType:    strPtr("action.http"),
			Timestamp:   time.Now().Add(-4*time.Minute - 50*time.Second).Format(time.RFC3339),
			Data:        map[string]interface{}{},
		},
		{
			ID:          uuid.New().String(),
			ExecutionID: execID.String(),
			EventType:   "api.call",
			NodeID:      strPtr("node_2"),
			Timestamp:   time.Now().Add(-4*time.Minute - 49*time.Second).Format(time.RFC3339),
			Data: map[string]interface{}{
				"method": "POST",
				"url":    "https://api.example.com/data",
			},
		},
		{
			ID:          uuid.New().String(),
			ExecutionID: execID.String(),
			EventType:   "api.response",
			NodeID:      strPtr("node_2"),
			Timestamp:   time.Now().Add(-4*time.Minute - 30*time.Second).Format(time.RFC3339),
			Data: map[string]interface{}{
				"status_code":  200,
				"duration_ms":  1900,
				"content_size": 2048,
			},
		},
		{
			ID:          uuid.New().String(),
			ExecutionID: execID.String(),
			EventType:   "execution.completed",
			Timestamp:   time.Now().Add(-4 * time.Minute).Format(time.RFC3339),
			Data: map[string]interface{}{
				"status":        "success",
				"total_nodes":   5,
				"duration_ms":   60000,
				"credits_used":  0.15,
			},
		},
	}

	common.Success(w, map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

func strPtr(s string) *string {
	return &s
}
