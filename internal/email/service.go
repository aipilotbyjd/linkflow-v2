package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

type Service struct {
	config      *Config
	templates   map[string]*template.Template
	queueClient *asynq.Client
}

type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
	UseTLS       bool
	UseSTARTTLS  bool
	TemplateDir  string
	QueueEnabled bool
}

type Email struct {
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []Attachment
	Headers     map[string]string
	ReplyTo     string
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type TemplateData map[string]interface{}

func NewService(config *Config, queueClient *asynq.Client) *Service {
	s := &Service{
		config:      config,
		templates:   make(map[string]*template.Template),
		queueClient: queueClient,
	}
	s.loadBuiltinTemplates()
	return s
}

func (s *Service) loadBuiltinTemplates() {
	templates := map[string]string{
		"welcome":              welcomeTemplate,
		"email_verification":   emailVerificationTemplate,
		"password_reset":       passwordResetTemplate,
		"workspace_invitation": workspaceInvitationTemplate,
		"execution_failed":     executionFailedTemplate,
		"usage_warning":        usageWarningTemplate,
		"billing_alert":        billingAlertTemplate,
	}

	for name, content := range templates {
		tmpl, err := template.New(name).Parse(content)
		if err == nil {
			s.templates[name] = tmpl
		}
	}
}

func (s *Service) Send(ctx context.Context, email *Email) error {
	if s.config.QueueEnabled && s.queueClient != nil {
		return s.enqueue(email)
	}
	return s.sendDirect(email)
}

func (s *Service) enqueue(email *Email) error {
	payload, err := json.Marshal(email)
	if err != nil {
		return err
	}

	task := asynq.NewTask("email:send", payload)
	_, err = s.queueClient.Enqueue(task,
		asynq.Queue("emails"),
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Second),
	)
	return err
}

func (s *Service) sendDirect(email *Email) error {
	msg := s.buildMessage(email)

	var auth smtp.Auth
	if s.config.SMTPUser != "" {
		auth = smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)
	}

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	if s.config.UseTLS {
		return s.sendWithTLS(addr, auth, email.To, msg)
	}

	if s.config.UseSTARTTLS {
		return s.sendWithSTARTTLS(addr, auth, email.To, msg)
	}

	return smtp.SendMail(addr, auth, s.config.FromEmail, email.To, msg)
}

func (s *Service) sendWithTLS(addr string, auth smtp.Auth, to []string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName: s.config.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(s.config.FromEmail); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

func (s *Service) sendWithSTARTTLS(addr string, auth smtp.Auth, to []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName: s.config.SMTPHost,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return err
	}

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(s.config.FromEmail); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

func (s *Service) buildMessage(email *Email) []byte {
	var buf bytes.Buffer

	from := s.config.FromEmail
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	}

	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))

	if len(email.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.CC, ", ")))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))

	if email.ReplyTo != "" {
		buf.WriteString(fmt.Sprintf("Reply-To: %s\r\n", email.ReplyTo))
	}

	for k, v := range email.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	boundary := "===============BOUNDARY==============="

	if email.HTMLBody != "" || len(email.Attachments) > 0 {
		buf.WriteString("MIME-Version: 1.0\r\n")
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
		buf.WriteString("\r\n")

		if email.HTMLBody != "" && email.Body != "" {
			altBoundary := "===============ALT==============="
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", altBoundary))

			buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
			buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
			buf.WriteString(email.Body)
			buf.WriteString("\r\n")

			buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
			buf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
			buf.WriteString(email.HTMLBody)
			buf.WriteString("\r\n")

			buf.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))
		} else if email.HTMLBody != "" {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
			buf.WriteString(email.HTMLBody)
			buf.WriteString("\r\n")
		} else {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
			buf.WriteString(email.Body)
			buf.WriteString("\r\n")
		}

		for _, att := range email.Attachments {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.ContentType, att.Filename))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Filename))
			buf.Write(att.Data)
			buf.WriteString("\r\n")
		}

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
		buf.WriteString(email.Body)
	}

	return buf.Bytes()
}

