package integrations

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

type EmailNode struct{}

func (n *EmailNode) Type() string {
	return "integration.email"
}

func (n *EmailNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	operation := getString(config, "operation", "send")

	switch operation {
	case "send":
		return n.sendEmail(ctx, execCtx, config)
	case "sendHTML":
		return n.sendHTMLEmail(ctx, execCtx, config)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *EmailNode) sendEmail(ctx context.Context, execCtx *core.ExecutionContext, config map[string]interface{}) (map[string]interface{}, error) {
	// Get SMTP credential
	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		return nil, fmt.Errorf("credential is required")
	}

	credID, err := uuid.Parse(credIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid credential ID")
	}

	cred, err := execCtx.GetCredential(credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	// SMTP config from credential
	host := getString(config, "host", "")
	port := getString(config, "port", "587")
	username := cred.Username
	password := cred.Password

	if host == "" {
		host = cred.Custom["host"]
	}
	if port == "587" && cred.Custom["port"] != "" {
		port = cred.Custom["port"]
	}

	// Email parameters
	from := getString(config, "from", username)
	to := getString(config, "to", "")
	cc := getString(config, "cc", "")
	bcc := getString(config, "bcc", "")
	replyTo := getString(config, "replyTo", "")
	subject := getString(config, "subject", "")
	body := getString(config, "body", "")

	// Build recipients list
	var recipients []string
	if to != "" {
		recipients = append(recipients, parseEmails(to)...)
	}
	if cc != "" {
		recipients = append(recipients, parseEmails(cc)...)
	}
	if bcc != "" {
		recipients = append(recipients, parseEmails(bcc)...)
	}

	if len(recipients) == 0 {
		return nil, fmt.Errorf("no recipients specified")
	}

	// Build message
	msg := buildPlainTextEmail(from, to, cc, replyTo, subject, body)

	// Send email
	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", username, password, host)

	var sendErr error
	if port == "465" {
		// SSL
		sendErr = sendMailSSL(addr, auth, from, recipients, msg)
	} else {
		// TLS (STARTTLS)
		sendErr = sendMailTLS(addr, auth, from, recipients, msg, host)
	}

	if sendErr != nil {
		return nil, fmt.Errorf("failed to send email: %w", sendErr)
	}

	return map[string]interface{}{
		"sent":       true,
		"to":         to,
		"cc":         cc,
		"bcc":        bcc,
		"subject":    subject,
		"recipients": len(recipients),
	}, nil
}

func (n *EmailNode) sendHTMLEmail(ctx context.Context, execCtx *core.ExecutionContext, config map[string]interface{}) (map[string]interface{}, error) {
	// Get SMTP credential
	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		return nil, fmt.Errorf("credential is required")
	}

	credID, err := uuid.Parse(credIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid credential ID")
	}

	cred, err := execCtx.GetCredential(credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	host := getString(config, "host", "")
	port := getString(config, "port", "587")
	username := cred.Username
	password := cred.Password

	if host == "" && cred.Custom != nil {
		host = cred.Custom["host"]
	}

	from := getString(config, "from", username)
	to := getString(config, "to", "")
	cc := getString(config, "cc", "")
	bcc := getString(config, "bcc", "")
	replyTo := getString(config, "replyTo", "")
	subject := getString(config, "subject", "")
	htmlBody := getString(config, "html", "")
	textBody := getString(config, "text", "")

	var recipients []string
	if to != "" {
		recipients = append(recipients, parseEmails(to)...)
	}
	if cc != "" {
		recipients = append(recipients, parseEmails(cc)...)
	}
	if bcc != "" {
		recipients = append(recipients, parseEmails(bcc)...)
	}

	if len(recipients) == 0 {
		return nil, fmt.Errorf("no recipients specified")
	}

	msg := buildHTMLEmail(from, to, cc, replyTo, subject, textBody, htmlBody)

	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", username, password, host)

	var sendErr error
	if port == "465" {
		sendErr = sendMailSSL(addr, auth, from, recipients, msg)
	} else {
		sendErr = sendMailTLS(addr, auth, from, recipients, msg, host)
	}

	if sendErr != nil {
		return nil, fmt.Errorf("failed to send email: %w", sendErr)
	}

	return map[string]interface{}{
		"sent":       true,
		"to":         to,
		"cc":         cc,
		"subject":    subject,
		"recipients": len(recipients),
	}, nil
}

func parseEmails(emails string) []string {
	var result []string
	for _, email := range strings.Split(emails, ",") {
		email = strings.TrimSpace(email)
		if email != "" {
			result = append(result, email)
		}
	}
	return result
}

func buildPlainTextEmail(from, to, cc, replyTo, subject, body string) []byte {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	if cc != "" {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", cc))
	}
	if replyTo != "" {
		msg.WriteString(fmt.Sprintf("Reply-To: %s\r\n", replyTo))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)
	return []byte(msg.String())
}

func buildHTMLEmail(from, to, cc, replyTo, subject, textBody, htmlBody string) []byte {
	boundary := "boundary-linkflow-email"

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	if cc != "" {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", cc))
	}
	if replyTo != "" {
		msg.WriteString(fmt.Sprintf("Reply-To: %s\r\n", replyTo))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")

	// Plain text part
	if textBody != "" {
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(textBody)
		msg.WriteString("\r\n")
	}

	// HTML part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n")

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return []byte(msg.String())
}

func sendMailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	conn, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// STARTTLS
	tlsConfig := &tls.Config{ServerName: host}
	if err := conn.StartTLS(tlsConfig); err != nil {
		return err
	}

	if err := conn.Auth(auth); err != nil {
		return err
	}

	if err := conn.Mail(from); err != nil {
		return err
	}

	for _, recipient := range to {
		if err := conn.Rcpt(recipient); err != nil {
			return err
		}
	}

	w, err := conn.Data()
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

	return conn.Quit()
}

func sendMailSSL(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return err
	}

	if err := client.Mail(from); err != nil {
		return err
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
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

var _ core.Node = (*EmailNode)(nil)
