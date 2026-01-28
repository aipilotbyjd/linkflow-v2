package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/providers"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// StructuredNode extracts structured JSON data from text
type StructuredNode struct {
	factory *providers.Factory
}

// NewStructuredNode creates a new structured output node
func NewStructuredNode() *StructuredNode {
	return &StructuredNode{
		factory: providers.NewFactory(),
	}
}

func (n *StructuredNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	// Get credentials
	apiKey, _ := params["api_key"].(string)
	if apiKey == "" {
		if credRef, ok := params["credential"].(map[string]interface{}); ok {
			if key, ok := credRef["api_key"].(string); ok {
				apiKey = key
			}
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Get provider
	providerStr, _ := params["provider"].(string)
	provider := ai.ProviderOpenAI
	if providerStr != "" {
		if p, ok := ai.ParseProvider(providerStr); ok {
			provider = p
		}
	}

	// Get model
	model, _ := params["model"].(string)
	if model == "" {
		model = n.factory.GetDefaultModel(provider)
	}

	// Get input text
	text, _ := params["text"].(string)
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}

	// Get schema
	var schema map[string]interface{}
	if s, ok := params["schema"].(map[string]interface{}); ok {
		schema = s
	} else if schemaStr, ok := params["schema"].(string); ok {
		if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
			return nil, fmt.Errorf("invalid schema JSON: %w", err)
		}
	}

	if schema == nil {
		return nil, fmt.Errorf("schema is required")
	}

	// Build system message with extraction instructions
	schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
	systemMessage := fmt.Sprintf(`You are a data extraction assistant. Extract the requested information from the provided text and return it as valid JSON matching the following schema:

%s

Important:
- Return ONLY valid JSON, no explanations
- If a field cannot be extracted, use null
- Follow the schema types exactly`, string(schemaJSON))

	// Build messages
	messages := []ai.Message{
		ai.NewSystemMessage(systemMessage),
		ai.NewUserMessage(text),
	}

	// Build request with JSON mode
	req := &ai.ChatRequest{
		Messages: messages,
		Model:    model,
		ResponseFormat: &ai.ResponseFormat{
			Type: "json_object",
		},
	}

	if maxTokens, ok := params["max_tokens"].(float64); ok {
		req.MaxTokens = int(maxTokens)
	}

	// Use low temperature for structured extraction
	temp := 0.1
	req.Temperature = &temp

	// Create provider adapter
	config := &ai.ProviderConfig{
		Provider: provider,
		APIKey:   apiKey,
	}

	if orgID, ok := params["org_id"].(string); ok {
		config.OrgID = orgID
	}

	adapter, err := n.factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider adapter: %w", err)
	}

	// Execute request
	resp, err := adapter.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("extraction request failed: %w", err)
	}

	// Parse response JSON
	var extracted map[string]interface{}
	responseText := resp.GetText()
	if err := json.Unmarshal([]byte(responseText), &extracted); err != nil {
		return nil, fmt.Errorf("failed to parse extracted JSON: %w", err)
	}

	return types.JSON{
		"data":     extracted,
		"raw":      responseText,
		"model":    resp.Model,
		"provider": resp.Provider.String(),
		"usage": map[string]interface{}{
			"input_tokens":  resp.Usage.InputTokens,
			"output_tokens": resp.Usage.OutputTokens,
			"total_tokens":  resp.Usage.TotalTokens,
		},
		"cost_usd":   resp.CostUSD,
		"latency_ms": resp.LatencyMS,
	}, nil
}

func (n *StructuredNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.structured",
		Name:        "AI Structured Output",
		Description: "Extract structured JSON data from text using a schema",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "CodeSquare",
		Color:       "#14B8A6",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "provider",
				DisplayName: "Provider",
				Type:        "options",
				Required:    true,
				Default:     "openai",
				Options: []wtypes.ParamOption{
					{Name: "OpenAI", Value: "openai"},
					{Name: "Anthropic", Value: "anthropic"},
					{Name: "Google", Value: "google"},
				},
			},
			{
				Name:        "api_key",
				DisplayName: "API Key",
				Type:        "credential",
				Required:    true,
			},
			{
				Name:        "model",
				DisplayName: "Model",
				Type:        "string",
				Required:    false,
			},
			{
				Name:        "text",
				DisplayName: "Input Text",
				Type:        "string",
				Required:    true,
				Description: "Text to extract data from",
			},
			{
				Name:        "schema",
				DisplayName: "JSON Schema",
				Type:        "json",
				Required:    true,
				Description: "JSON schema defining the structure to extract",
			},
			{
				Name:        "max_tokens",
				DisplayName: "Max Tokens",
				Type:        "number",
				Required:    false,
				Default:     2048,
			},
		},
		Credentials: []string{"openai", "anthropic", "google_ai"},
	}
}
