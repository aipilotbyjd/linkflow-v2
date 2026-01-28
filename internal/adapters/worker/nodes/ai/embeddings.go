package ai

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/providers"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// EmbeddingsNode generates vector embeddings for text
type EmbeddingsNode struct {
	factory *providers.Factory
}

// NewEmbeddingsNode creates a new embeddings node
func NewEmbeddingsNode() *EmbeddingsNode {
	return &EmbeddingsNode{
		factory: providers.NewFactory(),
	}
}

func (n *EmbeddingsNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
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

	// Get provider (OpenAI or Google support embeddings)
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
		switch provider {
		case ai.ProviderOpenAI:
			model = "text-embedding-3-small"
		case ai.ProviderGoogle:
			model = "text-embedding-004"
		default:
			model = "text-embedding-3-small"
		}
	}

	// Get input text(s)
	var inputs []string

	if text, ok := params["text"].(string); ok && text != "" {
		inputs = []string{text}
	} else if texts, ok := params["texts"].([]interface{}); ok {
		for _, t := range texts {
			if s, ok := t.(string); ok {
				inputs = append(inputs, s)
			}
		}
	}

	if len(inputs) == 0 {
		return nil, fmt.Errorf("at least one text input is required")
	}

	// Build request
	req := &ai.EmbeddingRequest{
		Input: inputs,
		Model: model,
	}

	// Optional dimensions
	if dimensions, ok := params["dimensions"].(float64); ok && dimensions > 0 {
		req.Dimensions = int(dimensions)
	}

	// Create provider adapter
	config := &ai.ProviderConfig{
		Provider: provider,
		APIKey:   apiKey,
	}

	if baseURL, ok := params["base_url"].(string); ok {
		config.BaseURL = baseURL
	}

	adapter, err := n.factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider adapter: %w", err)
	}

	// Execute request
	resp, err := adapter.Embed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}

	// Return single embedding or array
	var embeddings interface{}
	if len(resp.Embeddings) == 1 {
		embeddings = resp.Embeddings[0]
	} else {
		embeddings = resp.Embeddings
	}

	return types.JSON{
		"embeddings": embeddings,
		"model":      resp.Model,
		"provider":   resp.Provider.String(),
		"dimensions": resp.Dimensions,
		"count":      len(resp.Embeddings),
		"usage": map[string]interface{}{
			"total_tokens": resp.Usage.TotalTokens,
		},
		"cost_usd":   resp.CostUSD,
		"latency_ms": resp.LatencyMS,
	}, nil
}

func (n *EmbeddingsNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.embeddings",
		Name:        "AI Embeddings",
		Description: "Generate vector embeddings for text (for semantic search, RAG, etc.)",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "Database01",
		Color:       "#F59E0B",
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
				Type:        "options",
				Required:    false,
				Options: []wtypes.ParamOption{
					{Name: "text-embedding-3-small (OpenAI)", Value: "text-embedding-3-small"},
					{Name: "text-embedding-3-large (OpenAI)", Value: "text-embedding-3-large"},
					{Name: "text-embedding-004 (Google)", Value: "text-embedding-004"},
				},
			},
			{
				Name:        "text",
				DisplayName: "Text",
				Type:        "string",
				Required:    false,
				Description: "Single text to embed",
			},
			{
				Name:        "texts",
				DisplayName: "Texts",
				Type:        "json",
				Required:    false,
				Description: "Array of texts to embed",
			},
			{
				Name:        "dimensions",
				DisplayName: "Dimensions",
				Type:        "number",
				Required:    false,
				Description: "Output dimensions (OpenAI v3 models only)",
			},
		},
		Credentials: []string{"openai", "google_ai"},
	}
}
