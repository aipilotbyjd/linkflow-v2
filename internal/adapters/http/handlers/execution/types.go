package execution

import (
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// ExecutionResponse represents execution in responses
type ExecutionResponse struct {
	ID           string     `json:"id"`
	WorkflowID   string     `json:"workflow_id"`
	WorkspaceID  string     `json:"workspace_id"`
	WorkflowName string     `json:"workflow_name,omitempty"`
	Status       string     `json:"status"`
	TriggerType  string     `json:"trigger_type"`
	TriggeredBy  *string    `json:"triggered_by,omitempty"`
	InputData    types.JSON `json:"input_data,omitempty"`
	OutputData   types.JSON `json:"output_data,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	StartedAt    *string    `json:"started_at,omitempty"`
	CompletedAt  *string    `json:"completed_at,omitempty"`
	DurationMs   *int64     `json:"duration_ms,omitempty"`
	CreatedAt    string     `json:"created_at"`
}

// ExecutionStatsResponse represents execution statistics
type ExecutionStatsResponse struct {
	Total         int64            `json:"total"`
	ByStatus      map[string]int64 `json:"by_status"`
	AvgDurationMs int64            `json:"avg_duration_ms"`
	Period        string           `json:"period"`
	StartDate     string           `json:"start_date"`
	EndDate       string           `json:"end_date"`
}

// NodeExecutionResponse represents a node execution in responses
type NodeExecutionResponse struct {
	ID          string                 `json:"id"`
	ExecutionID string                 `json:"executionId"`
	NodeID      string                 `json:"nodeId"`
	NodeType    string                 `json:"nodeType"`
	NodeName    string                 `json:"nodeName"`
	Status      string                 `json:"status"`
	StartedAt   *time.Time             `json:"startedAt,omitempty"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
	Duration    *int64                 `json:"durationMs,omitempty"`
	InputData   map[string]interface{} `json:"inputData,omitempty"`
	OutputData  map[string]interface{} `json:"outputData,omitempty"`
	Error       *string                `json:"error,omitempty"`
	RetryCount  int                    `json:"retryCount"`
}

// WaitingExecution represents a waiting execution
type WaitingExecution struct {
	ID           string     `json:"id"`
	WorkflowID   string     `json:"workflowId"`
	WorkflowName string     `json:"workflowName"`
	NodeID       string     `json:"nodeId"`
	NodeName     string     `json:"nodeName"`
	WaitType     string     `json:"waitType"`
	WaitUntil    *time.Time `json:"waitUntil,omitempty"`
	ResumeToken  string     `json:"resumeToken"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// StartRequest represents execution start request
type StartRequest struct {
	InputData types.JSON `json:"input_data,omitempty"`
}

// ReplayRequest represents replay execution request
type ReplayRequest struct {
	UseOriginalInput bool                   `json:"useOriginalInput"`
	InputData        map[string]interface{} `json:"input_data,omitempty"`
}

// ReplayFromNodeRequest represents replay from node request
type ReplayFromNodeRequest struct {
	NodeID string `json:"node_id"`
}

// BulkDeleteRequest represents bulk delete request
type BulkDeleteRequest struct {
	ExecutionIDs []string `json:"execution_ids,omitempty"`
	OlderThan    *string  `json:"older_than,omitempty"`
	Status       *string  `json:"status,omitempty"`
}

// BulkDeleteResponse represents bulk delete response
type BulkDeleteResponse struct {
	Deleted int64 `json:"deleted"`
}

// ResumeRequest represents resume execution request
type ResumeRequest struct {
	Data map[string]interface{} `json:"data,omitempty"`
}

// ResumeResponse represents resume execution response
type ResumeResponse struct {
	ExecutionID string    `json:"executionId"`
	Status      string    `json:"status"`
	ResumedAt   time.Time `json:"resumedAt"`
}

// ResumeStatusResponse represents resume status response
type ResumeStatusResponse struct {
	Token       string     `json:"token"`
	ExecutionID string     `json:"executionId"`
	Status      string     `json:"status"`
	NodeID      string     `json:"nodeId"`
	WaitType    string     `json:"waitType"`
	CreatedAt   time.Time  `json:"createdAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

// ToExecutionResponse converts domain execution to response
func ToExecutionResponse(e *execution.Execution) ExecutionResponse {
	resp := ExecutionResponse{
		ID:          e.ID.String(),
		WorkflowID:  e.WorkflowID.String(),
		WorkspaceID: e.WorkspaceID.String(),
		Status:      string(e.Status),
		TriggerType: e.TriggerType,
		InputData:   e.InputData,
		OutputData:  e.OutputData,
		CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if e.TriggeredBy != nil {
		s := e.TriggeredBy.String()
		resp.TriggeredBy = &s
	}
	if e.ErrorMessage != nil {
		resp.ErrorMessage = e.ErrorMessage
	}
	if e.StartedAt != nil {
		s := e.StartedAt.Format("2006-01-02T15:04:05Z")
		resp.StartedAt = &s
	}
	if e.CompletedAt != nil {
		s := e.CompletedAt.Format("2006-01-02T15:04:05Z")
		resp.CompletedAt = &s
	}
	durationMs := e.DurationMs()
	if durationMs > 0 {
		resp.DurationMs = &durationMs
	}

	return resp
}
