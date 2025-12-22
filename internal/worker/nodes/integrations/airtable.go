package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// AirtableNode handles Airtable operations
type AirtableNode struct{}

func (n *AirtableNode) Type() string {
	return "integrations.airtable"
}

func (n *AirtableNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	apiKey := core.GetString(config, "apiKey", "")
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey is required")
	}

	baseId := core.GetString(config, "baseId", "")
	tableName := core.GetString(config, "tableName", "")

	operation := core.GetString(config, "operation", "list")

	switch operation {
	case "list":
		return n.listRecords(ctx, config, baseId, tableName, apiKey)
	case "get":
		return n.getRecord(ctx, config, baseId, tableName, apiKey)
	case "create":
		return n.createRecord(ctx, config, baseId, tableName, apiKey, execCtx.Input)
	case "update":
		return n.updateRecord(ctx, config, baseId, tableName, apiKey, execCtx.Input)
	case "delete":
		return n.deleteRecord(ctx, config, baseId, tableName, apiKey)
	case "search":
		return n.searchRecords(ctx, config, baseId, tableName, apiKey)
	case "listBases":
		return n.listBases(ctx, apiKey)
	case "listTables":
		return n.listTables(ctx, config, apiKey)
	default:
		return n.listRecords(ctx, config, baseId, tableName, apiKey)
	}
}

func (n *AirtableNode) listRecords(ctx context.Context, config map[string]interface{}, baseId, tableName, apiKey string) (map[string]interface{}, error) {
	if baseId == "" || tableName == "" {
		return nil, fmt.Errorf("baseId and tableName are required")
	}

	endpoint := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", baseId, url.PathEscape(tableName))

	params := url.Values{}
	if maxRecords := core.GetInt(config, "maxRecords", 0); maxRecords > 0 {
		params.Set("maxRecords", fmt.Sprintf("%d", maxRecords))
	}
	if view := core.GetString(config, "view", ""); view != "" {
		params.Set("view", view)
	}
	if filterFormula := core.GetString(config, "filterByFormula", ""); filterFormula != "" {
		params.Set("filterByFormula", filterFormula)
	}
	if sort := core.GetString(config, "sort", ""); sort != "" {
		params.Set("sort[0][field]", sort)
		params.Set("sort[0][direction]", core.GetString(config, "sortDirection", "asc"))
	}
	if offset := core.GetString(config, "offset", ""); offset != "" {
		params.Set("offset", offset)
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	result, err := n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"records": result["records"],
		"offset":  result["offset"],
	}, nil
}

func (n *AirtableNode) getRecord(ctx context.Context, config map[string]interface{}, baseId, tableName, apiKey string) (map[string]interface{}, error) {
	if baseId == "" || tableName == "" {
		return nil, fmt.Errorf("baseId and tableName are required")
	}

	recordId := core.GetString(config, "recordId", "")
	if recordId == "" {
		return nil, fmt.Errorf("recordId is required")
	}

	endpoint := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s", baseId, url.PathEscape(tableName), recordId)
	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *AirtableNode) createRecord(ctx context.Context, config map[string]interface{}, baseId, tableName, apiKey string, input map[string]interface{}) (map[string]interface{}, error) {
	if baseId == "" || tableName == "" {
		return nil, fmt.Errorf("baseId and tableName are required")
	}

	fields := config["fields"]
	if fields == nil {
		fields = input["fields"]
	}
	if fields == nil {
		fields = input
	}

	body := map[string]interface{}{
		"fields": fields,
	}

	// Support batch create
	if records, ok := config["records"].([]interface{}); ok {
		body = map[string]interface{}{
			"records": records,
		}
	}

	endpoint := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", baseId, url.PathEscape(tableName))
	bodyJSON, _ := json.Marshal(body)

	result, err := n.makeRequest(ctx, "POST", endpoint, bodyJSON, apiKey)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"created": true,
		"id":      result["id"],
		"fields":  result["fields"],
		"records": result["records"],
	}, nil
}

func (n *AirtableNode) updateRecord(ctx context.Context, config map[string]interface{}, baseId, tableName, apiKey string, input map[string]interface{}) (map[string]interface{}, error) {
	if baseId == "" || tableName == "" {
		return nil, fmt.Errorf("baseId and tableName are required")
	}

	recordId := core.GetString(config, "recordId", "")
	if recordId == "" {
		return nil, fmt.Errorf("recordId is required")
	}

	fields := config["fields"]
	if fields == nil {
		fields = input["fields"]
	}
	if fields == nil {
		fields = input
	}

	body := map[string]interface{}{
		"fields": fields,
	}

	endpoint := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s", baseId, url.PathEscape(tableName), recordId)
	bodyJSON, _ := json.Marshal(body)

	result, err := n.makeRequest(ctx, "PATCH", endpoint, bodyJSON, apiKey)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"updated":  true,
		"id":       result["id"],
		"fields":   result["fields"],
	}, nil
}

func (n *AirtableNode) deleteRecord(ctx context.Context, config map[string]interface{}, baseId, tableName, apiKey string) (map[string]interface{}, error) {
	if baseId == "" || tableName == "" {
		return nil, fmt.Errorf("baseId and tableName are required")
	}

	recordId := core.GetString(config, "recordId", "")
	if recordId == "" {
		return nil, fmt.Errorf("recordId is required")
	}

	endpoint := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s", baseId, url.PathEscape(tableName), recordId)

	result, err := n.makeRequest(ctx, "DELETE", endpoint, nil, apiKey)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"deleted": result["deleted"],
		"id":      result["id"],
	}, nil
}

func (n *AirtableNode) searchRecords(ctx context.Context, config map[string]interface{}, baseId, tableName, apiKey string) (map[string]interface{}, error) {
	if baseId == "" || tableName == "" {
		return nil, fmt.Errorf("baseId and tableName are required")
	}

	field := core.GetString(config, "field", "")
	value := core.GetString(config, "value", "")

	if field == "" || value == "" {
		return nil, fmt.Errorf("field and value are required for search")
	}

	formula := fmt.Sprintf("{%s}='%s'", field, value)

	endpoint := fmt.Sprintf("https://api.airtable.com/v0/%s/%s?filterByFormula=%s",
		baseId, url.PathEscape(tableName), url.QueryEscape(formula))

	result, err := n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"records": result["records"],
		"query":   formula,
	}, nil
}

func (n *AirtableNode) listBases(ctx context.Context, apiKey string) (map[string]interface{}, error) {
	endpoint := "https://api.airtable.com/v0/meta/bases"
	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *AirtableNode) listTables(ctx context.Context, config map[string]interface{}, apiKey string) (map[string]interface{}, error) {
	baseId := core.GetString(config, "baseId", "")
	if baseId == "" {
		return nil, fmt.Errorf("baseId is required")
	}

	endpoint := fmt.Sprintf("https://api.airtable.com/v0/meta/bases/%s/tables", baseId)
	return n.makeRequest(ctx, "GET", endpoint, nil, apiKey)
}

func (n *AirtableNode) makeRequest(ctx context.Context, method, endpoint string, body []byte, apiKey string) (map[string]interface{}, error) {
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

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode >= 400 {
		errMsg := "Airtable API error"
		if err, ok := result["error"].(map[string]interface{}); ok {
			if msg, ok := err["message"].(string); ok {
				errMsg = msg
			}
		}
		return nil, fmt.Errorf("%s (status %d)", errMsg, resp.StatusCode)
	}

	return result, nil
}

// Note: AirtableNode is already registered in integrations/init.go
