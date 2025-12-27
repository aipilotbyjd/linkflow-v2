package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =============================================================================
// EXECUTION FEATURES
// =============================================================================

// ExecutionQueue represents priority-based execution queues
type ExecutionQueue struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Priority    int       `gorm:"default:5;index" json:"priority"` // 1-10, higher = more priority
	Concurrency int       `gorm:"default:1" json:"concurrency"`    // Max concurrent executions
	RateLimit   int       `gorm:"default:0" json:"rate_limit"`     // Executions per minute, 0 = unlimited
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (ExecutionQueue) TableName() string {
	return "execution_queues"
}

// ExecutionShare represents shareable execution links
type ExecutionShare struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExecutionID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"execution_id"`
	WorkspaceID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy    uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	Token        string     `gorm:"size:64;uniqueIndex;not null" json:"token"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Password     *string    `gorm:"size:255" json:"-"` // Optional password protection
	ViewCount    int        `gorm:"default:0" json:"view_count"`
	MaxViews     *int       `json:"max_views,omitempty"` // Limit views
	AllowDownload bool      `gorm:"default:false" json:"allow_download"`
	IncludeLogs  bool       `gorm:"default:true" json:"include_logs"`
	IncludeData  bool       `gorm:"default:false" json:"include_data"` // Include input/output data
	CreatedAt    time.Time  `json:"created_at"`

	Execution Execution `gorm:"foreignKey:ExecutionID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator   User      `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (ExecutionShare) TableName() string {
	return "execution_shares"
}

// =============================================================================
// ALERTS & MONITORING
// =============================================================================

// AlertType constants
const (
	AlertTypeEmail   = "email"
	AlertTypeSlack   = "slack"
	AlertTypeWebhook = "webhook"
	AlertTypeSMS     = "sms"
)

// AlertTrigger constants
const (
	AlertTriggerOnFailure    = "on_failure"
	AlertTriggerOnSuccess    = "on_success"
	AlertTriggerOnTimeout    = "on_timeout"
	AlertTriggerOnLongRun    = "on_long_run"
	AlertTriggerOnQuotaLimit = "on_quota_limit"
)

// Alert represents configurable alerts for workflows/executions
type Alert struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	WorkflowID   *uuid.UUID     `gorm:"type:uuid;index" json:"workflow_id,omitempty"` // nil = workspace-wide
	CreatedBy    uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	Type         string         `gorm:"size:20;not null" json:"type"`       // email, slack, webhook, sms
	Trigger      string         `gorm:"size:30;not null" json:"trigger"`    // on_failure, on_success, etc.
	Config       JSON           `gorm:"type:jsonb;not null" json:"config"`  // Type-specific config
	Conditions   JSON           `gorm:"type:jsonb" json:"conditions"`       // Additional conditions
	CooldownMins int            `gorm:"default:5" json:"cooldown_mins"`     // Min minutes between alerts
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastFiredAt  *time.Time     `json:"last_fired_at,omitempty"`
	FireCount    int            `gorm:"default:0" json:"fire_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Workflow  *Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
	Creator   User      `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (Alert) TableName() string {
	return "alerts"
}

// AlertLog tracks fired alerts
type AlertLog struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AlertID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"alert_id"`
	ExecutionID *uuid.UUID `gorm:"type:uuid;index" json:"execution_id,omitempty"`
	Status      string     `gorm:"size:20;not null" json:"status"` // sent, failed, skipped
	Message     string     `gorm:"type:text" json:"message"`
	Response    *string    `gorm:"type:text" json:"response,omitempty"`
	Error       *string    `gorm:"type:text" json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	Alert     Alert      `gorm:"foreignKey:AlertID" json:"-"`
	Execution *Execution `gorm:"foreignKey:ExecutionID" json:"-"`
}

func (AlertLog) TableName() string {
	return "alert_logs"
}

