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

// NotionNode handles Notion operations
type NotionNode struct{}

func (n *NotionNode) Type() string {
	return "integrations.notion"
}

func (n *NotionNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	apiKey := core.GetString(config, "apiKey", "")
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey (integration token) is required")
	}

	operation := core.GetString(config, "operation", "getPage")

	switch operation {
	case "getPage":
		return n.getPage(ctx, config, apiKey)
	case "createPage":
		return n.createPage(ctx, config, apiKey, execCtx.Input)
	case "updatePage":
		return n.updatePage(ctx, config, apiKey, execCtx.Input)
	case "getDatabase":
		return n.getDatabase(ctx, config, apiKey)
	case "queryDatabase":
		return n.queryDatabase(ctx, config, apiKey)
	case "createDatabase":
		return n.createDatabase(ctx, config, apiKey)
	case "getBlock":
		return n.getBlock(ctx, config, apiKey)
	case "getBlockChildren":
		return n.getBlockChildren(ctx, config, apiKey)
	case "appendBlockChildren":
		return n.appendBlockChildren(ctx, config, apiKey, execCtx.Input)
	case "search":
		return n.search(ctx, config, apiKey)
	case "getUser":
		return n.getUser(ctx, config, apiKey)
	case "listUsers":
		return n.listUsers(ctx, apiKey)
	default:
		return n.getPage(ctx, config, apiKey)
	}
}

func (n *NotionNode) getPage(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	pageId := core.GetString(config, "pageId", "")
	if pageId == "" {
		return nil, fmt.Errorf("pageId is required")
	}

	endpoint := fmt.Sprintf("https://api.notion.com/v1/pages/%s", pageId)
	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *NotionNode) createPage(ctx context.Context, config map[string]interface{}, apiKey string, input map[string]interface{}) (map[string]interface{}, error) {
	parentType := core.GetString(config, "parentType", "page")
	parentId := core.GetString(config, "parentId", "")

	if parentId == "" {
		return nil, fmt.Errorf("parentId is required")
	}

	body := map[string]interface{}{}

	if parentType == "database" {
		body["parent"] = map[string]string{
			"type":        "database_id",
			"database_id": parentId,
		}
	} else {
		body["parent"] = map[string]string{
			"type":    "page_id",
			"page_id": parentId,
		}
	}

	// Add title property
	if title := core.GetString(config, "title", ""); title != "" {
		body["properties"] = map[string]interface{}{
			"title": map[string]interface{}{
				"title": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]string{
							"content": title,
						},
					},
				},
			},
		}
	}

	// Add properties from config
	if props, ok := config["properties"].(map[string]interface{}); ok {
		body["properties"] = props
	}

	// Add children blocks
	if children, ok := config["children"].([]interface{}); ok {
		body["children"] = children
	}

	// Add content as paragraph blocks
	if content := core.GetString(config, "content", ""); content != "" {
		body["children"] = []map[string]interface{}{
			{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": content,
							},
						},
					},
				},
			},
		}
	}

	endpoint := "https://api.notion.com/v1/pages"
	bodyJSON, _ := json.Marshal(body)

	return n.makeRequest(ctx, "POST", endpoint, bodyJSON, apiKey)
}

func (n *NotionNode) updatePage(ctx context.Context, config map[string]interface{}, apiKey string, input map[string]interface{}) (map[string]interface{}, error) {
	pageId := core.GetString(config, "pageId", "")
	if pageId == "" {
		return nil, fmt.Errorf("pageId is required")
	}

	body := map[string]interface{}{}

	if props, ok := config["properties"].(map[string]interface{}); ok {
		body["properties"] = props
	}

	if archived := config["archived"]; archived != nil {
		body["archived"] = archived
	}

	endpoint := fmt.Sprintf("https://api.notion.com/v1/pages/%s", pageId)
	bodyJSON, _ := json.Marshal(body)

	return n.makeRequest(ctx, "PATCH", endpoint, bodyJSON, apiKey)
}

func (n *NotionNode) getDatabase(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	databaseId := core.GetString(config, "databaseId", "")
	if databaseId == "" {
		return nil, fmt.Errorf("databaseId is required")
	}

	endpoint := fmt.Sprintf("https://api.notion.com/v1/databases/%s", databaseId)
	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *NotionNode) queryDatabase(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	databaseId := core.GetString(config, "databaseId", "")
	if databaseId == "" {
		return nil, fmt.Errorf("databaseId is required")
	}

	body := map[string]interface{}{}

	if filter, ok := config["filter"].(map[string]interface{}); ok {
		body["filter"] = filter
	}

	if sorts, ok := config["sorts"].([]interface{}); ok {
		body["sorts"] = sorts
	}

	if pageSize := core.GetInt(config, "pageSize", 0); pageSize > 0 {
		body["page_size"] = pageSize
	}

	if startCursor := core.GetString(config, "startCursor", ""); startCursor != "" {
		body["start_cursor"] = startCursor
	}

	endpoint := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", databaseId)
	bodyJSON, _ := json.Marshal(body)

	return n.makeRequest(ctx, "POST", endpoint, bodyJSON, apiKey)
}

