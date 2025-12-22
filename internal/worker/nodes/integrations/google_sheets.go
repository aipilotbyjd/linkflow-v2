package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// GoogleSheetsNode handles Google Sheets operations
type GoogleSheetsNode struct{}

func (n *GoogleSheetsNode) Type() string {
	return "integration.googleSheets"
}

func (n *GoogleSheetsNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	operation := getString(config, "operation", "read")

	// Get credential
	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		return nil, fmt.Errorf("Google OAuth credential is required")
	}

	credID, err := uuid.Parse(credIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid credential ID")
	}

	cred, err := execCtx.GetCredential(credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	accessToken := cred.AccessToken
	if accessToken == "" {
		return nil, fmt.Errorf("Google OAuth access token is required")
	}

	switch operation {
	case "read":
		return n.readSheet(ctx, config, accessToken)
	case "append":
		return n.appendToSheet(ctx, config, execCtx.Input, accessToken)
	case "update":
		return n.updateSheet(ctx, config, execCtx.Input, accessToken)
	case "clear":
		return n.clearSheet(ctx, config, accessToken)
	case "create":
		return n.createSpreadsheet(ctx, config, accessToken)
	case "getSheets":
		return n.getSheets(ctx, config, accessToken)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (n *GoogleSheetsNode) readSheet(ctx context.Context, config map[string]interface{}, token string) (map[string]interface{}, error) {
	spreadsheetID := getString(config, "spreadsheetId", "")
	sheetRange := getString(config, "range", "Sheet1")

	if spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheetId is required")
	}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s",
		spreadsheetID, sheetRange)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Transform to array of objects if headers are present
	if getBool(config, "parseHeaders", true) {
		result["data"] = n.parseRowsWithHeaders(result)
	}

	return result, nil
}

func (n *GoogleSheetsNode) parseRowsWithHeaders(result map[string]interface{}) []map[string]interface{} {
	values, ok := result["values"].([]interface{})
	if !ok || len(values) < 2 {
		return nil
	}

	headers, ok := values[0].([]interface{})
	if !ok {
		return nil
	}

	data := make([]map[string]interface{}, 0, len(values)-1)
	for i := 1; i < len(values); i++ {
		row, ok := values[i].([]interface{})
		if !ok {
			continue
		}

		obj := make(map[string]interface{})
		for j, header := range headers {
			headerStr, _ := header.(string)
			if j < len(row) {
				obj[headerStr] = row[j]
			} else {
				obj[headerStr] = ""
			}
		}
		data = append(data, obj)
	}
	return data
}

func (n *GoogleSheetsNode) appendToSheet(ctx context.Context, config map[string]interface{}, input map[string]interface{}, token string) (map[string]interface{}, error) {
	spreadsheetID := getString(config, "spreadsheetId", "")
	sheetRange := getString(config, "range", "Sheet1")
	values := getArray(config, "values")

	if spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheetId is required")
	}

	// If values not in config, try to get from input
	if len(values) == 0 {
		if inputValues, ok := input["values"].([]interface{}); ok {
			values = inputValues
		}
	}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s:append?valueInputOption=USER_ENTERED",
		spreadsheetID, sheetRange)

	payload := map[string]interface{}{
		"values": values,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var result map[string]interface{}
	_ = json.Unmarshal(respBody, &result)

	return result, nil
}

func (n *GoogleSheetsNode) updateSheet(ctx context.Context, config map[string]interface{}, input map[string]interface{}, token string) (map[string]interface{}, error) {
	spreadsheetID := getString(config, "spreadsheetId", "")
	sheetRange := getString(config, "range", "")
	values := getArray(config, "values")

	if spreadsheetID == "" || sheetRange == "" {
		return nil, fmt.Errorf("spreadsheetId and range are required")
	}

	if len(values) == 0 {
		if inputValues, ok := input["values"].([]interface{}); ok {
			values = inputValues
		}
	}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s?valueInputOption=USER_ENTERED",
		spreadsheetID, sheetRange)

	payload := map[string]interface{}{
		"values": values,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var result map[string]interface{}
	_ = json.Unmarshal(respBody, &result)

	return result, nil
}

func (n *GoogleSheetsNode) clearSheet(ctx context.Context, config map[string]interface{}, token string) (map[string]interface{}, error) {
	spreadsheetID := getString(config, "spreadsheetId", "")
	sheetRange := getString(config, "range", "")

	if spreadsheetID == "" || sheetRange == "" {
		return nil, fmt.Errorf("spreadsheetId and range are required")
	}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s:clear",
		spreadsheetID, sheetRange)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	return map[string]interface{}{
		"success":       true,
		"clearedRange":  sheetRange,
		"spreadsheetId": spreadsheetID,
	}, nil
}

func (n *GoogleSheetsNode) createSpreadsheet(ctx context.Context, config map[string]interface{}, token string) (map[string]interface{}, error) {
	title := getString(config, "title", "New Spreadsheet")

	url := "https://sheets.googleapis.com/v4/spreadsheets"

	payload := map[string]interface{}{
		"properties": map[string]interface{}{
			"title": title,
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var result map[string]interface{}
	_ = json.Unmarshal(respBody, &result)

	return result, nil
}

func (n *GoogleSheetsNode) getSheets(ctx context.Context, config map[string]interface{}, token string) (map[string]interface{}, error) {
	spreadsheetID := getString(config, "spreadsheetId", "")

	if spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheetId is required")
	}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s?fields=sheets.properties",
		spreadsheetID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var result map[string]interface{}
	_ = json.Unmarshal(body, &result)

	return result, nil
}
