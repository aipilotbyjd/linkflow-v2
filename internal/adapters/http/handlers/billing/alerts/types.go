package alerts

import (
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// CreateAlertRequest for creating a new alert
type CreateAlertRequest struct {
	AlertType string         `json:"alert_type" validate:"required,oneof=operations ai_credits storage data_transfer"`
	Threshold int            `json:"threshold" validate:"required,oneof=50 75 90 100"`
	Channels  ChannelsConfig `json:"channels" validate:"required"`
}

// UpdateAlertRequest for updating an alert
type UpdateAlertRequest struct {
	Enabled  *bool          `json:"enabled,omitempty"`
	Channels *ChannelsConfig `json:"channels,omitempty"`
}

// ChannelsConfig for alert delivery channels
type ChannelsConfig struct {
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

// AlertResponse for API responses
type AlertResponse struct {
	ID          string         `json:"id"`
	AlertType   string         `json:"alert_type"`
	Threshold   int            `json:"threshold"`
	Severity    string         `json:"severity"`
	Enabled     bool           `json:"enabled"`
	Channels    ChannelsConfig `json:"channels"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// AlertLogResponse for alert history
type AlertLogResponse struct {
	ID             string     `json:"id"`
	AlertType      string     `json:"alert_type"`
	Severity       string     `json:"severity"`
	Threshold      int        `json:"threshold"`
	CurrentUsage   int64      `json:"current_usage"`
	Limit          int64      `json:"limit"`
	Percentage     float64    `json:"percentage"`
	Message        string     `json:"message"`
	SentAt         time.Time  `json:"sent_at"`
	ChannelsSent   []string   `json:"channels_sent"`
	Acknowledged   bool       `json:"acknowledged"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
}

// ToAlertResponse converts domain to response
func ToAlertResponse(a *billing.UsageAlert) AlertResponse {
	return AlertResponse{
		ID:        a.ID.String(),
		AlertType: string(a.AlertType),
		Threshold: a.Threshold,
		Severity:  string(a.Severity),
		Enabled:   a.Enabled,
		Channels: ChannelsConfig{
			Email:        a.Channels.Email,
			EmailAddrs:   a.Channels.EmailAddrs,
			Slack:        a.Channels.Slack,
			SlackWebhook: a.Channels.SlackWebhook,
			Webhook:      a.Channels.Webhook,
			WebhookURL:   a.Channels.WebhookURL,
			InApp:        a.Channels.InApp,
			SMS:          a.Channels.SMS,
			PhoneNumbers: a.Channels.PhoneNumbers,
		},
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// ToAlertLogResponse converts domain to response
func ToAlertLogResponse(l *billing.UsageAlertLog) AlertLogResponse {
	return AlertLogResponse{
		ID:             l.ID.String(),
		AlertType:      string(l.AlertType),
		Severity:       string(l.Severity),
		Threshold:      l.Threshold,
		CurrentUsage:   l.CurrentUsage,
		Limit:          l.Limit,
		Percentage:     l.Percentage,
		Message:        l.Message,
		SentAt:         l.SentAt,
		ChannelsSent:   l.ChannelsSent,
		Acknowledged:   l.Acknowledged,
		AcknowledgedAt: l.AcknowledgedAt,
	}
}
