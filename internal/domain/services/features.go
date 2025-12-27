package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

// =============================================================================
// ERRORS
// =============================================================================

var (
	ErrAlertNotFound           = errors.New("alert not found")
	ErrCommentNotFound         = errors.New("comment not found")
	ErrShareNotFound           = errors.New("share not found")
	ErrShareExpired            = errors.New("share link has expired")
	ErrShareMaxViews           = errors.New("share link has reached maximum views")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrEnvVarNotFound          = errors.New("environment variable not found")
	ErrQueueNotFound           = errors.New("execution queue not found")
	ErrRateLimitExceeded       = errors.New("rate limit exceeded")
	ErrWebhookSignatureInvalid = errors.New("webhook signature invalid")
)

// =============================================================================
// AUDIT LOG SERVICE
// =============================================================================

type AuditLogService struct {
	repo *repositories.BaseRepository[models.AuditLog]
}

func NewAuditLogService(repo *repositories.BaseRepository[models.AuditLog]) *AuditLogService {
	return &AuditLogService{repo: repo}
}

type AuditLogInput struct {
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	ResourceName *string
	OldValue     interface{}
	NewValue     interface{}
	Metadata     map[string]interface{}
	IPAddress    *string
	UserAgent    *string
}

func (s *AuditLogService) Log(ctx context.Context, input AuditLogInput) error {
	var oldVal, newVal, meta models.JSON

	if input.OldValue != nil {
		if m, ok := input.OldValue.(map[string]interface{}); ok {
			oldVal = m
		} else if b, err := json.Marshal(input.OldValue); err == nil {
			json.Unmarshal(b, &oldVal)
		}
	}
	if input.NewValue != nil {
		if m, ok := input.NewValue.(map[string]interface{}); ok {
			newVal = m
		} else if b, err := json.Marshal(input.NewValue); err == nil {
			json.Unmarshal(b, &newVal)
		}
	}
	if input.Metadata != nil {
		meta = input.Metadata
	}

	log := &models.AuditLog{
		WorkspaceID:  input.WorkspaceID,
		UserID:       input.UserID,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		ResourceName: input.ResourceName,
		OldValue:     oldVal,
		NewValue:     newVal,
		Metadata:     meta,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
	}

	return s.repo.Create(ctx, log)
}

func (s *AuditLogService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := s.repo.DB().WithContext(ctx).Model(&models.AuditLog{}).Where("workspace_id = ?", workspaceID)
	query.Count(&total)

	if opts != nil {
		if opts.Limit > 0 {
			query = query.Limit(opts.Limit)
		}
		if opts.Offset > 0 {
			query = query.Offset(opts.Offset)
		}
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (s *AuditLogService) Search(ctx context.Context, workspaceID uuid.UUID, action, resourceType string, userID *uuid.UUID, start, end *time.Time, opts *repositories.ListOptions) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := s.repo.DB().WithContext(ctx).Model(&models.AuditLog{}).Where("workspace_id = ?", workspaceID)

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if start != nil {
		query = query.Where("created_at >= ?", *start)
	}
	if end != nil {
		query = query.Where("created_at <= ?", *end)
	}

	query.Count(&total)

	if opts != nil {
		if opts.Limit > 0 {
			query = query.Limit(opts.Limit)
		}
		if opts.Offset > 0 {
			query = query.Offset(opts.Offset)
		}
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

// =============================================================================
// ALERT SERVICE
// =============================================================================

type AlertService struct {
	repo    *repositories.BaseRepository[models.Alert]
	logRepo *repositories.BaseRepository[models.AlertLog]
}

func NewAlertService(repo *repositories.BaseRepository[models.Alert], logRepo *repositories.BaseRepository[models.AlertLog]) *AlertService {
	return &AlertService{repo: repo, logRepo: logRepo}
}

type CreateAlertInput struct {
	WorkspaceID  uuid.UUID
	WorkflowID   *uuid.UUID
	CreatedBy    uuid.UUID
	Name         string
	Type         string
	Trigger      string
	Config       map[string]interface{}
	Conditions   map[string]interface{}
	CooldownMins int
}

func (s *AlertService) Create(ctx context.Context, input CreateAlertInput) (*models.Alert, error) {
	alert := &models.Alert{
		WorkspaceID:  input.WorkspaceID,
		WorkflowID:   input.WorkflowID,
		CreatedBy:    input.CreatedBy,
		Name:         input.Name,
		Type:         input.Type,
		Trigger:      input.Trigger,
		Config:       input.Config,
		Conditions:   input.Conditions,
		CooldownMins: input.CooldownMins,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, alert); err != nil {
		return nil, err
	}
	return alert, nil
}

func (s *AlertService) GetByID(ctx context.Context, id uuid.UUID) (*models.Alert, error) {
	alert, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrAlertNotFound
	}
	return alert, nil
}

func (s *AlertService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]models.Alert, error) {
	var alerts []models.Alert
	err := s.repo.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&alerts).Error
	return alerts, err
}

