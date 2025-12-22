package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// SalesforceNode handles Salesforce operations
type SalesforceNode struct{}

func (n *SalesforceNode) Type() string {
	return "integrations.salesforce"
}

func (n *SalesforceNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	instanceUrl := core.GetString(config, "instanceUrl", "")
	accessToken := core.GetString(config, "accessToken", "")

	if instanceUrl == "" || accessToken == "" {
		return nil, fmt.Errorf("instanceUrl and accessToken are required")
	}

	instanceUrl = strings.TrimSuffix(instanceUrl, "/")
	apiVersion := core.GetString(config, "apiVersion", "v58.0")

	operation := core.GetString(config, "operation", "query")

	switch operation {
	case "query":
		return n.query(ctx, config, instanceUrl, apiVersion, accessToken)
	case "get":
		return n.getRecord(ctx, config, instanceUrl, apiVersion, accessToken)
	case "create":
		return n.createRecord(ctx, config, instanceUrl, apiVersion, accessToken, execCtx.Input)
	case "update":
		return n.updateRecord(ctx, config, instanceUrl, apiVersion, accessToken, execCtx.Input)
	case "delete":
		return n.deleteRecord(ctx, config, instanceUrl, apiVersion, accessToken)
	case "upsert":
		return n.upsertRecord(ctx, config, instanceUrl, apiVersion, accessToken, execCtx.Input)
	case "describe":
		return n.describeObject(ctx, config, instanceUrl, apiVersion, accessToken)
	case "describeGlobal":
		return n.describeGlobal(ctx, instanceUrl, apiVersion, accessToken)
	case "search":
		return n.search(ctx, config, instanceUrl, apiVersion, accessToken)
	default:
		return n.query(ctx, config, instanceUrl, apiVersion, accessToken)
	}
}

func (n *SalesforceNode) query(ctx context.Context, config map[string]interface{}, instanceUrl, apiVersion, accessToken string) (map[string]interface{}, error) {
	soql := core.GetString(config, "query", "")
	if soql == "" {
		return nil, fmt.Errorf("query (SOQL) is required")
	}

	endpoint := fmt.Sprintf("%s/services/data/%s/query?q=%s", instanceUrl, apiVersion, url.QueryEscape(soql))

	result, err := n.makeRequest(ctx, "GET", endpoint, nil, accessToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"records":    result["records"],
		"totalSize":  result["totalSize"],
		"done":       result["done"],
		"nextRecordsUrl": result["nextRecordsUrl"],
	}, nil
}

func (n *SalesforceNode) getRecord(ctx context.Context, config map[string]interface{}, instanceUrl, apiVersion, accessToken string) (map[string]interface{}, error) {
	objectType := core.GetString(config, "objectType", "")
	recordId := core.GetString(config, "recordId", "")

	if objectType == "" || recordId == "" {
		return nil, fmt.Errorf("objectType and recordId are required")
	}

	endpoint := fmt.Sprintf("%s/services/data/%s/sobjects/%s/%s", instanceUrl, apiVersion, objectType, recordId)

	fields := core.GetString(config, "fields", "")
	if fields != "" {
		endpoint += "?fields=" + fields
	}

	return n.makeRequest(ctx, "GET", endpoint, nil, accessToken)
}

func (n *SalesforceNode) createRecord(ctx context.Context, config map[string]interface{}, instanceUrl, apiVersion, accessToken string, input map[string]interface{}) (map[string]interface{}, error) {
	objectType := core.GetString(config, "objectType", "")
	if objectType == "" {
		return nil, fmt.Errorf("objectType is required")
	}

	data := config["data"]
	if data == nil {
		data = input["data"]
	}
	if data == nil {
		data = input
	}

	endpoint := fmt.Sprintf("%s/services/data/%s/sobjects/%s", instanceUrl, apiVersion, objectType)

	bodyJSON, _ := json.Marshal(data)
	result, err := n.makeRequest(ctx, "POST", endpoint, bodyJSON, accessToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"created": true,
		"id":      result["id"],
		"success": result["success"],
		"errors":  result["errors"],
	}, nil
}

func (n *SalesforceNode) updateRecord(ctx context.Context, config map[string]interface{}, instanceUrl, apiVersion, accessToken string, input map[string]interface{}) (map[string]interface{}, error) {
	objectType := core.GetString(config, "objectType", "")
	recordId := core.GetString(config, "recordId", "")

	if objectType == "" || recordId == "" {
		return nil, fmt.Errorf("objectType and recordId are required")
	}

	data := config["data"]
	if data == nil {
		data = input["data"]
	}
	if data == nil {
		data = input
	}

	endpoint := fmt.Sprintf("%s/services/data/%s/sobjects/%s/%s", instanceUrl, apiVersion, objectType, recordId)

	bodyJSON, _ := json.Marshal(data)
	_, err := n.makeRequest(ctx, "PATCH", endpoint, bodyJSON, accessToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"updated":  true,
		"id":       recordId,
		"objectType": objectType,
	}, nil
}

