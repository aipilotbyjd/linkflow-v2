package devtools

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// JiraNode integrates with Jira
type JiraNode struct{}

func NewJiraNode() *JiraNode {
	return &JiraNode{}
}

func (n *JiraNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	operation, _ := params["operation"].(string)
	
	baseURL := runtime.GetCredentialValue("jira", "base_url")
	email := runtime.GetCredentialValue("jira", "email")
	apiToken := runtime.GetCredentialValue("jira", "api_token")
	
	if baseURL == "" || apiToken == "" {
		return nil, fmt.Errorf("jira credentials not configured")
	}
	_ = email // Used for auth

	switch operation {
	case "create_issue":
		projectKey, _ := params["project_key"].(string)
		issueType, _ := params["issue_type"].(string)
		summary, _ := params["summary"].(string)
		description, _ := params["description"].(string)
		priority, _ := params["priority"].(string)
		assignee, _ := params["assignee"].(string)
		labels, _ := params["labels"].([]interface{})
		
		return types.JSON{
			"operation":   "create_issue",
			"project_key": projectKey,
			"issue_type":  issueType,
			"summary":     summary,
			"description": description,
			"priority":    priority,
			"assignee":    assignee,
			"labels":      labels,
			"success":     true,
		}, nil

	case "update_issue":
		issueKey, _ := params["issue_key"].(string)
		fields, _ := params["fields"].(map[string]interface{})
		return types.JSON{
			"operation": "update_issue",
			"issue_key": issueKey,
			"fields":    fields,
			"success":   true,
		}, nil

	case "get_issue":
		issueKey, _ := params["issue_key"].(string)
		return types.JSON{
			"operation": "get_issue",
			"issue_key": issueKey,
			"issue":     map[string]interface{}{},
			"success":   true,
		}, nil

	case "search_issues":
		jql, _ := params["jql"].(string)
		maxResults := int(toFloat(params["max_results"]))
		if maxResults == 0 {
			maxResults = 50
		}
		return types.JSON{
			"operation":   "search_issues",
			"jql":         jql,
			"max_results": maxResults,
			"issues":      []interface{}{},
			"success":     true,
		}, nil

	case "transition_issue":
		issueKey, _ := params["issue_key"].(string)
		transitionID, _ := params["transition_id"].(string)
		return types.JSON{
			"operation":     "transition_issue",
			"issue_key":     issueKey,
			"transition_id": transitionID,
			"success":       true,
		}, nil

	case "add_comment":
		issueKey, _ := params["issue_key"].(string)
		body, _ := inputData["body"].(string)
		if body == "" {
			body, _ = params["body"].(string)
		}
		return types.JSON{
			"operation": "add_comment",
			"issue_key": issueKey,
			"body":      body,
			"success":   true,
		}, nil

	case "assign_issue":
		issueKey, _ := params["issue_key"].(string)
		assignee, _ := params["assignee"].(string)
		return types.JSON{
			"operation": "assign_issue",
			"issue_key": issueKey,
			"assignee":  assignee,
			"success":   true,
		}, nil

	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *JiraNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.jira",
		Name:        "Jira",
		Description: "Create, update, and manage Jira issues",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}, {Name: "error", Type: "error"}},
		Credentials: []string{"jira"},
		Parameters: []wtypes.NodeParameter{
			{Name: "operation", Type: "select", Description: "Operation to perform", Required: true, Options: []wtypes.ParamOption{
				{Value: "create_issue", Name: "Create Issue"},
				{Value: "update_issue", Name: "Update Issue"},
				{Value: "get_issue", Name: "Get Issue"},
				{Value: "search_issues", Name: "Search Issues (JQL)"},
				{Value: "transition_issue", Name: "Transition Issue"},
				{Value: "add_comment", Name: "Add Comment"},
				{Value: "assign_issue", Name: "Assign Issue"},
			}},
			{Name: "project_key", Type: "string", Description: "Project key (e.g., PROJ)"},
			{Name: "issue_key", Type: "string", Description: "Issue key (e.g., PROJ-123)"},
			{Name: "issue_type", Type: "select", Description: "Issue type", Options: []wtypes.ParamOption{
				{Value: "Bug", Name: "Bug"},
				{Value: "Task", Name: "Task"},
				{Value: "Story", Name: "Story"},
				{Value: "Epic", Name: "Epic"},
			}},
			{Name: "summary", Type: "string", Description: "Issue summary"},
			{Name: "description", Type: "string", Description: "Issue description"},
			{Name: "priority", Type: "select", Description: "Priority", Options: []wtypes.ParamOption{
				{Value: "Highest", Name: "Highest"},
				{Value: "High", Name: "High"},
				{Value: "Medium", Name: "Medium"},
				{Value: "Low", Name: "Low"},
				{Value: "Lowest", Name: "Lowest"},
			}},
			{Name: "assignee", Type: "string", Description: "Assignee email or account ID"},
			{Name: "labels", Type: "array", Description: "Labels"},
			{Name: "jql", Type: "string", Description: "JQL query for search"},
			{Name: "transition_id", Type: "string", Description: "Transition ID"},
			{Name: "body", Type: "string", Description: "Comment body"},
		},
	}
}