func (s *AlertService) GetByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.Alert, error) {
	var alerts []models.Alert
	err := s.repo.DB().WithContext(ctx).Where("workflow_id = ?", workflowID).Find(&alerts).Error
	return alerts, err
}

func (s *AlertService) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return s.repo.DB().WithContext(ctx).Model(&models.Alert{}).Where("id = ?", id).Updates(updates).Error
}

func (s *AlertService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Delete(&models.Alert{}, "id = ?", id).Error
}

func (s *AlertService) ShouldFire(ctx context.Context, alert *models.Alert) bool {
	if !alert.IsActive {
		return false
	}
	if alert.LastFiredAt != nil && alert.CooldownMins > 0 {
		cooldown := time.Duration(alert.CooldownMins) * time.Minute
		if time.Since(*alert.LastFiredAt) < cooldown {
			return false
		}
	}
	return true
}

func (s *AlertService) Fire(ctx context.Context, alertID uuid.UUID, executionID *uuid.UUID, message string) error {
	now := time.Now()
	if err := s.repo.DB().WithContext(ctx).Model(&models.Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"last_fired_at": now,
		}).Error; err != nil {
		return err
	}
	// Increment fire count separately
	s.repo.DB().WithContext(ctx).Model(&models.Alert{}).
		Where("id = ?", alertID).
		UpdateColumn("fire_count", s.repo.DB().Raw("fire_count + 1"))

	log := &models.AlertLog{
		AlertID:     alertID,
		ExecutionID: executionID,
		Status:      "sent",
		Message:     message,
	}
	return s.logRepo.Create(ctx, log)
}

// =============================================================================
// WORKFLOW COMMENT SERVICE
// =============================================================================

type WorkflowCommentService struct {
	repo *repositories.BaseRepository[models.WorkflowComment]
}

func NewWorkflowCommentService(repo *repositories.BaseRepository[models.WorkflowComment]) *WorkflowCommentService {
	return &WorkflowCommentService{repo: repo}
}

type CreateCommentInput struct {
	WorkflowID  uuid.UUID
	WorkspaceID uuid.UUID
	NodeID      *string
	ParentID    *uuid.UUID
	CreatedBy   uuid.UUID
	Content     string
}

func (s *WorkflowCommentService) Create(ctx context.Context, input CreateCommentInput) (*models.WorkflowComment, error) {
	comment := &models.WorkflowComment{
		WorkflowID:  input.WorkflowID,
		WorkspaceID: input.WorkspaceID,
		NodeID:      input.NodeID,
		ParentID:    input.ParentID,
		CreatedBy:   input.CreatedBy,
		Content:     input.Content,
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *WorkflowCommentService) GetByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.WorkflowComment, error) {
	var comments []models.WorkflowComment
	err := s.repo.DB().WithContext(ctx).
		Where("workflow_id = ? AND parent_id IS NULL", workflowID).
		Preload("Replies").
		Order("created_at DESC").
		Find(&comments).Error
	return comments, err
}

