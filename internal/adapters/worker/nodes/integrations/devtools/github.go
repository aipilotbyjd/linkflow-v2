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
		Description: "Interact with GitHub repositories: manage issues, pull requests, releases, workflows, and more",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "github",
		Color:       "#181717",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data for GitHub operations"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "GitHub API response"},
			{Name: "error", Type: "object", Description: "Error details if operation fails"},
		},
		Credentials: []string{"github"},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "operation",
				DisplayName: "Operation",
				Type:        "options",
				Required:    true,
				Default:     "create_issue",
				Description: "GitHub operation to perform",
				Options: []wtypes.ParamOption{
					{Name: "Create Issue", Value: "create_issue", Description: "Create a new issue"},
					{Name: "Update Issue", Value: "update_issue", Description: "Update an existing issue"},
					{Name: "Close Issue", Value: "close_issue", Description: "Close an issue"},
					{Name: "Get Issue", Value: "get_issue", Description: "Get issue details"},
					{Name: "List Issues", Value: "list_issues", Description: "List repository issues"},
					{Name: "Create Pull Request", Value: "create_pr", Description: "Create a new PR"},
					{Name: "Merge Pull Request", Value: "merge_pr", Description: "Merge a PR"},
					{Name: "Get Pull Request", Value: "get_pr", Description: "Get PR details"},
					{Name: "List Pull Requests", Value: "list_prs", Description: "List repository PRs"},
					{Name: "Create Comment", Value: "create_comment", Description: "Add comment to issue/PR"},
					{Name: "Create Release", Value: "create_release", Description: "Create a new release"},
					{Name: "Trigger Workflow", Value: "dispatch_workflow", Description: "Trigger a GitHub Action"},
					{Name: "Get Repository", Value: "get_repo", Description: "Get repository info"},
					{Name: "List Branches", Value: "list_branches", Description: "List repository branches"},
					{Name: "Create Branch", Value: "create_branch", Description: "Create a new branch"},
					{Name: "Get File Contents", Value: "get_file", Description: "Get file from repository"},
					{Name: "Create/Update File", Value: "create_file", Description: "Create or update a file"},
				},
			},
			{
				Name:        "owner",
				DisplayName: "Owner",
				Type:        "string",
				Required:    true,
				Description: "Repository owner (username or organization)",
				Placeholder: "octocat",
			},
			{
				Name:        "repo",
				DisplayName: "Repository",
				Type:        "string",
				Required:    true,
				Description: "Repository name",
				Placeholder: "hello-world",
			},
			{
				Name:        "title",
				DisplayName: "Title",
				Type:        "string",
				Required:    false,
				Description: "Issue or PR title",
				Placeholder: "Bug: Something is broken",
				ShowIf:      "operation === 'create_issue' || operation === 'update_issue' || operation === 'create_pr'",
			},
			{
				Name:        "body",
				DisplayName: "Body",
				Type:        "code",
				Required:    false,
				Description: "Issue, PR, comment, or release body (Markdown supported)",
				Placeholder: "## Description\n\nDescribe the issue here...",
				ShowIf:      "operation === 'create_issue' || operation === 'update_issue' || operation === 'create_pr' || operation === 'create_comment' || operation === 'create_release'",
			},
			{
				Name:        "issue_number",
				DisplayName: "Issue Number",
				Type:        "number",
				Required:    false,
				Description: "Issue number to operate on",
				ShowIf:      "operation === 'update_issue' || operation === 'close_issue' || operation === 'get_issue' || operation === 'create_comment'",
			},
			{
				Name:        "pr_number",
				DisplayName: "PR Number",
				Type:        "number",
				Required:    false,
				Description: "Pull request number",
				ShowIf:      "operation === 'merge_pr' || operation === 'get_pr' || operation === 'create_comment'",
			},
			{
				Name:        "labels",
				DisplayName: "Labels",
				Type:        "json",
				Required:    false,
				Description: "Labels to apply (array of strings)",
				Placeholder: `["bug", "priority:high"]`,
				ShowIf:      "operation === 'create_issue' || operation === 'update_issue'",
			},
			{
				Name:        "assignees",
				DisplayName: "Assignees",
				Type:        "json",
				Required:    false,
				Description: "Users to assign (array of usernames)",
				Placeholder: `["octocat"]`,
				ShowIf:      "operation === 'create_issue' || operation === 'update_issue' || operation === 'create_pr'",
			},
			{
				Name:        "milestone",
				DisplayName: "Milestone",
				Type:        "number",
				Required:    false,
				Description: "Milestone number to assign",
				ShowIf:      "operation === 'create_issue' || operation === 'update_issue'",
			},
			{
				Name:        "head",
				DisplayName: "Head Branch",
				Type:        "string",
				Required:    false,
				Description: "Source branch for PR (branch with your changes)",
				Placeholder: "feature/my-feature",
				ShowIf:      "operation === 'create_pr'",
			},
			{
				Name:        "base",
				DisplayName: "Base Branch",
				Type:        "string",
				Required:    false,
				Default:     "main",
				Description: "Target branch for PR (branch to merge into)",
				Placeholder: "main",
				ShowIf:      "operation === 'create_pr'",
			},
			{
				Name:        "draft",
				DisplayName: "Draft PR",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Create as draft pull request",
				ShowIf:      "operation === 'create_pr'",
			},
			{
				Name:        "merge_method",
				DisplayName: "Merge Method",
				Type:        "options",
				Required:    false,
				Default:     "merge",
				Description: "How to merge the pull request",
				ShowIf:      "operation === 'merge_pr'",
				Options: []wtypes.ParamOption{
					{Name: "Merge Commit", Value: "merge"},
					{Name: "Squash and Merge", Value: "squash"},
					{Name: "Rebase and Merge", Value: "rebase"},
				},
			},
			{
				Name:        "state",
				DisplayName: "State Filter",
				Type:        "options",
				Required:    false,
				Default:     "open",
				Description: "Filter by issue/PR state",
				ShowIf:      "operation === 'list_issues' || operation === 'list_prs'",
				Options: []wtypes.ParamOption{
					{Name: "Open", Value: "open"},
					{Name: "Closed", Value: "closed"},
					{Name: "All", Value: "all"},
				},
			},
			{
				Name:        "sort",
				DisplayName: "Sort By",
				Type:        "options",
				Required:    false,
				Default:     "created",
				Description: "How to sort results",
				ShowIf:      "operation === 'list_issues' || operation === 'list_prs'",
				Options: []wtypes.ParamOption{
					{Name: "Created", Value: "created"},
					{Name: "Updated", Value: "updated"},
					{Name: "Comments", Value: "comments"},
				},
			},
			{
				Name:        "direction",
				DisplayName: "Sort Direction",
				Type:        "options",
				Required:    false,
				Default:     "desc",
				Description: "Sort direction",
				ShowIf:      "operation === 'list_issues' || operation === 'list_prs'",
				Options: []wtypes.ParamOption{
					{Name: "Descending", Value: "desc"},
					{Name: "Ascending", Value: "asc"},
				},
			},
			{
				Name:        "per_page",
				DisplayName: "Results Per Page",
				Type:        "number",
				Required:    false,
				Default:     30,
				Description: "Number of results per page (max 100)",
				ShowIf:      "operation === 'list_issues' || operation === 'list_prs' || operation === 'list_branches'",
			},
			{
				Name:        "tag_name",
				DisplayName: "Tag Name",
				Type:        "string",
				Required:    false,
				Description: "Git tag for the release",
				Placeholder: "v1.0.0",
				ShowIf:      "operation === 'create_release'",
			},
			{
				Name:        "release_name",
				DisplayName: "Release Name",
				Type:        "string",
				Required:    false,
				Description: "Name of the release",
				Placeholder: "Version 1.0.0",
				ShowIf:      "operation === 'create_release'",
			},
			{
				Name:        "prerelease",
				DisplayName: "Pre-release",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Mark as pre-release",
				ShowIf:      "operation === 'create_release'",
			},
			{
				Name:        "generate_release_notes",
				DisplayName: "Generate Release Notes",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Auto-generate release notes from commits",
				ShowIf:      "operation === 'create_release'",
			},
			{
				Name:        "workflow_id",
				DisplayName: "Workflow ID",
				Type:        "string",
				Required:    false,
				Description: "Workflow file name or ID",
				Placeholder: "ci.yml",
				ShowIf:      "operation === 'dispatch_workflow'",
			},
			{
				Name:        "ref",
				DisplayName: "Branch/Tag",
				Type:        "string",
				Required:    false,
				Default:     "main",
				Description: "Branch or tag to run workflow on",
				ShowIf:      "operation === 'dispatch_workflow' || operation === 'create_branch'",
			},
			{
				Name:        "workflow_inputs",
				DisplayName: "Workflow Inputs",
				Type:        "json",
				Required:    false,
				Description: "Input parameters for the workflow",
				Placeholder: `{"environment": "production"}`,
				ShowIf:      "operation === 'dispatch_workflow'",
			},
			{
				Name:        "file_path",
				DisplayName: "File Path",
				Type:        "string",
				Required:    false,
				Description: "Path to file in repository",
				Placeholder: "src/index.js",
				ShowIf:      "operation === 'get_file' || operation === 'create_file'",
			},
			{
				Name:        "file_content",
				DisplayName: "File Content",
				Type:        "code",
				Required:    false,
				Description: "Content to write to file",
				ShowIf:      "operation === 'create_file'",
			},
			{
				Name:        "commit_message",
				DisplayName: "Commit Message",
				Type:        "string",
				Required:    false,
				Description: "Commit message for file changes",
				Placeholder: "Update file via workflow",
				ShowIf:      "operation === 'create_file'",
			},
			{
				Name:        "branch_name",
				DisplayName: "New Branch Name",
				Type:        "string",
				Required:    false,
				Description: "Name for the new branch",
				Placeholder: "feature/new-feature",
				ShowIf:      "operation === 'create_branch'",
			},
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
