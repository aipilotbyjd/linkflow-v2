package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

// WebhookManager handles webhook URL generation and management
type WebhookManager struct {
	webhookRepo *repositories.WebhookEndpointRepository
	baseURL     string
}

func NewWebhookManager(webhookRepo *repositories.WebhookEndpointRepository, baseURL string) *WebhookManager {
	return &WebhookManager{
		webhookRepo: webhookRepo,
		baseURL:     strings.TrimSuffix(baseURL, "/"),
	}
}

// GenerateWebhookInput contains parameters for generating a webhook
type GenerateWebhookInput struct {
	WorkflowID  uuid.UUID
	WorkspaceID uuid.UUID
	NodeID      string
	Method      string // GET, POST, etc.
	CustomPath  string // Optional custom path segment
}

// WebhookInfo contains the generated webhook information
type WebhookInfo struct {
	ID       uuid.UUID `json:"id"`
	URL      string    `json:"url"`
	TestURL  string    `json:"test_url"`
	Path     string    `json:"path"`
	Method   string    `json:"method"`
	Secret   string    `json:"secret,omitempty"`
	IsActive bool      `json:"is_active"`
}

// GenerateWebhook creates a new webhook endpoint with auto-generated URL
func (m *WebhookManager) GenerateWebhook(ctx context.Context, input GenerateWebhookInput) (*WebhookInfo, error) {
	// Generate unique path
	path, err := m.generateUniquePath(ctx, input.CustomPath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate path: %w", err)
	}

	// Generate secret
	secret, err := generateSecret(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	method := input.Method
	if method == "" {
		method = "POST"
	}

	endpoint := &models.WebhookEndpoint{
		WorkflowID:  input.WorkflowID,
		WorkspaceID: input.WorkspaceID,
		NodeID:      input.NodeID,
		Path:        path,
		Method:      method,
		IsActive:    true,
		Secret:      &secret,
	}

	if err := m.webhookRepo.Create(ctx, endpoint); err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return &WebhookInfo{
		ID:       endpoint.ID,
		URL:      fmt.Sprintf("%s/webhook/%s", m.baseURL, path),
		TestURL:  fmt.Sprintf("%s/webhook/test/%s", m.baseURL, path),
		Path:     path,
		Method:   method,
		Secret:   secret,
		IsActive: true,
	}, nil
}

// GetWebhookURL returns the full webhook URL for an endpoint
func (m *WebhookManager) GetWebhookURL(path string) string {
	return fmt.Sprintf("%s/webhook/%s", m.baseURL, path)
}

// GetWebhooksByWorkflow returns all webhooks for a workflow
func (m *WebhookManager) GetWebhooksByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]WebhookInfo, error) {
	endpoints, err := m.webhookRepo.FindByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	result := make([]WebhookInfo, len(endpoints))
	for i, ep := range endpoints {
		result[i] = WebhookInfo{
			ID:       ep.ID,
			URL:      fmt.Sprintf("%s/webhook/%s", m.baseURL, ep.Path),
			TestURL:  fmt.Sprintf("%s/webhook/test/%s", m.baseURL, ep.Path),
			Path:     ep.Path,
			Method:   ep.Method,
			IsActive: ep.IsActive,
		}
	}
	return result, nil
}

// RegenerateSecret generates a new secret for a webhook
func (m *WebhookManager) RegenerateSecret(ctx context.Context, endpointID uuid.UUID) (string, error) {
	secret, err := generateSecret(32)
	if err != nil {
		return "", err
	}

	endpoint, err := m.webhookRepo.FindByID(ctx, endpointID)
	if err != nil {
		return "", err
	}

	endpoint.Secret = &secret
	if err := m.webhookRepo.Update(ctx, endpoint); err != nil {
		return "", err
	}

	return secret, nil
}

// ActivateWebhook activates a webhook endpoint
func (m *WebhookManager) ActivateWebhook(ctx context.Context, endpointID uuid.UUID) error {
	return m.webhookRepo.SetActive(ctx, endpointID, true)
}

// DeactivateWebhook deactivates a webhook endpoint
func (m *WebhookManager) DeactivateWebhook(ctx context.Context, endpointID uuid.UUID) error {
	return m.webhookRepo.SetActive(ctx, endpointID, false)
}

// DeactivateAllForWorkflow deactivates all webhooks for a workflow
func (m *WebhookManager) DeactivateAllForWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	return m.webhookRepo.DeactivateByWorkflow(ctx, workflowID)
}