func (s *WorkflowCommentService) GetByNode(ctx context.Context, workflowID uuid.UUID, nodeID string) ([]models.WorkflowComment, error) {
	var comments []models.WorkflowComment
	err := s.repo.DB().WithContext(ctx).
		Where("workflow_id = ? AND node_id = ? AND parent_id IS NULL", workflowID, nodeID).
		Preload("Replies").
		Order("created_at DESC").
		Find(&comments).Error
	return comments, err
}

func (s *WorkflowCommentService) Update(ctx context.Context, id uuid.UUID, content string) error {
	return s.repo.DB().WithContext(ctx).Model(&models.WorkflowComment{}).
		Where("id = ?", id).
		Update("content", content).Error
}

func (s *WorkflowCommentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Delete(&models.WorkflowComment{}, "id = ?", id).Error
}

func (s *WorkflowCommentService) Resolve(ctx context.Context, id uuid.UUID, resolvedBy uuid.UUID) error {
	now := time.Now()
	return s.repo.DB().WithContext(ctx).Model(&models.WorkflowComment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_resolved": true,
			"resolved_by": resolvedBy,
			"resolved_at": now,
		}).Error
}

func (s *WorkflowCommentService) Unresolve(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Model(&models.WorkflowComment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_resolved": false,
			"resolved_by": nil,
			"resolved_at": nil,
		}).Error
}

// =============================================================================
// EXECUTION SHARE SERVICE
// =============================================================================

type ExecutionShareService struct {
	repo *repositories.BaseRepository[models.ExecutionShare]
}

func NewExecutionShareService(repo *repositories.BaseRepository[models.ExecutionShare]) *ExecutionShareService {
	return &ExecutionShareService{repo: repo}
}

type CreateShareInput struct {
	ExecutionID   uuid.UUID
	WorkspaceID   uuid.UUID
	CreatedBy     uuid.UUID
	ExpiresAt     *time.Time
	Password      *string
	MaxViews      *int
	AllowDownload bool
	IncludeLogs   bool
	IncludeData   bool
}

func (s *ExecutionShareService) Create(ctx context.Context, input CreateShareInput) (*models.ExecutionShare, error) {
	token := generateToken(32)

	share := &models.ExecutionShare{
		ExecutionID:   input.ExecutionID,
		WorkspaceID:   input.WorkspaceID,
		CreatedBy:     input.CreatedBy,
		Token:         token,
		ExpiresAt:     input.ExpiresAt,
		Password:      input.Password,
		MaxViews:      input.MaxViews,
		AllowDownload: input.AllowDownload,
		IncludeLogs:   input.IncludeLogs,
		IncludeData:   input.IncludeData,
	}

	if err := s.repo.Create(ctx, share); err != nil {
		return nil, err
	}
	return share, nil
}

func (s *ExecutionShareService) GetByToken(ctx context.Context, token string) (*models.ExecutionShare, error) {
	var share models.ExecutionShare
	err := s.repo.DB().WithContext(ctx).Where("token = ?", token).First(&share).Error
	if err != nil {
		return nil, ErrShareNotFound
	}
	return &share, nil
}

func (s *ExecutionShareService) ValidateAccess(ctx context.Context, share *models.ExecutionShare, password *string) error {
	if share.ExpiresAt != nil && time.Now().After(*share.ExpiresAt) {
		return ErrShareExpired
	}
	if share.MaxViews != nil && share.ViewCount >= *share.MaxViews {
		return ErrShareMaxViews
	}
	if share.Password != nil && (password == nil || *password != *share.Password) {
		return ErrInvalidPassword
	}
	return nil
}

func (s *ExecutionShareService) IncrementViews(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Model(&models.ExecutionShare{}).
		Where("id = ?", id).
		UpdateColumn("view_count", s.repo.DB().Raw("view_count + 1")).Error
}

func (s *ExecutionShareService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Delete(&models.ExecutionShare{}, "id = ?", id).Error
}

// =============================================================================
// ENVIRONMENT VARIABLE SERVICE
// =============================================================================

type EnvironmentVariableService struct {
	repo      *repositories.BaseRepository[models.EnvironmentVariable]
	encryptor Encryptor
}

