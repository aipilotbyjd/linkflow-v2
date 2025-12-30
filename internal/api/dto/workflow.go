package dto

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/pkg/validator"
)

// Workflow requests

type CreateWorkflowRequest struct {
	Name        string           `json:"name" validate:"required,min=1,max=255"`
	Description *string          `json:"description,omitempty" validate:"omitempty,max=1000"`
	Nodes       models.JSONArray `json:"nodes" validate:"required"`
	Connections models.JSONArray `json:"connections" validate:"required"`
	Settings    models.JSON      `json:"settings,omitempty"`
	Tags        []string         `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
}

type UpdateWorkflowRequest struct {
	Name        *string          `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string          `json:"description,omitempty" validate:"omitempty,max=1000"`
	Nodes       models.JSONArray `json:"nodes,omitempty"`
	Connections models.JSONArray `json:"connections,omitempty"`
	Settings    models.JSON      `json:"settings,omitempty"`
	Tags        []string         `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
}

type ExecuteWorkflowRequest struct {
	InputData models.JSON `json:"input_data,omitempty"`
}

type CloneWorkflowRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

// Workflow responses

type WorkflowResponse struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    *string     `json:"description,omitempty"`
	Status         string      `json:"status"`
	Version        int         `json:"version"`
	Nodes          interface{} `json:"nodes,omitempty"`
	Connections    interface{} `json:"connections,omitempty"`
	Settings       interface{} `json:"settings,omitempty"`
	Tags           []string    `json:"tags,omitempty"`
	ExecutionCount int         `json:"execution_count"`
	LastExecutedAt *int64      `json:"last_executed_at,omitempty"`
	CreatedAt      int64       `json:"created_at"`
	UpdatedAt      int64       `json:"updated_at"`
}

type WorkflowVersionResponse struct {
	ID            string      `json:"id"`
	Version       int         `json:"version"`
	Nodes         interface{} `json:"nodes"`
	Connections   interface{} `json:"connections"`
	Settings      interface{} `json:"settings,omitempty"`
	ChangeMessage *string     `json:"change_message,omitempty"`
	CreatedAt     int64       `json:"created_at"`
}

// WorkflowValidationError represents a workflow-specific validation error
type WorkflowValidationError struct {
	Field   string `json:"field"`
	NodeID  string `json:"node_id,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WorkflowValidationErrorResponse returns a workflow validation error response
func WorkflowValidationErrorResponse(w http.ResponseWriter, errors []WorkflowValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	// Convert to validation error format
	details := make([]validator.ValidationError, len(errors))
	for i, e := range errors {
		field := e.Field
		if e.NodeID != "" {
			field = "node:" + e.NodeID + "." + e.Field
		}
		details[i] = validator.ValidationError{
			Field:   field,
			Message: e.Message,
		}
	}

	response := Response{
		Success:   false,
		RequestID: getRequestID(w),
		Timestamp: time.Now().Unix(),
		Error: &ErrorData{
			Code:    "WORKFLOW_VALIDATION_ERROR",
			Message: "Workflow validation failed",
			Details: details,
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}
