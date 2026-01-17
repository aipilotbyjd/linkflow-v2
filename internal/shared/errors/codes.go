package errors

// Error codes for consistent error handling across the application
const (
	// Authentication errors (1xxx)
	CodeInvalidCredentials = "AUTH_1001"
	CodeTokenExpired       = "AUTH_1002"
	CodeTokenInvalid       = "AUTH_1003"
	CodeSessionExpired     = "AUTH_1004"
	CodeMFARequired        = "AUTH_1005"
	CodeMFAInvalid         = "AUTH_1006"
	CodeAccountLocked      = "AUTH_1007"
	CodeEmailNotVerified   = "AUTH_1008"
	CodePasswordTooWeak    = "AUTH_1009"
	CodeOAuthError         = "AUTH_1010"

	// Authorization errors (2xxx)
	CodeAccessDenied      = "AUTHZ_2001"
	CodeInsufficientRole  = "AUTHZ_2002"
	CodeResourceNotOwned  = "AUTHZ_2003"
	CodeWorkspaceRequired = "AUTHZ_2004"

	// Validation errors (3xxx)
	CodeValidationFailed = "VAL_3001"
	CodeRequiredField    = "VAL_3002"
	CodeInvalidFormat    = "VAL_3003"
	CodeOutOfRange       = "VAL_3004"
	CodeDuplicateValue   = "VAL_3005"
	CodeInvalidJSON      = "VAL_3006"

	// Resource errors (4xxx)
	CodeNotFound         = "RES_4001"
	CodeAlreadyExists    = "RES_4002"
	CodeConflict         = "RES_4003"
	CodeDependencyExists = "RES_4004"
	CodeResourceLocked   = "RES_4005"

	// Workflow errors (5xxx)
	CodeWorkflowInactive   = "WF_5001"
	CodeWorkflowInvalid    = "WF_5002"
	CodeNoTriggerNode      = "WF_5003"
	CodeCircularDependency = "WF_5004"
	CodeNodeNotFound       = "WF_5005"
	CodeConnectionInvalid  = "WF_5006"

	// Execution errors (6xxx)
	CodeExecutionFailed     = "EXEC_6001"
	CodeExecutionTimeout    = "EXEC_6002"
	CodeExecutionCancelled  = "EXEC_6003"
	CodeNodeExecutionFailed = "EXEC_6004"
	CodeRetryLimitExceeded  = "EXEC_6005"
	CodeInvalidTriggerData  = "EXEC_6006"

	// Credential errors (7xxx)
	CodeCredentialInvalid   = "CRED_7001"
	CodeCredentialExpired   = "CRED_7002"
	CodeCredentialNotShared = "CRED_7003"
	CodeEncryptionError     = "CRED_7004"

	// Rate limiting errors (8xxx)
	CodeRateLimited      = "RATE_8001"
	CodeQuotaExceeded    = "RATE_8002"
	CodePlanLimitReached = "RATE_8003"

	// Integration errors (9xxx)
	CodeIntegrationError = "INT_9001"
	CodeAPIError         = "INT_9002"
	CodeConnectionFailed = "INT_9003"
	CodeResponseInvalid  = "INT_9004"

	// Internal errors (10xxx)
	CodeInternalError = "INT_10001"
	CodeDatabaseError = "INT_10002"
	CodeCacheError    = "INT_10003"
	CodeQueueError    = "INT_10004"
)

// ErrorMessages maps error codes to human-readable messages
var ErrorMessages = map[string]string{
	CodeInvalidCredentials:  "Invalid email or password",
	CodeTokenExpired:        "Token has expired",
	CodeTokenInvalid:        "Invalid token",
	CodeSessionExpired:      "Session has expired",
	CodeMFARequired:         "MFA verification required",
	CodeMFAInvalid:          "Invalid MFA code",
	CodeAccountLocked:       "Account is locked",
	CodeEmailNotVerified:    "Email not verified",
	CodePasswordTooWeak:     "Password does not meet requirements",
	CodeOAuthError:          "OAuth authentication failed",
	CodeAccessDenied:        "Access denied",
	CodeInsufficientRole:    "Insufficient permissions",
	CodeResourceNotOwned:    "Resource not owned by user",
	CodeWorkspaceRequired:   "Workspace context required",
	CodeValidationFailed:    "Validation failed",
	CodeRequiredField:       "Required field missing",
	CodeInvalidFormat:       "Invalid format",
	CodeOutOfRange:          "Value out of range",
	CodeDuplicateValue:      "Duplicate value",
	CodeInvalidJSON:         "Invalid JSON",
	CodeNotFound:            "Resource not found",
	CodeAlreadyExists:       "Resource already exists",
	CodeConflict:            "Resource conflict",
	CodeDependencyExists:    "Dependent resources exist",
	CodeResourceLocked:      "Resource is locked",
	CodeWorkflowInactive:    "Workflow is not active",
	CodeWorkflowInvalid:     "Invalid workflow configuration",
	CodeNoTriggerNode:       "Workflow requires a trigger node",
	CodeCircularDependency:  "Circular dependency detected",
	CodeNodeNotFound:        "Node not found",
	CodeConnectionInvalid:   "Invalid connection",
	CodeExecutionFailed:     "Execution failed",
	CodeExecutionTimeout:    "Execution timed out",
	CodeExecutionCancelled:  "Execution canceled",
	CodeNodeExecutionFailed: "Node execution failed",
	CodeRetryLimitExceeded:  "Retry limit exceeded",
	CodeInvalidTriggerData:  "Invalid trigger data",
	CodeCredentialInvalid:   "Invalid credentials",
	CodeCredentialExpired:   "Credential expired",
	CodeCredentialNotShared: "Credential not shared with user",
	CodeEncryptionError:     "Encryption error",
	CodeRateLimited:         "Rate limit exceeded",
	CodeQuotaExceeded:       "Quota exceeded",
	CodePlanLimitReached:    "Plan limit reached",
	CodeIntegrationError:    "Integration error",
	CodeAPIError:            "External API error",
	CodeConnectionFailed:    "Connection failed",
	CodeResponseInvalid:     "Invalid response",
	CodeInternalError:       "Internal server error",
	CodeDatabaseError:       "Database error",
	CodeCacheError:          "Cache error",
	CodeQueueError:          "Queue error",
}

// GetErrorMessage returns the human-readable message for an error code
func GetErrorMessage(code string) string {
	if msg, ok := ErrorMessages[code]; ok {
		return msg
	}
	return "Unknown error"
}
