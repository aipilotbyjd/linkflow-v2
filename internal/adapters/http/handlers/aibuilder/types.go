package aibuilder

import (
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	appbuilder "github.com/linkflow-ai/linkflow/internal/core/application/aibuilder"
)

// GenerateRequest represents a workflow generation request
type GenerateRequest struct {
	Prompt           string            `json:"prompt" validate:"required,min=10,max=2000"`
	PreferredTrigger *string           `json:"preferred_trigger,omitempty"`
	AvailableCreds   []string          `json:"available_credentials,omitempty"`
	Constraints      map[string]string `json:"constraints,omitempty"`
}

// GenerateResponse represents a workflow generation response
type GenerateResponse struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Nodes       []map[string]interface{} `json:"nodes"`
	Connections []map[string]interface{} `json:"connections"`
	Settings    map[string]interface{}   `json:"settings,omitempty"`
	Explanation string                   `json:"explanation"`
	Suggestions []string                 `json:"suggestions,omitempty"`
}

// SuggestRequest represents a workflow improvement suggestion request
type SuggestRequest struct {
	WorkflowJSON string `json:"workflow_json" validate:"required"`
}

// SuggestResponse represents improvement suggestions
type SuggestResponse struct {
	Suggestions []string `json:"suggestions"`
}

// ExplainRequest represents a workflow explanation request
type ExplainRequest struct {
	WorkflowJSON string `json:"workflow_json" validate:"required"`
}

// ExplainResponse represents workflow explanation
type ExplainResponse struct {
	Explanation string `json:"explanation"`
}

// Service interface for dependency injection
type Service interface {
	GenerateWorkflow(ctx interface{}, workspaceID, userID interface{}, prompt string, genCtx interface{}) (interface{}, error)
	SuggestImprovements(ctx interface{}, workflowJSON string) ([]string, error)
	ExplainWorkflow(ctx interface{}, workflowJSON string) (string, error)
}

// AIBuilderService is the actual service type
type AIBuilderService = appbuilder.Service

// AIProvider is the AI provider interface
type AIProvider = ai.ProviderAdapter
