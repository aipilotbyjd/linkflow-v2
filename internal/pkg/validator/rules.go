package validator

import (
	"errors"
	"fmt"
	"strings"
)

// Business rule errors
var (
	// Workflow errors
	ErrWorkflowNotActive       = errors.New("workflow is not active")
	ErrWorkflowNotDraft        = errors.New("workflow is not in draft status")
	ErrWorkflowAlreadyActive   = errors.New("workflow is already active")
	ErrWorkflowAlreadyInactive = errors.New("workflow is already inactive")
	ErrWorkflowArchived        = errors.New("workflow is archived")
	ErrWorkflowNoTrigger       = errors.New("workflow must have at least one trigger node to be activated")
	ErrWorkflowHasSchedules    = errors.New("cannot deactivate workflow with active schedules")
	ErrWorkflowHasExecutions   = errors.New("cannot delete workflow with running executions")
	ErrWorkflowEmpty           = errors.New("workflow has no nodes")

	// Execution errors
	ErrExecutionNotRunning    = errors.New("execution is not running")
	ErrExecutionAlreadyDone   = errors.New("execution is already completed")
	ErrExecutionNotRetryable  = errors.New("only failed executions can be retried")
	ErrExecutionNotCancelable = errors.New("only running or queued executions can be cancelled")

	// Credential errors
	ErrCredentialInUse    = errors.New("credential is in use by one or more workflows")
	ErrCredentialExpired  = errors.New("credential has expired")
	ErrCredentialNotFound = errors.New("referenced credential not found")

	// Schedule errors
	ErrScheduleWorkflowInactive = errors.New("cannot create schedule for inactive workflow")
	ErrScheduleInvalidCron      = errors.New("invalid cron expression")
	ErrScheduleInvalidTimezone  = errors.New("invalid timezone")
	ErrScheduleAlreadyExists    = errors.New("schedule already exists for this workflow")

	// Webhook errors
	ErrWebhookPathExists  = errors.New("webhook path already exists in this workspace")
	ErrWebhookPathInvalid = errors.New("invalid webhook path")

	// Quota/Plan errors
	ErrQuotaWorkflowsExceeded   = errors.New("workflow limit exceeded for current plan")
	ErrQuotaExecutionsExceeded  = errors.New("execution limit exceeded for current plan")
	ErrQuotaMembersExceeded     = errors.New("member limit exceeded for current plan")
	ErrQuotaCredentialsExceeded = errors.New("credential limit exceeded for current plan")
	ErrQuotaSchedulesExceeded   = errors.New("schedule limit exceeded for current plan")

	// General errors
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("access denied")
)

// Workflow status constants
const (
	WorkflowStatusDraft    = "draft"
	WorkflowStatusActive   = "active"
	WorkflowStatusInactive = "inactive"
	WorkflowStatusArchived = "archived"
)

// Execution status constants
const (
	ExecutionStatusQueued    = "queued"
	ExecutionStatusRunning   = "running"
	ExecutionStatusCompleted = "completed"
	ExecutionStatusFailed    = "failed"
	ExecutionStatusCancelled = "cancelled"
	ExecutionStatusTimeout   = "timeout"
)

// RuleError wraps a rule violation with context
type RuleError struct {
	Rule    string
	Message string
	Err     error
}

func (e *RuleError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *RuleError) Unwrap() error {
	return e.Err
}

// NewRuleError creates a new rule error
func NewRuleError(rule string, err error) *RuleError {
	return &RuleError{Rule: rule, Err: err}
}

// NewRuleErrorWithMessage creates a new rule error with custom message
func NewRuleErrorWithMessage(rule, message string, err error) *RuleError {
	return &RuleError{Rule: rule, Message: message, Err: err}
}

// WorkflowRules validates workflow business rules
type WorkflowRules struct{}

// NewWorkflowRules creates a new workflow rules validator
func NewWorkflowRules() *WorkflowRules {
	return &WorkflowRules{}
}

// CanExecute checks if a workflow can be executed
func (r *WorkflowRules) CanExecute(status string, hasNodes bool) error {
	if !hasNodes {
		return NewRuleError("workflow.execute", ErrWorkflowEmpty)
	}

	switch status {
	case WorkflowStatusActive:
		return nil
	case WorkflowStatusDraft:
		// Allow executing draft workflows for testing
		return nil
	case WorkflowStatusInactive:
		return NewRuleError("workflow.execute", ErrWorkflowNotActive)
	case WorkflowStatusArchived:
		return NewRuleError("workflow.execute", ErrWorkflowArchived)
	default:
		return NewRuleError("workflow.execute", ErrWorkflowNotActive)
	}
}