type Encryptor interface {
	Encrypt(data string) (string, error)
	Decrypt(data string) (string, error)
}

func NewEnvironmentVariableService(repo *repositories.BaseRepository[models.EnvironmentVariable], encryptor Encryptor) *EnvironmentVariableService {
	return &EnvironmentVariableService{repo: repo, encryptor: encryptor}
}

type CreateEnvVarInput struct {
	WorkspaceID uuid.UUID
	CreatedBy   uuid.UUID
	Name        string
	Value       string
	IsSecret    bool
	Environment *string
	Description *string
}

func (s *EnvironmentVariableService) Create(ctx context.Context, input CreateEnvVarInput) (*models.EnvironmentVariable, error) {
	encrypted, err := s.encryptor.Encrypt(input.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt value: %w", err)
	}

	envVar := &models.EnvironmentVariable{
		WorkspaceID: input.WorkspaceID,
		CreatedBy:   input.CreatedBy,
		Name:        input.Name,
		Value:       encrypted,
		IsSecret:    input.IsSecret,
		Environment: input.Environment,
		Description: input.Description,
	}

	if err := s.repo.Create(ctx, envVar); err != nil {
		return nil, err
	}
	return envVar, nil
}

func (s *EnvironmentVariableService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, environment *string) ([]models.EnvironmentVariable, error) {
	var vars []models.EnvironmentVariable
	query := s.repo.DB().WithContext(ctx).Where("workspace_id = ?", workspaceID)
	if environment != nil {
		query = query.Where("environment = ? OR environment IS NULL", *environment)
	}
	err := query.Order("name").Find(&vars).Error
	return vars, err
}

func (s *EnvironmentVariableService) GetDecrypted(ctx context.Context, workspaceID uuid.UUID, environment *string) (map[string]string, error) {
	vars, err := s.GetByWorkspace(ctx, workspaceID, environment)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, v := range vars {
		decrypted, err := s.encryptor.Decrypt(v.Value)
		if err != nil {
			continue
		}
		result[v.Name] = decrypted
	}
	return result, nil
}

func (s *EnvironmentVariableService) Update(ctx context.Context, id uuid.UUID, value string) error {
	encrypted, err := s.encryptor.Encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt value: %w", err)
	}
	return s.repo.DB().WithContext(ctx).Model(&models.EnvironmentVariable{}).
		Where("id = ?", id).
		Update("value", encrypted).Error
}

func (s *EnvironmentVariableService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.DB().WithContext(ctx).Delete(&models.EnvironmentVariable{}, "id = ?", id).Error
}

// =============================================================================
// WEBHOOK SIGNATURE SERVICE
// =============================================================================

type WebhookSignatureService struct {
	repo      *repositories.BaseRepository[models.WebhookSignatureConfig]
	encryptor Encryptor
}

func NewWebhookSignatureService(repo *repositories.BaseRepository[models.WebhookSignatureConfig], encryptor Encryptor) *WebhookSignatureService {
	return &WebhookSignatureService{repo: repo, encryptor: encryptor}
}

func (s *WebhookSignatureService) Create(ctx context.Context, webhookID uuid.UUID, algorithm, headerName string, signaturePrefix *string) (*models.WebhookSignatureConfig, error) {
	secret := generateToken(32)
	encrypted, err := s.encryptor.Encrypt(secret)
	if err != nil {
		return nil, err
	}

	config := &models.WebhookSignatureConfig{
		WebhookID:       webhookID,
		Algorithm:       algorithm,
		Secret:          encrypted,
		HeaderName:      headerName,
		SignaturePrefix: signaturePrefix,
		IsActive:        true,
		FailOnInvalid:   true,
	}

	if err := s.repo.Create(ctx, config); err != nil {
		return nil, err
	}
	return config, nil
}

