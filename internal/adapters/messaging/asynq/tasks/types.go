package tasks

// Task type constants
const (
	TypeWorkflowExecution = "workflow:execute"
	TypeSendEmail         = "email:send"
	TypeTokenRefresh      = "token:refresh"
	TypeCleanup           = "cleanup:run"
	TypeWebhookProcess    = "webhook:process"
)

// Queue names
const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)
