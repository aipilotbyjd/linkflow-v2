package integrations

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// AWSS3Node handles AWS S3 operations
type AWSS3Node struct{}

func (n *AWSS3Node) Type() string {
	return "integrations.aws_s3"
}

func (n *AWSS3Node) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	nodeConfig := execCtx.Config

	client, err := n.createClient(ctx, nodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	operation := core.GetString(nodeConfig, "operation", "list")

	switch operation {
	case "list":
		return n.listObjects(ctx, client, nodeConfig)
	case "get":
		return n.getObject(ctx, client, nodeConfig)
	case "put":
		return n.putObject(ctx, client, nodeConfig, execCtx.Input)
	case "delete":
		return n.deleteObject(ctx, client, nodeConfig)
	case "copy":
		return n.copyObject(ctx, client, nodeConfig)
	case "getSignedUrl":
		return n.getSignedURL(ctx, client, nodeConfig)
	case "listBuckets":
		return n.listBuckets(ctx, client)
	default:
		return n.listObjects(ctx, client, nodeConfig)
	}
}

func (n *AWSS3Node) createClient(ctx context.Context, nodeConfig map[string]interface{}) (*s3.Client, error) {
	region := core.GetString(nodeConfig, "region", "us-east-1")
	accessKeyID := core.GetString(nodeConfig, "accessKeyId", "")
	secretAccessKey := core.GetString(nodeConfig, "secretAccessKey", "")
	endpoint := core.GetString(nodeConfig, "endpoint", "")

	var cfg aws.Config
	var err error

	if accessKeyID != "" && secretAccessKey != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKeyID,
				secretAccessKey,
				"",
			)),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}

	if err != nil {
		return nil, err
	}

	opts := []func(*s3.Options){}
	if endpoint != "" {
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	}

	return s3.NewFromConfig(cfg, opts...), nil
}

func (n *AWSS3Node) listObjects(ctx context.Context, client *s3.Client, nodeConfig map[string]interface{}) (map[string]interface{}, error) {
	bucket := core.GetString(nodeConfig, "bucket", "")
	if bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	prefix := core.GetString(nodeConfig, "prefix", "")
	maxKeys := int32(core.GetInt(nodeConfig, "maxKeys", 1000))

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(maxKeys),
	}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	result, err := client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	var objects []map[string]interface{}
	for _, obj := range result.Contents {
		objects = append(objects, map[string]interface{}{
			"key":          aws.ToString(obj.Key),
			"size":         obj.Size,
			"lastModified": obj.LastModified.Format(time.RFC3339),
			"etag":         strings.Trim(aws.ToString(obj.ETag), "\""),
			"storageClass": string(obj.StorageClass),
		})
	}

	return map[string]interface{}{
		"objects":     objects,
		"count":       len(objects),
		"isTruncated": aws.ToBool(result.IsTruncated),
		"bucket":      bucket,
		"prefix":      prefix,
	}, nil
}

func (n *AWSS3Node) getObject(ctx context.Context, client *s3.Client, nodeConfig map[string]interface{}) (map[string]interface{}, error) {
	bucket := core.GetString(nodeConfig, "bucket", "")
	key := core.GetString(nodeConfig, "key", "")

	if bucket == "" || key == "" {
		return nil, fmt.Errorf("bucket and key are required")
	}

	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get failed: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	// Try to return as string, otherwise base64 encode
	content := string(data)
	isBase64 := false
	if !isValidUTF8(data) {
		content = base64.StdEncoding.EncodeToString(data)
		isBase64 = true
	}

	return map[string]interface{}{
		"content":      content,
		"isBase64":     isBase64,
		"size":         len(data),
		"contentType":  aws.ToString(result.ContentType),
		"etag":         strings.Trim(aws.ToString(result.ETag), "\""),
		"lastModified": result.LastModified.Format(time.RFC3339),
		"bucket":       bucket,
		"key":          key,
	}, nil
}

func (n *AWSS3Node) putObject(ctx context.Context, client *s3.Client, nodeConfig map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	bucket := core.GetString(nodeConfig, "bucket", "")
	key := core.GetString(nodeConfig, "key", "")

	if bucket == "" || key == "" {
		return nil, fmt.Errorf("bucket and key are required")
	}

	content := core.GetString(nodeConfig, "content", "")
	if content == "" {
		if c, ok := input["content"].(string); ok {
			content = c
		} else if c, ok := input["data"].(string); ok {
			content = c
		}
	}

	contentType := core.GetString(nodeConfig, "contentType", "")
	if contentType == "" {
		contentType = detectContentType(key)
	}

	// Check if content is base64 encoded
	var data []byte
	if isBase64 := core.GetBool(nodeConfig, "isBase64", false); isBase64 {
		var err error
		data, err = base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
	} else {
		data = []byte(content)
	}

	putInput := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}
	if contentType != "" {
		putInput.ContentType = aws.String(contentType)
	}

	result, err := client.PutObject(ctx, putInput)
	if err != nil {
		return nil, fmt.Errorf("put failed: %w", err)
	}

	return map[string]interface{}{
		"uploaded": true,
		"bucket":   bucket,
		"key":      key,
		"size":     len(data),
		"etag":     strings.Trim(aws.ToString(result.ETag), "\""),
	}, nil
}

