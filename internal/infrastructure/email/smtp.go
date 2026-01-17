package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// SMTPProvider implements the Provider interface using SMTP
type SMTPProvider struct {
	config SMTPConfig
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
	Timeout  time.Duration
}

// NewSMTPProvider creates a new SMTP provider
func NewSMTPProvider(config SMTPConfig) *SMTPProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &SMTPProvider{config: config}
}

// Send sends an email using SMTP
func (p *SMTPProvider) Send(ctx context.Context, msg *Message) error {
	addr := net.JoinHostPort(p.config.Host, fmt.Sprintf("%d", p.config.Port))

	// Set up authentication
	var auth smtp.Auth
	if p.config.Username != "" {
		auth = smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)
	}

	// Build the message
	message := p.buildMessage(msg)

	// Get all recipients
	recipients := make([]string, 0, len(msg.To)+len(msg.CC)+len(msg.BCC))
	recipients = append(recipients, msg.To...)
	recipients = append(recipients, msg.CC...)
	recipients = append(recipients, msg.BCC...)

	// Connect with timeout
	conn, err := net.DialTimeout("tcp", addr, p.config.Timeout)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, p.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Start TLS if required
	if p.config.UseTLS {
		tlsConfig := &tls.Config{
			ServerName: p.config.Host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(msg.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", rcpt, err)
		}
	}

	// Send message body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %w", err)
	}

	if _, err := w.Write(message); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data connection: %w", err)
	}

	return client.Quit()
}

func (p *SMTPProvider) buildMessage(msg *Message) []byte {
	var sb strings.Builder

	// Headers
	sb.WriteString(fmt.Sprintf("From: %s\r\n", msg.From))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	if len(msg.CC) > 0 {
		sb.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.CC, ", ")))
	}
	if msg.ReplyTo != "" {
		sb.WriteString(fmt.Sprintf("Reply-To: %s\r\n", msg.ReplyTo))
	}
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	sb.WriteString("MIME-Version: 1.0\r\n")

	// Custom headers
	for key, value := range msg.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// Content
	if msg.HTMLBody != "" && msg.TextBody != "" {
		// Multipart message
		boundary := fmt.Sprintf("boundary-%d", time.Now().UnixNano())
		sb.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

		// Text part
		sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		sb.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
		sb.WriteString(msg.TextBody)
		sb.WriteString("\r\n")

		// HTML part
		sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		sb.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
		sb.WriteString(msg.HTMLBody)
		sb.WriteString("\r\n")

		sb.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if msg.HTMLBody != "" {
		sb.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
		sb.WriteString(msg.HTMLBody)
	} else {
		sb.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
		sb.WriteString(msg.TextBody)
	}

	return []byte(sb.String())
}
