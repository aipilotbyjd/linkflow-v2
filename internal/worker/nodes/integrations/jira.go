package integrations

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// JiraNode handles Jira operations
type JiraNode struct{}

func (n *JiraNode) Type() string {
	return "integrations.jira"
}

func (n *JiraNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	domain := core.GetString(config, "domain", "")
	email := core.GetString(config, "email", "")
	apiToken := core.GetString(config, "apiToken", "")

	if domain == "" || email == "" || apiToken == "" {
		return nil, fmt.Errorf("domain, email, and apiToken are required")
	}

	if !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain
	}
	domain = strings.TrimSuffix(domain, "/")

	operation := core.GetString(config, "operation", "getIssue")

	switch operation {
	case "getIssue":
		return n.getIssue(ctx, config, domain, email, apiToken)
	case "createIssue":
		return n.createIssue(ctx, config, domain, email, apiToken, execCtx.Input)
	case "updateIssue":
		return n.updateIssue(ctx, config, domain, email, apiToken, execCtx.Input)
	case "deleteIssue":
		return n.deleteIssue(ctx, config, domain, email, apiToken)
	case "searchIssues":
		return n.searchIssues(ctx, config, domain, email, apiToken)
	case "addComment":
		return n.addComment(ctx, config, domain, email, apiToken)
	case "getComments":
		return n.getComments(ctx, config, domain, email, apiToken)
	case "transition":
		return n.transitionIssue(ctx, config, domain, email, apiToken)
	case "assignIssue":
		return n.assignIssue(ctx, config, domain, email, apiToken)
	case "getProjects":
		return n.getProjects(ctx, domain, email, apiToken)
	case "getProject":
		return n.getProject(ctx, config, domain, email, apiToken)
	case "getTransitions":
		return n.getTransitions(ctx, config, domain, email, apiToken)
	case "getUsers":
		return n.getUsers(ctx, config, domain, email, apiToken)
	default:
		return n.getIssue(ctx, config, domain, email, apiToken)
	}
}

func (n *JiraNode) getIssue(ctx context.Context, config map[string]interface{}, domain, email, apiToken string) (map[string]interface{}, error) {
	issueKey := core.GetString(config, "issueKey", "")
	if issueKey == "" {
		return nil, fmt.Errorf("issueKey is required")
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/issue/%s", domain, issueKey)
	return n.makeRequest(ctx, "GET", endpoint, nil, email, apiToken)
}

func (n *JiraNode) createIssue(ctx context.Context, config map[string]interface{}, domain, email, apiToken string, input map[string]interface{}) (map[string]interface{}, error) {
	projectKey := core.GetString(config, "projectKey", "")
	issueType := core.GetString(config, "issueType", "Task")
	summary := core.GetString(config, "summary", "")

	if projectKey == "" || summary == "" {
		return nil, fmt.Errorf("projectKey and summary are required")
	}

	fields := map[string]interface{}{
		"project": map[string]string{
			"key": projectKey,
		},
		"issuetype": map[string]string{
			"name": issueType,
		},
		"summary": summary,
	}

	if description := core.GetString(config, "description", ""); description != "" {
		fields["description"] = map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": description,
						},
					},
				},
			},
		}
	}

	if priority := core.GetString(config, "priority", ""); priority != "" {
		fields["priority"] = map[string]string{"name": priority}
	}

	if assignee := core.GetString(config, "assignee", ""); assignee != "" {
		fields["assignee"] = map[string]string{"accountId": assignee}
	}

	if labels, ok := config["labels"].([]interface{}); ok {
		labelStrs := make([]string, len(labels))
		for i, l := range labels {
			labelStrs[i] = fmt.Sprintf("%v", l)
		}
		fields["labels"] = labelStrs
	}

	body := map[string]interface{}{
		"fields": fields,
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/issue", domain)
	bodyJSON, _ := json.Marshal(body)

	result, err := n.makeRequest(ctx, "POST", endpoint, bodyJSON, email, apiToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"created": true,
		"id":      result["id"],
		"key":     result["key"],
		"self":    result["self"],
	}, nil
}

