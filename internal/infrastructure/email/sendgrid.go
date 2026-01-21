package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SendGridProvider implements the Provider interface using SendGrid
type SendGridProvider struct {
	apiKey string
}

// NewSendGridProvider creates a new SendGrid provider
func NewSendGridProvider(apiKey string) *SendGridProvider {
	return &SendGridProvider{apiKey: apiKey}
}

// sendGridRequest represents the SendGrid API request structure
type sendGridRequest struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content,omitempty"`
	ReplyTo          *sendGridEmail            `json:"reply_to,omitempty"`
	Attachments      []sendGridAttachment      `json:"attachments,omitempty"`
}

type sendGridPersonalization struct {
	To  []sendGridEmail `json:"to"`
	CC  []sendGridEmail `json:"cc,omitempty"`
	BCC []sendGridEmail `json:"bcc,omitempty"`
}

type sendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type sendGridAttachment struct {
	Content     string `json:"content"`
	Filename    string `json:"filename"`
	Type        string `json:"type,omitempty"`
	Disposition string `json:"disposition,omitempty"`
}

// Send sends an email using SendGrid
func (p *SendGridProvider) Send(ctx context.Context, msg *Message) error {
	req := sendGridRequest{
		From:    sendGridEmail{Email: msg.From},
		Subject: msg.Subject,
	}

	// Add personalization
	pers := sendGridPersonalization{}
	for _, to := range msg.To {
		pers.To = append(pers.To, sendGridEmail{Email: to})
	}
	for _, cc := range msg.CC {
		pers.CC = append(pers.CC, sendGridEmail{Email: cc})
	}
	for _, bcc := range msg.BCC {
		pers.BCC = append(pers.BCC, sendGridEmail{Email: bcc})
	}
	req.Personalizations = []sendGridPersonalization{pers}

	// Add content
	if msg.TextBody != "" {
		req.Content = append(req.Content, sendGridContent{Type: "text/plain", Value: msg.TextBody})
	}
	if msg.HTMLBody != "" {
		req.Content = append(req.Content, sendGridContent{Type: "text/html", Value: msg.HTMLBody})
	}

	// Add reply-to
	if msg.ReplyTo != "" {
		req.ReplyTo = &sendGridEmail{Email: msg.ReplyTo}
	}

	// Add attachments
	for _, att := range msg.Attachments {
		req.Attachments = append(req.Attachments, sendGridAttachment{
			Content:  base64.StdEncoding.EncodeToString(att.Content),
			Filename: att.Filename,
			Type:     att.ContentType,
		})
	}

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sendgrid error: %s", string(respBody))
	}

	return nil
}
