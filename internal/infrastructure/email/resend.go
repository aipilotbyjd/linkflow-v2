package email

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

// ResendProvider implements Provider using Resend API
type ResendProvider struct {
	client *resend.Client
}

// NewResendProvider creates a new Resend provider
func NewResendProvider(apiKey string) *ResendProvider {
	return &ResendProvider{
		client: resend.NewClient(apiKey),
	}
}

// Send sends an email using Resend
func (p *ResendProvider) Send(ctx context.Context, msg *Message) error {
	params := &resend.SendEmailRequest{
		From:    msg.From,
		To:      msg.To,
		Subject: msg.Subject,
		Html:    msg.HTMLBody,
		Text:    msg.TextBody,
		Cc:      msg.CC,
		Bcc:     msg.BCC,
		ReplyTo: msg.ReplyTo,
	}

	// Add attachments if any
	if len(msg.Attachments) > 0 {
		attachments := make([]*resend.Attachment, len(msg.Attachments))
		for i, att := range msg.Attachments {
			attachments[i] = &resend.Attachment{
				Filename: att.Filename,
				Content:  att.Content,
			}
		}
		params.Attachments = attachments
	}

	// Add headers if any
	if len(msg.Headers) > 0 {
		params.Headers = msg.Headers
	}

	_, err := p.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}

	return nil
}
