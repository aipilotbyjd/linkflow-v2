package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/worker/nodes"
)

type GitHubNode struct{}

func NewGitHubNode() *GitHubNode {
	return &GitHubNode{}
}

func (n *GitHubNode) Type() string {
	return "integration.github"
}

func (n *GitHubNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	credID := getString(execCtx.Config, "credentialId", "")
	if credID == "" {
		return nil, fmt.Errorf("credential ID is required")
	}

	cred, err := execCtx.GetCredential(parseUUID(credID))
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	token := cred.Token
	if token == "" {
		token = cred.AccessToken
	}
	operation := getString(execCtx.Config, "operation", "getRepo")

	switch operation {
	case "getRepo":
		return n.getRepo(ctx, token, execCtx.Config)
	case "listRepos":
		return n.listRepos(ctx, token, execCtx.Config)
	case "createIssue":
		return n.createIssue(ctx, token, execCtx.Config)
	case "getIssue":
		return n.getIssue(ctx, token, execCtx.Config)
	case "listIssues":
		return n.listIssues(ctx, token, execCtx.Config)
	case "updateIssue":
		return n.updateIssue(ctx, token, execCtx.Config)
	case "createPR":
		return n.createPR(ctx, token, execCtx.Config)
	case "getPR":
		return n.getPR(ctx, token, execCtx.Config)
	case "listPRs":
		return n.listPRs(ctx, token, execCtx.Config)
	case "mergePR":
		return n.mergePR(ctx, token, execCtx.Config)
	case "createComment":
		return n.createComment(ctx, token, execCtx.Config)
	case "getUser":
		return n.getUser(ctx, token, execCtx.Config)
	case "createRelease":
		return n.createRelease(ctx, token, execCtx.Config)
	case "listBranches":
		return n.listBranches(ctx, token, execCtx.Config)
	case "getFile":
		return n.getFile(ctx, token, execCtx.Config)
	case "createFile":
		return n.createFile(ctx, token, execCtx.Config)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *GitHubNode) getRepo(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	return n.makeRequest(ctx, token, "GET", url, nil)
}

func (n *GitHubNode) listRepos(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	user := getString(config, "user", "")
	var url string
	if user == "" {
		url = "https://api.github.com/user/repos"
	} else {
		url = fmt.Sprintf("https://api.github.com/users/%s/repos", user)
	}
	return n.makeRequest(ctx, token, "GET", url, nil)
}

func (n *GitHubNode) createIssue(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)

	payload := map[string]interface{}{
		"title": getString(config, "title", ""),
		"body":  getString(config, "body", ""),
	}
	if labels, ok := config["labels"].([]interface{}); ok {
		payload["labels"] = labels
	}
	if assignees, ok := config["assignees"].([]interface{}); ok {
		payload["assignees"] = assignees
	}

	return n.makeRequest(ctx, token, "POST", url, payload)
}

func (n *GitHubNode) getIssue(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	number := getInt(config, "number", 0)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, number)
	return n.makeRequest(ctx, token, "GET", url, nil)
}

func (n *GitHubNode) listIssues(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	state := getString(config, "state", "open")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=%s", owner, repo, state)
	return n.makeRequest(ctx, token, "GET", url, nil)
}

func (n *GitHubNode) updateIssue(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	number := getInt(config, "number", 0)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, number)

	payload := make(map[string]interface{})
	if title := getString(config, "title", ""); title != "" {
		payload["title"] = title
	}
	if body := getString(config, "body", ""); body != "" {
		payload["body"] = body
	}
	if state := getString(config, "state", ""); state != "" {
		payload["state"] = state
	}

	return n.makeRequest(ctx, token, "PATCH", url, payload)
}

func (n *GitHubNode) createPR(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)

	payload := map[string]interface{}{
		"title": getString(config, "title", ""),
		"body":  getString(config, "body", ""),
		"head":  getString(config, "head", ""),
		"base":  getString(config, "base", "main"),
	}

	return n.makeRequest(ctx, token, "POST", url, payload)
}

func (n *GitHubNode) getPR(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	number := getInt(config, "number", 0)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, number)
	return n.makeRequest(ctx, token, "GET", url, nil)
}

func (n *GitHubNode) listPRs(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	state := getString(config, "state", "open")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=%s", owner, repo, state)
	return n.makeRequest(ctx, token, "GET", url, nil)
}

func (n *GitHubNode) mergePR(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	number := getInt(config, "number", 0)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/merge", owner, repo, number)

	payload := map[string]interface{}{
		"merge_method": getString(config, "mergeMethod", "merge"),
	}
	if msg := getString(config, "commitMessage", ""); msg != "" {
		payload["commit_message"] = msg
	}

	return n.makeRequest(ctx, token, "PUT", url, payload)
}

func (n *GitHubNode) createComment(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	number := getInt(config, "number", 0)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, number)

	payload := map[string]interface{}{
		"body": getString(config, "body", ""),
	}

	return n.makeRequest(ctx, token, "POST", url, payload)
}

func (n *GitHubNode) getUser(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	username := getString(config, "username", "")
	var url string
	if username == "" {
		url = "https://api.github.com/user"
	} else {
		url = fmt.Sprintf("https://api.github.com/users/%s", username)
	}
	return n.makeRequest(ctx, token, "GET", url, nil)
}

func (n *GitHubNode) createRelease(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)

	payload := map[string]interface{}{
		"tag_name":         getString(config, "tagName", ""),
		"name":             getString(config, "name", ""),
		"body":             getString(config, "body", ""),
		"draft":            getBool(config, "draft", false),
		"prerelease":       getBool(config, "prerelease", false),
		"generate_release_notes": getBool(config, "generateNotes", false),
	}

	return n.makeRequest(ctx, token, "POST", url, payload)
}

func (n *GitHubNode) listBranches(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches", owner, repo)
	return n.makeRequest(ctx, token, "GET", url, nil)
}

func (n *GitHubNode) getFile(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	path := getString(config, "path", "")
	ref := getString(config, "ref", "")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}
	return n.makeRequest(ctx, token, "GET", url, nil)
}

func (n *GitHubNode) createFile(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	owner := getString(config, "owner", "")
	repo := getString(config, "repo", "")
	path := getString(config, "path", "")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)

	payload := map[string]interface{}{
		"message": getString(config, "message", "Create file"),
		"content": getString(config, "content", ""),
	}
	if branch := getString(config, "branch", ""); branch != "" {
		payload["branch"] = branch
	}
	if sha := getString(config, "sha", ""); sha != "" {
		payload["sha"] = sha
	}

	return n.makeRequest(ctx, token, "PUT", url, payload)
}

func (n *GitHubNode) makeRequest(ctx context.Context, token, method, url string, payload map[string]interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(respBody))
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return map[string]interface{}{
			"statusCode": resp.StatusCode,
			"body":       string(respBody),
		}, nil
	}

	return map[string]interface{}{
		"statusCode": resp.StatusCode,
		"data":       result,
	}, nil
}
