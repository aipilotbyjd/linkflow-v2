package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"github.com/rs/zerolog/log"
)

// NotificationService handles sending usage alerts via various channels
type NotificationService struct {
	emailSender   EmailSender
	slackClient   SlackClient
	webhookClient *http.Client
	alertLogRepo  billing.UsageAlertLogRepository
}

// EmailSender interface for sending emails
type EmailSender interface {
	SendEmail(ctx context.Context, to []string, subject, body string) error
}

// SlackClient interface for sending Slack messages
type SlackClient interface {
	SendMessage(ctx context.Context, webhookURL, message string) error
}

// NewNotificationService creates a new notification service
func NewNotificationService(emailSender EmailSender, slackClient SlackClient, alertLogRepo billing.UsageAlertLogRepository) *NotificationService {
	return &NotificationService{
		emailSender:   emailSender,
		slackClient:   slackClient,
		webhookClient: &http.Client{Timeout: 10 * time.Second},
		alertLogRepo:  alertLogRepo,
	}
}

// SendAlert sends an alert through configured channels
func (s *NotificationService) SendAlert(ctx context.Context, alert *billing.UsageAlert, currentUsage, limit int64) error {
	message := alert.AlertMessage(currentUsage, limit)
	channels := alert.Channels
	sentChannels := []string{}

	// Create alert log
	alertLog := billing.NewAlertLog(
		alert.WorkspaceID,
		alert.ID,
		alert.AlertType,
		alert.Severity,
		alert.Threshold,
		currentUsage,
		limit,
	)

	// Send via Email
	if channels.Email && len(channels.EmailAddrs) > 0 {
		subject := s.getEmailSubject(alert)
		body := s.getEmailBody(alert, currentUsage, limit)
		if err := s.emailSender.SendEmail(ctx, channels.EmailAddrs, subject, body); err == nil {
			sentChannels = append(sentChannels, "email")
		}
	}

	// Send via Slack
	if channels.Slack && channels.SlackWebhook != "" {
		slackMsg := s.getSlackMessage(alert, currentUsage, limit)
		if err := s.slackClient.SendMessage(ctx, channels.SlackWebhook, slackMsg); err == nil {
			sentChannels = append(sentChannels, "slack")
		}
	}

	// Send via Webhook
	if channels.Webhook && channels.WebhookURL != "" {
		if err := s.sendWebhook(ctx, channels.WebhookURL, alert, currentUsage, limit); err == nil {
			sentChannels = append(sentChannels, "webhook")
		}
	}

	// Always mark in-app as sent if enabled
	if channels.InApp {
		sentChannels = append(sentChannels, "in_app")
	}

	// Save alert log
	alertLog.ChannelsSent = sentChannels
	alertLog.Message = message
	if s.alertLogRepo != nil {
		if err := s.alertLogRepo.Create(ctx, alertLog); err != nil {
			log.Error().Err(err).Msg("failed to create alert log")
		}
	}

	return nil
}

// SendOverageAlert sends a special overage notification
func (s *NotificationService) SendOverageAlert(ctx context.Context, workspaceID uuid.UUID, alertType billing.UsageAlertType, overageAmount int64, channels billing.AlertChannels) error {
	message := fmt.Sprintf("⚠️ OVERAGE ALERT: You have exceeded your %s limit by %d. Additional charges of $%.2f will apply at the end of your billing cycle.",
		alertType,
		overageAmount,
		float64(overageAmount)*float64(billing.DefaultOverageRates.OverageMultiplier)*0.01,
	)

	if channels.Email && len(channels.EmailAddrs) > 0 {
		subject := fmt.Sprintf("[ACTION REQUIRED] %s Overage Alert", alertType)
		if err := s.emailSender.SendEmail(ctx, channels.EmailAddrs, subject, message); err != nil {
			log.Error().Err(err).Msg("failed to send overage alert email")
		}
	}

	if channels.Slack && channels.SlackWebhook != "" {
		if err := s.slackClient.SendMessage(ctx, channels.SlackWebhook, message); err != nil {
			log.Error().Err(err).Msg("failed to send overage alert slack message")
		}
	}

	return nil
}

// getEmailSubject generates email subject based on severity
func (s *NotificationService) getEmailSubject(alert *billing.UsageAlert) string {
	prefix := ""
	switch alert.Severity {
	case billing.AlertSeverityUrgent:
		prefix = "🚨 [URGENT] "
	case billing.AlertSeverityCritical:
		prefix = "⚠️ [CRITICAL] "
	case billing.AlertSeverityWarning:
		prefix = "⚡ [WARNING] "
	default:
		prefix = "ℹ️ "
	}

	return fmt.Sprintf("%sUsage Alert: %d%% of %s limit reached", prefix, alert.Threshold, alert.AlertType)
}

