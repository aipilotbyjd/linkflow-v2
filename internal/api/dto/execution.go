package dto

// Execution responses

type ExecutionResponse struct {
	ID                string      `json:"id"`
	WorkflowID        string      `json:"workflow_id"`
	WorkspaceID       string      `json:"workspace_id"`
	TriggeredBy       *string     `json:"triggered_by,omitempty"`
	WorkflowVersion   int         `json:"workflow_version"`
	Status            string      `json:"status"`
	TriggerType       string      `json:"trigger_type"`
	TriggerData       interface{} `json:"trigger_data,omitempty"`
	InputData         interface{} `json:"input_data,omitempty"`
	OutputData        interface{} `json:"output_data,omitempty"`
	ErrorMessage      *string     `json:"error_message,omitempty"`
	ErrorNodeID       *string     `json:"error_node_id,omitempty"`
	NodesTotal        int         `json:"nodes_total"`
	NodesCompleted    int         `json:"nodes_completed"`
	RetryCount        int         `json:"retry_count"`
	MaxRetries        int         `json:"max_retries"`
	Priority          int         `json:"priority"`
	TimeoutSeconds    int         `json:"timeout_seconds"`
	ParentExecutionID *string     `json:"parent_execution_id,omitempty"`
	BatchID           *string     `json:"batch_id,omitempty"`
	QueuedAt          int64       `json:"queued_at"`
	StartedAt         *int64      `json:"started_at,omitempty"`
	CompletedAt       *int64      `json:"completed_at,omitempty"`
	PausedAt          *int64      `json:"paused_at,omitempty"`
	ResumedAt         *int64      `json:"resumed_at,omitempty"`
	CreatedAt         int64       `json:"created_at"`
}

type NodeExecutionResponse struct {
	ID           string      `json:"id"`
	ExecutionID  string      `json:"execution_id"`
	NodeID       string      `json:"node_id"`
	NodeType     string      `json:"node_type"`
	NodeName     *string     `json:"node_name,omitempty"`
	Status       string      `json:"status"`
	InputData    interface{} `json:"input_data,omitempty"`
	OutputData   interface{} `json:"output_data,omitempty"`
	ErrorMessage *string     `json:"error_message,omitempty"`
	DurationMs   *int        `json:"duration_ms,omitempty"`
	RetryCount   int         `json:"retry_count"`
	StartedAt    *int64      `json:"started_at,omitempty"`
	CompletedAt  *int64      `json:"completed_at,omitempty"`
	CreatedAt    int64       `json:"created_at"`
}