// CanActivate checks if a workflow can be activated
func (r *WorkflowRules) CanActivate(status string, hasTrigger bool) error {
	if status == WorkflowStatusActive {
		return NewRuleError("workflow.activate", ErrWorkflowAlreadyActive)
	}

	if status == WorkflowStatusArchived {
		return NewRuleError("workflow.activate", ErrWorkflowArchived)
	}

	if !hasTrigger {
		return NewRuleError("workflow.activate", ErrWorkflowNoTrigger)
	}

	return nil
}

// CanDeactivate checks if a workflow can be deactivated
func (r *WorkflowRules) CanDeactivate(status string, hasActiveSchedules bool) error {
	if status != WorkflowStatusActive {
		return NewRuleError("workflow.deactivate", ErrWorkflowAlreadyInactive)
	}

	if hasActiveSchedules {
		return NewRuleError("workflow.deactivate", ErrWorkflowHasSchedules)
	}

	return nil
}

// CanDelete checks if a workflow can be deleted
func (r *WorkflowRules) CanDelete(hasRunningExecutions bool) error {
	if hasRunningExecutions {
		return NewRuleError("workflow.delete", ErrWorkflowHasExecutions)
	}
	return nil
}

// CanArchive checks if a workflow can be archived
func (r *WorkflowRules) CanArchive(status string) error {
	if status == WorkflowStatusArchived {
		return NewRuleError("workflow.archive", errors.New("workflow is already archived"))
	}
	return nil
}

// ExecutionRules validates execution business rules
type ExecutionRules struct{}

// NewExecutionRules creates a new execution rules validator
func NewExecutionRules() *ExecutionRules {
	return &ExecutionRules{}
}

// CanCancel checks if an execution can be cancelled
func (r *ExecutionRules) CanCancel(status string) error {
	switch status {
	case ExecutionStatusQueued, ExecutionStatusRunning:
		return nil
	case ExecutionStatusCompleted, ExecutionStatusFailed, ExecutionStatusCancelled, ExecutionStatusTimeout:
		return NewRuleError("execution.cancel", ErrExecutionNotCancelable)
	default:
		return NewRuleError("execution.cancel", ErrExecutionNotCancelable)
	}
}

// CanRetry checks if an execution can be retried
func (r *ExecutionRules) CanRetry(status string) error {
	switch status {
	case ExecutionStatusFailed, ExecutionStatusTimeout, ExecutionStatusCancelled:
		return nil
	case ExecutionStatusCompleted:
		return NewRuleError("execution.retry", ErrExecutionAlreadyDone)
	case ExecutionStatusQueued, ExecutionStatusRunning:
		return NewRuleError("execution.retry", ErrExecutionNotRunning)
	default:
		return NewRuleError("execution.retry", ErrExecutionNotRetryable)
	}
}

// IsTerminal checks if an execution is in a terminal state
func (r *ExecutionRules) IsTerminal(status string) bool {
	switch status {
	case ExecutionStatusCompleted, ExecutionStatusFailed, ExecutionStatusCancelled, ExecutionStatusTimeout:
		return true
	default:
		return false
	}
}

// CredentialRules validates credential business rules
type CredentialRules struct{}

// NewCredentialRules creates a new credential rules validator
func NewCredentialRules() *CredentialRules {
	return &CredentialRules{}
}

// CanDelete checks if a credential can be deleted
func (r *CredentialRules) CanDelete(workflowsUsingCredential int) error {
	if workflowsUsingCredential > 0 {
		return NewRuleErrorWithMessage(
			"credential.delete",
			fmt.Sprintf("credential is used by %d workflow(s)", workflowsUsingCredential),
			ErrCredentialInUse,
		)
	}
	return nil
}

// ScheduleRules validates schedule business rules
type ScheduleRules struct{}

// NewScheduleRules creates a new schedule rules validator
func NewScheduleRules() *ScheduleRules {
	return &ScheduleRules{}
}

// CanCreate checks if a schedule can be created
func (r *ScheduleRules) CanCreate(workflowStatus string) error {
	if workflowStatus != WorkflowStatusActive && workflowStatus != WorkflowStatusDraft {
		return NewRuleError("schedule.create", ErrScheduleWorkflowInactive)
	}
	return nil
}

// QuotaRules validates plan quota rules
type QuotaRules struct{}

// NewQuotaRules creates a new quota rules validator
func NewQuotaRules() *QuotaRules {
	return &QuotaRules{}
}

// CheckWorkflowQuota checks if workspace can create more workflows
func (r *QuotaRules) CheckWorkflowQuota(current, limit int) error {
	if limit > 0 && current >= limit {
		return NewRuleErrorWithMessage(
			"quota.workflows",
			fmt.Sprintf("workflow limit reached (%d/%d)", current, limit),
			ErrQuotaWorkflowsExceeded,
		)
	}
	return nil
}

