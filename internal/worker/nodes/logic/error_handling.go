package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// TryCatchNode wraps node execution with error handling
type TryCatchNode struct{}

func (n *TryCatchNode) Type() string {
	return "logic.try_catch"
}

func (n *TryCatchNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	continueOnFail := core.GetBool(config, "continueOnFail", true)
	errorOutput := core.GetString(config, "errorOutput", "error")

	// The try block execution is handled by the executor
	// This node just passes through input and handles error routing

	// Check if there's an error from previous execution
	if errMsg, hasError := input["$error"].(string); hasError {
		if !continueOnFail {
			return nil, fmt.Errorf("try block failed: %s", errMsg)
		}

		// Route to error output
		return map[string]interface{}{
			"output":      errorOutput,
			"hasError":    true,
			"error":       errMsg,
			"errorData":   input["$errorData"],
			"originalInput": input["$originalInput"],
		}, nil
	}

	// No error, pass through
	return map[string]interface{}{
		"output":   "success",
		"hasError": false,
		"data":     input,
	}, nil
}

// RetryNode retries failed operations with configurable backoff
type RetryNode struct{}

func (n *RetryNode) Type() string {
	return "logic.retry"
}

func (n *RetryNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	maxRetries := core.GetInt(config, "maxRetries", 3)
	initialDelay := core.GetInt(config, "initialDelay", 1000) // ms
	maxDelay := core.GetInt(config, "maxDelay", 30000)        // ms
	backoffType := core.GetString(config, "backoffType", "exponential")
	retryOn := core.GetStringArray(config, "retryOn") // specific error types to retry

	// Get current retry count
	retryCount := 0
	if count, ok := input["$retryCount"].(float64); ok {
		retryCount = int(count)
	}

	// Check if we have an error to retry
	errMsg, hasError := input["$error"].(string)
	if !hasError {
		// No error, pass through
		return input, nil
	}

	// Check if we should retry this error type
	if len(retryOn) > 0 && !containsError(retryOn, errMsg) {
		return nil, fmt.Errorf("non-retryable error: %s", errMsg)
	}

	// Check max retries
	if retryCount >= maxRetries {
		return nil, fmt.Errorf("max retries (%d) exceeded: %s", maxRetries, errMsg)
	}

	// Calculate delay
	delay := calculateDelay(backoffType, initialDelay, maxDelay, retryCount)

	// Wait before retry
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Duration(delay) * time.Millisecond):
	}

	// Signal retry
	return map[string]interface{}{
		"$retry":      true,
		"$retryCount": retryCount + 1,
		"$retryDelay": delay,
		"data":        input["$originalInput"],
	}, nil
}

func calculateDelay(backoffType string, initialDelay, maxDelay, retryCount int) int {
	var delay int
	switch backoffType {
	case "fixed":
		delay = initialDelay
	case "linear":
		delay = initialDelay * (retryCount + 1)
	case "exponential":
		delay = initialDelay * (1 << retryCount)
	default:
		delay = initialDelay
	}
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

func containsError(patterns []string, errMsg string) bool {
	for _, pattern := range patterns {
		if pattern == "*" || contains(errMsg, pattern) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ThrowErrorNode throws a custom error
type ThrowErrorNode struct{}

func (n *ThrowErrorNode) Type() string {
	return "logic.throw_error"
}

func (n *ThrowErrorNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config

	errorMessage := core.GetString(config, "errorMessage", "Custom error")
	errorType := core.GetString(config, "errorType", "Error")

	return nil, &CustomError{
		Type:    errorType,
		Message: errorMessage,
	}
}

type CustomError struct {
	Type    string
	Message string
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// ContinueOnFailNode allows workflow to continue even if a node fails
type ContinueOnFailNode struct{}

func (n *ContinueOnFailNode) Type() string {
	return "logic.continue_on_fail"
}

func (n *ContinueOnFailNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	input := execCtx.Input

	// Check for error from previous node
	if errMsg, hasError := input["$error"].(string); hasError {
		return map[string]interface{}{
			"success":   false,
			"error":     errMsg,
			"errorData": input["$errorData"],
			"continued": true,
		}, nil
	}

	return map[string]interface{}{
		"success":   true,
		"data":      input,
		"continued": false,
	}, nil
}

// TimeoutNode adds timeout to node execution
type TimeoutNode struct{}

func (n *TimeoutNode) Type() string {
	return "logic.timeout"
}

func (n *TimeoutNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	timeoutMs := core.GetInt(config, "timeout", 30000)
	onTimeout := core.GetString(config, "onTimeout", "error") // error, continue, default

	// Check if timeout occurred
	if timedOut, ok := input["$timedOut"].(bool); ok && timedOut {
		switch onTimeout {
		case "error":
			return nil, fmt.Errorf("operation timed out after %dms", timeoutMs)
		case "continue":
			return map[string]interface{}{
				"timedOut":     true,
				"data":         nil,
				"timeoutAfter": timeoutMs,
			}, nil
		case "default":
			defaultValue := config["defaultValue"]
			return map[string]interface{}{
				"timedOut":     true,
				"data":         defaultValue,
				"timeoutAfter": timeoutMs,
			}, nil
		}
	}

	// Pass configuration for executor to enforce timeout
	return map[string]interface{}{
		"$timeout": timeoutMs,
		"data":     input,
	}, nil
}

// FallbackNode provides fallback behavior when a node fails
type FallbackNode struct{}

func (n *FallbackNode) Type() string {
	return "logic.fallback"
}

func (n *FallbackNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	fallbackValue := config["fallbackValue"]
	useFallbackOn := core.GetStringArray(config, "useFallbackOn") // error types

	// Check for error
	errMsg, hasError := input["$error"].(string)
	if !hasError {
		return input, nil
	}

	// Check if we should use fallback for this error
	if len(useFallbackOn) > 0 && !containsError(useFallbackOn, errMsg) {
		return nil, fmt.Errorf("error without fallback: %s", errMsg)
	}

	return map[string]interface{}{
		"data":         fallbackValue,
		"usedFallback": true,
		"originalError": errMsg,
	}, nil
}


