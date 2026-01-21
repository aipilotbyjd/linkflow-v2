package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// ApprovalNode pauses workflow execution until human approval
type ApprovalNode struct{}

// NewApprovalNode creates a new approval node
func NewApprovalNode() *ApprovalNode {
	return &ApprovalNode{}
}

// Execute pauses for approval and returns approval status
func (n *ApprovalNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	// Check if this is a resume with approval response
	if approvalResponse, ok := inputData["_approval_response"].(map[string]interface{}); ok {
		approved, _ := approvalResponse["approved"].(bool)
		approvedBy, _ := approvalResponse["approved_by"].(string)
		approvedAt, _ := approvalResponse["approved_at"].(string)
		comments, _ := approvalResponse["comments"].(string)

		return types.JSON{
			"approved":    approved,
			"approved_by": approvedBy,
			"approved_at": approvedAt,
			"comments":    comments,
			"data":        inputData["data"],
		}, nil
	}

	// First execution - create approval request
	title, _ := params["title"].(string)
	description, _ := params["description"].(string)
	approvers, _ := params["approvers"].([]interface{})
	timeoutHours, _ := params["timeout_hours"].(float64)
	notifyVia, _ := params["notify_via"].([]interface{})
	requireAll, _ := params["require_all_approvers"].(bool)

	if title == "" {
		title = "Approval Required"
	}
	if timeoutHours == 0 {
		timeoutHours = 72 // Default 3 days
	}

	// Calculate timeout
	timeout := time.Now().Add(time.Duration(timeoutHours) * time.Hour)

	// Create approval request through runtime
	approvalID, resumeURL, err := runtime.CreateApprovalRequest(ctx, executor.ApprovalRequest{
		Title:       title,
		Description: description,
		Approvers:   toStringSlice(approvers),
		TimeoutAt:   timeout,
		NotifyVia:   toStringSlice(notifyVia),
		RequireAll:  requireAll,
		Data:        inputData,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create approval request: %w", err)
	}

	// Return waiting status - execution will be paused
	return types.JSON{
		"_wait_for_approval": true,
		"approval_id":        approvalID,
		"resume_url":         resumeURL,
		"title":              title,
		"description":        description,
		"approvers":          approvers,
		"timeout_at":         timeout.Format(time.RFC3339),
		"status":             "pending",
	}, nil
}

// Metadata returns node metadata
func (n *ApprovalNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.approval",
		Name:        "Wait for Approval",
		Description: "Pause workflow until human approval is received",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "approved", Type: "any"}, {Name: "rejected", Type: "any"}, {Name: "timeout", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "title", Type: "string", Description: "Approval request title", Required: true},
			{Name: "description", Type: "string", Description: "Description of what needs approval"},
			{Name: "approvers", Type: "array", Description: "List of approver emails or user IDs"},
			{Name: "timeout_hours", Type: "number", Description: "Hours before timeout", Default: 72},
			{Name: "notify_via", Type: "array", Description: "Notification channels (email, slack, etc.)"},
			{Name: "require_all_approvers", Type: "boolean", Description: "Require all approvers to approve", Default: false},
		},
	}
}

// WaitNode pauses execution for a specified time or condition
type WaitForEventNode struct{}

// NewWaitForEventNode creates a new wait for event node
func NewWaitForEventNode() *WaitForEventNode {
	return &WaitForEventNode{}
}

// Execute waits for an external event
func (n *WaitForEventNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	// Check if this is a resume with event data
	if eventData, ok := inputData["_event_data"].(map[string]interface{}); ok {
		return types.JSON{
			"event":      eventData,
			"resumed":    true,
			"resumed_at": time.Now().Format(time.RFC3339),
		}, nil
	}

	// First execution - set up event listener
	eventType, _ := params["event_type"].(string)
	eventFilter, _ := params["event_filter"].(map[string]interface{})
	timeoutHours, _ := params["timeout_hours"].(float64)

	if timeoutHours == 0 {
		timeoutHours = 168 // Default 1 week
	}

	timeout := time.Now().Add(time.Duration(timeoutHours) * time.Hour)

	// Register event listener through runtime
	listenerID, webhookURL, err := runtime.RegisterEventListener(ctx, executor.EventListener{
		EventType:  eventType,
		Filter:     eventFilter,
		TimeoutAt:  timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register event listener: %w", err)
	}

	return types.JSON{
		"_wait_for_event": true,
		"listener_id":     listenerID,
		"webhook_url":     webhookURL,
		"event_type":      eventType,
		"timeout_at":      timeout.Format(time.RFC3339),
		"status":          "waiting",
	}, nil
}

// Metadata returns node metadata
func (n *WaitForEventNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.wait_for_event",
		Name:        "Wait for Event",
		Description: "Pause workflow until an external event is received",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "event", Type: "any"}, {Name: "timeout", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "event_type", Type: "string", Description: "Type of event to wait for", Required: true},
			{Name: "event_filter", Type: "json", Description: "Filter conditions for the event"},
			{Name: "timeout_hours", Type: "number", Description: "Hours before timeout", Default: 168},
		},
	}
}

func toStringSlice(arr []interface{}) []string {
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