func (n *AWSS3Node) deleteObject(ctx context.Context, client *s3.Client, nodeConfig map[string]interface{}) (map[string]interface{}, error) {
	bucket := core.GetString(nodeConfig, "bucket", "")
	key := core.GetString(nodeConfig, "key", "")

	if bucket == "" || key == "" {
		return nil, fmt.Errorf("bucket and key are required")
	}

	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	return map[string]interface{}{
		"deleted": true,
		"bucket":  bucket,
		"key":     key,
	}, nil
}

func (n *AWSS3Node) copyObject(ctx context.Context, client *s3.Client, nodeConfig map[string]interface{}) (map[string]interface{}, error) {
	sourceBucket := core.GetString(nodeConfig, "sourceBucket", "")
	sourceKey := core.GetString(nodeConfig, "sourceKey", "")
	destBucket := core.GetString(nodeConfig, "destBucket", "")
	destKey := core.GetString(nodeConfig, "destKey", "")

	if sourceBucket == "" || sourceKey == "" {
		return nil, fmt.Errorf("sourceBucket and sourceKey are required")
	}
	if destBucket == "" {
		destBucket = sourceBucket
	}
	if destKey == "" {
		return nil, fmt.Errorf("destKey is required")
	}

	copySource := fmt.Sprintf("%s/%s", sourceBucket, sourceKey)

	result, err := client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(destBucket),
		Key:        aws.String(destKey),
		CopySource: aws.String(copySource),
	})
	if err != nil {
		return nil, fmt.Errorf("copy failed: %w", err)
	}

	return map[string]interface{}{
		"copied":       true,
		"sourceBucket": sourceBucket,
		"sourceKey":    sourceKey,
		"destBucket":   destBucket,
		"destKey":      destKey,
		"etag":         strings.Trim(aws.ToString(result.CopyObjectResult.ETag), "\""),
	}, nil
}

func (n *AWSS3Node) getSignedURL(ctx context.Context, client *s3.Client, nodeConfig map[string]interface{}) (map[string]interface{}, error) {
	bucket := core.GetString(nodeConfig, "bucket", "")
	key := core.GetString(nodeConfig, "key", "")
	expiration := core.GetInt(nodeConfig, "expiration", 3600) // 1 hour default

	if bucket == "" || key == "" {
		return nil, fmt.Errorf("bucket and key are required")
	}

	presignClient := s3.NewPresignClient(client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expiration) * time.Second
	})
	if err != nil {
		return nil, fmt.Errorf("presign failed: %w", err)
	}

	return map[string]interface{}{
		"url":        request.URL,
		"bucket":     bucket,
		"key":        key,
		"expiration": expiration,
		"expiresAt":  time.Now().Add(time.Duration(expiration) * time.Second).Format(time.RFC3339),
	}, nil
}

func (n *AWSS3Node) listBuckets(ctx context.Context, client *s3.Client) (map[string]interface{}, error) {
	result, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("list buckets failed: %w", err)
	}

	var buckets []map[string]interface{}
	for _, bucket := range result.Buckets {
		buckets = append(buckets, map[string]interface{}{
			"name":         aws.ToString(bucket.Name),
			"creationDate": bucket.CreationDate.Format(time.RFC3339),
		})
	}

	return map[string]interface{}{
		"buckets": buckets,
		"count":   len(buckets),
	}, nil
}

func isValidUTF8(data []byte) bool {
	for i := 0; i < len(data); {
		if data[i] < 0x80 {
			i++
			continue
		}
		if data[i]&0xE0 == 0xC0 {
			if i+1 >= len(data) || data[i+1]&0xC0 != 0x80 {
				return false
			}
			i += 2
		} else if data[i]&0xF0 == 0xE0 {
			if i+2 >= len(data) || data[i+1]&0xC0 != 0x80 || data[i+2]&0xC0 != 0x80 {
				return false
			}
			i += 3
		} else if data[i]&0xF8 == 0xF0 {
			if i+3 >= len(data) || data[i+1]&0xC0 != 0x80 || data[i+2]&0xC0 != 0x80 || data[i+3]&0xC0 != 0x80 {
				return false
			}
			i += 4
		} else {
			return false
		}
	}
	return true
}

func detectContentType(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".txt":
		return "text/plain"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

// Note: AWSS3Node is registered in integrations/init.go