func (s *WebhookSignatureService) GetByWebhook(ctx context.Context, webhookID uuid.UUID) (*models.WebhookSignatureConfig, error) {
	var config models.WebhookSignatureConfig
	err := s.repo.DB().WithContext(ctx).Where("webhook_id = ?", webhookID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *WebhookSignatureService) Verify(ctx context.Context, webhookID uuid.UUID, payload []byte, signature string) error {
	config, err := s.GetByWebhook(ctx, webhookID)
	if err != nil {
		return nil // No signature config = pass
	}
	if !config.IsActive {
		return nil
	}

	secret, err := s.encryptor.Decrypt(config.Secret)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Remove prefix if present
	sig := signature
	if config.SignaturePrefix != nil && len(signature) > len(*config.SignaturePrefix) {
		sig = signature[len(*config.SignaturePrefix):]
	}

	// Calculate expected signature
	var expected string
	switch config.Algorithm {
	case models.WebhookSigHMACSHA256:
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		expected = hex.EncodeToString(mac.Sum(nil))
	default:
		return fmt.Errorf("unsupported algorithm: %s", config.Algorithm)
	}

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		if config.FailOnInvalid {
			return ErrWebhookSignatureInvalid
		}
	}
	return nil
}

// =============================================================================
// CREDENTIAL RATE LIMIT SERVICE
// =============================================================================

type CredentialRateLimitService struct {
	repo *repositories.BaseRepository[models.CredentialRateLimit]
}

func NewCredentialRateLimitService(repo *repositories.BaseRepository[models.CredentialRateLimit]) *CredentialRateLimitService {
	return &CredentialRateLimitService{repo: repo}
}

func (s *CredentialRateLimitService) Create(ctx context.Context, credentialID uuid.UUID, reqPerMin, reqPerHour, reqPerDay, burst int) (*models.CredentialRateLimit, error) {
	limit := &models.CredentialRateLimit{
		CredentialID:    credentialID,
		RequestsPerMin:  reqPerMin,
		RequestsPerHour: reqPerHour,
		RequestsPerDay:  reqPerDay,
		BurstLimit:      burst,
		IsActive:        true,
		LastResetMin:    time.Now(),
		LastResetHour:   time.Now(),
		LastResetDay:    time.Now(),
	}

	if err := s.repo.Create(ctx, limit); err != nil {
		return nil, err
	}
	return limit, nil
}

func (s *CredentialRateLimitService) Check(ctx context.Context, credentialID uuid.UUID) error {
	var limit models.CredentialRateLimit
	err := s.repo.DB().WithContext(ctx).Where("credential_id = ?", credentialID).First(&limit).Error
	if err != nil {
		return nil // No rate limit configured
	}
	if !limit.IsActive {
		return nil
	}

	now := time.Now()

	// Reset counters if needed
	if now.Sub(limit.LastResetMin) >= time.Minute {
		limit.CurrentMinute = 0
		limit.LastResetMin = now
	}
	if now.Sub(limit.LastResetHour) >= time.Hour {
		limit.CurrentHour = 0
		limit.LastResetHour = now
	}
	if now.Sub(limit.LastResetDay) >= 24*time.Hour {
		limit.CurrentDay = 0
		limit.LastResetDay = now
	}

	// Check limits
	if limit.RequestsPerMin > 0 && limit.CurrentMinute >= limit.RequestsPerMin {
		return ErrRateLimitExceeded
	}
	if limit.RequestsPerHour > 0 && limit.CurrentHour >= limit.RequestsPerHour {
		return ErrRateLimitExceeded
	}
	if limit.RequestsPerDay > 0 && limit.CurrentDay >= limit.RequestsPerDay {
		return ErrRateLimitExceeded
	}

	// Increment counters
	limit.CurrentMinute++
	limit.CurrentHour++
	limit.CurrentDay++

	return s.repo.DB().WithContext(ctx).Save(&limit).Error
}

// =============================================================================
// ANALYTICS SERVICE
// =============================================================================

type AnalyticsService struct {
	workspaceRepo *repositories.BaseRepository[models.WorkspaceAnalytics]
	workflowRepo  *repositories.BaseRepository[models.WorkflowAnalytics]
	execRepo      *repositories.ExecutionRepository
}

