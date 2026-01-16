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
	// TODO: Implement AWS S3 get object using AWS SDK
	return types.JSON{
		"bucket": bucket,
		"key":    key,
		"body":   nil,
	}, nil
}

func (n *AWSS3Node) putObject(ctx context.Context, bucket, key string, params map[string]interface{}) (types.JSON, error) {
	body := params["body"]
	contentType, _ := params["content_type"].(string)

	// TODO: Implement AWS S3 put object using AWS SDK
	return types.JSON{
		"bucket":       bucket,
		"key":          key,
		"content_type": contentType,
		"size":         len(fmt.Sprint(body)),
	}, nil
}

func (n *AWSS3Node) deleteObject(ctx context.Context, bucket, key string) (types.JSON, error) {
	// TODO: Implement AWS S3 delete object using AWS SDK
	return types.JSON{
		"bucket":  bucket,
		"key":     key,
		"deleted": true,
	}, nil
}

func (n *AWSS3Node) listObjects(ctx context.Context, bucket string, params map[string]interface{}) (types.JSON, error) {
	prefix, _ := params["prefix"].(string)
	maxKeys, _ := params["max_keys"].(float64)
	if maxKeys == 0 {
		maxKeys = 1000
	}

	// TODO: Implement AWS S3 list objects using AWS SDK
	return types.JSON{
		"bucket":  bucket,
		"prefix":  prefix,
		"objects": []interface{}{},
	}, nil
}

func (n *AWSS3Node) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.aws_s3",
		Name:        "AWS S3",
		Description: "Interact with AWS S3 storage",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
