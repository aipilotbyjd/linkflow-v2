package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for file storage operations
type Storage interface {
	// Put stores a file
	Put(ctx context.Context, path string, reader io.Reader, contentType string) error

	// Get retrieves a file
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists
	Exists(ctx context.Context, path string) (bool, error)

	// URL returns a URL for accessing the file
	URL(ctx context.Context, path string) (string, error)

	// SignedURL returns a signed URL for temporary access
	SignedURL(ctx context.Context, path string, expiry time.Duration) (string, error)

	// List lists files in a directory
	List(ctx context.Context, prefix string) ([]FileInfo, error)

	// Copy copies a file
	Copy(ctx context.Context, src, dst string) error

	// Move moves a file
	Move(ctx context.Context, src, dst string) error
}

// FileInfo represents file metadata
type FileInfo struct {
	Path         string
	Size         int64
	ContentType  string
	LastModified time.Time
	ETag         string
}

// Config holds storage configuration
type Config struct {
	Type      string // "local", "s3"
	LocalPath string
	S3        S3Config
}

// S3Config holds S3-specific configuration
type S3Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
	UsePathStyle    bool
}
