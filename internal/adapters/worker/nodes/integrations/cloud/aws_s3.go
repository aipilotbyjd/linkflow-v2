package cloud

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type AWSS3Node struct{}

func NewAWSS3Node() *AWSS3Node {
	return &AWSS3Node{}
}

func (n *AWSS3Node) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	operation, _ := params["operation"].(string)
	bucket, _ := params["bucket"].(string)
	key, _ := params["key"].(string)

	switch operation {
	case "get":
		return n.getObject(ctx, bucket, key, params)
	case "put":
		return n.putObject(ctx, bucket, key, params)
	case "delete":
		return n.deleteObject(ctx, bucket, key)
	case "list":
		return n.listObjects(ctx, bucket, params)
	default:
		return nil, fmt.Errorf("unsupported S3 operation: %s", operation)
	}
}

func (n *AWSS3Node) getObject(ctx context.Context, bucket, key string, params map[string]interface{}) (types.JSON, error) {
	// S3 operations require credentials from the runtime
	// This returns a placeholder - full implementation requires AWS SDK client
	return types.JSON{
		"bucket":  bucket,
		"key":     key,
		"body":    nil,
		"message": "S3 get object requires AWS credentials configuration",
	}, nil
}

func (n *AWSS3Node) putObject(ctx context.Context, bucket, key string, params map[string]interface{}) (types.JSON, error) {
	body := params["body"]
	contentType, _ := params["content_type"].(string)

	// S3 operations require credentials from the runtime
	// This returns a placeholder - full implementation requires AWS SDK client
	return types.JSON{
		"bucket":       bucket,
		"key":          key,
		"content_type": contentType,
		"size":         len(fmt.Sprint(body)),
		"message":      "S3 put object requires AWS credentials configuration",
	}, nil
}

func (n *AWSS3Node) deleteObject(ctx context.Context, bucket, key string) (types.JSON, error) {
	// S3 operations require credentials from the runtime
	// This returns a placeholder - full implementation requires AWS SDK client
	return types.JSON{
		"bucket":  bucket,
		"key":     key,
		"deleted": false,
		"message": "S3 delete object requires AWS credentials configuration",
	}, nil
}

func (n *AWSS3Node) listObjects(ctx context.Context, bucket string, params map[string]interface{}) (types.JSON, error) {
	prefix, _ := params["prefix"].(string)
	maxKeys, _ := params["max_keys"].(float64)
	if maxKeys == 0 {
		maxKeys = 1000
	}

	// S3 operations require credentials from the runtime
	// This returns a placeholder - full implementation requires AWS SDK client
	return types.JSON{
		"bucket":   bucket,
		"prefix":   prefix,
		"max_keys": int(maxKeys),
		"objects":  []interface{}{},
		"message":  "S3 list objects requires AWS credentials configuration",
	}, nil
}

func (n *AWSS3Node) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.aws_s3",
		Name:        "AWS S3",
		Description: "Interact with AWS S3 storage",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "CloudUpload",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
