package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/lib/pq"
)

// JSON type for JSONB columns
type JSON map[string]interface{}

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSON: not a byte slice")
	}
	return json.Unmarshal(bytes, j)
}

// JSONArray type for JSONB array columns
type JSONArray []interface{}

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONArray: not a byte slice")
	}
	return json.Unmarshal(bytes, j)
}

// StringArray type for text[] columns
type StringArray = pq.StringArray

// Workflow status constants
const (
	WorkflowStatusDraft    = "draft"
	WorkflowStatusActive   = "active"
	WorkflowStatusInactive = "inactive"
	WorkflowStatusArchived = "archived"
)

// Execution status constants
const (
	ExecutionStatusQueued    = "queued"
	ExecutionStatusRunning   = "running"
	ExecutionStatusCompleted = "completed"
	ExecutionStatusFailed    = "failed"
	ExecutionStatusCancelled = "cancelled"
	ExecutionStatusTimeout   = "timeout"
)

// Node execution status constants
const (
	NodeStatusPending   = "pending"
	NodeStatusRunning   = "running"
	NodeStatusCompleted = "completed"
	NodeStatusFailed    = "failed"
	NodeStatusSkipped   = "skipped"
)

// User status constants
const (
	UserStatusActive    = "active"
	UserStatusSuspended = "suspended"
	UserStatusDeleted   = "deleted"
)

// Workspace member roles
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

// Trigger types
const (
	TriggerManual   = "manual"
	TriggerSchedule = "schedule"
	TriggerWebhook  = "webhook"
	TriggerAPI      = "api"
)

// Credential types
const (
	CredentialTypeAPIKey  = "api_key"
	CredentialTypeOAuth2  = "oauth2"
	CredentialTypeBasic   = "basic"
	CredentialTypeBearer  = "bearer"
	CredentialTypeCustom  = "custom"
)
