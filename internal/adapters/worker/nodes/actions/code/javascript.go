package code

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type JavaScriptNode struct{}

func NewJavaScriptNode() *JavaScriptNode {
	return &JavaScriptNode{}
}

func (n *JavaScriptNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	code, _ := params["code"].(string)
	if code == "" {
		return nil, fmt.Errorf("JavaScript code is required")
	}

	// TODO: Implement JavaScript execution with goja or similar
	return types.JSON{
		"executed": true,
		"language": "javascript",
		"warning":  "JavaScript execution not implemented",
	}, nil
}

func (n *JavaScriptNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.code",
		Name:        "JavaScript",
		Description: "Execute JavaScript code",
		Category:    "action",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "code", DisplayName: "Code", Type: "code", Required: true},
		},
	}
}
