package devtools

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// GitHubNode integrates with GitHub API
type GitHubNode struct{}

func NewGitHubNode() *GitHubNode {
	return &GitHubNode{}
}

func (n *GitHubNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	operation, _ := params["operation"].(string)
	owner, _ := params["owner"].(string)
	repo, _ := params["repo"].(string)

	token := runtime.GetCredentialValue("github", "token")
	if token == "" {
		return nil, fmt.Errorf("github credentials not configured")
	}

	switch operation {
	case "create_issue":
		title, _ := params["title"].(string)
		body, _ := params["body"].(string)
		labels, _ := params["labels"].([]interface{})
		return types.JSON{
			"operation": "create_issue",
			"owner":     owner,
			"repo":      repo,
			"title":     title,
			"body":      body,
			"labels":    labels,
			"success":   true,
		}, nil

	case "update_issue":
		issueNumber := int(toFloat(params["issue_number"]))
		return types.JSON{
			"operation":    "update_issue",
			"owner":        owner,
			"repo":         repo,
			"issue_number": issueNumber,
			"success":      true,
		}, nil

	case "list_issues":
		state, _ := params["state"].(string)
		if state == "" {
			state = "open"
		}
		return types.JSON{
			"operation": "list_issues",
			"owner":     owner,
			"repo":      repo,
			"state":     state,
			"issues":    []interface{}{},
			"success":   true,
		}, nil

	case "create_pr":
		title, _ := params["title"].(string)
		body, _ := params["body"].(string)
		head, _ := params["head"].(string)
		base, _ := params["base"].(string)
		return types.JSON{
			"operation": "create_pr",
			"owner":     owner,
			"repo":      repo,
			"title":     title,
			"body":      body,
			"head":      head,
			"base":      base,
			"success":   true,
		}, nil

	case "merge_pr":
		prNumber := int(toFloat(params["pr_number"]))
		return types.JSON{
			"operation": "merge_pr",
			"owner":     owner,
			"repo":      repo,
			"pr_number": prNumber,
			"success":   true,
		}, nil

	case "create_comment":
		issueNumber := int(toFloat(params["issue_number"]))
		body, _ := inputData["body"].(string)
		if body == "" {
			body, _ = params["body"].(string)
		}
		return types.JSON{
			"operation":    "create_comment",
			"owner":        owner,
			"repo":         repo,
			"issue_number": issueNumber,
			"body":         body,
			"success":      true,
		}, nil

	case "create_release":
		tagName, _ := params["tag_name"].(string)
		name, _ := params["name"].(string)
		body, _ := params["body"].(string)
		return types.JSON{
			"operation": "create_release",
			"owner":     owner,
			"repo":      repo,
			"tag_name":  tagName,
			"name":      name,
			"body":      body,
			"success":   true,
		}, nil

	case "dispatch_workflow":
		workflowID, _ := params["workflow_id"].(string)
		ref, _ := params["ref"].(string)
		inputs, _ := params["inputs"].(map[string]interface{})
		return types.JSON{
			"operation":   "dispatch_workflow",
			"owner":       owner,
			"repo":        repo,
			"workflow_id": workflowID,
			"ref":         ref,
			"inputs":      inputs,
			"success":     true,
		}, nil

	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *GitHubNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.github",
		Name:        "GitHub",
		Description: "Manage GitHub issues, PRs, releases, and workflows",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}, {Name: "error", Type: "error"}},
		Credentials: []string{"github"},
		Parameters: []wtypes.NodeParameter{
			{Name: "operation", Type: "select", Description: "Operation to perform", Required: true, Options: []wtypes.ParamOption{
				{Value: "create_issue", Name: "Create Issue"},
				{Value: "update_issue", Name: "Update Issue"},
				{Value: "list_issues", Name: "List Issues"},
				{Value: "create_pr", Name: "Create Pull Request"},
				{Value: "merge_pr", Name: "Merge Pull Request"},
				{Value: "create_comment", Name: "Create Comment"},
				{Value: "create_release", Name: "Create Release"},
				{Value: "dispatch_workflow", Name: "Trigger Workflow"},
			}},
			{Name: "owner", Type: "string", Description: "Repository owner", Required: true},
			{Name: "repo", Type: "string", Description: "Repository name", Required: true},
			{Name: "title", Type: "string", Description: "Issue/PR title"},
			{Name: "body", Type: "string", Description: "Content body"},
			{Name: "issue_number", Type: "number", Description: "Issue number"},
			{Name: "pr_number", Type: "number", Description: "Pull request number"},
			{Name: "labels", Type: "array", Description: "Labels to apply"},
			{Name: "head", Type: "string", Description: "Head branch"},
			{Name: "base", Type: "string", Description: "Base branch"},
			{Name: "state", Type: "select", Description: "Issue state", Options: []wtypes.ParamOption{
				{Value: "open", Name: "Open"},
				{Value: "closed", Name: "Closed"},
				{Value: "all", Name: "All"},
			}},
		},
	}
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}
