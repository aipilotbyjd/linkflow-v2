package workspace

import (
	"time"

	"github.com/google/uuid"
)

// Event types
const (
	EventWorkspaceCreated       = "workspace.created"
	EventWorkspaceUpdated       = "workspace.updated"
	EventWorkspaceDeleted       = "workspace.deleted"
	EventMemberInvited          = "workspace.member_invited"
	EventMemberJoined           = "workspace.member_joined"
	EventMemberRemoved          = "workspace.member_removed"
	EventMemberRoleChanged      = "workspace.member_role_changed"
	EventWorkspacePlanChanged   = "workspace.plan_changed"
)

// WorkspaceCreated event
type WorkspaceCreated struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Plan        string    `json:"plan"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e WorkspaceCreated) EventType() string    { return EventWorkspaceCreated }
func (e WorkspaceCreated) AggregateID() string  { return e.WorkspaceID.String() }
func (e WorkspaceCreated) OccurredAt() time.Time { return e.Timestamp }

// WorkspaceUpdated event
type WorkspaceUpdated struct {
	WorkspaceID uuid.UUID         `json:"workspace_id"`
	Changes     map[string]string `json:"changes"`
	UpdatedBy   uuid.UUID         `json:"updated_by"`
	Timestamp   time.Time         `json:"timestamp"`
}

func (e WorkspaceUpdated) EventType() string    { return EventWorkspaceUpdated }
func (e WorkspaceUpdated) AggregateID() string  { return e.WorkspaceID.String() }
func (e WorkspaceUpdated) OccurredAt() time.Time { return e.Timestamp }

// WorkspaceDeleted event
type WorkspaceDeleted struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	DeletedBy   uuid.UUID `json:"deleted_by"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e WorkspaceDeleted) EventType() string    { return EventWorkspaceDeleted }
func (e WorkspaceDeleted) AggregateID() string  { return e.WorkspaceID.String() }
func (e WorkspaceDeleted) OccurredAt() time.Time { return e.Timestamp }

// MemberInvited event
type MemberInvited struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	InvitedBy   uuid.UUID `json:"invited_by"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MemberInvited) EventType() string    { return EventMemberInvited }
func (e MemberInvited) AggregateID() string  { return e.WorkspaceID.String() }
func (e MemberInvited) OccurredAt() time.Time { return e.Timestamp }

// MemberJoined event
type MemberJoined struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MemberJoined) EventType() string    { return EventMemberJoined }
func (e MemberJoined) AggregateID() string  { return e.WorkspaceID.String() }
func (e MemberJoined) OccurredAt() time.Time { return e.Timestamp }

// MemberRemoved event
type MemberRemoved struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	UserID      uuid.UUID `json:"user_id"`
	RemovedBy   uuid.UUID `json:"removed_by"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MemberRemoved) EventType() string    { return EventMemberRemoved }
func (e MemberRemoved) AggregateID() string  { return e.WorkspaceID.String() }
func (e MemberRemoved) OccurredAt() time.Time { return e.Timestamp }

// MemberRoleChanged event
type MemberRoleChanged struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	UserID      uuid.UUID `json:"user_id"`
	OldRole     string    `json:"old_role"`
	NewRole     string    `json:"new_role"`
	ChangedBy   uuid.UUID `json:"changed_by"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MemberRoleChanged) EventType() string    { return EventMemberRoleChanged }
func (e MemberRoleChanged) AggregateID() string  { return e.WorkspaceID.String() }
func (e MemberRoleChanged) OccurredAt() time.Time { return e.Timestamp }

// WorkspacePlanChanged event
type WorkspacePlanChanged struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	OldPlan     string    `json:"old_plan"`
	NewPlan     string    `json:"new_plan"`
	ChangedBy   uuid.UUID `json:"changed_by"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e WorkspacePlanChanged) EventType() string    { return EventWorkspacePlanChanged }
func (e WorkspacePlanChanged) AggregateID() string  { return e.WorkspaceID.String() }
func (e WorkspacePlanChanged) OccurredAt() time.Time { return e.Timestamp }
