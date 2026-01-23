package replay

import (
	"github.com/linkflow-ai/linkflow/internal/core/domain/replay"
)

// CreateReplayRequest represents a request to create a replay session
type CreateReplayRequest struct {
	ExecutionID      string                            `json:"execution_id" validate:"required"`
	Mode             string                            `json:"mode" validate:"required"`
	StartFromNodeID  *string                           `json:"start_from_node_id,omitempty"`
	EndAtNodeID      *string                           `json:"end_at_node_id,omitempty"`
	SkipNodes        []string                          `json:"skip_nodes,omitempty"`
	ModifyInputs     map[string]interface{}            `json:"modify_inputs,omitempty"`
	ModifyNodeParams map[string]map[string]interface{} `json:"modify_node_params,omitempty"`
	Breakpoints      []string                          `json:"breakpoints,omitempty"`
}

// ReplaySessionResponse represents a replay session response
type ReplaySessionResponse struct {
	ID             string   `json:"id"`
	OriginalExecID string   `json:"original_execution_id"`
	WorkflowID     string   `json:"workflow_id"`
	Status         string   `json:"status"`
	Mode           string   `json:"mode"`
	CurrentNodeID  *string  `json:"current_node_id,omitempty"`
	Breakpoints    []string `json:"breakpoints"`
	NewExecutionID *string  `json:"new_execution_id,omitempty"`
	CreatedAt      string   `json:"created_at"`
	StartedAt      *string  `json:"started_at,omitempty"`
	CompletedAt    *string  `json:"completed_at,omitempty"`
}

// StepRequest represents a request to step through replay
type StepRequest struct {
	Action string `json:"action" validate:"required"` // step, continue, pause
}

// BreakpointRequest represents a request to add/remove breakpoint
type BreakpointRequest struct {
	NodeID string `json:"node_id" validate:"required"`
	Action string `json:"action" validate:"required"` // add, remove
}

// EventLogResponse represents an event log entry
type EventLogResponse struct {
	ID          string                 `json:"id"`
	ExecutionID string                 `json:"execution_id"`
	EventType   string                 `json:"event_type"`
	NodeID      *string                `json:"node_id,omitempty"`
	NodeType    *string                `json:"node_type,omitempty"`
	Timestamp   string                 `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

// ToReplaySessionResponse converts domain to response
func ToReplaySessionResponse(s *replay.ReplaySession) ReplaySessionResponse {
	resp := ReplaySessionResponse{
		ID:             s.ID.String(),
		OriginalExecID: s.OriginalExecID.String(),
		WorkflowID:     s.WorkflowID.String(),
		Status:         string(s.Status),
		Mode:           string(s.Mode),
		CurrentNodeID:  s.CurrentNodeID,
		Breakpoints:    s.Breakpoints,
		CreatedAt:      s.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if s.NewExecutionID != nil {
		id := s.NewExecutionID.String()
		resp.NewExecutionID = &id
	}
	if s.StartedAt != nil {
		t := s.StartedAt.Format("2006-01-02T15:04:05Z")
		resp.StartedAt = &t
	}
	if s.CompletedAt != nil {
		t := s.CompletedAt.Format("2006-01-02T15:04:05Z")
		resp.CompletedAt = &t
	}
	return resp
}
