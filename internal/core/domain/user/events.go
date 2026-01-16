package user

import (
	"time"

	"github.com/google/uuid"
)

// Event types
const (
	EventUserRegistered      = "user.registered"
	EventUserLoggedIn        = "user.logged_in"
	EventUserLoggedOut       = "user.logged_out"
	EventUserUpdated         = "user.updated"
	EventUserDeleted         = "user.deleted"
	EventUserEmailVerified   = "user.email_verified"
	EventUserPasswordChanged = "user.password_changed"
	EventUserMFAEnabled      = "user.mfa_enabled"
	EventUserMFADisabled     = "user.mfa_disabled"
)

// UserRegistered event
type UserRegistered struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Timestamp time.Time `json:"timestamp"`
}

func (e UserRegistered) EventType() string { return EventUserRegistered }
func (e UserRegistered) AggregateID() string { return e.UserID.String() }
func (e UserRegistered) OccurredAt() time.Time { return e.Timestamp }

// UserLoggedIn event
type UserLoggedIn struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
}

func (e UserLoggedIn) EventType() string { return EventUserLoggedIn }
func (e UserLoggedIn) AggregateID() string { return e.UserID.String() }
func (e UserLoggedIn) OccurredAt() time.Time { return e.Timestamp }

// UserLoggedOut event
type UserLoggedOut struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e UserLoggedOut) EventType() string { return EventUserLoggedOut }
func (e UserLoggedOut) AggregateID() string { return e.UserID.String() }
func (e UserLoggedOut) OccurredAt() time.Time { return e.Timestamp }

// UserUpdated event
type UserUpdated struct {
	UserID    uuid.UUID         `json:"user_id"`
	Changes   map[string]string `json:"changes"`
	Timestamp time.Time         `json:"timestamp"`
}

func (e UserUpdated) EventType() string { return EventUserUpdated }
func (e UserUpdated) AggregateID() string { return e.UserID.String() }
func (e UserUpdated) OccurredAt() time.Time { return e.Timestamp }

// UserEmailVerified event
type UserEmailVerified struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

func (e UserEmailVerified) EventType() string { return EventUserEmailVerified }
func (e UserEmailVerified) AggregateID() string { return e.UserID.String() }
func (e UserEmailVerified) OccurredAt() time.Time { return e.Timestamp }

// UserPasswordChanged event
type UserPasswordChanged struct {
	UserID    uuid.UUID `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e UserPasswordChanged) EventType() string { return EventUserPasswordChanged }
func (e UserPasswordChanged) AggregateID() string { return e.UserID.String() }
func (e UserPasswordChanged) OccurredAt() time.Time { return e.Timestamp }

// UserMFAEnabled event
type UserMFAEnabled struct {
	UserID    uuid.UUID `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e UserMFAEnabled) EventType() string { return EventUserMFAEnabled }
func (e UserMFAEnabled) AggregateID() string { return e.UserID.String() }
func (e UserMFAEnabled) OccurredAt() time.Time { return e.Timestamp }

// UserMFADisabled event
type UserMFADisabled struct {
	UserID    uuid.UUID `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e UserMFADisabled) EventType() string { return EventUserMFADisabled }
func (e UserMFADisabled) AggregateID() string { return e.UserID.String() }
func (e UserMFADisabled) OccurredAt() time.Time { return e.Timestamp }