// getEmailBody generates HTML email body
func (s *NotificationService) getEmailBody(alert *billing.UsageAlert, currentUsage, limit int64) string {
	percentage := float64(currentUsage) / float64(limit) * 100
	resourceName := string(alert.AlertType)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: %s; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 20px; border-radius: 0 0 8px 8px; }
        .progress-bar { background: #e0e0e0; height: 20px; border-radius: 10px; overflow: hidden; }
        .progress { background: %s; height: 100%%; width: %.1f%%; }
        .stats { margin: 20px 0; }
        .stat { display: inline-block; margin-right: 30px; }
        .stat-value { font-size: 24px; font-weight: bold; }
        .cta { text-align: center; margin-top: 20px; }
        .btn { display: inline-block; background: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Usage Alert</h1>
            <p>%d%% of your monthly %s limit reached</p>
        </div>
        <div class="content">
            <div class="progress-bar">
                <div class="progress"></div>
            </div>
            <div class="stats">
                <div class="stat">
                    <div class="stat-value">%d</div>
                    <div>Used</div>
                </div>
                <div class="stat">
                    <div class="stat-value">%d</div>
                    <div>Limit</div>
                </div>
                <div class="stat">
                    <div class="stat-value">%d</div>
                    <div>Remaining</div>
                </div>
            </div>
            <p>%s</p>
            <div class="cta">
                <a href="https://app.linkflow.ai/settings/billing" class="btn">Manage Plan</a>
            </div>
        </div>
    </div>
</body>
</html>`,
		s.getSeverityColor(alert.Severity),
		s.getProgressColor(percentage),
		percentage,
		alert.Threshold,
		resourceName,
		currentUsage,
		limit,
		limit-currentUsage,
		alert.AlertMessage(currentUsage, limit),
	)
}

// getSlackMessage generates Slack message with blocks
func (s *NotificationService) getSlackMessage(alert *billing.UsageAlert, currentUsage, limit int64) string {
	percentage := float64(currentUsage) / float64(limit) * 100
	emoji := s.getSeverityEmoji(alert.Severity)

	return fmt.Sprintf(`{
		"blocks": [
			{
				"type": "header",
				"text": {
					"type": "plain_text",
					"text": "%s Usage Alert: %s",
					"emoji": true
				}
			},
			{
				"type": "section",
				"fields": [
					{"type": "mrkdwn", "text": "*Resource:*\n%s"},
					{"type": "mrkdwn", "text": "*Usage:*\n%.1f%%"},
					{"type": "mrkdwn", "text": "*Used:*\n%d"},
					{"type": "mrkdwn", "text": "*Limit:*\n%d"}
				]
			},
			{
				"type": "section",
				"text": {
					"type": "mrkdwn",
					"text": "%s"
				}
			},
			{
				"type": "actions",
				"elements": [
					{
						"type": "button",
						"text": {"type": "plain_text", "text": "View Usage"},
						"url": "https://app.linkflow.ai/settings/usage"
					},
					{
						"type": "button",
						"text": {"type": "plain_text", "text": "Upgrade Plan"},
						"url": "https://app.linkflow.ai/settings/billing",
						"style": "primary"
					}
				]
			}
		]
	}`, emoji, alert.AlertType, alert.AlertType, percentage, currentUsage, limit, alert.AlertMessage(currentUsage, limit))
}

// sendWebhook sends alert to custom webhook
func (s *NotificationService) sendWebhook(ctx context.Context, url string, alert *billing.UsageAlert, currentUsage, limit int64) error {
	payload := map[string]interface{}{
		"event":         "usage_alert",
		"workspace_id":  alert.WorkspaceID.String(),
		"alert_type":    alert.AlertType,
		"severity":      alert.Severity,
		"threshold":     alert.Threshold,
		"current_usage": currentUsage,
		"limit":         limit,
		"percentage":    float64(currentUsage) / float64(limit) * 100,
		"message":       alert.AlertMessage(currentUsage, limit),
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "LinkFlow-Webhook/1.0")

	resp, err := s.webhookClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *NotificationService) getSeverityColor(severity billing.AlertSeverity) string {
	switch severity {
	case billing.AlertSeverityUrgent:
		return "#DC2626" // Red
	case billing.AlertSeverityCritical:
		return "#F59E0B" // Orange
	case billing.AlertSeverityWarning:
		return "#FBBF24" // Yellow
	default:
		return "#3B82F6" // Blue
	}
}

func (s *NotificationService) getProgressColor(percentage float64) string {
	if percentage >= 100 {
		return "#DC2626"
	} else if percentage >= 90 {
		return "#F59E0B"
	} else if percentage >= 75 {
		return "#FBBF24"
	}
	return "#10B981"
}

func (s *NotificationService) getSeverityEmoji(severity billing.AlertSeverity) string {
	switch severity {
	case billing.AlertSeverityUrgent:
		return "🚨"
	case billing.AlertSeverityCritical:
		return "⚠️"
	case billing.AlertSeverityWarning:
		return "⚡"
	default:
		return "ℹ️"
	}
}
