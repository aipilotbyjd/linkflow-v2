package webhook

import (
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Event represents a webhook call log entry
type Event struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EndpointID uuid.UUID  `gorm:"type:uuid;index;not null" json:"endpoint_id"`
	Method     string     `gorm:"size:10;not null" json:"method"`
	Path       string     `gorm:"size:255;not null" json:"path"`
	Headers    types.JSON `gorm:"type:jsonb" json:"headers,omitempty"`
	Body       *string    `gorm:"type:text" json:"body,omitempty"`
	IPAddress  *string    `gorm:"size:45" json:"ip_address,omitempty"`
	StatusCode int        `gorm:"not null" json:"status_code"`
	Response   *string    `gorm:"type:text" json:"response,omitempty"`
	DurationMs int        `gorm:"not null" json:"duration_ms"`
	CreatedAt  time.Time  `json:"created_at"`

	Endpoint Endpoint `gorm:"foreignKey:EndpointID" json:"-"`
}

func (Event) TableName() string {
	return "webhook_logs"
}

// NewEvent creates a new webhook event
func NewEvent(endpointID uuid.UUID, method, path string) *Event {
	return &Event{
		ID:         uuid.New(),
		EndpointID: endpointID,
		Method:     method,
		Path:       path,
		CreatedAt:  time.Now(),
	}
}

// WithHeaders sets the request headers
func (e *Event) WithHeaders(headers types.JSON) *Event {
	e.Headers = headers
	return e
}

// WithBody sets the request body
func (e *Event) WithBody(body string) *Event {
	e.Body = &body
	return e
}

// WithIPAddress sets the client IP address
func (e *Event) WithIPAddress(ip string) *Event {
	e.IPAddress = &ip
	return e
}

// SetResponse sets the response information
func (e *Event) SetResponse(statusCode int, response string, durationMs int) {
	e.StatusCode = statusCode
	e.Response = &response
	e.DurationMs = durationMs
}

// IsSuccess checks if the webhook call was successful
func (e *Event) IsSuccess() bool {
	return e.StatusCode >= 200 && e.StatusCode < 300
}

// Duration returns the duration as time.Duration
func (e *Event) Duration() time.Duration {
	return time.Duration(e.DurationMs) * time.Millisecond
}

// MarkFailed marks the event as failed with an error message
func (e *Event) MarkFailed(errorMsg string) {
	e.StatusCode = 403 // Forbidden for security failures
	e.Response = &errorMsg
}
