package logic

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type WaitNode struct{}

func NewWaitNode() *WaitNode {
	return &WaitNode{}
}

func (n *WaitNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	seconds := 1.0
	if s, ok := params["seconds"].(float64); ok {
		seconds = s
	}

	select {
	case <-time.After(time.Duration(seconds) * time.Second):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return types.JSON{
		"waited_seconds": seconds,
	}, nil
}

func (n *WaitNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.wait",
		Name:        "Wait",
		Description: "Pause execution for a specified time",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "seconds", DisplayName: "Seconds", Type: "number", Required: false, Default: 1},
		},
	}
}
