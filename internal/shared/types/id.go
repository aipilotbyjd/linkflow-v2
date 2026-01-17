package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// ID is a typed UUID wrapper for domain identifiers
type ID uuid.UUID

var NilID = ID(uuid.Nil)

func NewID() ID {
	return ID(uuid.New())
}

func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return NilID, fmt.Errorf("invalid ID format: %w", err)
	}
	return ID(id), nil
}

func MustParseID(s string) ID {
	id, err := ParseID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

func (id ID) IsNil() bool {
	return uuid.UUID(id) == uuid.Nil
}

func (id ID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.UUID(id).String())
}

func (id *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return err
	}
	*id = ID(parsed)
	return nil
}

func (id ID) Value() (driver.Value, error) {
	return uuid.UUID(id).String(), nil
}

func (id *ID) Scan(value interface{}) error {
	if value == nil {
		*id = NilID
		return nil
	}
	switch v := value.(type) {
	case []byte:
		parsed, err := uuid.ParseBytes(v)
		if err != nil {
			return err
		}
		*id = ID(parsed)
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		*id = ID(parsed)
	default:
		return fmt.Errorf("cannot scan type %T into ID", value)
	}
	return nil
}

// Domain-specific ID types
type (
	UserID         ID
	WorkspaceID    ID
	WorkflowID     ID
	ExecutionID    ID
	CredentialID   ID
	ScheduleID     ID
	WebhookID      ID
	SessionID      ID
	APIKeyID       ID
	FolderID       ID
	TemplateID     ID
	SubscriptionID ID
)

// Conversion helpers
func (id UserID) UUID() uuid.UUID         { return uuid.UUID(id) }
func (id WorkspaceID) UUID() uuid.UUID    { return uuid.UUID(id) }
func (id WorkflowID) UUID() uuid.UUID     { return uuid.UUID(id) }
func (id ExecutionID) UUID() uuid.UUID    { return uuid.UUID(id) }
func (id CredentialID) UUID() uuid.UUID   { return uuid.UUID(id) }
func (id ScheduleID) UUID() uuid.UUID     { return uuid.UUID(id) }
func (id WebhookID) UUID() uuid.UUID      { return uuid.UUID(id) }
func (id SessionID) UUID() uuid.UUID      { return uuid.UUID(id) }
func (id APIKeyID) UUID() uuid.UUID       { return uuid.UUID(id) }
func (id FolderID) UUID() uuid.UUID       { return uuid.UUID(id) }
func (id TemplateID) UUID() uuid.UUID     { return uuid.UUID(id) }
func (id SubscriptionID) UUID() uuid.UUID { return uuid.UUID(id) }

func (id UserID) String() string         { return uuid.UUID(id).String() }
func (id WorkspaceID) String() string    { return uuid.UUID(id).String() }
func (id WorkflowID) String() string     { return uuid.UUID(id).String() }
func (id ExecutionID) String() string    { return uuid.UUID(id).String() }
func (id CredentialID) String() string   { return uuid.UUID(id).String() }
func (id ScheduleID) String() string     { return uuid.UUID(id).String() }
func (id WebhookID) String() string      { return uuid.UUID(id).String() }
func (id SessionID) String() string      { return uuid.UUID(id).String() }
func (id APIKeyID) String() string       { return uuid.UUID(id).String() }
func (id FolderID) String() string       { return uuid.UUID(id).String() }
func (id TemplateID) String() string     { return uuid.UUID(id).String() }
func (id SubscriptionID) String() string { return uuid.UUID(id).String() }

func (id UserID) IsNil() bool         { return uuid.UUID(id) == uuid.Nil }
func (id WorkspaceID) IsNil() bool    { return uuid.UUID(id) == uuid.Nil }
func (id WorkflowID) IsNil() bool     { return uuid.UUID(id) == uuid.Nil }
func (id ExecutionID) IsNil() bool    { return uuid.UUID(id) == uuid.Nil }
func (id CredentialID) IsNil() bool   { return uuid.UUID(id) == uuid.Nil }
func (id ScheduleID) IsNil() bool     { return uuid.UUID(id) == uuid.Nil }
func (id WebhookID) IsNil() bool      { return uuid.UUID(id) == uuid.Nil }
func (id SessionID) IsNil() bool      { return uuid.UUID(id) == uuid.Nil }
func (id APIKeyID) IsNil() bool       { return uuid.UUID(id) == uuid.Nil }
func (id FolderID) IsNil() bool       { return uuid.UUID(id) == uuid.Nil }
func (id TemplateID) IsNil() bool     { return uuid.UUID(id) == uuid.Nil }
func (id SubscriptionID) IsNil() bool { return uuid.UUID(id) == uuid.Nil }