func NewAnalyticsService(
	workspaceRepo *repositories.BaseRepository[models.WorkspaceAnalytics],
	workflowRepo *repositories.BaseRepository[models.WorkflowAnalytics],
	execRepo *repositories.ExecutionRepository,
) *AnalyticsService {
	return &AnalyticsService{
		workspaceRepo: workspaceRepo,
		workflowRepo:  workflowRepo,
		execRepo:      execRepo,
	}
}

func (s *AnalyticsService) GetWorkspaceAnalytics(ctx context.Context, workspaceID uuid.UUID, start, end time.Time) ([]models.WorkspaceAnalytics, error) {
	var analytics []models.WorkspaceAnalytics
	err := s.workspaceRepo.DB().WithContext(ctx).
		Where("workspace_id = ? AND date >= ? AND date <= ?", workspaceID, start, end).
		Order("date").
		Find(&analytics).Error
	return analytics, err
}

func (s *AnalyticsService) GetWorkflowAnalytics(ctx context.Context, workflowID uuid.UUID, start, end time.Time) ([]models.WorkflowAnalytics, error) {
	var analytics []models.WorkflowAnalytics
	err := s.workflowRepo.DB().WithContext(ctx).
		Where("workflow_id = ? AND date >= ? AND date <= ?", workflowID, start, end).
		Order("date").
		Find(&analytics).Error
	return analytics, err
}

func (s *AnalyticsService) AggregateWorkspaceDaily(ctx context.Context, workspaceID uuid.UUID, date time.Time) error {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	stats, err := s.execRepo.GetStats(ctx, workspaceID, startOfDay, endOfDay)
	if err != nil {
		return err
	}

	analytics := &models.WorkspaceAnalytics{
		WorkspaceID:         workspaceID,
		Date:                startOfDay,
		ExecutionsTotal:     getIntStat(stats, "total"),
		ExecutionsSuccess:   getIntStat(stats, "success"),
		ExecutionsFailed:    getIntStat(stats, "failed"),
		ExecutionsCancelled: getIntStat(stats, "cancelled"),
	}

	// Upsert
	return s.workspaceRepo.DB().WithContext(ctx).
		Where("workspace_id = ? AND date = ?", workspaceID, startOfDay).
		Assign(analytics).
		FirstOrCreate(analytics).Error
}

// =============================================================================
// WORKFLOW IMPORT/EXPORT SERVICE
// =============================================================================

type WorkflowExportService struct {
	exportRepo   *repositories.BaseRepository[models.WorkflowExport]
	importRepo   *repositories.BaseRepository[models.WorkflowImport]
	workflowRepo *repositories.WorkflowRepository
}

func NewWorkflowExportService(
	exportRepo *repositories.BaseRepository[models.WorkflowExport],
	importRepo *repositories.BaseRepository[models.WorkflowImport],
	workflowRepo *repositories.WorkflowRepository,
) *WorkflowExportService {
	return &WorkflowExportService{
		exportRepo:   exportRepo,
		importRepo:   importRepo,
		workflowRepo: workflowRepo,
	}
}

type WorkflowExportData struct {
	Version     string                   `json:"version"`
	ExportedAt  time.Time                `json:"exported_at"`
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	Nodes       []map[string]interface{} `json:"nodes"`
	Connections []map[string]interface{} `json:"connections"`
	Settings    map[string]interface{}   `json:"settings,omitempty"`
	Variables   []map[string]interface{} `json:"variables,omitempty"`
}

