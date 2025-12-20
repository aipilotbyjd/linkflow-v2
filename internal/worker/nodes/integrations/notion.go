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

type NotionNode struct{}

func NewNotionNode() *NotionNode {
	return &NotionNode{}
}

func (n *NotionNode) Type() string {
	return "integration.notion"
}

func (n *NotionNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	credID := getString(execCtx.Config, "credentialId", "")
	if credID == "" {
		return nil, fmt.Errorf("credential ID is required")
	}

	cred, err := execCtx.GetCredential(parseUUID(credID))
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	apiKey := cred.APIKey
	if apiKey == "" {
		apiKey = cred.Token
	}
	operation := getString(execCtx.Config, "operation", "getPage")

	switch operation {
	case "getPage":
		return n.getPage(ctx, apiKey, execCtx.Config)
	case "createPage":
		return n.createPage(ctx, apiKey, execCtx.Config)
	case "updatePage":
		return n.updatePage(ctx, apiKey, execCtx.Config)
	case "getDatabase":
		return n.getDatabase(ctx, apiKey, execCtx.Config)
	case "queryDatabase":
		return n.queryDatabase(ctx, apiKey, execCtx.Config)
	case "createDatabase":
		return n.createDatabase(ctx, apiKey, execCtx.Config)
	case "getBlock":
		return n.getBlock(ctx, apiKey, execCtx.Config)
	case "getBlockChildren":
		return n.getBlockChildren(ctx, apiKey, execCtx.Config)
	case "appendBlockChildren":
		return n.appendBlockChildren(ctx, apiKey, execCtx.Config)
	case "deleteBlock":
		return n.deleteBlock(ctx, apiKey, execCtx.Config)
	case "search":
		return n.search(ctx, apiKey, execCtx.Config)
	case "getUser":
		return n.getUser(ctx, apiKey, execCtx.Config)
	case "listUsers":
		return n.listUsers(ctx, apiKey)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *NotionNode) getPage(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	pageID := getString(config, "pageId", "")
	url := fmt.Sprintf("https://api.notion.com/v1/pages/%s", pageID)
	return n.makeRequest(ctx, apiKey, "GET", url, nil)
}

func (n *NotionNode) createPage(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"parent":     config["parent"],
		"properties": config["properties"],
	}

	if children, ok := config["children"]; ok {
		payload["children"] = children
	}
	if icon, ok := config["icon"]; ok {
		payload["icon"] = icon
	}
	if cover, ok := config["cover"]; ok {
		payload["cover"] = cover
	}

	return n.makeRequest(ctx, apiKey, "POST", "https://api.notion.com/v1/pages", payload)
}

func (n *NotionNode) updatePage(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	pageID := getString(config, "pageId", "")
	url := fmt.Sprintf("https://api.notion.com/v1/pages/%s", pageID)

	payload := map[string]interface{}{
		"properties": config["properties"],
	}

	if icon, ok := config["icon"]; ok {
		payload["icon"] = icon
	}
	if cover, ok := config["cover"]; ok {
		payload["cover"] = cover
	}
	if archived, ok := config["archived"].(bool); ok {
		payload["archived"] = archived
	}

	return n.makeRequest(ctx, apiKey, "PATCH", url, payload)
}

func (n *NotionNode) getDatabase(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	databaseID := getString(config, "databaseId", "")
	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s", databaseID)
	return n.makeRequest(ctx, apiKey, "GET", url, nil)
}

func (n *NotionNode) queryDatabase(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	databaseID := getString(config, "databaseId", "")
	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", databaseID)

	payload := make(map[string]interface{})
	if filter, ok := config["filter"]; ok {
		payload["filter"] = filter
	}
	if sorts, ok := config["sorts"]; ok {
		payload["sorts"] = sorts
	}
	if startCursor := getString(config, "startCursor", ""); startCursor != "" {
		payload["start_cursor"] = startCursor
	}
	if pageSize := getInt(config, "pageSize", 0); pageSize > 0 {
		payload["page_size"] = pageSize
	}

	return n.makeRequest(ctx, apiKey, "POST", url, payload)
}

func (n *NotionNode) createDatabase(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"parent":     config["parent"],
		"title":      config["title"],
		"properties": config["properties"],
	}

	return n.makeRequest(ctx, apiKey, "POST", "https://api.notion.com/v1/databases", payload)
}

func (n *NotionNode) getBlock(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	blockID := getString(config, "blockId", "")
	url := fmt.Sprintf("https://api.notion.com/v1/blocks/%s", blockID)
	return n.makeRequest(ctx, apiKey, "GET", url, nil)
}

func (n *NotionNode) getBlockChildren(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	blockID := getString(config, "blockId", "")
	url := fmt.Sprintf("https://api.notion.com/v1/blocks/%s/children", blockID)

	if startCursor := getString(config, "startCursor", ""); startCursor != "" {
		url += "?start_cursor=" + startCursor
	}
	if pageSize := getInt(config, "pageSize", 0); pageSize > 0 {
		url += fmt.Sprintf("&page_size=%d", pageSize)
	}

	return n.makeRequest(ctx, apiKey, "GET", url, nil)
}

func (n *NotionNode) appendBlockChildren(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	blockID := getString(config, "blockId", "")
	url := fmt.Sprintf("https://api.notion.com/v1/blocks/%s/children", blockID)

	payload := map[string]interface{}{
		"children": config["children"],
	}

	return n.makeRequest(ctx, apiKey, "PATCH", url, payload)
}

func (n *NotionNode) deleteBlock(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	blockID := getString(config, "blockId", "")
	url := fmt.Sprintf("https://api.notion.com/v1/blocks/%s", blockID)
	return n.makeRequest(ctx, apiKey, "DELETE", url, nil)
}

func (n *NotionNode) search(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := make(map[string]interface{})

	if query := getString(config, "query", ""); query != "" {
		payload["query"] = query
	}
	if filter, ok := config["filter"]; ok {
		payload["filter"] = filter
	}
	if sort, ok := config["sort"]; ok {
		payload["sort"] = sort
	}
	if startCursor := getString(config, "startCursor", ""); startCursor != "" {
		payload["start_cursor"] = startCursor
	}
	if pageSize := getInt(config, "pageSize", 0); pageSize > 0 {
		payload["page_size"] = pageSize
	}

	return n.makeRequest(ctx, apiKey, "POST", "https://api.notion.com/v1/search", payload)
}

func (n *NotionNode) getUser(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	userID := getString(config, "userId", "")
	url := fmt.Sprintf("https://api.notion.com/v1/users/%s", userID)
	return n.makeRequest(ctx, apiKey, "GET", url, nil)
}

func (n *NotionNode) listUsers(ctx context.Context, apiKey string) (map[string]interface{}, error) {
	return n.makeRequest(ctx, apiKey, "GET", "https://api.notion.com/v1/users", nil)
}

func (n *NotionNode) makeRequest(ctx context.Context, apiKey, method, url string, payload map[string]interface{}) (map[string]interface{}, error) {
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

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Notion-Version", "2022-06-28")
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
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("Notion API error: %v - %s", errResp["message"], errResp["code"])
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result, nil
}
