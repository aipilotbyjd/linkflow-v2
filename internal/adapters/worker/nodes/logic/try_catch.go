package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// TryCatchNode provides error handling with try/catch/finally semantics
type TryCatchNode struct{}

// NewTryCatchNode creates a new try-catch node
func NewTryCatchNode() *TryCatchNode {
	return &TryCatchNode{}
}

// Execute handles error flow control
func (n *TryCatchNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	inputData := runtime.GetInputData()
	params, _ := node["parameters"].(map[string]interface{})

	// Check if we're receiving an error from upstream
	errorData, hasError := inputData["error"].(map[string]interface{})
	
	// Get retry settings
	maxRetries, _ := params["max_retries"].(float64)
	retryDelay, _ := params["retry_delay"].(float64)
	retryOn, _ := params["retry_on"].([]interface{})

	// Track retry count
	retryCount := 0
	if rc, ok := inputData["_retry_count"].(float64); ok {
		retryCount = int(rc)
	}

	result := types.JSON{
		"has_error":     hasError,
		"retry_count":   retryCount,
		"max_retries":   int(maxRetries),
		"retry_delay":   int(retryDelay),
		"retry_on":      retryOn,
		"continue_on_fail": params["continue_on_fail"],
	}

	if hasError {
		result["error"] = errorData
		result["error_message"] = errorData["message"]
		result["error_type"] = errorData["type"]
		
		// Check if should retry
		if retryCount < int(maxRetries) {
			result["should_retry"] = true
			result["_retry_count"] = retryCount + 1
		} else {
			result["should_retry"] = false
		}
	} else {
		result["data"] = inputData
	}

	return result, nil
}

// Metadata returns node metadata
func (n *TryCatchNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.try_catch",
		Name:        "Try/Catch",
		Description: "Handle errors gracefully with try/catch/finally semantics. Supports automatic retries and error recovery.",
		Category:    "logic",
		Version:     "1.0.0",
		Icon:        "shield",
		Color:       "#EF4444",
		Inputs: []wtypes.NodePort{
			{Name: "try", Type: "any", Description: "Main execution path (connect nodes to try)"},
			{Name: "error", Type: "object", Description: "Error input from failed nodes"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "success", Type: "any", Description: "Output on successful execution"},
			{Name: "catch", Type: "any", Description: "Output when error is caught (includes error details)"},
			{Name: "finally", Type: "any", Description: "Always executes after try/catch"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "continue_on_fail",
				DisplayName: "Continue on Failure",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Continue workflow execution even when an error occurs",
			},
			{
				Name:        "max_retries",
				DisplayName: "Max Retries",
				Type:        "number",
				Required:    false,
				Default:     0,
				Description: "Number of times to retry failed operation (0 = no retry)",
			},
			{
				Name:        "retry_delay",
				DisplayName: "Retry Delay (seconds)",
				Type:        "number",
				Required:    false,
				Default:     1,
				Description: "Wait time between retry attempts",
				ShowIf:      "max_retries > 0",
			},
			{
				Name:        "retry_backoff",
				DisplayName: "Exponential Backoff",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Double delay after each retry",
				ShowIf:      "max_retries > 0",
			},
			{
				Name:        "retry_on",
				DisplayName: "Retry On Error Types",
				Type:        "json",
				Required:    false,
				Description: "Only retry for specific error types (empty = retry all)",
				Placeholder: `["network_error", "timeout"]`,
				ShowIf:      "max_retries > 0",
			},
			{
				Name:        "error_output",
				DisplayName: "Error Output Mode",
				Type:        "options",
				Required:    false,
				Default:     "full",
				Description: "What to include in catch output",
				Options: []wtypes.ParamOption{
					{Name: "Full Error Details", Value: "full"},
					{Name: "Message Only", Value: "message"},
					{Name: "Sanitized (no stack trace)", Value: "sanitized"},
				},
			},
			{
				Name:        "fallback_value",
				DisplayName: "Fallback Value",
				Type:        "json",
				Required:    false,
				Description: "Default value to use when error occurs (instead of error object)",
			},
		},
	}
}

// ErrorThrowNode explicitly throws an error
type ErrorThrowNode struct{}

// NewErrorThrowNode creates a new error throw node
func NewErrorThrowNode() *ErrorThrowNode {
	return &ErrorThrowNode{}
}

// Execute throws an error with specified message and type
func (n *ErrorThrowNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	
	errorMessage, _ := params["message"].(string)
	errorType, _ := params["error_type"].(string)
	errorCode, _ := params["error_code"].(string)

	if errorMessage == "" {
		errorMessage = "Error thrown by workflow"
	}
	if errorType == "" {
		errorType = "custom"
	}

	return types.JSON{
		"error": map[string]interface{}{
			"message": errorMessage,
			"type":    errorType,
			"code":    errorCode,
			"thrown":  true,
		},
	}, nil
}

// Metadata returns node metadata
func (n *ErrorThrowNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.error_throw",
		Name:        "Throw Error",
		Description: "Explicitly throw an error in the workflow",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "error", Type: "error"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "message", Type: "string", Description: "Error message", Required: true},
			{Name: "error_type", Type: "string", Description: "Error type/category", Default: "custom"},
			{Name: "error_code", Type: "string", Description: "Error code for programmatic handling"},
		},
	}
}
