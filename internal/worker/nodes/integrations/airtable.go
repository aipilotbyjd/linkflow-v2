package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

type AirtableNode struct{}

func NewAirtableNode() *AirtableNode {
	return &AirtableNode{}
}

func (n *AirtableNode) Type() string {
	return "integration.airtable"
}

func (n *AirtableNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
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
	operation := getString(execCtx.Config, "operation", "listRecords")

	switch operation {
	case "listRecords":
		return n.listRecords(ctx, apiKey, execCtx.Config)
	case "getRecord":
		return n.getRecord(ctx, apiKey, execCtx.Config)
	case "createRecord":
		return n.createRecord(ctx, apiKey, execCtx.Config)
	case "createRecords":
		return n.createRecords(ctx, apiKey, execCtx.Config)
	case "updateRecord":
		return n.updateRecord(ctx, apiKey, execCtx.Config)
	case "updateRecords":
		return n.updateRecords(ctx, apiKey, execCtx.Config)
	case "deleteRecord":
		return n.deleteRecord(ctx, apiKey, execCtx.Config)
	case "deleteRecords":
		return n.deleteRecords(ctx, apiKey, execCtx.Config)
	case "listBases":
		return n.listBases(ctx, apiKey)
	case "getBaseSchema":
		return n.getBaseSchema(ctx, apiKey, execCtx.Config)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *AirtableNode) listRecords(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	baseID := getString(config, "baseId", "")
	tableID := getString(config, "tableId", "")
	baseURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", baseID, url.PathEscape(tableID))

	params := url.Values{}
	if pageSize := getInt(config, "pageSize", 0); pageSize > 0 {
		params.Set("pageSize", fmt.Sprintf("%d", pageSize))
	}
	if offset := getString(config, "offset", ""); offset != "" {
		params.Set("offset", offset)
	}
	if view := getString(config, "view", ""); view != "" {
		params.Set("view", view)
	}
	if formula := getString(config, "filterByFormula", ""); formula != "" {
		params.Set("filterByFormula", formula)
	}
	if fields, ok := config["fields"].([]interface{}); ok {
		for _, f := range fields {
			params.Add("fields[]", fmt.Sprintf("%v", f))
		}
	}
	if sort, ok := config["sort"].([]interface{}); ok {
		for i, s := range sort {
			if sortMap, ok := s.(map[string]interface{}); ok {
				if field, ok := sortMap["field"].(string); ok {
					params.Add(fmt.Sprintf("sort[%d][field]", i), field)
				}
				if direction, ok := sortMap["direction"].(string); ok {
					params.Add(fmt.Sprintf("sort[%d][direction]", i), direction)
				}
			}
		}
	}

	reqURL := baseURL
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	return n.makeRequest(ctx, apiKey, "GET", reqURL, nil)
}

func (n *AirtableNode) getRecord(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	baseID := getString(config, "baseId", "")
	tableID := getString(config, "tableId", "")
	recordID := getString(config, "recordId", "")
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s", baseID, url.PathEscape(tableID), recordID)
	return n.makeRequest(ctx, apiKey, "GET", url, nil)
}

func (n *AirtableNode) createRecord(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	baseID := getString(config, "baseId", "")
	tableID := getString(config, "tableId", "")
	reqURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", baseID, url.PathEscape(tableID))

	payload := map[string]interface{}{
		"fields": config["fields"],
	}

	if typecast := getBool(config, "typecast", false); typecast {
		payload["typecast"] = true
	}

	return n.makeRequest(ctx, apiKey, "POST", reqURL, payload)
}

func (n *AirtableNode) createRecords(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	baseID := getString(config, "baseId", "")
	tableID := getString(config, "tableId", "")
	reqURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", baseID, url.PathEscape(tableID))

	records, ok := config["records"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("records must be an array")
	}

	payload := map[string]interface{}{
		"records": records,
	}

	if typecast := getBool(config, "typecast", false); typecast {
		payload["typecast"] = true
	}

	return n.makeRequest(ctx, apiKey, "POST", reqURL, payload)
}

func (n *AirtableNode) updateRecord(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	baseID := getString(config, "baseId", "")
	tableID := getString(config, "tableId", "")
	recordID := getString(config, "recordId", "")
	reqURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s", baseID, url.PathEscape(tableID), recordID)

	payload := map[string]interface{}{
		"fields": config["fields"],
	}

	if typecast := getBool(config, "typecast", false); typecast {
		payload["typecast"] = true
	}

	method := "PATCH"
	if replace := getBool(config, "replaceAllFields", false); replace {
		method = "PUT"
	}

	return n.makeRequest(ctx, apiKey, method, reqURL, payload)
}

func (n *AirtableNode) updateRecords(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	baseID := getString(config, "baseId", "")
	tableID := getString(config, "tableId", "")
	reqURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", baseID, url.PathEscape(tableID))

	records, ok := config["records"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("records must be an array")
	}

	payload := map[string]interface{}{
		"records": records,
	}

	if typecast := getBool(config, "typecast", false); typecast {
		payload["typecast"] = true
	}

	method := "PATCH"
	if replace := getBool(config, "replaceAllFields", false); replace {
		method = "PUT"
	}

	return n.makeRequest(ctx, apiKey, method, reqURL, payload)
}

func (n *AirtableNode) deleteRecord(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	baseID := getString(config, "baseId", "")
	tableID := getString(config, "tableId", "")
	recordID := getString(config, "recordId", "")
	reqURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s/%s", baseID, url.PathEscape(tableID), recordID)
	return n.makeRequest(ctx, apiKey, "DELETE", reqURL, nil)
}

func (n *AirtableNode) deleteRecords(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	baseID := getString(config, "baseId", "")
	tableID := getString(config, "tableId", "")
	recordIDs, ok := config["recordIds"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("recordIds must be an array")
	}

	params := url.Values{}
	for _, id := range recordIDs {
		params.Add("records[]", fmt.Sprintf("%v", id))
	}

	reqURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s?%s", baseID, url.PathEscape(tableID), params.Encode())
	return n.makeRequest(ctx, apiKey, "DELETE", reqURL, nil)
}

func (n *AirtableNode) listBases(ctx context.Context, apiKey string) (map[string]interface{}, error) {
	return n.makeRequest(ctx, apiKey, "GET", "https://api.airtable.com/v0/meta/bases", nil)
}

func (n *AirtableNode) getBaseSchema(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	baseID := getString(config, "baseId", "")
	url := fmt.Sprintf("https://api.airtable.com/v0/meta/bases/%s/tables", baseID)
	return n.makeRequest(ctx, apiKey, "GET", url, nil)
}

func (n *AirtableNode) makeRequest(ctx context.Context, apiKey, method, reqURL string, payload map[string]interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
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
		_ = json.Unmarshal(respBody, &errResp)
		if errMsg, ok := errResp["error"].(map[string]interface{}); ok {
			return nil, fmt.Errorf("Airtable API error: %v - %v", errMsg["type"], errMsg["message"])
		}
		return nil, fmt.Errorf("Airtable API error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result, nil
}
