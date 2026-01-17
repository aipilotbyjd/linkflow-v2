package email

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// TemplateEngine manages email templates
type TemplateEngine struct {
	htmlTemplates map[string]*template.Template
	textTemplates map[string]*template.Template
	subjects      map[string]string
}

// NewTemplateEngine creates a new template engine with built-in templates
func NewTemplateEngine() (*TemplateEngine, error) {
	e := &TemplateEngine{
		htmlTemplates: make(map[string]*template.Template),
		textTemplates: make(map[string]*template.Template),
		subjects:      make(map[string]string),
	}

	// Register built-in templates
	templates := []struct {
		name    string
		subject string
		html    string
		text    string
	}{
		{
			name:    "welcome",
			subject: "Welcome to LinkFlow!",
			html:    welcomeHTMLTemplate,
			text:    welcomeTextTemplate,
		},
		{
			name:    "reset_password",
			subject: "Reset Your Password",
			html:    resetPasswordHTMLTemplate,
			text:    resetPasswordTextTemplate,
		},
		{
			name:    "invitation",
			subject: "You've been invited to join {{.WorkspaceName}}",
			html:    invitationHTMLTemplate,
			text:    invitationTextTemplate,
		},
		{
			name:    "execution_failed",
			subject: "Workflow Execution Failed: {{.WorkflowName}}",
			html:    executionFailedHTMLTemplate,
			text:    executionFailedTextTemplate,
		},
		{
			name:    "verify_email",
			subject: "Verify Your Email Address",
			html:    verifyEmailHTMLTemplate,
			text:    verifyEmailTextTemplate,
		},
	}

	for _, t := range templates {
		if err := e.RegisterTemplate(t.name, t.subject, t.html, t.text); err != nil {
			return nil, fmt.Errorf("failed to register template %s: %w", t.name, err)
		}
	}

	return e, nil
}

// RegisterTemplate registers a new template
func (e *TemplateEngine) RegisterTemplate(name, subject, html, text string) error {
	e.subjects[name] = subject

	if html != "" {
		tmpl, err := template.New(name).Parse(html)
		if err != nil {
			return fmt.Errorf("failed to parse HTML template: %w", err)
		}
		e.htmlTemplates[name] = tmpl
	}

	if text != "" {
		tmpl, err := template.New(name).Parse(text)
		if err != nil {
			return fmt.Errorf("failed to parse text template: %w", err)
		}
		e.textTemplates[name] = tmpl
	}

	return nil
}

// RenderHTML renders an HTML template
func (e *TemplateEngine) RenderHTML(name string, data map[string]interface{}) (string, error) {
	tmpl, ok := e.htmlTemplates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderText renders a text template
func (e *TemplateEngine) RenderText(name string, data map[string]interface{}) (string, error) {
	tmpl, ok := e.textTemplates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetSubject returns the subject for a template
func (e *TemplateEngine) GetSubject(name string, data map[string]interface{}) string {
	subject, ok := e.subjects[name]
	if !ok {
		return ""
	}

	// Replace template variables in subject
	for key, value := range data {
		placeholder := "{{." + key + "}}"
		subject = strings.ReplaceAll(subject, placeholder, fmt.Sprint(value))
	}

	return subject
}

// Built-in templates
var welcomeHTMLTemplate = `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1>Welcome to LinkFlow, {{.Name}}!</h1>
<p>Thank you for signing up. We're excited to have you on board!</p>
<p>Get started by creating your first workflow.</p>
<p>Best,<br>The LinkFlow Team</p>
</body>
</html>`

var welcomeTextTemplate = `Welcome to LinkFlow, {{.Name}}!

Thank you for signing up. We're excited to have you on board!

Get started by creating your first workflow.

Best,
The LinkFlow Team`

var resetPasswordHTMLTemplate = `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1>Reset Your Password</h1>
<p>Hi {{.Name}},</p>
<p>We received a request to reset your password. Click the link below to create a new password:</p>
<p><a href="{{.ResetURL}}" style="background: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">Reset Password</a></p>
<p>This link will expire in 1 hour.</p>
<p>If you didn't request this, please ignore this email.</p>
</body>
</html>`

var resetPasswordTextTemplate = `Reset Your Password

Hi {{.Name}},

We received a request to reset your password. Visit the link below to create a new password:

{{.ResetURL}}

This link will expire in 1 hour.

If you didn't request this, please ignore this email.`

var invitationHTMLTemplate = `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1>You're Invited!</h1>
<p>{{.InviterName}} has invited you to join <strong>{{.WorkspaceName}}</strong> on LinkFlow.</p>
<p><a href="{{.InviteURL}}" style="background: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">Accept Invitation</a></p>
</body>
</html>`

var invitationTextTemplate = `You're Invited!

{{.InviterName}} has invited you to join {{.WorkspaceName}} on LinkFlow.

Accept the invitation: {{.InviteURL}}`

var executionFailedHTMLTemplate = `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1 style="color: #DC2626;">Workflow Execution Failed</h1>
<p>Your workflow <strong>{{.WorkflowName}}</strong> failed to execute.</p>
<p><strong>Execution ID:</strong> {{.ExecutionID}}</p>
<p><strong>Error:</strong></p>
<pre style="background: #F3F4F6; padding: 12px; border-radius: 4px;">{{.ErrorMessage}}</pre>
</body>
</html>`

var executionFailedTextTemplate = `Workflow Execution Failed

Your workflow "{{.WorkflowName}}" failed to execute.

Execution ID: {{.ExecutionID}}

Error: {{.ErrorMessage}}`

var verifyEmailHTMLTemplate = `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1>Verify Your Email</h1>
<p>Hi {{.Name}},</p>
<p>Please verify your email address by clicking the button below:</p>
<p><a href="{{.VerifyURL}}" style="background: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">Verify Email</a></p>
</body>
</html>`

var verifyEmailTextTemplate = `Verify Your Email

Hi {{.Name}},

Please verify your email address by visiting:

{{.VerifyURL}}`