// WorkspaceAnalytics stores aggregated analytics
type WorkspaceAnalytics struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID          uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Date                 time.Time `gorm:"type:date;index;not null" json:"date"`
	ExecutionsTotal      int       `gorm:"default:0" json:"executions_total"`
	ExecutionsSuccess    int       `gorm:"default:0" json:"executions_success"`
	ExecutionsFailed     int       `gorm:"default:0" json:"executions_failed"`
	ExecutionsCancelled  int       `gorm:"default:0" json:"executions_cancelled"`
	AvgExecutionDuration int       `gorm:"default:0" json:"avg_execution_duration_ms"`
	TotalOperations      int       `gorm:"default:0" json:"total_operations"`
	CreditsUsed          int       `gorm:"default:0" json:"credits_used"`
	WebhooksReceived     int       `gorm:"default:0" json:"webhooks_received"`
	SchedulesTriggered   int       `gorm:"default:0" json:"schedules_triggered"`
	ErrorsByType         JSON      `gorm:"type:jsonb" json:"errors_by_type"`   // {error_type: count}
	TopWorkflows         JSON      `gorm:"type:jsonb" json:"top_workflows"`    // [{workflow_id, count}]
	TopErrors            JSON      `gorm:"type:jsonb" json:"top_errors"`       // [{error, count}]
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (WorkspaceAnalytics) TableName() string {
	return "workspace_analytics"
}

// WorkflowAnalytics stores per-workflow analytics
type WorkflowAnalytics struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID           uuid.UUID `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID          uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Date                 time.Time `gorm:"type:date;index;not null" json:"date"`
	ExecutionsTotal      int       `gorm:"default:0" json:"executions_total"`
	ExecutionsSuccess    int       `gorm:"default:0" json:"executions_success"`
	ExecutionsFailed     int       `gorm:"default:0" json:"executions_failed"`
	AvgDurationMs        int       `gorm:"default:0" json:"avg_duration_ms"`
	MinDurationMs        int       `gorm:"default:0" json:"min_duration_ms"`
	MaxDurationMs        int       `gorm:"default:0" json:"max_duration_ms"`
	P95DurationMs        int       `gorm:"default:0" json:"p95_duration_ms"`
	CreditsUsed          int       `gorm:"default:0" json:"credits_used"`
	SuccessRate          float64   `gorm:"default:0" json:"success_rate"`
	NodePerformance      JSON      `gorm:"type:jsonb" json:"node_performance"` // {node_id: {avg_ms, success_rate}}
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	Workflow  Workflow  `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (WorkflowAnalytics) TableName() string {
	return "workflow_analytics"
}

// =============================================================================
// TEAM & COLLABORATION
// =============================================================================

// WorkflowComment represents annotations on workflows/nodes
type WorkflowComment struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	NodeID      *string        `gorm:"size:100;index" json:"node_id,omitempty"` // nil = workflow-level comment
	ParentID    *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id,omitempty"` // For replies
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	IsResolved  bool           `gorm:"default:false" json:"is_resolved"`
	ResolvedBy  *uuid.UUID     `gorm:"type:uuid" json:"resolved_by,omitempty"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Workflow  Workflow          `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace Workspace         `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator   User              `gorm:"foreignKey:CreatedBy" json:"-"`
	Resolver  *User             `gorm:"foreignKey:ResolvedBy" json:"-"`
	Parent    *WorkflowComment  `gorm:"foreignKey:ParentID" json:"-"`
	Replies   []WorkflowComment `gorm:"foreignKey:ParentID" json:"-"`
}

func (WorkflowComment) TableName() string {
	return "workflow_comments"
}

// AuditAction constants
const (
	AuditActionCreate   = "create"
	AuditActionUpdate   = "update"
	AuditActionDelete   = "delete"
	AuditActionExecute  = "execute"
	AuditActionActivate = "activate"
	AuditActionDeactivate = "deactivate"
	AuditActionLogin    = "login"
	AuditActionLogout   = "logout"
	AuditActionInvite   = "invite"
	AuditActionExport   = "export"
	AuditActionImport   = "import"
)

// AuditLog tracks all user actions for compliance
type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Action       string     `gorm:"size:30;not null;index" json:"action"`
	ResourceType string     `gorm:"size:50;not null;index" json:"resource_type"` // workflow, credential, etc.
	ResourceID   *uuid.UUID `gorm:"type:uuid;index" json:"resource_id,omitempty"`
	ResourceName *string    `gorm:"size:255" json:"resource_name,omitempty"`
	OldValue     JSON       `gorm:"type:jsonb" json:"old_value,omitempty"` // Before state
	NewValue     JSON       `gorm:"type:jsonb" json:"new_value,omitempty"` // After state
	Metadata     JSON       `gorm:"type:jsonb" json:"metadata,omitempty"`  // Additional context
	IPAddress    *string    `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent    *string    `gorm:"size:500" json:"user_agent,omitempty"`
	CreatedAt    time.Time  `gorm:"index" json:"created_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// Permission represents granular RBAC permissions
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	Resource    string    `gorm:"size:50;not null;index" json:"resource"` // workflow, credential, etc.
	Action      string    `gorm:"size:30;not null" json:"action"`         // create, read, update, delete, execute
	CreatedAt   time.Time `json:"created_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

