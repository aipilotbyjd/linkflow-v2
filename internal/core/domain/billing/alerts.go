package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UsageAlertType defines the type of usage being tracked
type UsageAlertType string

const (
	AlertTypeOperations   UsageAlertType = "operations"
	AlertTypeAICredits    UsageAlertType = "ai_credits"
	AlertTypeStorage      UsageAlertType = "storage"
	AlertTypeDataTransfer UsageAlertType = "data_transfer"
)

// AlertSeverity defines urgency level of alerts
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityUrgent   AlertSeverity = "urgent"
)

// UsageAlert represents a usage threshold alert configuration
type UsageAlert struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	AlertType   UsageAlertType `gorm:"size:50;not null" json:"alert_type"`
	Threshold   int            `gorm:"not null" json:"threshold"` // Percentage (50, 75, 90, 100)
	Severity    AlertSeverity  `gorm:"size:20;not null" json:"severity"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	Channels    AlertChannels  `gorm:"type:jsonb" json:"channels"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UsageAlert) TableName() string {
	return "usage_alerts"
}

// AlertChannels defines where alerts are sent
type AlertChannels struct {
	Email        bool     `json:"email"`
	EmailAddrs   []string `json:"email_addresses,omitempty"`
	Slack        bool     `json:"slack"`
	SlackWebhook string   `json:"slack_webhook,omitempty"`
	Webhook      bool     `json:"webhook"`
	WebhookURL   string   `json:"webhook_url,omitempty"`
	InApp        bool     `json:"in_app"`
	SMS          bool     `json:"sms"`
	PhoneNumbers []string `json:"phone_numbers,omitempty"`
}

// UsageAlertLog records when alerts were triggered
type UsageAlertLog struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	AlertID        uuid.UUID      `gorm:"type:uuid;index" json:"alert_id"`
	AlertType      UsageAlertType `gorm:"size:50;not null" json:"alert_type"`
	Severity       AlertSeverity  `gorm:"size:20" json:"severity"`
	Threshold      int            `json:"threshold"`
	CurrentUsage   int64          `json:"current_usage"`
	Limit          int64          `json:"limit"`
	Percentage     float64        `json:"percentage"`
	Message        string         `gorm:"size:1000" json:"message"`
	SentAt         time.Time      `gorm:"index" json:"sent_at"`
	ChannelsSent   []string       `gorm:"type:jsonb" json:"channels_sent"`
	Acknowledged   bool           `gorm:"default:false" json:"acknowledged"`
	AcknowledgedAt *time.Time     `json:"acknowledged_at,omitempty"`
	AcknowledgedBy *uuid.UUID     `gorm:"type:uuid" json:"acknowledged_by,omitempty"`
}

func (UsageAlertLog) TableName() string {
	return "usage_alert_logs"
}

// Default alert thresholds with severity mapping
var DefaultAlertThresholds = []struct {
	Threshold int
	Severity  AlertSeverity
}{
	{50, AlertSeverityInfo},
	{75, AlertSeverityWarning},
	{90, AlertSeverityCritical},
	{100, AlertSeverityUrgent},
}