func (n *JiraNode) updateIssue(ctx context.Context, config map[string]interface{}, domain, email, apiToken string, input map[string]interface{}) (map[string]interface{}, error) {
	issueKey := core.GetString(config, "issueKey", "")
	if issueKey == "" {
		return nil, fmt.Errorf("issueKey is required")
	}

	fields := map[string]interface{}{}

	if summary := core.GetString(config, "summary", ""); summary != "" {
		fields["summary"] = summary
	}

	if description := core.GetString(config, "description", ""); description != "" {
		fields["description"] = map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": description,
						},
					},
				},
			},
		}
	}

	if priority := core.GetString(config, "priority", ""); priority != "" {
		fields["priority"] = map[string]string{"name": priority}
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	body := map[string]interface{}{
		"fields": fields,
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/issue/%s", domain, issueKey)
	bodyJSON, _ := json.Marshal(body)

	_, err := n.makeRequest(ctx, "PUT", endpoint, bodyJSON, email, apiToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"updated":  true,
		"issueKey": issueKey,
	}, nil
}

func (n *JiraNode) deleteIssue(ctx context.Context, config map[string]interface{}, domain, email, apiToken string) (map[string]interface{}, error) {
	issueKey := core.GetString(config, "issueKey", "")
	if issueKey == "" {
		return nil, fmt.Errorf("issueKey is required")
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/issue/%s", domain, issueKey)

	_, err := n.makeRequest(ctx, "DELETE", endpoint, nil, email, apiToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"deleted":  true,
		"issueKey": issueKey,
	}, nil
}

func (n *JiraNode) searchIssues(ctx context.Context, config map[string]interface{}, domain, email, apiToken string) (map[string]interface{}, error) {
	jql := core.GetString(config, "jql", "")
	if jql == "" {
		return nil, fmt.Errorf("jql is required")
	}

	maxResults := core.GetInt(config, "maxResults", 50)
	startAt := core.GetInt(config, "startAt", 0)

	body := map[string]interface{}{
		"jql":        jql,
		"maxResults": maxResults,
		"startAt":    startAt,
		"fields":     []string{"summary", "status", "assignee", "priority", "created", "updated", "issuetype", "project"},
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/search", domain)
	bodyJSON, _ := json.Marshal(body)

	result, err := n.makeRequest(ctx, "POST", endpoint, bodyJSON, email, apiToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"issues":     result["issues"],
		"total":      result["total"],
		"maxResults": result["maxResults"],
		"startAt":    result["startAt"],
	}, nil
}

func (n *JiraNode) addComment(ctx context.Context, config map[string]interface{}, domain, email, apiToken string) (map[string]interface{}, error) {
	issueKey := core.GetString(config, "issueKey", "")
	comment := core.GetString(config, "comment", "")

	if issueKey == "" || comment == "" {
		return nil, fmt.Errorf("issueKey and comment are required")
	}

	body := map[string]interface{}{
		"body": map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": comment,
						},
					},
				},
			},
		},
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/issue/%s/comment", domain, issueKey)
	bodyJSON, _ := json.Marshal(body)

	result, err := n.makeRequest(ctx, "POST", endpoint, bodyJSON, email, apiToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"added":     true,
		"commentId": result["id"],
		"issueKey":  issueKey,
	}, nil
}