func (n *SalesforceNode) deleteRecord(ctx context.Context, config map[string]interface{}, instanceUrl, apiVersion, accessToken string) (map[string]interface{}, error) {
	objectType := core.GetString(config, "objectType", "")
	recordId := core.GetString(config, "recordId", "")

	if objectType == "" || recordId == "" {
		return nil, fmt.Errorf("objectType and recordId are required")
	}

	endpoint := fmt.Sprintf("%s/services/data/%s/sobjects/%s/%s", instanceUrl, apiVersion, objectType, recordId)

	_, err := n.makeRequest(ctx, "DELETE", endpoint, nil, accessToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"deleted":    true,
		"id":         recordId,
		"objectType": objectType,
	}, nil
}

func (n *SalesforceNode) upsertRecord(ctx context.Context, config map[string]interface{}, instanceUrl, apiVersion, accessToken string, input map[string]interface{}) (map[string]interface{}, error) {
	objectType := core.GetString(config, "objectType", "")
	externalIdField := core.GetString(config, "externalIdField", "")
	externalIdValue := core.GetString(config, "externalIdValue", "")

	if objectType == "" || externalIdField == "" || externalIdValue == "" {
		return nil, fmt.Errorf("objectType, externalIdField, and externalIdValue are required")
	}

	data := config["data"]
	if data == nil {
		data = input["data"]
	}
	if data == nil {
		data = input
	}

	endpoint := fmt.Sprintf("%s/services/data/%s/sobjects/%s/%s/%s",
		instanceUrl, apiVersion, objectType, externalIdField, externalIdValue)

	bodyJSON, _ := json.Marshal(data)
	result, err := n.makeRequest(ctx, "PATCH", endpoint, bodyJSON, accessToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"upserted": true,
		"id":       result["id"],
		"created":  result["created"],
	}, nil
}

func (n *SalesforceNode) describeObject(ctx context.Context, config map[string]interface{}, instanceUrl, apiVersion, accessToken string) (map[string]interface{}, error) {
	objectType := core.GetString(config, "objectType", "")
	if objectType == "" {
		return nil, fmt.Errorf("objectType is required")
	}

	endpoint := fmt.Sprintf("%s/services/data/%s/sobjects/%s/describe", instanceUrl, apiVersion, objectType)
	return n.makeRequest(ctx, "GET", endpoint, nil, accessToken)
}

func (n *SalesforceNode) describeGlobal(ctx context.Context, instanceUrl, apiVersion, accessToken string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/services/data/%s/sobjects", instanceUrl, apiVersion)
	return n.makeRequest(ctx, "GET", endpoint, nil, accessToken)
}

func (n *SalesforceNode) search(ctx context.Context, config map[string]interface{}, instanceUrl, apiVersion, accessToken string) (map[string]interface{}, error) {
	sosl := core.GetString(config, "search", "")
	if sosl == "" {
		return nil, fmt.Errorf("search (SOSL) is required")
	}

	endpoint := fmt.Sprintf("%s/services/data/%s/search?q=%s", instanceUrl, apiVersion, url.QueryEscape(sosl))
	return n.makeRequest(ctx, "GET", endpoint, nil, accessToken)
}

func (n *SalesforceNode) makeRequest(ctx context.Context, method, endpoint string, body []byte, accessToken string) (map[string]interface{}, error) {
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

	req.Header.Set("Authorization", "Bearer "+accessToken)
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

	// Handle empty response (204 No Content)
	if len(respBody) == 0 {
		return map[string]interface{}{"success": true}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Try as array (search results)
		var arr []interface{}
		if err2 := json.Unmarshal(respBody, &arr); err2 == nil {
			return map[string]interface{}{"searchRecords": arr}, nil
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode >= 400 {
		errMsg := "Salesforce API error"
		if msg, ok := result["message"].(string); ok {
			errMsg = msg
		}
		if errs, ok := result["errors"].([]interface{}); ok && len(errs) > 0 {
			errMsg = fmt.Sprintf("%v", errs)
		}
		return nil, fmt.Errorf("%s (status %d)", errMsg, resp.StatusCode)
	}

	return result, nil
}

// Note: SalesforceNode is registered in integrations/init.go
