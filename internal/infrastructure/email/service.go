package email

import (
	"context"
	"fmt"
)

// Service provides email sending functionality
type Service struct {
	provider    Provider
	templates   *TemplateEngine
	defaultFrom string
}

// Config holds email service configuration
type Config struct {
	Provider     string
	DefaultFrom  string
	APIKey       string
	Domain       string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
	ResendAPIKey string
}

// NewService creates a new email service
func NewService(config Config) (*Service, error) {
	var provider Provider

	switch config.Provider {
	case "sendgrid":
		provider = NewSendGridProvider(config.APIKey)
	case "resend":
		provider = NewResendProvider(config.ResendAPIKey)
	case "smtp":
		provider = NewSMTPProvider(SMTPConfig{
			Host:     config.SMTPHost,
			Port:     config.SMTPPort,
			Username: config.SMTPUser,
			Password: config.SMTPPass,
		})
	case "console", "log":
		provider = &ConsoleProvider{}
	default:
		provider = &NoopProvider{}
	}

	templates, err := NewTemplateEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize templates: %w", err)
	}

	return &Service{
		provider:    provider,
		templates:   templates,
		defaultFrom: config.DefaultFrom,
	}, nil
}

// Message represents an email message
type Message struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	ReplyTo     string
	Subject     string
	TextBody    string
	HTMLBody    string
	Attachments []Attachment
	Headers     map[string]string
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

// Send sends an email message
func (s *Service) Send(ctx context.Context, msg *Message) error {
	if msg.From == "" {
		msg.From = s.defaultFrom
	}
	return s.provider.Send(ctx, msg)
}

// SendTemplate sends a templated email
func (s *Service) SendTemplate(ctx context.Context, to []string, templateName string, data map[string]interface{}) error {
	html, err := s.templates.RenderHTML(templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	text, _ := s.templates.RenderText(templateName, data)

	subject := s.templates.GetSubject(templateName, data)

	return s.Send(ctx, &Message{
		To:       to,
		Subject:  subject,
		HTMLBody: html,
		TextBody: text,
	})
}

// SendWelcome sends a welcome email
func (s *Service) SendWelcome(ctx context.Context, email, name string) error {
	return s.SendTemplate(ctx, []string{email}, "welcome", map[string]interface{}{
		"Name": name,
	})
}

// SendPasswordReset sends a password reset email
func (s *Service) SendPasswordReset(ctx context.Context, email, name, resetURL string) error {
	return s.SendTemplate(ctx, []string{email}, "reset_password", map[string]interface{}{
		"Name":     name,
		"ResetURL": resetURL,
	})
}

// SendInvitation sends a workspace invitation email
func (s *Service) SendInvitation(ctx context.Context, email, inviterName, workspaceName, inviteURL string) error {
	return s.SendTemplate(ctx, []string{email}, "invitation", map[string]interface{}{
		"InviterName":   inviterName,
		"WorkspaceName": workspaceName,
		"InviteURL":     inviteURL,
	})
}

// SendExecutionFailed sends an execution failure notification
func (s *Service) SendExecutionFailed(ctx context.Context, email, workflowName, executionID, errorMessage string) error {
	return s.SendTemplate(ctx, []string{email}, "execution_failed", map[string]interface{}{
		"WorkflowName": workflowName,
		"ExecutionID":  executionID,
		"ErrorMessage": errorMessage,
	})
}

// SendVerification sends an email verification email
func (s *Service) SendVerification(ctx context.Context, email, name, verifyURL string) error {
	return s.SendTemplate(ctx, []string{email}, "verify_email", map[string]interface{}{
		"Name":      name,
		"VerifyURL": verifyURL,
	})
}

// Provider is the interface for email providers
type Provider interface {
	Send(ctx context.Context, msg *Message) error
}

// NoopProvider is a no-op email provider
type NoopProvider struct{}

func (p *NoopProvider) Send(ctx context.Context, msg *Message) error {
	return nil
}

// ConsoleProvider logs emails to console (for development)
type ConsoleProvider struct{}

func (p *ConsoleProvider) Send(ctx context.Context, msg *Message) error {
	fmt.Printf("=== Email ===\nFrom: %s\nTo: %v\nSubject: %s\nBody:\n%s\n=============\n",
		msg.From, msg.To, msg.Subject, msg.TextBody)
	return nil
}
