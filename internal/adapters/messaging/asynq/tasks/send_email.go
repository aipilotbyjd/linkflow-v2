package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// SendEmailPayload contains data for email sending task
type SendEmailPayload struct {
	To          []string               `json:"to"`
	Subject     string                 `json:"subject"`
	Template    string                 `json:"template"`
	Data        map[string]interface{} `json:"data,omitempty"`
	HTMLContent string                 `json:"html_content,omitempty"`
	TextContent string                 `json:"text_content,omitempty"`
}

// NewSendEmailTask creates a new send email task
func NewSendEmailTask(payload SendEmailPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return asynq.NewTask(
		TypeSendEmail,
		data,
		asynq.MaxRetry(5),
		asynq.Timeout(2*time.Minute),
		asynq.Queue(QueueDefault),
	), nil
}

// SendEmailHandler handles email sending tasks
type SendEmailHandler struct {
	sender EmailSender
}

// EmailSender interface for sending emails
type EmailSender interface {
	Send(ctx context.Context, to []string, subject, template string, data map[string]interface{}) error
	SendHTML(ctx context.Context, to []string, subject, htmlContent, textContent string) error
}

// NewSendEmailHandler creates a new handler
func NewSendEmailHandler(sender EmailSender) *SendEmailHandler {
	return &SendEmailHandler{sender: sender}
}

// ProcessTask processes an email sending task
func (h *SendEmailHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload SendEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.Template != "" {
		return h.sender.Send(ctx, payload.To, payload.Subject, payload.Template, payload.Data)
	}

	return h.sender.SendHTML(ctx, payload.To, payload.Subject, payload.HTMLContent, payload.TextContent)
}