// generateUniquePath generates a unique webhook path
func (m *WebhookManager) generateUniquePath(ctx context.Context, customPath string) (string, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		var path string
		if customPath != "" && i == 0 {
			path = sanitizePath(customPath)
		} else {
			randomPart, err := generateRandomString(12)
			if err != nil {
				return "", err
			}
			if customPath != "" {
				path = fmt.Sprintf("%s-%s", sanitizePath(customPath), randomPart)
			} else {
				path = randomPart
			}
		}

		exists, err := m.webhookRepo.ExistsByPath(ctx, path)
		if err != nil {
			return "", err
		}
		if !exists {
			return path, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique path after %d attempts", maxAttempts)
}

func sanitizePath(path string) string {
	path = strings.ToLower(path)
	path = strings.ReplaceAll(path, " ", "-")
	// Remove any characters that aren't alphanumeric or hyphens
	var result strings.Builder
	for _, r := range path {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func generateSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2+1)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// WaitResumeManager handles wait/resume execution patterns
type WaitResumeManager struct {
	waitingRepo *repositories.WaitingExecutionRepository
	baseURL     string
}

func NewWaitResumeManager(waitingRepo *repositories.WaitingExecutionRepository, baseURL string) *WaitResumeManager {
	return &WaitResumeManager{
		waitingRepo: waitingRepo,
		baseURL:     strings.TrimSuffix(baseURL, "/"),
	}
}

// WaitForWebhookInput contains parameters for creating a wait-for-webhook
type WaitForWebhookInput struct {
	ExecutionID   uuid.UUID
	WorkflowID    uuid.UUID
	WorkspaceID   uuid.UUID
	NodeID        string
	Timeout       time.Duration
	ExecutionData models.JSON
}

// WaitInfo contains information about the waiting execution
type WaitInfo struct {
	ResumeToken string    `json:"resume_token"`
	ResumeURL   string    `json:"resume_url"`
	WebhookPath string    `json:"webhook_path"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// CreateWaitForWebhook creates a waiting execution that can be resumed via webhook
func (m *WaitResumeManager) CreateWaitForWebhook(ctx context.Context, input WaitForWebhookInput) (*WaitInfo, error) {
	// Generate unique resume token
	token, err := generateRandomString(32)
	if err != nil {
		return nil, err
	}

	// Generate webhook path
	webhookPath := fmt.Sprintf("resume/%s", token)

	timeout := input.Timeout
	if timeout == 0 {
		timeout = 24 * time.Hour // Default 24 hour timeout
	}
	expiresAt := time.Now().Add(timeout)

	waiting := &models.WaitingExecution{
		ExecutionID:   input.ExecutionID,
		WorkflowID:    input.WorkflowID,
		WorkspaceID:   input.WorkspaceID,
		NodeID:        input.NodeID,
		ResumeToken:   token,
		ResumeType:    "webhook",
		WebhookPath:   &webhookPath,
		TimeoutAt:     &expiresAt,
		ExecutionData: input.ExecutionData,
		Status:        "waiting",
	}

	if err := m.waitingRepo.Create(ctx, waiting); err != nil {
		return nil, fmt.Errorf("failed to create waiting execution: %w", err)
	}

	return &WaitInfo{
		ResumeToken: token,
		ResumeURL:   fmt.Sprintf("%s/webhook/%s", m.baseURL, webhookPath),
		WebhookPath: webhookPath,
		ExpiresAt:   expiresAt,
	}, nil
}

// ResumeExecution resumes a waiting execution with data
func (m *WaitResumeManager) ResumeExecution(ctx context.Context, token string, data models.JSON) (*models.WaitingExecution, error) {
	waiting, err := m.waitingRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("waiting execution not found: %w", err)
	}

	if waiting.Status != "waiting" {
		return nil, fmt.Errorf("execution already resumed or expired")
	}

	if waiting.TimeoutAt != nil && time.Now().After(*waiting.TimeoutAt) {
		waiting.Status = "expired"
		_ = m.waitingRepo.Update(ctx, waiting)
		return nil, fmt.Errorf("resume token has expired")
	}

	now := time.Now()
	waiting.Status = "resumed"
	waiting.ResumedAt = &now
	waiting.ResumeData = data

	if err := m.waitingRepo.Update(ctx, waiting); err != nil {
		return nil, err
	}

	return waiting, nil
}

// GetWaitingExecution retrieves a waiting execution by token
func (m *WaitResumeManager) GetWaitingExecution(ctx context.Context, token string) (*models.WaitingExecution, error) {
	return m.waitingRepo.FindByToken(ctx, token)
}

// GetWaitingByExecution retrieves waiting executions for an execution
func (m *WaitResumeManager) GetWaitingByExecution(ctx context.Context, executionID uuid.UUID) ([]models.WaitingExecution, error) {
	return m.waitingRepo.FindByExecutionID(ctx, executionID)
}

// ExpireOldWaiting expires old waiting executions
func (m *WaitResumeManager) ExpireOldWaiting(ctx context.Context) (int64, error) {
	return m.waitingRepo.ExpireOld(ctx)
}
