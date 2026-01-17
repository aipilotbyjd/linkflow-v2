package communication

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type TelegramNode struct{}

func NewTelegramNode() *TelegramNode {
	return &TelegramNode{}
}

func (n *TelegramNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "send_message":
		return n.sendMessage(ctx, params)
	case "send_photo":
		return n.sendPhoto(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported Telegram operation: %s", operation)
	}
}

func (n *TelegramNode) sendMessage(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	chatID, _ := params["chat_id"].(string)
	text, _ := params["text"].(string)

	return types.JSON{
		"success": true,
		"chat_id": chatID,
		"text":    text,
	}, nil
}

func (n *TelegramNode) sendPhoto(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	chatID, _ := params["chat_id"].(string)
	photoURL, _ := params["photo_url"].(string)
	caption, _ := params["caption"].(string)

	return types.JSON{
		"success":   true,
		"chat_id":   chatID,
		"photo_url": photoURL,
		"caption":   caption,
	}, nil
}

func (n *TelegramNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.telegram",
		Name:        "Telegram",
		Description: "Send messages via Telegram",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
