package transform

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// CSVParseNode parses CSV strings into arrays
type CSVParseNode struct{}

func NewCSVParseNode() *CSVParseNode {
	return &CSVParseNode{}
}

func (n *CSVParseNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	csvStr, _ := params["csv_string"].(string)
	if csvStr == "" {
		if str, ok := inputData["data"].(string); ok {
			csvStr = str
		}
	}

	if csvStr == "" {
		return nil, fmt.Errorf("csv_string is required")
	}

	delimiter := ","
	if d, ok := params["delimiter"].(string); ok && d != "" {
		delimiter = d
	}

	hasHeader, _ := params["has_header"].(bool)

	reader := csv.NewReader(strings.NewReader(csvStr))
	reader.Comma = rune(delimiter[0])

	records, err := reader.ReadAll()
	if err != nil {
		return types.JSON{"error": err.Error(), "success": false}, nil
	}

	if len(records) == 0 {
		return types.JSON{"data": []interface{}{}, "success": true}, nil
	}

	var result []map[string]interface{}
	var headers []string

	if hasHeader && len(records) > 0 {
		headers = records[0]
		records = records[1:]
	}

	for _, record := range records {
		row := make(map[string]interface{})
		for i, value := range record {
			if hasHeader && i < len(headers) {
				row[headers[i]] = value
			} else {
				row[fmt.Sprintf("col_%d", i)] = value
			}
		}
		result = append(result, row)
	}

	return types.JSON{
		"data":    result,
		"headers": headers,
		"count":   len(result),
		"success": true,
	}, nil
}

func (n *CSVParseNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "transform.csv_parse",
		Name:        "CSV Parse",
		Description: "Parse CSV string into array of objects",
		Category:    "transform",
		Version:     "1.0.0",
		Icon:        "File02",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "string"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "array"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "csv_string", Type: "string", Description: "CSV string to parse"},
			{Name: "delimiter", Type: "string", Description: "Column delimiter", Default: ","},
			{Name: "has_header", Type: "boolean", Description: "First row is header", Default: true},
		},
	}
}

// CSVStringifyNode converts arrays to CSV strings
type CSVStringifyNode struct{}

func NewCSVStringifyNode() *CSVStringifyNode {
	return &CSVStringifyNode{}
}

func (n *CSVStringifyNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	data, _ := inputData["data"].([]interface{})
	if data == nil {
		if items, ok := inputData["items"].([]interface{}); ok {
			data = items
		}
	}

	if len(data) == 0 {
		return types.JSON{"csv": "", "success": true}, nil
	}

	delimiter := ","
	if d, ok := params["delimiter"].(string); ok && d != "" {
		delimiter = d
	}

	includeHeader, _ := params["include_header"].(bool)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = rune(delimiter[0])

	// Get headers from first row
	var headers []string
	if first, ok := data[0].(map[string]interface{}); ok {
		for k := range first {
			headers = append(headers, k)
		}
	}

	if includeHeader && len(headers) > 0 {
		writer.Write(headers)
	}

	for _, item := range data {
		if row, ok := item.(map[string]interface{}); ok {
			var record []string
			for _, h := range headers {
				record = append(record, fmt.Sprintf("%v", row[h]))
			}
			writer.Write(record)
		}
	}

	writer.Flush()

	return types.JSON{"csv": buf.String(), "success": true}, nil
}

func (n *CSVStringifyNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "transform.csv_stringify",
		Name:        "CSV Stringify",
		Description: "Convert array of objects to CSV string",
		Category:    "transform",
		Version:     "1.0.0",
		Icon:        "File02",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "array"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "string"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "delimiter", Type: "string", Description: "Column delimiter", Default: ","},
			{Name: "include_header", Type: "boolean", Description: "Include header row", Default: true},
		},
	}
}
