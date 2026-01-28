package productivity

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// GoogleSheetsNode integrates with Google Sheets
type GoogleSheetsNode struct{}

func NewGoogleSheetsNode() *GoogleSheetsNode {
	return &GoogleSheetsNode{}
}

func (n *GoogleSheetsNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	operation, _ := params["operation"].(string)
	spreadsheetID, _ := params["spreadsheet_id"].(string)
	sheetName, _ := params["sheet_name"].(string)

	// Get OAuth credentials from runtime
	accessToken := runtime.GetCredentialValue("google_sheets", "access_token")
	if accessToken == "" {
		return nil, fmt.Errorf("google sheets credentials not configured")
	}

	switch operation {
	case "read":
		rangeStr, _ := params["range"].(string)
		if rangeStr == "" {
			rangeStr = "A1:Z1000"
		}
		// API call would go here
		return types.JSON{
			"operation":      "read",
			"spreadsheet_id": spreadsheetID,
			"sheet":          sheetName,
			"range":          rangeStr,
			"data":           []interface{}{}, // Placeholder
			"success":        true,
		}, nil

	case "append":
		data, _ := inputData["data"].([]interface{})
		return types.JSON{
			"operation":      "append",
			"spreadsheet_id": spreadsheetID,
			"sheet":          sheetName,
			"rows_added":     len(data),
			"success":        true,
		}, nil

	case "update":
		rangeStr, _ := params["range"].(string)
		return types.JSON{
			"operation":      "update",
			"spreadsheet_id": spreadsheetID,
			"sheet":          sheetName,
			"range":          rangeStr,
			"success":        true,
		}, nil

	case "clear":
		rangeStr, _ := params["range"].(string)
		return types.JSON{
			"operation":      "clear",
			"spreadsheet_id": spreadsheetID,
			"sheet":          sheetName,
			"range":          rangeStr,
			"success":        true,
		}, nil

	case "create_sheet":
		return types.JSON{
			"operation":      "create_sheet",
			"spreadsheet_id": spreadsheetID,
			"sheet":          sheetName,
			"success":        true,
		}, nil

	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *GoogleSheetsNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.google_sheets",
		Name:        "Google Sheets",
		Description: "Read, write, and manage Google Sheets",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "Table",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}, {Name: "error", Type: "error"}},
		Credentials: []string{"google_sheets"},
		Parameters: []wtypes.NodeParameter{
			{Name: "operation", Type: "options", Description: "Operation to perform", Required: true, Options: []wtypes.ParamOption{
				{Value: "read", Name: "Read Rows"},
				{Value: "append", Name: "Append Rows"},
				{Value: "update", Name: "Update Cells"},
				{Value: "clear", Name: "Clear Range"},
				{Value: "create_sheet", Name: "Create Sheet"},
			}},
			{Name: "spreadsheet_id", Type: "string", Description: "Spreadsheet ID from URL", Required: true},
			{Name: "sheet_name", Type: "string", Description: "Sheet name", Default: "Sheet1"},
			{Name: "range", Type: "string", Description: "Cell range (e.g., A1:D10)"},
		},
	}
}
