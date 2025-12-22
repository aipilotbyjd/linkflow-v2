package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// SendGridNode handles SendGrid email operations
type SendGridNode struct{}

func (n *SendGridNode) Type() string {
	return "integrations.sendgrid"
}

func (n *SendGridNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	apiKey := core.GetString(config, "apiKey", "")
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey is required")
	}

	operation := core.GetString(config, "operation", "send")

	switch operation {
	case "send":
		return n.sendEmail(ctx, config, apiKey, execCtx.Input)
	case "sendTemplate":
		return n.sendTemplateEmail(ctx, config, apiKey, execCtx.Input)
	case "getContacts":
		return n.getContacts(ctx, config, apiKey)
	case "addContact":
		return n.addContact(ctx, config, apiKey, execCtx.Input)
	case "deleteContact":
		return n.deleteContact(ctx, config, apiKey)
	case "getLists":
		return n.getLists(ctx, apiKey)
	case "createList":
		return n.createList(ctx, config, apiKey)
	default:
		return n.sendEmail(ctx, config, apiKey, execCtx.Input)
	}
}

func (n *SendGridNode) sendEmail(ctx context.Context, config map[string]interface{}, apiKey string, input map[string]interface{}) (map[string]interface{}, error) {
	to := core.GetString(config, "to", "")
	from := core.GetString(config, "from", "")
	subject := core.GetString(config, "subject", "")
	text := core.GetString(config, "text", "")
	html := core.GetString(config, "html", "")

	if to == "" || from == "" || subject == "" {
		return nil, fmt.Errorf("to, from, and subject are required")
	}

	if text == "" && html == "" {
		if t, ok := input["text"].(string); ok {
			text = t
		}
		if h, ok := input["html"].(string); ok {
			html = h
		}
	}

	if text == "" && html == "" {
		return nil, fmt.Errorf("text or html content is required")
	}

	// Build personalizations
	personalizations := []map[string]interface{}{
		{
			"to": []map[string]string{
				{"email": to},
			},
		},
	}

	// Add CC if provided
	if cc := core.GetString(config, "cc", ""); cc != "" {
		personalizations[0]["cc"] = []map[string]string{{"email": cc}}
	}

	// Add BCC if provided
	if bcc := core.GetString(config, "bcc", ""); bcc != "" {
		personalizations[0]["bcc"] = []map[string]string{{"email": bcc}}
	}

	// Build content array
	var content []map[string]string
	if text != "" {
		content = append(content, map[string]string{"type": "text/plain", "value": text})
	}
	if html != "" {
		content = append(content, map[string]string{"type": "text/html", "value": html})
	}

	body := map[string]interface{}{
		"personalizations": personalizations,
		"from":             map[string]string{"email": from},
		"subject":          subject,
		"content":          content,
	}

	// Add reply-to if provided
	if replyTo := core.GetString(config, "replyTo", ""); replyTo != "" {
		body["reply_to"] = map[string]string{"email": replyTo}
	}

	endpoint := "https://api.sendgrid.com/v3/mail/send"
	bodyJSON, _ := json.Marshal(body)

	_, err := n.makeRequest(ctx, "POST", endpoint, bodyJSON, apiKey)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"sent":    true,
		"to":      to,
		"from":    from,
		"subject": subject,
	}, nil
}

func (n *SendGridNode) sendTemplateEmail(ctx context.Context, config map[string]interface{}, apiKey string, input map[string]interface{}) (map[string]interface{}, error) {
	to := core.GetString(config, "to", "")
	from := core.GetString(config, "from", "")
	templateId := core.GetString(config, "templateId", "")

	if to == "" || from == "" || templateId == "" {
		return nil, fmt.Errorf("to, from, and templateId are required")
	}

	dynamicData := config["dynamicData"]
	if dynamicData == nil {
		dynamicData = input["dynamicData"]
	}
	if dynamicData == nil {
		dynamicData = map[string]interface{}{}
	}

	personalizations := []map[string]interface{}{
		{
			"to": []map[string]string{
				{"email": to},
			},
			"dynamic_template_data": dynamicData,
		},
	}

	body := map[string]interface{}{
		"personalizations": personalizations,
		"from":             map[string]string{"email": from},
		"template_id":      templateId,
	}

	endpoint := "https://api.sendgrid.com/v3/mail/send"
	bodyJSON, _ := json.Marshal(body)

	_, err := n.makeRequest(ctx, "POST", endpoint, bodyJSON, apiKey)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"sent":       true,
		"to":         to,
		"from":       from,
		"templateId": templateId,
	}, nil
}

