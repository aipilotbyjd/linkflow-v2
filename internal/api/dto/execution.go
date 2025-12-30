package dto

// Execution responses

type ExecutionResponse struct {
	ID              string      `json:"id"`
	WorkflowID      string      `json:"workflow_id"`
	WorkflowVersion int         `json:"workflow_version"`
	Status          string      `json:"status"`
	TriggerType     string      `json:"trigger_type"`
	InputData       interface{} `json:"input_data,omitempty"`
	OutputData      interface{} `json:"output_data,omitempty"`
	ErrorMessage    *string     `json:"error_message,omitempty"`
	ErrorNodeID     *string     `json:"error_node_id,omitempty"`
	NodesTotal      int         `json:"nodes_total"`
	NodesCompleted  int         `json:"nodes_completed"`
	QueuedAt        int64       `json:"queued_at"`
	StartedAt       *int64      `json:"started_at,omitempty"`
	CompletedAt     *int64      `json:"completed_at,omitempty"`
}

type NodeExecutionResponse struct {
	ID           string      `json:"id"`
	NodeID       string      `json:"node_id"`
	NodeType     string      `json:"node_type"`
	NodeName     *string     `json:"node_name,omitempty"`
	Status       string      `json:"status"`
	InputData    interface{} `json:"input_data,omitempty"`
	OutputData   interface{} `json:"output_data,omitempty"`
	ErrorMessage *string     `json:"error_message,omitempty"`
	DurationMs   *int        `json:"duration_ms,omitempty"`
	StartedAt    *int64      `json:"started_at,omitempty"`
	CompletedAt  *int64      `json:"completed_at,omitempty"`
}