func (s *Service) SendTemplate(ctx context.Context, templateName string, to []string, subject string, data TemplateData) error {
	tmpl, ok := s.templates[templateName]
	if !ok {
		return fmt.Errorf("template %s not found", templateName)
	}

	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		return err
	}

	email := &Email{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBuf.String(),
	}

	return s.Send(ctx, email)
}

func (s *Service) SendWelcome(ctx context.Context, to, name string) error {
	return s.SendTemplate(ctx, "welcome", []string{to}, "Welcome to LinkFlow!", TemplateData{
		"Name":     name,
		"AppName":  "LinkFlow",
		"LoginURL": "https://app.linkflow.ai/login",
	})
}

func (s *Service) SendEmailVerification(ctx context.Context, to, name, token string) error {
	return s.SendTemplate(ctx, "email_verification", []string{to}, "Verify your email", TemplateData{
		"Name":      name,
		"AppName":   "LinkFlow",
		"VerifyURL": fmt.Sprintf("https://app.linkflow.ai/verify-email?token=%s", token),
		"ExpiresIn": "24 hours",
	})
}

func (s *Service) SendPasswordReset(ctx context.Context, to, name, token string) error {
	return s.SendTemplate(ctx, "password_reset", []string{to}, "Reset your password", TemplateData{
		"Name":     name,
		"AppName":  "LinkFlow",
		"ResetURL": fmt.Sprintf("https://app.linkflow.ai/reset-password?token=%s", token),
		"ExpiresIn": "1 hour",
	})
}

func (s *Service) SendWorkspaceInvitation(ctx context.Context, to, inviterName, workspaceName, token string) error {
	return s.SendTemplate(ctx, "workspace_invitation", []string{to}, fmt.Sprintf("You've been invited to %s", workspaceName), TemplateData{
		"InviterName":   inviterName,
		"WorkspaceName": workspaceName,
		"AppName":       "LinkFlow",
		"AcceptURL":     fmt.Sprintf("https://app.linkflow.ai/accept-invite?token=%s", token),
		"ExpiresIn":     "7 days",
	})
}

func (s *Service) SendExecutionFailed(ctx context.Context, to, workflowName, executionID, errorMsg string) error {
	return s.SendTemplate(ctx, "execution_failed", []string{to}, fmt.Sprintf("Workflow '%s' failed", workflowName), TemplateData{
		"WorkflowName": workflowName,
		"ExecutionID":  executionID,
		"ErrorMessage": errorMsg,
		"AppName":      "LinkFlow",
		"ViewURL":      fmt.Sprintf("https://app.linkflow.ai/executions/%s", executionID),
		"Timestamp":    time.Now().Format(time.RFC1123),
	})
}

func (s *Service) SendUsageWarning(ctx context.Context, to, workspaceName string, usagePercent int, resourceType string) error {
	return s.SendTemplate(ctx, "usage_warning", []string{to}, fmt.Sprintf("Usage warning: %s at %d%%", resourceType, usagePercent), TemplateData{
		"WorkspaceName": workspaceName,
		"UsagePercent":  usagePercent,
		"ResourceType":  resourceType,
		"AppName":       "LinkFlow",
		"UpgradeURL":    "https://app.linkflow.ai/settings/billing",
	})
}

func (s *Service) SendBillingAlert(ctx context.Context, to string, alertType string, data TemplateData) error {
	subjects := map[string]string{
		"payment_failed":      "Payment failed",
		"subscription_ending": "Your subscription is ending soon",
		"invoice_ready":       "Your invoice is ready",
	}

	subject := subjects[alertType]
	if subject == "" {
		subject = "Billing notification"
	}

	data["AppName"] = "LinkFlow"
	data["BillingURL"] = "https://app.linkflow.ai/settings/billing"

	return s.SendTemplate(ctx, "billing_alert", []string{to}, subject, data)
}
