package billing

import (
	"time"

	"github.com/google/uuid"
)

// UsageAlert represents a usage threshold alert configuration
type UsageAlert struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	WorkspaceID uuid.UUID       `gorm:"type:uuid;index;not null" json:"workspace_id"`
	AlertType   UsageAlertType  `gorm:"size:50;not null" json:"alert_type"`
	Threshold   int             `gorm:"not null" json:"threshold"` // Percentage (50, 75, 90, 100)
	Enabled     bool            `gorm:"default:true" json:"enabled"`
	Channels    AlertChannels   `gorm:"type:jsonb" json:"channels"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type UsageAlertType string

const (
	AlertTypeOperations  UsageAlertType = "operations"
	AlertTypeAICredits   UsageAlertType = "ai_credits"
	AlertTypeStorage     UsageAlertType = "storage"
	AlertTypeDataTransfer UsageAlertType = "data_transfer"
)

// AlertChannels defines where alerts are sent
type AlertChannels struct {
	Email       bool     `json:"email"`
	EmailAddrs  []string `json:"email_addresses,omitempty"`
	Slack       bool     `json:"slack"`
	SlackWebhook string  `json:"slack_webhook,omitempty"`
	Webhook     bool     `json:"webhook"`
	WebhookURL  string   `json:"webhook_url,omitempty"`
	InApp       bool     `json:"in_app"`
}

// UsageAlertLog records when alerts were triggered
type UsageAlertLog struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	AlertID     uuid.UUID      `gorm:"type:uuid;index" json:"alert_id"`
	AlertType   UsageAlertType `gorm:"size:50;not null" json:"alert_type"`
	Threshold   int            `json:"threshold"`
	CurrentUsage int64         `json:"current_usage"`
	Limit       int64          `json:"limit"`
	Percentage  float64        `json:"percentage"`
	Message     string         `gorm:"size:500" json:"message"`
	SentAt      time.Time      `json:"sent_at"`
	Channels    []string       `gorm:"type:jsonb" json:"channels_sent"`
}

// Default alert thresholds
var DefaultAlertThresholds = []int{50, 75, 90, 100}

// NewUsageAlert creates a new usage alert
func NewUsageAlert(workspaceID uuid.UUID, alertType UsageAlertType, threshold int) *UsageAlert {
	return &UsageAlert{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		AlertType:   alertType,
		Threshold:   threshold,
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
	
	for _, threshold := range DefaultAlertThresholds {
		alerts = append(alerts, NewUsageAlert(workspaceID, AlertTypeOperations, threshold))
		alerts = append(alerts, NewUsageAlert(workspaceID, AlertTypeAICredits, threshold))
	}
	
	return alerts
}

// AlertMessage generates alert message based on type and threshold
func (a *UsageAlert) AlertMessage(currentUsage, limit int64) string {
	percentage := float64(currentUsage) / float64(limit) * 100
	
	switch a.Threshold {
	case 50:
		return "You've used 50% of your monthly " + string(a.AlertType) + ". Consider monitoring your usage."
	case 75:
		return "Warning: You've used 75% of your monthly " + string(a.AlertType) + ". You may need to upgrade soon."
	case 90:
		return "Critical: You've used 90% of your monthly " + string(a.AlertType) + ". Upgrade now to avoid interruption."
	case 100:
		return "You've reached 100% of your monthly " + string(a.AlertType) + " limit."
	default:
		return "Usage alert: " + string(a.AlertType) + " at " + string(rune(percentage)) + "%"
	}
}
