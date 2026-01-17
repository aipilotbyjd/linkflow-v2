package websocket

import "github.com/google/uuid"

const (
	// Client events
	EventSubscribe   = "subscribe"
	EventUnsubscribe = "unsubscribe"
	EventPing        = "ping"
	EventPong        = "pong"

	// Workflow events
	EventWorkflowCreated     = "workflow.created"
	EventWorkflowUpdated     = "workflow.updated"
	EventWorkflowDeleted     = "workflow.deleted"
	EventWorkflowActivated   = "workflow.activated"
	EventWorkflowDeactivated = "workflow.deactivated"

	// Execution events
	EventExecutionStarted   = "execution.started"
	EventExecutionCompleted = "execution.completed"
	EventExecutionFailed    = "execution.failed"
	EventExecutionCancelled = "execution.cancelled"
	EventNodeStarted        = "node.started"
	EventNodeCompleted      = "node.completed"
	EventNodeFailed         = "node.failed"

	// Credential events
	EventCredentialCreated = "credential.created"
	EventCredentialUpdated = "credential.updated"
	EventCredentialDeleted = "credential.deleted"

	// Schedule events
	EventScheduleCreated   = "schedule.created"
	EventScheduleUpdated   = "schedule.updated"
	EventScheduleDeleted   = "schedule.deleted"
	EventSchedulePaused    = "schedule.paused"
	EventScheduleResumed   = "schedule.resumed"
	EventScheduleTriggered = "schedule.triggered"

	// Webhook events
	EventWebhookTriggered = "webhook.triggered"

	// Workspace events
	EventWorkspaceUpdated  = "workspace.updated"
	EventMemberAdded       = "workspace.member_added"
	EventMemberRemoved     = "workspace.member_removed"
	EventMemberRoleChanged = "workspace.member_role_changed"

	// Notification events
	EventNotification = "notification"
)

type Message struct {
	Event       string      `json:"event"`
	Data        interface{} `json:"data,omitempty"`
	WorkspaceID uuid.UUID   `json:"workspace_id,omitempty"`
	UserID      uuid.UUID   `json:"user_id,omitempty"`
	Timestamp   int64       `json:"timestamp,omitempty"`
}

// Event data structures
type WorkflowEventData struct {
	WorkflowID   uuid.UUID `json:"workflow_id"`
	WorkflowName string    `json:"workflow_name"`
	Version      int       `json:"version,omitempty"`
}

type ExecutionEventData struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	Status      string    `json:"status"`
	TriggerType string    `json:"trigger_type,omitempty"`
	Error       string    `json:"error,omitempty"`
	Duration    int64     `json:"duration_ms,omitempty"`
}

type NodeEventData struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	NodeID      string    `json:"node_id"`
	NodeType    string    `json:"node_type"`
	NodeName    string    `json:"node_name"`
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
	Duration    int64     `json:"duration_ms,omitempty"`
}

type NotificationData struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    string `json:"type"` // info, success, warning, error
	Link    string `json:"link,omitempty"`
}
