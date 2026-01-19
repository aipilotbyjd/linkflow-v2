package webhook

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Endpoint entity (aggregate root)
type Endpoint struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID   uuid.UUID      `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	NodeID       string         `gorm:"size:100;not null" json:"node_id"`
	Path         string         `gorm:"size:255;uniqueIndex;not null" json:"path"`
	Method       string         `gorm:"size:10;default:POST" json:"method"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Secret       *string        `gorm:"size:255" json:"-"`
	LastCalledAt *time.Time     `json:"last_called_at,omitempty"`
	CallCount    int            `gorm:"default:0" json:"call_count"`

	// Security settings
	RequireSignature   bool   `gorm:"default:false" json:"require_signature"`
	SignatureHeader    string `gorm:"size:100;default:'X-Webhook-Signature'" json:"signature_header,omitempty"`
	AllowedIPs         string `gorm:"type:text" json:"allowed_ips,omitempty"`          // Comma-separated IPs/CIDRs
	RequireTimestamp   bool   `gorm:"default:false" json:"require_timestamp"`
	TimestampHeader    string `gorm:"size:100;default:'X-Webhook-Timestamp'" json:"timestamp_header,omitempty"`
	TimestampMaxAgeSec int    `gorm:"default:300" json:"timestamp_max_age_sec,omitempty"` // Default 5 min
	RequireNonce       bool   `gorm:"default:false" json:"require_nonce"`
	NonceHeader        string `gorm:"size:100;default:'X-Webhook-Nonce'" json:"nonce_header,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Endpoint) TableName() string {
	return "webhook_endpoints"
}

// NewEndpoint creates a new webhook endpoint
func NewEndpoint(workflowID, workspaceID uuid.UUID, nodeID, path string) *Endpoint {
	return &Endpoint{
		ID:          uuid.New(),
		WorkflowID:  workflowID,
		WorkspaceID: workspaceID,
		NodeID:      nodeID,
		Path:        path,
		Method:      "POST",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// GetWorkspaceID implements the WorkspaceOwned interface
func (e *Endpoint) GetWorkspaceID() uuid.UUID {
	return e.WorkspaceID
}

// WithMethod sets the HTTP method
func (e *Endpoint) WithMethod(method string) *Endpoint {
	e.Method = method
	return e
}

// WithSecret sets the webhook secret
func (e *Endpoint) WithSecret(secret string) *Endpoint {
	e.Secret = &secret
	return e
}

// Activate activates the endpoint
func (e *Endpoint) Activate() {
	e.IsActive = true
	e.UpdatedAt = time.Now()
}

// Deactivate deactivates the endpoint
func (e *Endpoint) Deactivate() {
	e.IsActive = false
	e.UpdatedAt = time.Now()
}

// RegenerateSecret sets a new secret
func (e *Endpoint) RegenerateSecret(secret string) {
	e.Secret = &secret
	e.UpdatedAt = time.Now()
}

// RecordCall records a webhook call
func (e *Endpoint) RecordCall() {
	now := time.Now()
	e.LastCalledAt = &now
	e.CallCount++
	e.UpdatedAt = now
}

// UpdatePath updates the webhook path
func (e *Endpoint) UpdatePath(path string) {
	e.Path = path
	e.UpdatedAt = time.Now()
}

// HasSecret checks if the endpoint has a secret configured
func (e *Endpoint) HasSecret() bool {
	return e.Secret != nil && *e.Secret != ""
}

// GetURL returns the full webhook URL
func (e *Endpoint) GetURL(baseURL string) string {
	return baseURL + "/webhooks/" + e.Path
}

// GetAllowedIPsList returns allowed IPs as a slice
func (e *Endpoint) GetAllowedIPsList() []string {
	if e.AllowedIPs == "" {
		return nil
	}
	ips := make([]string, 0)
	for _, ip := range splitAndTrim(e.AllowedIPs, ",") {
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

// SetAllowedIPs sets allowed IPs from a slice
func (e *Endpoint) SetAllowedIPs(ips []string) {
	e.AllowedIPs = joinNonEmpty(ips, ",")
	e.UpdatedAt = time.Now()
}

// EnableSignatureVerification enables HMAC signature verification
func (e *Endpoint) EnableSignatureVerification(secret string) {
	e.RequireSignature = true
	e.Secret = &secret
	e.UpdatedAt = time.Now()
}

// DisableSignatureVerification disables signature verification
func (e *Endpoint) DisableSignatureVerification() {
	e.RequireSignature = false
	e.UpdatedAt = time.Now()
}

// EnableReplayProtection enables timestamp and nonce validation
func (e *Endpoint) EnableReplayProtection(maxAgeSec int) {
	e.RequireTimestamp = true
	e.RequireNonce = true
	if maxAgeSec > 0 {
		e.TimestampMaxAgeSec = maxAgeSec
	}
	e.UpdatedAt = time.Now()
}

// DisableReplayProtection disables replay protection
func (e *Endpoint) DisableReplayProtection() {
	e.RequireTimestamp = false
	e.RequireNonce = false
	e.UpdatedAt = time.Now()
}

// IsSecured returns true if any security measure is enabled
func (e *Endpoint) IsSecured() bool {
	return e.RequireSignature || e.RequireTimestamp || e.RequireNonce || e.AllowedIPs != ""
}

// Helper functions
func splitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := make([]string, 0)
	for _, p := range stringsSplit(s, sep) {
		trimmed := stringsTrim(p)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func stringsSplit(s, sep string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func stringsTrim(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func joinNonEmpty(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if p != "" {
			if i > 0 && result != "" {
				result += sep
			}
			result += p
		}
	}
	return result
}