// Role represents custom workspace roles
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID *uuid.UUID     `gorm:"type:uuid;index" json:"workspace_id,omitempty"` // nil = system role
	Name        string         `gorm:"size:50;not null" json:"name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"` // Cannot be deleted
	Color       *string        `gorm:"size:7" json:"color,omitempty"`  // Hex color
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace   *Workspace       `gorm:"foreignKey:WorkspaceID" json:"-"`
	Permissions []RolePermission `gorm:"foreignKey:RoleID" json:"-"`
}

func (Role) TableName() string {
	return "roles"
}

// RolePermission links roles to permissions
type RolePermission struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoleID       uuid.UUID `gorm:"type:uuid;index;not null" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;index;not null" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`

	Role       Role       `gorm:"foreignKey:RoleID" json:"-"`
	Permission Permission `gorm:"foreignKey:PermissionID" json:"-"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// =============================================================================
// ENVIRONMENT VARIABLES
// =============================================================================

// EnvironmentVariable represents workspace-level secrets/variables
type EnvironmentVariable struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Value       string         `gorm:"type:text;not null" json:"-"`   // Encrypted
	IsSecret    bool           `gorm:"default:false" json:"is_secret"` // If true, mask in logs
	Environment *string        `gorm:"size:20" json:"environment,omitempty"` // dev, staging, prod
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Creator   User      `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (EnvironmentVariable) TableName() string {
	return "environment_variables"
}

// =============================================================================
// WEBHOOK FEATURES
// =============================================================================

// WebhookSignatureAlgorithm constants
const (
	WebhookSigHMACSHA256 = "hmac-sha256"
	WebhookSigHMACSHA1   = "hmac-sha1"
	WebhookSigHMACSHA512 = "hmac-sha512"
)

// WebhookSignatureConfig stores signature verification settings
type WebhookSignatureConfig struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WebhookID       uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"webhook_id"`
	Algorithm       string     `gorm:"size:20;not null;default:hmac-sha256" json:"algorithm"`
	Secret          string     `gorm:"size:255;not null" json:"-"` // Encrypted
	HeaderName      string     `gorm:"size:100;not null;default:X-Signature" json:"header_name"`
	SignaturePrefix *string    `gorm:"size:20" json:"signature_prefix,omitempty"` // e.g., "sha256="
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	FailOnInvalid   bool       `gorm:"default:true" json:"fail_on_invalid"` // Reject invalid signatures
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	Webhook WebhookEndpoint `gorm:"foreignKey:WebhookID" json:"-"`
}

func (WebhookSignatureConfig) TableName() string {
	return "webhook_signature_configs"
}

// =============================================================================
// RATE LIMITING
// =============================================================================

// CredentialRateLimit stores per-credential rate limiting config
type CredentialRateLimit struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CredentialID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"credential_id"`
	RequestsPerMin  int       `gorm:"default:60" json:"requests_per_min"`
	RequestsPerHour int       `gorm:"default:1000" json:"requests_per_hour"`
	RequestsPerDay  int       `gorm:"default:10000" json:"requests_per_day"`
	BurstLimit      int       `gorm:"default:10" json:"burst_limit"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	CurrentMinute   int       `gorm:"default:0" json:"-"`
	CurrentHour     int       `gorm:"default:0" json:"-"`
	CurrentDay      int       `gorm:"default:0" json:"-"`
	LastResetMin    time.Time `json:"-"`
	LastResetHour   time.Time `json:"-"`
	LastResetDay    time.Time `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Credential Credential `gorm:"foreignKey:CredentialID" json:"-"`
}

