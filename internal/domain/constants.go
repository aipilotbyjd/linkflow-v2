package domain

// Workflow statuses
const (
	WorkflowStatusDraft    = "draft"
	WorkflowStatusActive   = "active"
	WorkflowStatusInactive = "inactive"
	WorkflowStatusArchived = "archived"
)

// Execution statuses
const (
	ExecStatusPending   = "pending"
	ExecStatusRunning   = "running"
	ExecStatusCompleted = "completed"
	ExecStatusFailed    = "failed"
	ExecStatusCancelled = "cancelled"
	ExecStatusTimeout   = "timeout"
	ExecStatusWaiting   = "waiting"
)

// Node execution statuses
const (
	NodeStatusPending   = "pending"
	NodeStatusRunning   = "running"
	NodeStatusCompleted = "completed"
	NodeStatusFailed    = "failed"
	NodeStatusSkipped   = "skipped"
)

// Trigger types
const (
	TriggerTypeManual   = "manual"
	TriggerTypeWebhook  = "webhook"
	TriggerTypeSchedule = "schedule"
	TriggerTypeReplay   = "replay"
	TriggerTypeAPI      = "api"
)

// Workspace roles
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

// Plan IDs
const (
	PlanFree       = "free"
	PlanStarter    = "starter"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"
)

// Credential types
const (
	CredTypeOAuth2     = "oauth2"
	CredTypeAPIKey     = "api_key"
	CredTypeBasicAuth  = "basic_auth"
	CredTypeBearerToken = "bearer_token"
	CredTypeCustom     = "custom"
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Resource types (for audit logs)
const (
	ResourceWorkflow   = "workflow"
	ResourceExecution  = "execution"
	ResourceCredential = "credential"
	ResourceSchedule   = "schedule"
	ResourceWebhook    = "webhook"
	ResourceWorkspace  = "workspace"
	ResourceUser       = "user"
	ResourceMember     = "member"
)

// Pagination defaults
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Time formats
const (
	TimeFormatISO8601 = "2006-01-02T15:04:05Z07:00"
	TimeFormatDate    = "2006-01-02"
)
