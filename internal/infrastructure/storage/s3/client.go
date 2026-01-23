package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/storage"
)

type S3Storage struct {
	client   *s3.Client
	presign  *s3.PresignClient
	bucket   string
	region   string
	endpoint string
}

func NewS3Storage(cfg storage.S3Config) (*S3Storage, error) {
	ctx := context.Background()

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	clientOpts := []func(*s3.Options){}
	if cfg.Endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.UsePathStyle
		})
	}

	client := s3.NewFromConfig(awsCfg, clientOpts...)
	presign := s3.NewPresignClient(client)

	return &S3Storage{
		client:   client,
		presign:  presign,
		bucket:   cfg.Bucket,
		region:   cfg.Region,
		endpoint: cfg.Endpoint,
	}, nil
}

// Upload implements StorageService.Upload
func (s *S3Storage) Upload(ctx context.Context, workspaceID uuid.UUID, fileName string, reader io.Reader, size int64) (string, error) {
	fileID := uuid.New().String()
	ext := filepath.Ext(fileName)
	storagePath := fmt.Sprintf("%s/%s%s", workspaceID.String(), fileID, ext)

	// Simple content type detection (could be improved)
	contentType := "application/octet-stream"
	if ext == ".png" {
		contentType = "image/png"
	} else if ext == ".jpg" || ext == ".jpeg" {
		contentType = "image/jpeg"
	} else if ext == ".json" {
		contentType = "application/json"
	} else if ext == ".pdf" {
		contentType = "application/pdf"
	} else if ext == ".txt" {
		contentType = "text/plain"
	}

	if err := s.Put(ctx, storagePath, reader, contentType); err != nil {
		return "", err
	}
	return storagePath, nil
}

// Download implements StorageService.Download
func (s *S3Storage) Download(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	return s.Get(ctx, storagePath)
}

// GetURL implements StorageService.GetURL
func (s *S3Storage) GetURL(ctx context.Context, storagePath string) (string, error) {
	return s.SignedURL(ctx, storagePath, 1*time.Hour)
}

func (s *S3Storage) Put(ctx context.Context, path string, reader io.Reader, contentType string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	_, err = s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}
	return nil
}

func (s *S3Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return result.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, path string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		var nsk *types.NotFound
		if ok := errors.As(err, &nsk); ok {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

func (s *S3Storage) URL(ctx context.Context, path string) (string, error) {
	if s.endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, path), nil
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, path), nil
}

func (s *S3Storage) SignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}

	presignResult, err := s.presign.PresignGetObject(ctx, input, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignResult.URL, nil
}

func (s *S3Storage) List(ctx context.Context, prefix string) ([]storage.FileInfo, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}

	var files []storage.FileInfo
	paginator := s3.NewListObjectsV2Paginator(s.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			files = append(files, storage.FileInfo{
				Path:         aws.ToString(obj.Key),
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
				ETag:         aws.ToString(obj.ETag),
			})
		}
	}
	return files, nil
}

func (s *S3Storage) Copy(ctx context.Context, src, dst string) error {
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s", s.bucket, src)),
		Key:        aws.String(dst),
	}

	_, err := s.client.CopyObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}
	return nil
}

func (s *S3Storage) Move(ctx context.Context, src, dst string) error {
	if err := s.Copy(ctx, src, dst); err != nil {
		return err
	}
	return s.Delete(ctx, src)
}