// CheckExecutionQuota checks if workspace can run more executions
func (r *QuotaRules) CheckExecutionQuota(current, limit int) error {
	if limit > 0 && current >= limit {
		return NewRuleErrorWithMessage(
			"quota.executions",
			fmt.Sprintf("execution limit reached (%d/%d)", current, limit),
			ErrQuotaExecutionsExceeded,
		)
	}
	return nil
}

// CheckMemberQuota checks if workspace can add more members
func (r *QuotaRules) CheckMemberQuota(current, limit int) error {
	if limit > 0 && current >= limit {
		return NewRuleErrorWithMessage(
			"quota.members",
			fmt.Sprintf("member limit reached (%d/%d)", current, limit),
			ErrQuotaMembersExceeded,
		)
	}
	return nil
}

// CheckCredentialQuota checks if workspace can create more credentials
func (r *QuotaRules) CheckCredentialQuota(current, limit int) error {
	if limit > 0 && current >= limit {
		return NewRuleErrorWithMessage(
			"quota.credentials",
			fmt.Sprintf("credential limit reached (%d/%d)", current, limit),
			ErrQuotaCredentialsExceeded,
		)
	}
	return nil
}

// WebhookRules validates webhook business rules
type WebhookRules struct{}

// NewWebhookRules creates a new webhook rules validator
func NewWebhookRules() *WebhookRules {
	return &WebhookRules{}
}

// ValidatePath validates webhook path
func (r *WebhookRules) ValidatePath(path string) error {
	if path == "" {
		return nil // Empty path will be auto-generated
	}

	// Check for invalid characters
	if !webhookPathRegex.MatchString(path) {
		return NewRuleError("webhook.path", ErrWebhookPathInvalid)
	}

	// Check for reserved paths
	reserved := []string{"health", "metrics", "api", "admin", "system"}
	lower := strings.ToLower(path)
	for _, r := range reserved {
		if lower == r {
			return NewRuleErrorWithMessage(
				"webhook.path",
				fmt.Sprintf("'%s' is a reserved path", path),
				ErrWebhookPathInvalid,
			)
		}
	}

	return nil
}

// Package-level convenience functions

var (
	workflowRules   = NewWorkflowRules()
	executionRules  = NewExecutionRules()
	credentialRules = NewCredentialRules()
	scheduleRules   = NewScheduleRules()
	quotaRules      = NewQuotaRules()
	webhookRules    = NewWebhookRules()
)

// CanExecuteWorkflow checks if a workflow can be executed
func CanExecuteWorkflow(status string, hasNodes bool) error {
	return workflowRules.CanExecute(status, hasNodes)
}

// CanActivateWorkflow checks if a workflow can be activated
func CanActivateWorkflow(status string, hasTrigger bool) error {
	return workflowRules.CanActivate(status, hasTrigger)
}

// CanDeactivateWorkflow checks if a workflow can be deactivated
func CanDeactivateWorkflow(status string, hasActiveSchedules bool) error {
	return workflowRules.CanDeactivate(status, hasActiveSchedules)
}

// CanDeleteWorkflow checks if a workflow can be deleted
func CanDeleteWorkflow(hasRunningExecutions bool) error {
	return workflowRules.CanDelete(hasRunningExecutions)
}

// CanCancelExecution checks if an execution can be cancelled
func CanCancelExecution(status string) error {
	return executionRules.CanCancel(status)
}

// CanRetryExecution checks if an execution can be retried
func CanRetryExecution(status string) error {
	return executionRules.CanRetry(status)
}

// CanDeleteCredential checks if a credential can be deleted
func CanDeleteCredential(workflowsUsingCredential int) error {
	return credentialRules.CanDelete(workflowsUsingCredential)
}

// CanCreateSchedule checks if a schedule can be created
func CanCreateSchedule(workflowStatus string) error {
	return scheduleRules.CanCreate(workflowStatus)
}

// CheckQuota checks various quota limits
func CheckQuota(quotaType string, current, limit int) error {
	switch quotaType {
	case "workflows":
		return quotaRules.CheckWorkflowQuota(current, limit)
	case "executions":
		return quotaRules.CheckExecutionQuota(current, limit)
	case "members":
		return quotaRules.CheckMemberQuota(current, limit)
	case "credentials":
		return quotaRules.CheckCredentialQuota(current, limit)
	default:
		return nil
	}
}

// ValidateWebhookPath validates a webhook path
func ValidateWebhookPath(path string) error {
	return webhookRules.ValidatePath(path)
}