func (n *JiraNode) getComments(ctx context.Context, config map[string]interface{}, domain, email, apiToken string) (map[string]interface{}, error) {
	issueKey := core.GetString(config, "issueKey", "")
	if issueKey == "" {
		return nil, fmt.Errorf("issueKey is required")
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/issue/%s/comment", domain, issueKey)
	return n.makeRequest(ctx, "GET", endpoint, nil, email, apiToken)
}

func (n *JiraNode) transitionIssue(ctx context.Context, config map[string]interface{}, domain, email, apiToken string) (map[string]interface{}, error) {
	issueKey := core.GetString(config, "issueKey", "")
	transitionId := core.GetString(config, "transitionId", "")

	if issueKey == "" || transitionId == "" {
		return nil, fmt.Errorf("issueKey and transitionId are required")
	}

	body := map[string]interface{}{
		"transition": map[string]string{
			"id": transitionId,
		},
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", domain, issueKey)
	bodyJSON, _ := json.Marshal(body)

	_, err := n.makeRequest(ctx, "POST", endpoint, bodyJSON, email, apiToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"transitioned": true,
		"issueKey":     issueKey,
		"transitionId": transitionId,
	}, nil
}

func (n *JiraNode) assignIssue(ctx context.Context, config map[string]interface{}, domain, email, apiToken string) (map[string]interface{}, error) {
	issueKey := core.GetString(config, "issueKey", "")
	accountId := core.GetString(config, "accountId", "")

	if issueKey == "" {
		return nil, fmt.Errorf("issueKey is required")
	}

	body := map[string]interface{}{}
	if accountId != "" {
		body["accountId"] = accountId
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/issue/%s/assignee", domain, issueKey)
	bodyJSON, _ := json.Marshal(body)

	_, err := n.makeRequest(ctx, "PUT", endpoint, bodyJSON, email, apiToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"assigned":  true,
		"issueKey":  issueKey,
		"accountId": accountId,
	}, nil
}

func (n *JiraNode) getProjects(ctx context.Context, domain, email, apiToken string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/rest/api/3/project/search", domain)
	return n.makeRequest(ctx, "GET", endpoint, nil, email, apiToken)
}

func (n *JiraNode) getProject(ctx context.Context, config map[string]interface{}, domain, email, apiToken string) (map[string]interface{}, error) {
	projectKey := core.GetString(config, "projectKey", "")
	if projectKey == "" {
		return nil, fmt.Errorf("projectKey is required")
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/project/%s", domain, projectKey)
	return n.makeRequest(ctx, "GET", endpoint, nil, email, apiToken)
}

func (n *JiraNode) getTransitions(ctx context.Context, config map[string]interface{}, domain, email, apiToken string) (map[string]interface{}, error) {
	issueKey := core.GetString(config, "issueKey", "")
	if issueKey == "" {
		return nil, fmt.Errorf("issueKey is required")
	}

	endpoint := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", domain, issueKey)
	return n.makeRequest(ctx, "GET", endpoint, nil, email, apiToken)
}

func (n *JiraNode) getUsers(ctx context.Context, config map[string]interface{}, domain, email, apiToken string) (map[string]interface{}, error) {
	query := core.GetString(config, "query", "")

	endpoint := fmt.Sprintf("%s/rest/api/3/user/search", domain)
	if query != "" {
		endpoint += "?query=" + url.QueryEscape(query)
	}

	result, err := n.makeRequest(ctx, "GET", endpoint, nil, email, apiToken)
	if err != nil {
		// Response is an array, not object
		return nil, err
	}

	return result, nil
}

func (n *JiraNode) makeRequest(ctx context.Context, method, endpoint string, body []byte, email, apiToken string) (map[string]interface{}, error) {
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

	// Basic auth with email:apiToken
	auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + apiToken))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

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

	// Handle empty response (204 No Content)
	if len(respBody) == 0 {
		return map[string]interface{}{"success": true}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Try as array
		var arr []interface{}
		if err2 := json.Unmarshal(respBody, &arr); err2 == nil {
			return map[string]interface{}{"items": arr}, nil
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode >= 400 {
		errMsgs := result["errorMessages"]
		if errMsgs == nil {
			errMsgs = result["errors"]
		}
		return nil, fmt.Errorf("Jira API error (status %d): %v", resp.StatusCode, errMsgs)
	}

	return result, nil
}

// Note: JiraNode is registered in integrations/init.go
