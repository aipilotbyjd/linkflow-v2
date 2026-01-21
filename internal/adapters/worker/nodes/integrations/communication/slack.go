package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SlackNode struct {
	client *http.Client
}

func NewSlackNode() *SlackNode {
	return &SlackNode{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (n *SlackNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	webhookURL, _ := params["webhook_url"].(string)
	channel, _ := params["channel"].(string)
	message, _ := params["message"].(string)

	if webhookURL == "" {
		return nil, fmt.Errorf("Slack webhook URL is required")
	}
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}

	payload := map[string]interface{}{
		"text": message,
	}
	if channel != "" {
		payload["channel"] = channel
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	return types.JSON{
		"status":      "sent",
		"status_code": resp.StatusCode,
		"channel":     channel,
	}, nil
}

func (n *SlackNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.slack",
		Name:        "Slack",
		Description: "Send messages to Slack",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
