package cloud

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type GoogleDriveNode struct{}

func NewGoogleDriveNode() *GoogleDriveNode {
	return &GoogleDriveNode{}
}

func (n *GoogleDriveNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "list":
		return n.listFiles(ctx, params)
	case "get":
		return n.getFile(ctx, params)
	case "create":
		return n.createFile(ctx, params)
	case "update":
		return n.updateFile(ctx, params)
	case "delete":
		return n.deleteFile(ctx, params)
	case "download":
		return n.downloadFile(ctx, params)
	case "upload":
		return n.uploadFile(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported Google Drive operation: %s", operation)
	}
}

func (n *GoogleDriveNode) listFiles(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	folderId, _ := params["folder_id"].(string)
	query, _ := params["query"].(string)

	return types.JSON{
		"folder_id": folderId,
		"query":     query,
		"files":     []interface{}{},
	}, nil
}

func (n *GoogleDriveNode) getFile(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	fileId, _ := params["file_id"].(string)

	return types.JSON{
		"file_id": fileId,
		"name":    "",
		"mime":    "",
	}, nil
}

func (n *GoogleDriveNode) createFile(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	name, _ := params["name"].(string)
	mimeType, _ := params["mime_type"].(string)
	parentId, _ := params["parent_id"].(string)

	return types.JSON{
		"file_id":   "",
		"name":      name,
		"mime_type": mimeType,
		"parent_id": parentId,
	}, nil
}

func (n *GoogleDriveNode) updateFile(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	fileId, _ := params["file_id"].(string)

	return types.JSON{
		"file_id": fileId,
		"updated": true,
	}, nil
}

func (n *GoogleDriveNode) deleteFile(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	fileId, _ := params["file_id"].(string)

	return types.JSON{
		"file_id": fileId,
		"deleted": true,
	}, nil
}

func (n *GoogleDriveNode) downloadFile(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	fileId, _ := params["file_id"].(string)

	return types.JSON{
		"file_id": fileId,
		"content": nil,
	}, nil
}

func (n *GoogleDriveNode) uploadFile(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	name, _ := params["name"].(string)
	content := params["content"]
	parentId, _ := params["parent_id"].(string)

	return types.JSON{
		"file_id":   "",
		"name":      name,
		"parent_id": parentId,
		"size":      len(fmt.Sprint(content)),
	}, nil
}

func (n *GoogleDriveNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.google_drive",
		Name:        "Google Drive",
		Description: "Interact with Google Drive",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