func (CredentialRateLimit) TableName() string {
	return "credential_rate_limits"
}

// =============================================================================
// SUB-WORKFLOWS
// =============================================================================

// SubWorkflowExecution tracks sub-workflow executions
type SubWorkflowExecution struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ParentExecutionID uuid.UUID  `gorm:"type:uuid;index;not null" json:"parent_execution_id"`
	ChildExecutionID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"child_execution_id"`
	ParentNodeID      string     `gorm:"size:100;not null" json:"parent_node_id"`
	InputMapping      JSON       `gorm:"type:jsonb" json:"input_mapping,omitempty"`
	OutputMapping     JSON       `gorm:"type:jsonb" json:"output_mapping,omitempty"`
	Status            string     `gorm:"size:20;not null;default:pending" json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`

	ParentExecution Execution `gorm:"foreignKey:ParentExecutionID" json:"-"`
	ChildExecution  Execution `gorm:"foreignKey:ChildExecutionID" json:"-"`
}

func (SubWorkflowExecution) TableName() string {
	return "sub_workflow_executions"
}

// =============================================================================
// WORKFLOW IMPORT/EXPORT
// =============================================================================

// WorkflowExport tracks workflow exports
type WorkflowExport struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID  uuid.UUID `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	ExportedBy  uuid.UUID `gorm:"type:uuid;not null" json:"exported_by"`
	Version     int       `gorm:"not null" json:"version"`
	Format      string    `gorm:"size:20;not null;default:json" json:"format"` // json, yaml
	FileSize    int       `gorm:"default:0" json:"file_size"`
	IncludeCredentials bool `gorm:"default:false" json:"include_credentials"`
	CreatedAt   time.Time `json:"created_at"`

	Workflow  Workflow  `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Exporter  User      `gorm:"foreignKey:ExportedBy" json:"-"`
}

func (WorkflowExport) TableName() string {
	return "workflow_exports"
}

// WorkflowImport tracks workflow imports
type WorkflowImport struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID  *uuid.UUID `gorm:"type:uuid;index" json:"workflow_id,omitempty"` // Created workflow
	WorkspaceID uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	ImportedBy  uuid.UUID  `gorm:"type:uuid;not null" json:"imported_by"`
	SourceName  *string    `gorm:"size:255" json:"source_name,omitempty"` // Original workflow name
	SourceType  string     `gorm:"size:20;not null" json:"source_type"`   // file, url, template
	Status      string     `gorm:"size:20;not null;default:pending" json:"status"`
	Error       *string    `gorm:"type:text" json:"error,omitempty"`
	Warnings    JSON       `gorm:"type:jsonb" json:"warnings,omitempty"` // Non-fatal issues
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	Workflow  *Workflow `gorm:"foreignKey:WorkflowID" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Importer  User      `gorm:"foreignKey:ImportedBy" json:"-"`
}

func (WorkflowImport) TableName() string {
	return "workflow_imports"
}

// =============================================================================
// SYSTEM DEFAULTS
// =============================================================================

// DefaultPermissions returns the system-default permissions
func DefaultPermissions() []Permission {
	resources := []string{"workflow", "credential", "execution", "schedule", "webhook", "member", "settings", "billing"}
	actions := []string{"create", "read", "update", "delete", "execute"}

	var perms []Permission
	for _, resource := range resources {
		for _, action := range actions {
			// Skip non-applicable combinations
			if action == "execute" && resource != "workflow" {
				continue
			}
			perms = append(perms, Permission{
				Name:     resource + ":" + action,
				Resource: resource,
				Action:   action,
			})
		}
	}
	return perms
}

// DefaultRoles returns the system-default roles
func DefaultRoles() []Role {
	return []Role{
		{
			Name:        "owner",
			Description: strPtr("Full access to workspace"),
			IsSystem:    true,
		},
		{
			Name:        "admin",
			Description: strPtr("Administrative access"),
			IsSystem:    true,
		},
		{
			Name:        "editor",
			Description: strPtr("Can create and edit workflows"),
			IsSystem:    true,
		},
		{
			Name:        "viewer",
			Description: strPtr("Read-only access"),
			IsSystem:    true,
		},
		{
			Name:        "executor",
			Description: strPtr("Can only execute workflows"),
			IsSystem:    true,
		},
	}
}


