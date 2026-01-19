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

// EvaluateNode uses LLM-as-a-judge to evaluate text quality
type EvaluateNode struct {
	factory *providers.Factory
}

// NewEvaluateNode creates a new evaluate node
func NewEvaluateNode() *EvaluateNode {
	return &EvaluateNode{
		factory: providers.NewFactory(),
	}
}

func (n *EvaluateNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
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

	// Get content to evaluate
	content, _ := params["content"].(string)
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Get evaluation type
	evalType, _ := params["evaluation_type"].(string)
	if evalType == "" {
		evalType = "quality"
	}

	// Get criteria
	criteria, _ := params["criteria"].(string)

	// Get reference (for comparison evaluations)
	reference, _ := params["reference"].(string)

	// Build evaluation prompt
	var systemMessage string
	var userMessage string

	switch evalType {
	case "quality":
		systemMessage = `You are an expert content evaluator. Evaluate the given content on these dimensions:
1. Clarity (1-10): How clear and understandable is the content?
2. Accuracy (1-10): How factually accurate does the content appear?
3. Completeness (1-10): How thoroughly does it cover the topic?
4. Coherence (1-10): How well does it flow and make sense?
5. Overall (1-10): Overall quality score

Return your evaluation as JSON with the format:
{
  "scores": {
    "clarity": <number>,
    "accuracy": <number>,
    "completeness": <number>,
    "coherence": <number>,
    "overall": <number>
  },
  "feedback": "<brief explanation>",
  "strengths": ["<strength1>", "<strength2>"],
  "improvements": ["<improvement1>", "<improvement2>"]
}`
		userMessage = fmt.Sprintf("Evaluate this content:\n\n%s", content)

	case "comparison":
		if reference == "" {
			return nil, fmt.Errorf("reference is required for comparison evaluation")
		}
		systemMessage = `You are an expert content evaluator. Compare two pieces of content and determine which is better.

Return your evaluation as JSON with the format:
{
  "winner": "A" or "B",
  "confidence": <0.0-1.0>,
  "reasoning": "<explanation>",
  "a_strengths": ["<strength1>"],
  "b_strengths": ["<strength1>"],
  "recommendation": "<which to use and why>"
}`
		userMessage = fmt.Sprintf("Compare these two pieces of content:\n\nContent A:\n%s\n\nContent B:\n%s", content, reference)

	case "factuality":
		systemMessage = `You are a fact-checking expert. Evaluate the factual accuracy of the given content.

Return your evaluation as JSON with the format:
{
  "factuality_score": <1-10>,
  "claims": [
    {"claim": "<claim>", "assessment": "verified|unverified|false", "confidence": <0.0-1.0>}
  ],
  "overall_assessment": "<summary>",
  "concerns": ["<concern1>"]
}`
		userMessage = fmt.Sprintf("Evaluate the factual accuracy of:\n\n%s", content)

	case "custom":
		if criteria == "" {
			return nil, fmt.Errorf("criteria is required for custom evaluation")
		}
		systemMessage = fmt.Sprintf(`You are an expert evaluator. Evaluate content based on these criteria:

%s

Return your evaluation as JSON with scores and explanations.`, criteria)
		userMessage = fmt.Sprintf("Evaluate this content:\n\n%s", content)

	default:
		return nil, fmt.Errorf("unsupported evaluation type: %s", evalType)
	}

	// Build messages
	messages := []ai.Message{
		ai.NewSystemMessage(systemMessage),
		ai.NewUserMessage(userMessage),
	}

	// Build request
	req := &ai.ChatRequest{
		Messages: messages,
		Model:    model,
		ResponseFormat: &ai.ResponseFormat{
			Type: "json_object",
		},
	}

	// Use low temperature for evaluation
	temp := 0.2
	req.Temperature = &temp

	if maxTokens, ok := params["max_tokens"].(float64); ok {
		req.MaxTokens = int(maxTokens)
	}

	// Create provider adapter
	config := &ai.ProviderConfig{
		Provider: provider,
		APIKey:   apiKey,
	}

	adapter, err := n.factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider adapter: %w", err)
	}

	// Execute request
	resp, err := adapter.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("evaluation request failed: %w", err)
	}

	// Parse evaluation result
	var evaluation map[string]interface{}
	responseText := resp.GetText()
	if err := json.Unmarshal([]byte(responseText), &evaluation); err != nil {
		// Return raw if JSON parsing fails
		evaluation = map[string]interface{}{
			"raw": responseText,
		}
	}

	return types.JSON{
		"evaluation":      evaluation,
		"evaluation_type": evalType,
		"model":           resp.Model,
		"provider":        resp.Provider.String(),
		"usage": map[string]interface{}{
			"input_tokens":  resp.Usage.InputTokens,
			"output_tokens": resp.Usage.OutputTokens,
			"total_tokens":  resp.Usage.TotalTokens,
		},
		"cost_usd":   resp.CostUSD,
		"latency_ms": resp.LatencyMS,
	}, nil
}

func (n *EvaluateNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.evaluate",
		Name:        "AI Evaluate",
		Description: "Use LLM-as-a-judge to evaluate content quality, accuracy, or custom criteria",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "check-circle",
		Color:       "#22C55E",
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
				Name:        "content",
				DisplayName: "Content",
				Type:        "string",
				Required:    true,
				Description: "Content to evaluate",
			},
			{
				Name:        "evaluation_type",
				DisplayName: "Evaluation Type",
				Type:        "options",
				Required:    true,
				Default:     "quality",
				Options: []wtypes.ParamOption{
					{Name: "Quality (Clarity, Accuracy, etc.)", Value: "quality"},
					{Name: "Comparison (A vs B)", Value: "comparison"},
					{Name: "Factuality Check", Value: "factuality"},
					{Name: "Custom Criteria", Value: "custom"},
				},
			},
			{
				Name:        "reference",
				DisplayName: "Reference Content",
				Type:        "string",
				Required:    false,
				Description: "Reference content for comparison",
				ShowIf:      "evaluation_type=comparison",
			},
			{
				Name:        "criteria",
				DisplayName: "Custom Criteria",
				Type:        "string",
				Required:    false,
				Description: "Custom evaluation criteria",
				ShowIf:      "evaluation_type=custom",
			},
			{
				Name:        "max_tokens",
				DisplayName: "Max Tokens",
				Type:        "number",
				Required:    false,
				Default:     1024,
			},
		},
		Credentials: []string{"openai", "anthropic", "google_ai"},
	}
}
