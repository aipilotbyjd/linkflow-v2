package events

import (
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event
type Event interface {
	EventName() string
	OccurredAt() time.Time
	AggregateID() uuid.UUID
	AggregateType() string
}

// BaseEvent provides common event fields
type BaseEvent struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Timestamp     time.Time `json:"timestamp"`
	AggregateUUID uuid.UUID `json:"aggregate_id"`
	AggregateName string    `json:"aggregate_type"`
}

func NewBaseEvent(name string, aggregateID uuid.UUID, aggregateType string) BaseEvent {
	return BaseEvent{
		ID:            uuid.New(),
		Name:          name,
		Timestamp:     time.Now().UTC(),
		AggregateUUID: aggregateID,
		AggregateName: aggregateType,
	}
}

func (e BaseEvent) EventName() string       { return e.Name }
func (e BaseEvent) OccurredAt() time.Time   { return e.Timestamp }
func (e BaseEvent) AggregateID() uuid.UUID  { return e.AggregateUUID }
func (e BaseEvent) AggregateType() string   { return e.AggregateName }

// Common domain events

// UserRegistered event
type UserRegistered struct {
	BaseEvent
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

// UserLoggedIn event
type UserLoggedIn struct {
	BaseEvent
	UserID    uuid.UUID `json:"user_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

// WorkspaceCreated event
type WorkspaceCreated struct {
	BaseEvent
	WorkspaceID uuid.UUID `json:"workspace_id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
}

// WorkspaceMemberInvited event
type WorkspaceMemberInvited struct {
	BaseEvent
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	InvitedBy   uuid.UUID `json:"invited_by"`
}

// WorkflowCreated event
type WorkflowCreated struct {
	BaseEvent
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	CreatedBy   uuid.UUID `json:"created_by"`
}

// WorkflowActivated event
type WorkflowActivated struct {
	BaseEvent
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Version     int       `json:"version"`
}

// WorkflowDeactivated event
type WorkflowDeactivated struct {
	BaseEvent
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
}

// ExecutionStarted event
type ExecutionStarted struct {
	BaseEvent
	ExecutionID uuid.UUID `json:"execution_id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	TriggerType string    `json:"trigger_type"`
}

// ExecutionCompleted event
type ExecutionCompleted struct {
	BaseEvent
	ExecutionID uuid.UUID `json:"execution_id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Status      string    `json:"status"`
	DurationMs  int64     `json:"duration_ms"`
}

// ExecutionFailed event
type ExecutionFailed struct {
	BaseEvent
	ExecutionID  uuid.UUID `json:"execution_id"`
	WorkflowID   uuid.UUID `json:"workflow_id"`
	WorkspaceID  uuid.UUID `json:"workspace_id"`
	ErrorMessage string    `json:"error_message"`
	ErrorNodeID  string    `json:"error_node_id,omitempty"`
}

// CredentialCreated event
type CredentialCreated struct {
	BaseEvent
	CredentialID uuid.UUID `json:"credential_id"`
	WorkspaceID  uuid.UUID `json:"workspace_id"`
	Type         string    `json:"type"`
	CreatedBy    uuid.UUID `json:"created_by"`
}

// ScheduleCreated event
type ScheduleCreated struct {
	BaseEvent
	ScheduleID     uuid.UUID `json:"schedule_id"`
	WorkflowID     uuid.UUID `json:"workflow_id"`
	WorkspaceID    uuid.UUID `json:"workspace_id"`
	CronExpression string    `json:"cron_expression"`
}

// ScheduleTriggered event
type ScheduleTriggered struct {
	BaseEvent
	ScheduleID  uuid.UUID `json:"schedule_id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	ExecutionID uuid.UUID `json:"execution_id"`
}

// WebhookTriggered event
type WebhookTriggered struct {
	BaseEvent
	EndpointID  uuid.UUID `json:"endpoint_id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	ExecutionID uuid.UUID `json:"execution_id"`
	Method      string    `json:"method"`
}
