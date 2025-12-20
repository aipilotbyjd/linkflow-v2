package logic

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

type WaitNode struct{}

func NewWaitNode() *WaitNode {
	return &WaitNode{}
}

func (n *WaitNode) Type() string {
	return "logic.wait"
}

func (n *WaitNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	unit, _ := execCtx.Config["unit"].(string)
	if unit == "" {
		unit = "seconds"
	}

	amount := core.GetInt(execCtx.Config, "amount", 1)
	maxWait := core.GetInt(execCtx.Config, "maxWait", 3600)

	var duration time.Duration
	switch unit {
	case "milliseconds":
		duration = time.Duration(amount) * time.Millisecond
	case "seconds":
		duration = time.Duration(amount) * time.Second
	case "minutes":
		duration = time.Duration(amount) * time.Minute
	case "hours":
		duration = time.Duration(amount) * time.Hour
	default:
		duration = time.Duration(amount) * time.Second
	}

	if duration > time.Duration(maxWait)*time.Second {
		duration = time.Duration(maxWait) * time.Second
	}

	startTime := time.Now()

	select {
	case <-ctx.Done():
		return map[string]interface{}{
			"waited":      false,
			"interrupted": true,
			"waitedMs":    time.Since(startTime).Milliseconds(),
			"reason":      "cancelled",
		}, ctx.Err()
	case <-time.After(duration):
		return map[string]interface{}{
			"waited":      true,
			"interrupted": false,
			"waitedMs":    duration.Milliseconds(),
			"data":        execCtx.Input["$json"],
		}, nil
	}
}

type ErrorTriggerNode struct{}

func NewErrorTriggerNode() *ErrorTriggerNode {
	return &ErrorTriggerNode{}
}

func (n *ErrorTriggerNode) Type() string {
	return "trigger.error"
}

func (n *ErrorTriggerNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	errorData := execCtx.Input["$error"]
	if errorData == nil {
		errorData = map[string]interface{}{
			"message": "No error data available",
		}
	}

	return map[string]interface{}{
		"error":       errorData,
		"triggered":   true,
		"executionId": execCtx.ExecutionID,
		"workflowId":  execCtx.WorkflowID,
	}, nil
}

type StopErrorNode struct{}

func NewStopErrorNode() *StopErrorNode {
	return &StopErrorNode{}
}

func (n *StopErrorNode) Type() string {
	return "action.stopError"
}

func (n *StopErrorNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	errorMsg, _ := execCtx.Config["message"].(string)
	if errorMsg == "" {
		errorMsg = "Workflow stopped by Stop and Error node"
	}

	errorType, _ := execCtx.Config["errorType"].(string)
	if errorType == "" {
		errorType = "workflow_error"
	}

	return map[string]interface{}{
		"error":     true,
		"message":   errorMsg,
		"errorType": errorType,
		"stopped":   true,
	}, &WorkflowError{
		Message: errorMsg,
		Type:    errorType,
	}
}

type WorkflowError struct {
	Message string
	Type    string
}

func (e *WorkflowError) Error() string {
	return e.Message
}

type RespondWebhookNode struct{}

func NewRespondWebhookNode() *RespondWebhookNode {
	return &RespondWebhookNode{}
}

func (n *RespondWebhookNode) Type() string {
	return "action.respondWebhook"
}

func (n *RespondWebhookNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	statusCode := core.GetInt(execCtx.Config, "statusCode", 200)
	body := execCtx.Config["body"]
	headers, _ := execCtx.Config["headers"].(map[string]interface{})
	contentType, _ := execCtx.Config["contentType"].(string)
	if contentType == "" {
		contentType = "application/json"
	}

	if headers == nil {
		headers = make(map[string]interface{})
	}
	headers["Content-Type"] = contentType

	response := map[string]interface{}{
		"statusCode": statusCode,
		"body":       body,
		"headers":    headers,
		"responded":  true,
	}

	if execCtx.Variables != nil {
		execCtx.Variables["$webhookResponse"] = response
	}

	return response, nil
}

type NoOpNode struct{}

func NewNoOpNode() *NoOpNode {
	return &NoOpNode{}
}

func (n *NoOpNode) Type() string {
	return "logic.noop"
}

func (n *NoOpNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"data": execCtx.Input["$json"],
	}, nil
}