// NewUsageAlert creates a new usage alert configuration
func NewUsageAlert(workspaceID uuid.UUID, alertType UsageAlertType, threshold int) *UsageAlert {
	severity := AlertSeverityInfo
	for _, t := range DefaultAlertThresholds {
		if t.Threshold == threshold {
			severity = t.Severity
			break
		}
	}

	return &UsageAlert{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		AlertType:   alertType,
		Threshold:   threshold,
		Severity:    severity,
		Enabled:     true,
		Channels: AlertChannels{
			Email: true,
			InApp: true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateDefaultAlerts creates default alerts for a workspace
func CreateDefaultAlerts(workspaceID uuid.UUID) []*UsageAlert {
	var alerts []*UsageAlert

	alertTypes := []UsageAlertType{AlertTypeOperations, AlertTypeAICredits}

	for _, alertType := range alertTypes {
		for _, t := range DefaultAlertThresholds {
			alert := NewUsageAlert(workspaceID, alertType, t.Threshold)
			alert.Severity = t.Severity
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// GenerateAlertMessage generates a human-readable alert message
func GenerateAlertMessage(alertType UsageAlertType, threshold int, currentUsage, limit int64) string {
	percentage := float64(currentUsage) / float64(limit) * 100
	resourceName := formatResourceName(alertType)

	switch threshold {
	case 50:
		return fmt.Sprintf("You've used 50%% (%d/%d) of your monthly %s. Keep an eye on your usage.",
			currentUsage, limit, resourceName)
	case 75:
		return fmt.Sprintf("Warning: You've used 75%% (%d/%d) of your monthly %s. Consider upgrading your plan.",
			currentUsage, limit, resourceName)
	case 90:
		return fmt.Sprintf("Critical: You've used 90%% (%d/%d) of your monthly %s. Upgrade now to avoid service interruption.",
			currentUsage, limit, resourceName)
	case 100:
		return fmt.Sprintf("You've reached 100%% of your monthly %s limit (%d). Overage charges may apply or service may be limited.",
			resourceName, limit)
	default:
		return fmt.Sprintf("Usage alert: %s at %.1f%% (%d/%d)",
			resourceName, percentage, currentUsage, limit)
	}
}

// AlertMessage generates alert message (method version)
func (a *UsageAlert) AlertMessage(currentUsage, limit int64) string {
	return GenerateAlertMessage(a.AlertType, a.Threshold, currentUsage, limit)
}

// NewAlertLog creates a new alert log entry
func NewAlertLog(workspaceID, alertID uuid.UUID, alertType UsageAlertType, severity AlertSeverity, threshold int, currentUsage, limit int64) *UsageAlertLog {
	percentage := float64(currentUsage) / float64(limit) * 100
	message := GenerateAlertMessage(alertType, threshold, currentUsage, limit)

	return &UsageAlertLog{
		ID:           uuid.New(),
		WorkspaceID:  workspaceID,
		AlertID:      alertID,
		AlertType:    alertType,
		Severity:     severity,
		Threshold:    threshold,
		CurrentUsage: currentUsage,
		Limit:        limit,
		Percentage:   percentage,
		Message:      message,
		SentAt:       time.Now(),
		ChannelsSent: []string{},
	}
}

// Acknowledge marks an alert as acknowledged
func (a *UsageAlertLog) Acknowledge(userID uuid.UUID) {
	now := time.Now()
	a.Acknowledged = true
	a.AcknowledgedAt = &now
	a.AcknowledgedBy = &userID
}

// formatResourceName returns a human-readable name for the resource
func formatResourceName(alertType UsageAlertType) string {
	switch alertType {
	case AlertTypeOperations:
		return "operations"
	case AlertTypeAICredits:
		return "AI credits"
	case AlertTypeStorage:
		return "storage"
	case AlertTypeDataTransfer:
		return "data transfer"
	default:
		return string(alertType)
	}
}

// UsageAlertRepository defines the interface for alert persistence
type UsageAlertRepository interface {
	Create(ctx context.Context, alert *UsageAlert) error
	Update(ctx context.Context, alert *UsageAlert) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*UsageAlert, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]*UsageAlert, error)
	FindByWorkspaceAndType(ctx context.Context, workspaceID uuid.UUID, alertType UsageAlertType) ([]*UsageAlert, error)
}

// UsageAlertLogRepository defines the interface for alert log persistence
type UsageAlertLogRepository interface {
	Create(ctx context.Context, log *UsageAlertLog) error
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*UsageAlertLog, int64, error)
	FindUnacknowledged(ctx context.Context, workspaceID uuid.UUID) ([]*UsageAlertLog, error)
	Acknowledge(ctx context.Context, id, userID uuid.UUID) error
}
