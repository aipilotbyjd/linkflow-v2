package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LocalStorage implements binarydata.StorageService for local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage service
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0750); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &LocalStorage{basePath: basePath}, nil
}

// Upload uploads a file to local storage
func (s *LocalStorage) Upload(ctx context.Context, workspaceID uuid.UUID, fileName string, reader io.Reader, size int64) (string, error) {
	// Create workspace directory
	wsDir := filepath.Join(s.basePath, workspaceID.String())
	if err := os.MkdirAll(wsDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Generate unique file name
	fileID := uuid.New().String()
	ext := filepath.Ext(fileName)
	storagePath := filepath.Join(workspaceID.String(), fileID+ext)
	fullPath := filepath.Join(s.basePath, storagePath)

	// Create file
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(fullPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return storagePath, nil
}

// Download downloads a file from local storage
func (s *LocalStorage) Download(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, storagePath)

	// Prevent path traversal
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, s.basePath) {
		return nil, fmt.Errorf("invalid storage path")
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// Delete deletes a file from local storage
func (s *LocalStorage) Delete(ctx context.Context, storagePath string) error {
	fullPath := filepath.Join(s.basePath, storagePath)

	// Prevent path traversal
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, s.basePath) {
		return fmt.Errorf("invalid storage path")
	}

	return os.Remove(cleanPath)
}

// GetURL returns a URL for the file (for local storage, returns the path)
func (s *LocalStorage) GetURL(ctx context.Context, storagePath string) (string, error) {
	return "/api/v1/binary-data/download/" + storagePath, nil
}
