package transform

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SetNode struct{}

func NewSetNode() *SetNode {
	return &SetNode{}
}

func (n *SetNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	values, _ := params["values"].(map[string]interface{})

	result := make(types.JSON)
	for k, v := range values {
		result[k] = v
	}

	return result, nil
}

func (n *SetNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.set",
		Name:        "Set",
		Description: "Set, transform, or rename fields in your data. Create new objects with computed values.",
		Category:    "action",
		Version:     "1.0.0",
		Icon:        "edit-3",
		Color:       "#F97316",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data to transform"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "Object with set values"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "mode",
				DisplayName: "Mode",
				Type:        "options",
				Required:    true,
				Default:     "manual",
				Description: "How to define the output values",
				Options: []wtypes.ParamOption{
					{Name: "Manual Entry", Value: "manual", Description: "Define each field manually"},
					{Name: "JSON", Value: "json", Description: "Define values as JSON object"},
					{Name: "Expression", Value: "expression", Description: "Use JavaScript to transform"},
				},
			},
			{
				Name:        "values",
				DisplayName: "Values",
				Type:        "json",
				Required:    true,
				Description: "Key-value pairs to set (supports expressions in values)",
				Placeholder: `{"name": "{{$input.firstName}}", "total": "{{$input.price * $input.quantity}}"}`,
				ShowIf:      "mode === 'json' || mode === 'manual'",
			},
			{
				Name:        "expression",
				DisplayName: "Expression",
				Type:        "code",
				Required:    true,
				Description: "JavaScript code that returns an object",
				Placeholder: "return { fullName: $input.firstName + ' ' + $input.lastName, total: $input.items.length }",
				ShowIf:      "mode === 'expression'",
			},
			{
				Name:        "keep_existing",
				DisplayName: "Keep Existing Fields",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Merge new values with input data (true) or output only new values (false)",
			},
			{
				Name:        "dot_notation",
				DisplayName: "Support Dot Notation",
				Type:        "boolean",
				Required:    false,
				Default:     true,
				Description: "Allow setting nested fields with dot notation (e.g., 'user.name')",
			},
			{
				Name:        "remove_fields",
				DisplayName: "Fields to Remove",
				Type:        "json",
				Required:    false,
				Description: "Array of field names to remove from output",
				Placeholder: `["password", "secret"]`,
				ShowIf:      "keep_existing === true",
			},
		},
	}
}
