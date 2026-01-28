package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type MergeNode struct{}

func NewMergeNode() *MergeNode {
	return &MergeNode{}
}

func (n *MergeNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	result := types.JSON{
		"merged": true,
	}
	return result, nil
}

func (n *MergeNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.merge",
		Name:        "Merge",
		Description: "Combine data from multiple branches into a single output. Useful for joining parallel paths.",
		Category:    "logic",
		Version:     "1.0.0",
		Icon:        "GitMerge",
		Color:       "#A855F7",
		Inputs: []wtypes.NodePort{
			{Name: "input1", Type: "any", Description: "First input branch"},
			{Name: "input2", Type: "any", Description: "Second input branch"},
			{Name: "input3", Type: "any", Description: "Third input branch (optional)"},
			{Name: "input4", Type: "any", Description: "Fourth input branch (optional)"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Merged output data"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "mode",
				DisplayName: "Merge Mode",
				Type:        "options",
				Required:    true,
				Default:     "combine",
				Description: "How to merge the inputs",
				Options: []wtypes.ParamOption{
					{Name: "Combine Objects", Value: "combine", Description: "Merge objects into one (later overwrites earlier)"},
					{Name: "Append to Array", Value: "append", Description: "Create array with all inputs"},
					{Name: "Merge Arrays", Value: "concat", Description: "Concatenate arrays from all inputs"},
					{Name: "Wait for All", Value: "wait_all", Description: "Wait for all inputs before outputting"},
					{Name: "Pass First", Value: "first", Description: "Output immediately when first input arrives"},
					{Name: "Zip", Value: "zip", Description: "Pair items from arrays by index"},
					{Name: "Multiplex", Value: "multiplex", Description: "Create all combinations of inputs"},
				},
			},
			{
				Name:        "output_as",
				DisplayName: "Output Format",
				Type:        "options",
				Required:    false,
				Default:     "auto",
				Description: "How to format the merged output",
				Options: []wtypes.ParamOption{
					{Name: "Auto", Value: "auto"},
					{Name: "Object", Value: "object"},
					{Name: "Array", Value: "array"},
				},
			},
			{
				Name:        "deep_merge",
				DisplayName: "Deep Merge",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Recursively merge nested objects (only for combine mode)",
				ShowIf:      "mode === 'combine'",
			},
			{
				Name:        "include_empty",
				DisplayName: "Include Empty Inputs",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Include inputs that have no data (null/undefined)",
			},
			{
				Name:        "key_field",
				DisplayName: "Key Field",
				Type:        "string",
				Required:    false,
				Description: "Field to use as key when merging arrays of objects",
				Placeholder: "id",
				ShowIf:      "mode === 'concat'",
			},
		},
	}
}
