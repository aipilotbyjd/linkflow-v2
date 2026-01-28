package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// JSONParseNode parses JSON strings
type JSONParseNode struct{}

func NewJSONParseNode() *JSONParseNode {
	return &JSONParseNode{}
}

func (n *JSONParseNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	// Get JSON string from input or parameter
	jsonStr, _ := params["json_string"].(string)
	if jsonStr == "" {
		if str, ok := inputData["data"].(string); ok {
			jsonStr = str
		}
	}

	if jsonStr == "" {
		return nil, fmt.Errorf("json_string is required")
	}

	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return types.JSON{"error": err.Error(), "success": false}, nil
	}

	return types.JSON{"data": result, "success": true}, nil
}

func (n *JSONParseNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "transform.json_parse",
		Name:        "JSON Parse",
		Description: "Parse a JSON string into an object",
		Category:    "transform",
		Version:     "1.0.0",
		Icon:        "BracketsCheck",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "string"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "json_string", Type: "string", Description: "JSON string to parse"},
		},
	}
}

// JSONStringifyNode converts objects to JSON strings
type JSONStringifyNode struct{}

func NewJSONStringifyNode() *JSONStringifyNode {
	return &JSONStringifyNode{}
}

func (n *JSONStringifyNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	data := inputData
	if d, ok := params["data"]; ok {
		data = types.JSON{"data": d}
	}

	pretty, _ := params["pretty"].(bool)

	var result []byte
	var err error
	if pretty {
		result, err = json.MarshalIndent(data, "", "  ")
	} else {
		result, err = json.Marshal(data)
	}

	if err != nil {
		return types.JSON{"error": err.Error(), "success": false}, nil
	}

	return types.JSON{"json": string(result), "success": true}, nil
}

func (n *JSONStringifyNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "transform.json_stringify",
		Name:        "JSON Stringify",
		Description: "Convert an object to a JSON string",
		Category:    "transform",
		Version:     "1.0.0",
		Icon:        "BracketsCheck",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "string"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "pretty", Type: "boolean", Description: "Pretty print with indentation", Default: false},
		},
	}
}
