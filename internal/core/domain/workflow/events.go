package workflow

import (
	"time"

	"github.com/google/uuid"
)

// Event types
const (
	EventWorkflowCreated     = "workflow.created"
	EventWorkflowUpdated     = "workflow.updated"
	EventWorkflowDeleted     = "workflow.deleted"
	EventWorkflowActivated   = "workflow.activated"
	EventWorkflowDeactivated = "workflow.deactivated"
	EventWorkflowCloned      = "workflow.cloned"
	EventVersionCreated      = "workflow.version_created"
	EventVersionRolledBack   = "workflow.version_rolled_back"
)

// WorkflowCreated event
type WorkflowCreated struct {
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	CreatedBy   uuid.UUID `json:"created_by"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e WorkflowCreated) EventType() string     { return EventWorkflowCreated }
func (e WorkflowCreated) AggregateID() string   { return e.WorkflowID.String() }
func (e WorkflowCreated) OccurredAt() time.Time { return e.Timestamp }

// WorkflowUpdated event
type WorkflowUpdated struct {
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Version     int       `json:"version"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e WorkflowUpdated) EventType() string     { return EventWorkflowUpdated }
func (e WorkflowUpdated) AggregateID() string   { return e.WorkflowID.String() }
func (e WorkflowUpdated) OccurredAt() time.Time { return e.Timestamp }

// WorkflowDeleted event
type WorkflowDeleted struct {
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	DeletedBy   uuid.UUID `json:"deleted_by"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e WorkflowDeleted) EventType() string     { return EventWorkflowDeleted }
func (e WorkflowDeleted) AggregateID() string   { return e.WorkflowID.String() }
func (e WorkflowDeleted) OccurredAt() time.Time { return e.Timestamp }

// WorkflowActivated event
type WorkflowActivated struct {
	WorkflowID  uuid.UUID `json:"workflow_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	ActivatedBy uuid.UUID `json:"activated_by"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e WorkflowActivated) EventType() string     { return EventWorkflowActivated }
func (e WorkflowActivated) AggregateID() string   { return e.WorkflowID.String() }
func (e WorkflowActivated) OccurredAt() time.Time { return e.Timestamp }

// WorkflowDeactivated event
type WorkflowDeactivated struct {
	WorkflowID    uuid.UUID `json:"workflow_id"`
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	DeactivatedBy uuid.UUID `json:"deactivated_by"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e WorkflowDeactivated) EventType() string     { return EventWorkflowDeactivated }
func (e WorkflowDeactivated) AggregateID() string   { return e.WorkflowID.String() }
func (e WorkflowDeactivated) OccurredAt() time.Time { return e.Timestamp }

// WorkflowCloned event
type WorkflowCloned struct {
	WorkflowID       uuid.UUID `json:"workflow_id"`
	SourceWorkflowID uuid.UUID `json:"source_workflow_id"`
	WorkspaceID      uuid.UUID `json:"workspace_id"`
	Name             string    `json:"name"`
	ClonedBy         uuid.UUID `json:"cloned_by"`
	Timestamp        time.Time `json:"timestamp"`
}

func (e WorkflowCloned) EventType() string     { return EventWorkflowCloned }
func (e WorkflowCloned) AggregateID() string   { return e.WorkflowID.String() }
func (e WorkflowCloned) OccurredAt() time.Time { return e.Timestamp }

// VersionCreated event
type VersionCreated struct {
	WorkflowID uuid.UUID `json:"workflow_id"`
	VersionID  uuid.UUID `json:"version_id"`
	Version    int       `json:"version"`
	CreatedBy  uuid.UUID `json:"created_by"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e VersionCreated) EventType() string     { return EventVersionCreated }
func (e VersionCreated) AggregateID() string   { return e.WorkflowID.String() }
func (e VersionCreated) OccurredAt() time.Time { return e.Timestamp }

// VersionRolledBack event
type VersionRolledBack struct {
	WorkflowID   uuid.UUID `json:"workflow_id"`
	FromVersion  int       `json:"from_version"`
	ToVersion    int       `json:"to_version"`
	RolledBackBy uuid.UUID `json:"rolled_back_by"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e VersionRolledBack) EventType() string     { return EventVersionRolledBack }
func (e VersionRolledBack) AggregateID() string   { return e.WorkflowID.String() }
func (e VersionRolledBack) OccurredAt() time.Time { return e.Timestamp }