func (n *SendGridNode) getContacts(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	endpoint := "https://api.sendgrid.com/v3/marketing/contacts"

	if email := core.GetString(config, "email", ""); email != "" {
		// Search for specific contact
		searchBody := map[string]string{
			"query": fmt.Sprintf("email = '%s'", email),
		}
		bodyJSON, _ := json.Marshal(searchBody)
		return n.makeRequest(ctx, "POST", "https://api.sendgrid.com/v3/marketing/contacts/search", bodyJSON, apiKey)
	}

	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *SendGridNode) addContact(ctx context.Context, config map[string]interface{}, apiKey string, input map[string]interface{}) (map[string]interface{}, error) {
	email := core.GetString(config, "email", "")
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	contact := map[string]interface{}{
		"email": email,
	}

	if firstName := core.GetString(config, "firstName", ""); firstName != "" {
		contact["first_name"] = firstName
	}
	if lastName := core.GetString(config, "lastName", ""); lastName != "" {
		contact["last_name"] = lastName
	}

	// Add to list if specified
	var listIds []string
	if listId := core.GetString(config, "listId", ""); listId != "" {
		listIds = []string{listId}
	}

	body := map[string]interface{}{
		"contacts": []map[string]interface{}{contact},
	}
	if len(listIds) > 0 {
		body["list_ids"] = listIds
	}

	endpoint := "https://api.sendgrid.com/v3/marketing/contacts"
	bodyJSON, _ := json.Marshal(body)

	result, err := n.makeRequest(ctx, "PUT", endpoint, bodyJSON, apiKey)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"added":  true,
		"email":  email,
		"jobId":  result["job_id"],
	}, nil
}

func (n *SendGridNode) deleteContact(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	contactId := core.GetString(config, "contactId", "")
	if contactId == "" {
		return nil, fmt.Errorf("contactId is required")
	}

	endpoint := fmt.Sprintf("https://api.sendgrid.com/v3/marketing/contacts?ids=%s", contactId)

	_, err := n.makeRequest(ctx, "DELETE", endpoint, nil, apiKey)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"deleted":   true,
		"contactId": contactId,
	}, nil
}

func (n *SendGridNode) getLists(ctx context.Context, apiKey string) (map[string]interface{}, error) {
	endpoint := "https://api.sendgrid.com/v3/marketing/lists"
	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *SendGridNode) createList(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	name := core.GetString(config, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	body := map[string]string{"name": name}
	endpoint := "https://api.sendgrid.com/v3/marketing/lists"
	bodyJSON, _ := json.Marshal(body)

	return n.makeRequest(ctx, "POST", endpoint, bodyJSON, apiKey)
}

func (n *SendGridNode) makeRequest(ctx context.Context, method, endpoint string, body []byte, apiKey string) (map[string]interface{}, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, endpoint, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle 202 Accepted (email sent)
	if resp.StatusCode == 202 {
		return map[string]interface{}{"accepted": true}, nil
	}

	// Handle empty response
	if len(respBody) == 0 {
		return map[string]interface{}{"success": true}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode >= 400 {
		errMsg := "SendGrid API error"
		if errs, ok := result["errors"].([]interface{}); ok && len(errs) > 0 {
			errMsg = fmt.Sprintf("%v", errs)
		}
		return nil, fmt.Errorf("%s (status %d)", errMsg, resp.StatusCode)
	}

	return result, nil
}

// Note: SendGridNode is registered in integrations/init.go