func (n *NotionNode) createDatabase(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	parentPageId := core.GetString(config, "parentPageId", "")
	title := core.GetString(config, "title", "New Database")

	if parentPageId == "" {
		return nil, fmt.Errorf("parentPageId is required")
	}

	body := map[string]interface{}{
		"parent": map[string]string{
			"type":    "page_id",
			"page_id": parentPageId,
		},
		"title": []map[string]interface{}{
			{
				"type": "text",
				"text": map[string]string{
					"content": title,
				},
			},
		},
		"properties": map[string]interface{}{
			"Name": map[string]interface{}{
				"title": map[string]interface{}{},
			},
		},
	}

	// Add custom properties
	if props, ok := config["properties"].(map[string]interface{}); ok {
		body["properties"] = props
	}

	endpoint := "https://api.notion.com/v1/databases"
	bodyJSON, _ := json.Marshal(body)

	return n.makeRequest(ctx, "POST", endpoint, bodyJSON, apiKey)
}

func (n *NotionNode) getBlock(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	blockId := core.GetString(config, "blockId", "")
	if blockId == "" {
		return nil, fmt.Errorf("blockId is required")
	}

	endpoint := fmt.Sprintf("https://api.notion.com/v1/blocks/%s", blockId)
	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *NotionNode) getBlockChildren(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	blockId := core.GetString(config, "blockId", "")
	if blockId == "" {
		return nil, fmt.Errorf("blockId is required")
	}

	endpoint := fmt.Sprintf("https://api.notion.com/v1/blocks/%s/children", blockId)

	if pageSize := core.GetInt(config, "pageSize", 0); pageSize > 0 {
		endpoint += fmt.Sprintf("?page_size=%d", pageSize)
	}

	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *NotionNode) appendBlockChildren(ctx context.Context, config map[string]interface{}, apiKey string, input map[string]interface{}) (map[string]interface{}, error) {
	blockId := core.GetString(config, "blockId", "")
	if blockId == "" {
		return nil, fmt.Errorf("blockId is required")
	}

	children := config["children"]
	if children == nil {
		children = input["children"]
	}

	// Support simple text content
	if content := core.GetString(config, "content", ""); content != "" && children == nil {
		children = []map[string]interface{}{
			{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]string{
								"content": content,
							},
						},
					},
				},
			},
		}
	}

	if children == nil {
		return nil, fmt.Errorf("children blocks are required")
	}

	body := map[string]interface{}{
		"children": children,
	}

	endpoint := fmt.Sprintf("https://api.notion.com/v1/blocks/%s/children", blockId)
	bodyJSON, _ := json.Marshal(body)

	return n.makeRequest(ctx, "PATCH", endpoint, bodyJSON, apiKey)
}

func (n *NotionNode) search(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	body := map[string]interface{}{}

	if query := core.GetString(config, "query", ""); query != "" {
		body["query"] = query
	}

	if filter, ok := config["filter"].(map[string]interface{}); ok {
		body["filter"] = filter
	}

	if sort, ok := config["sort"].(map[string]interface{}); ok {
		body["sort"] = sort
	}

	if pageSize := core.GetInt(config, "pageSize", 0); pageSize > 0 {
		body["page_size"] = pageSize
	}

	endpoint := "https://api.notion.com/v1/search"
	bodyJSON, _ := json.Marshal(body)

	return n.makeRequest(ctx, "POST", endpoint, bodyJSON, apiKey)
}

func (n *NotionNode) getUser(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	userId := core.GetString(config, "userId", "")
	if userId == "" {
		return nil, fmt.Errorf("userId is required")
	}

	endpoint := fmt.Sprintf("https://api.notion.com/v1/users/%s", userId)
	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *NotionNode) listUsers(ctx context.Context, apiKey string) (map[string]interface{}, error) {
	endpoint := "https://api.notion.com/v1/users"
	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *NotionNode) makeRequest(ctx context.Context, method, endpoint string, body []byte, apiKey string) (map[string]interface{}, error) {
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
	req.Header.Set("Notion-Version", "2022-06-28")

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

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode >= 400 {
		errMsg := "Notion API error"
		if msg, ok := result["message"].(string); ok {
			errMsg = msg
		}
		return nil, fmt.Errorf("%s (status %d)", errMsg, resp.StatusCode)
	}

	return result, nil
}

// Note: NotionNode is already registered in integrations/init.go
