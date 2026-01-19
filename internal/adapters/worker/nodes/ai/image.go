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

// ImageNode generates images from text prompts
type ImageNode struct {
	factory *providers.Factory
}

// NewImageNode creates a new image generation node
func NewImageNode() *ImageNode {
	return &ImageNode{
		factory: providers.NewFactory(),
	}
}

func (n *ImageNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
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

	// Currently only OpenAI supports image generation
	provider := ai.ProviderOpenAI

	// Get prompt
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	// Get model
	model, _ := params["model"].(string)
	if model == "" {
		model = "dall-e-3"
	}

	// Build request
	req := &ai.ImageRequest{
		Prompt: prompt,
		Model:  model,
	}

	// Optional parameters
	if n, ok := params["n"].(float64); ok {
		req.N = int(n)
	} else {
		req.N = 1
	}

	if size, ok := params["size"].(string); ok {
		req.Size = size
	} else {
		req.Size = "1024x1024"
	}

	if quality, ok := params["quality"].(string); ok {
		req.Quality = quality
	}

	if style, ok := params["style"].(string); ok {
		req.Style = style
	}

	// Response format
	returnBase64, _ := params["return_base64"].(bool)
	if returnBase64 {
		req.ResponseFormat = "b64_json"
	} else {
		req.ResponseFormat = "url"
	}

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
	resp, err := adapter.GenerateImage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("image generation failed: %w", err)
	}

	// Build response
	images := make([]map[string]interface{}, len(resp.Images))
	for i, img := range resp.Images {
		images[i] = map[string]interface{}{}
		if img.URL != "" {
			images[i]["url"] = img.URL
		}
		if img.Base64 != "" {
			images[i]["base64"] = img.Base64
		}
		if img.RevisedPrompt != "" {
			images[i]["revised_prompt"] = img.RevisedPrompt
		}
	}

	// Return single image or array
	var imageResult interface{}
	if len(images) == 1 {
		imageResult = images[0]
	} else {
		imageResult = images
	}

	return types.JSON{
		"images":     imageResult,
		"model":      resp.Model,
		"provider":   resp.Provider.String(),
		"count":      len(resp.Images),
		"cost_usd":   resp.CostUSD,
		"latency_ms": resp.LatencyMS,
	}, nil
}

func (n *ImageNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.image",
		Name:        "AI Image Generate",
		Description: "Generate images from text prompts using DALL-E",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "image",
		Color:       "#06B6D4",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "api_key",
				DisplayName: "OpenAI API Key",
				Type:        "credential",
				Required:    true,
			},
			{
				Name:        "model",
				DisplayName: "Model",
				Type:        "options",
				Required:    false,
				Default:     "dall-e-3",
				Options: []wtypes.ParamOption{
					{Name: "DALL-E 3", Value: "dall-e-3"},
					{Name: "DALL-E 2", Value: "dall-e-2"},
				},
			},
			{
				Name:        "prompt",
				DisplayName: "Prompt",
				Type:        "string",
				Required:    true,
				Description: "Description of the image to generate",
			},
			{
				Name:        "size",
				DisplayName: "Size",
				Type:        "options",
				Required:    false,
				Default:     "1024x1024",
				Options: []wtypes.ParamOption{
					{Name: "1024x1024", Value: "1024x1024"},
					{Name: "1024x1792 (Portrait)", Value: "1024x1792"},
					{Name: "1792x1024 (Landscape)", Value: "1792x1024"},
					{Name: "512x512 (DALL-E 2)", Value: "512x512"},
					{Name: "256x256 (DALL-E 2)", Value: "256x256"},
				},
			},
			{
				Name:        "quality",
				DisplayName: "Quality",
				Type:        "options",
				Required:    false,
				Default:     "standard",
				Options: []wtypes.ParamOption{
					{Name: "Standard", Value: "standard"},
					{Name: "HD", Value: "hd"},
				},
			},
			{
				Name:        "style",
				DisplayName: "Style",
				Type:        "options",
				Required:    false,
				Default:     "vivid",
				Options: []wtypes.ParamOption{
					{Name: "Vivid", Value: "vivid"},
					{Name: "Natural", Value: "natural"},
				},
			},
			{
				Name:        "n",
				DisplayName: "Number of Images",
				Type:        "number",
				Required:    false,
				Default:     1,
				Description: "Number of images to generate (1-10, DALL-E 2 only for n>1)",
			},
			{
				Name:        "return_base64",
				DisplayName: "Return Base64",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Return base64-encoded image instead of URL",
			},
		},
		Credentials: []string{"openai"},
	}
}