func (s *WorkflowExportService) Export(ctx context.Context, workflowID uuid.UUID, exportedBy uuid.UUID, includeCredentials bool) (*WorkflowExportData, error) {
	workflow, err := s.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	// Convert JSONArray to []map[string]interface{}
	var nodes []map[string]interface{}
	for _, n := range workflow.Nodes {
		if m, ok := n.(map[string]interface{}); ok {
			nodes = append(nodes, m)
		}
	}

	var connections []map[string]interface{}
	for _, c := range workflow.Connections {
		if m, ok := c.(map[string]interface{}); ok {
			connections = append(connections, m)
		}
	}

	settings := map[string]interface{}(workflow.Settings)

	// Remove credential references if not including them
	if !includeCredentials {
		for i := range nodes {
			if params, ok := nodes[i]["parameters"].(map[string]interface{}); ok {
				delete(params, "credential_id")
			}
		}
	}

	data := &WorkflowExportData{
		Version:     "1.0",
		ExportedAt:  time.Now(),
		Name:        workflow.Name,
		Description: workflow.Description,
		Nodes:       nodes,
		Connections: connections,
		Settings:    settings,
	}

	// Log export
	exportLog := &models.WorkflowExport{
		WorkflowID:         workflowID,
		WorkspaceID:        workflow.WorkspaceID,
		ExportedBy:         exportedBy,
		Version:            workflow.Version,
		Format:             "json",
		IncludeCredentials: includeCredentials,
	}
	s.exportRepo.Create(ctx, exportLog)

	return data, nil
}

func (s *WorkflowExportService) Import(ctx context.Context, workspaceID, importedBy uuid.UUID, data *WorkflowExportData) (*models.Workflow, error) {
	// Convert []map[string]interface{} to JSONArray
	var nodes models.JSONArray
	for _, n := range data.Nodes {
		nodes = append(nodes, n)
	}

	var connections models.JSONArray
	for _, c := range data.Connections {
		connections = append(connections, c)
	}

	workflow := &models.Workflow{
		WorkspaceID: workspaceID,
		CreatedBy:   importedBy,
		Name:        data.Name + " (Imported)",
		Description: data.Description,
		Status:      "draft",
		Version:     1,
		Nodes:       nodes,
		Connections: connections,
		Settings:    data.Settings,
	}

	if err := s.workflowRepo.Create(ctx, workflow); err != nil {
		return nil, err
	}

	// Log import
	importLog := &models.WorkflowImport{
		WorkflowID:  &workflow.ID,
		WorkspaceID: workspaceID,
		ImportedBy:  importedBy,
		SourceName:  &data.Name,
		SourceType:  "file",
		Status:      "completed",
	}
	now := time.Now()
	importLog.CompletedAt = &now
	s.importRepo.Create(ctx, importLog)

	return workflow, nil
}

// =============================================================================
// EXECUTION REPLAY SERVICE
// =============================================================================

type ExecutionReplayService struct {
	execService *ExecutionService
	execRepo    *repositories.ExecutionRepository
}

func NewExecutionReplayService(execService *ExecutionService, execRepo *repositories.ExecutionRepository) *ExecutionReplayService {
	return &ExecutionReplayService{execService: execService, execRepo: execRepo}
}

func (s *ExecutionReplayService) Replay(ctx context.Context, executionID uuid.UUID, triggeredBy *uuid.UUID) (*models.Execution, error) {
	original, err := s.execRepo.FindByID(ctx, executionID)
	if err != nil {
		return nil, err
	}

	return s.execService.Create(ctx, CreateExecutionInput{
		WorkflowID:  original.WorkflowID,
		WorkspaceID: original.WorkspaceID,
		TriggeredBy: triggeredBy,
		TriggerType: "replay",
		TriggerData: original.TriggerData,
		InputData:   original.InputData,
	})
}

func (s *ExecutionReplayService) ReplayFromNode(ctx context.Context, executionID uuid.UUID, startNodeID string, triggeredBy *uuid.UUID) (*models.Execution, error) {
	original, err := s.execRepo.FindByID(ctx, executionID)
	if err != nil {
		return nil, err
	}

	// Create execution with partial flag
	triggerData := models.JSON{
		"replay_from":        startNodeID,
		"original_execution": executionID.String(),
	}

	return s.execService.Create(ctx, CreateExecutionInput{
		WorkflowID:  original.WorkflowID,
		WorkspaceID: original.WorkspaceID,
		TriggeredBy: triggeredBy,
		TriggerType: "partial_replay",
		TriggerData: triggerData,
		InputData:   original.InputData,
	})
}

// =============================================================================
// HELPERS
// =============================================================================

func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getIntStat(stats map[string]interface{}, key string) int {
	if v, ok := stats[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}
