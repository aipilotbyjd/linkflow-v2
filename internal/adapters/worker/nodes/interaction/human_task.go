package interaction

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// HumanTaskNode pauses workflow execution until a user completes a form
type HumanTaskNode struct{}

// NewHumanTaskNode creates a new human task node
func NewHumanTaskNode() *HumanTaskNode {
	return &HumanTaskNode{}
}

// Execute pauses for input and returns the input data
func (n *HumanTaskNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	// Check if this is a resume with response data
	if taskResponse, ok := inputData["_human_task_response"].(map[string]interface{}); ok {
		completedBy, _ := taskResponse["completed_by"].(string)
		completedAt, _ := taskResponse["completed_at"].(string)
		formData, _ := taskResponse["form_data"].(map[string]interface{})

		return types.JSON{
			"completed_by": completedBy,
			"completed_at": completedAt,
			"data":         formData,
			"input_data":   inputData["data"], // Preserve original input if needed
		}, nil
	}

	// First execution - create task request
	title, _ := params["title"].(string)
	description, _ := params["description"].(string)
	assignees, _ := params["assignees"].([]interface{})
	timeoutHours, _ := params["timeout_hours"].(float64)
	notifyVia, _ := params["notify_via"].([]interface{})
	
	// Form configuration
	var formFields []map[string]interface{}
	if fields, ok := params["form_fields"].([]interface{}); ok {
		for _, f := range fields {
			if fieldMap, ok := f.(map[string]interface{}); ok {
				formFields = append(formFields, fieldMap)
			}
		}
	}

	if title == "" {
		title = "Action Required"
	}
	if timeoutHours == 0 {
		timeoutHours = 72 // Default 3 days
	}

	// Calculate timeout
	timeout := time.Now().Add(time.Duration(timeoutHours) * time.Hour)

	// In a real implementation, we would call a service to create the task record
	// For now, we simulate the request creation via runtime or just return the wait signal
	
	// Generate a unique ID for this task (in reality, the backend would do this)
	// We'll rely on the runtime to handle the "wait" state and persistence
	
	return types.JSON{
		"_wait_for_human_task": true,
		"task_config": map[string]interface{}{
			"title":       title,
			"description": description,
			"assignees":   assignees,
			"form_fields": formFields,
			"timeout_at":  timeout.Format(time.RFC3339),
			"notify_via":  notifyVia,
			"data_context": inputData,
		},
		"status": "waiting",
	}, nil
}

// Metadata returns node metadata
func (n *HumanTaskNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "interaction.human_task",
		Name:        "Human Task",
		Description: "Pause workflow and request human input via a form",
		Category:    "interaction",
		Version:     "1.0.0",
		Icon:        "user-check", 
		Color:       "#EC4899", // Pink/Magenta
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}, {Name: "timeout", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "title",
				DisplayName: "Task Title",
				Type:        "string",
				Required:    true,
				Description: "Title of the task for the user",
			},
			{
				Name:        "description",
				DisplayName: "Instructions",
				Type:        "string",
				Required:    true,
				Description: "Detailed instructions for the user",
			},
			{
				Name:        "assignees",
				DisplayName: "Assignees",
				Type:        "string", // In UI this would be a multi-select user/email list
				Description: "Users or emails to assign this task to",
			},
			{
				Name:        "form_fields",
				DisplayName: "Form Fields",
				Type:        "json", // In UI this would be a form builder
				Required:    true,
				Description: "Array of fields to collect (name, type, label, required)",
				Default: `[
					{
						"name": "approval", 
						"label": "Approve?", 
						"type": "boolean", 
						"required": true
					},
					{
						"name": "comments", 
						"label": "Comments", 
						"type": "textarea", 
						"required": false
					}
				]`,
			},
			{
				Name:        "timeout_hours",
				DisplayName: "Timeout (Hours)",
				Type:        "number",
				Default:     72,
			},
			{
				Name:        "notify_via",
				DisplayName: "Notify Via",
				Type:        "multi-select",
				Options: []wtypes.ParamOption{
					{Name: "Email", Value: "email"},
					{Name: "Slack", Value: "slack"},
					{Name: "In-App", Value: "in_app"},
				},
				Default: []string{"email", "in_app"},
			},
		},
	}
}
