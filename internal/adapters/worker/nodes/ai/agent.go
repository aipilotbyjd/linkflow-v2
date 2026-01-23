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

// AgentNode is an autonomous AI agent that can use tools iteratively
type AgentNode struct {
	factory *providers.Factory
}

// NewAgentNode creates a new AI agent node
func NewAgentNode() *AgentNode {
	return &AgentNode{
		factory: providers.NewFactory(),
	}
}

func (n *AgentNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
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

	// Get model (must support tools)
	model, _ := params["model"].(string)
	if model == "" {
		switch provider {
		case ai.ProviderOpenAI:
			model = "gpt-4o"
		case ai.ProviderAnthropic:
			model = "claude-3-5-sonnet-20241022"
		default:
			model = "gpt-4o"
		}
	}

	// Get task
	task, _ := params["task"].(string)
	if task == "" {
		return nil, fmt.Errorf("task is required")
	}

	// Get tools
	var tools []ai.Tool
	if toolsData, ok := params["tools"].([]interface{}); ok {
		for _, t := range toolsData {
			if toolMap, ok := t.(map[string]interface{}); ok {
				name, _ := toolMap["name"].(string)
				desc, _ := toolMap["description"].(string)
				parameters, _ := toolMap["parameters"].(map[string]interface{})
				tools = append(tools, ai.NewTool(name, desc, parameters))
			}
		}
	}

	if len(tools) == 0 {
		return nil, fmt.Errorf("at least one tool is required for agent")
	}

	// Get max iterations
	maxIterations := 10
	if mi, ok := params["max_iterations"].(float64); ok && mi > 0 {
		maxIterations = int(mi)
	}

	// Tool results handler (placeholder - in real implementation, this would be a callback)
	toolResultsHandler, _ := params["tool_handler"].(func(string, map[string]interface{}) (string, error))

	// Build system message
	systemMessage, _ := params["system_message"].(string)
	if systemMessage == "" {
		systemMessage = `You are an autonomous AI agent that can use tools to accomplish tasks. 
When you need to perform an action, use the appropriate tool.
Think step by step and use tools as needed to complete the task.
When the task is complete, provide a final summary.`
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

	// Initialize conversation
	messages := []ai.Message{
		ai.NewSystemMessage(systemMessage),
		ai.NewUserMessage(task),
	}

	// Track iterations and tool calls
	iterations := 0
	var allToolCalls []map[string]interface{}
	var totalUsage ai.Usage
	var totalCost float64

	// Agent loop
	for iterations < maxIterations {
		iterations++

		// Build request
		req := &ai.ChatRequest{
			Messages:   messages,
			Model:      model,
			Tools:      tools,
			ToolChoice: "auto",
		}

		if maxTokens, ok := params["max_tokens"].(float64); ok {
			req.MaxTokens = int(maxTokens)
		}

		// Execute request
		resp, err := adapter.Chat(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("agent request failed at iteration %d: %w", iterations, err)
		}

		// Update usage
		totalUsage.InputTokens += resp.Usage.InputTokens
		totalUsage.OutputTokens += resp.Usage.OutputTokens
		totalUsage.TotalTokens += resp.Usage.TotalTokens
		totalCost += resp.CostUSD

		// Add assistant message to history
		messages = append(messages, resp.Message)

		// Check if we have tool calls
		if !resp.HasToolCalls() {
			// No tool calls - agent is done
			return types.JSON{
				"response":   resp.GetText(),
				"iterations": iterations,
				"tool_calls": allToolCalls,
				"model":      resp.Model,
				"provider":   resp.Provider.String(),
				"usage": map[string]interface{}{
					"input_tokens":  totalUsage.InputTokens,
					"output_tokens": totalUsage.OutputTokens,
					"total_tokens":  totalUsage.TotalTokens,
				},
				"cost_usd": totalCost,
				"status":   "completed",
			}, nil
		}

		// Process tool calls
		for _, tc := range resp.Message.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				args = map[string]interface{}{"raw": tc.Function.Arguments}
			}

			toolCallInfo := map[string]interface{}{
				"id":        tc.ID,
				"name":      tc.Function.Name,
				"arguments": args,
				"iteration": iterations,
			}

			// Execute tool (if handler provided)
			var toolResult string
			if toolResultsHandler != nil {
				result, err := toolResultsHandler(tc.Function.Name, args)
				if err != nil {
					toolResult = fmt.Sprintf("Error: %s", err.Error())
				} else {
					toolResult = result
				}
			} else {
				// Placeholder result - in real implementation, tools would be executed
				toolResult = fmt.Sprintf("Tool '%s' was called with arguments: %v. (Tool execution not implemented - provide tool_handler)", tc.Function.Name, args)
			}

			toolCallInfo["result"] = toolResult
			allToolCalls = append(allToolCalls, toolCallInfo)

			// Add tool result to messages
			messages = append(messages, ai.NewToolResultMessage(tc.ID, toolResult))
		}
	}

	// Max iterations reached
	return types.JSON{
		"response":   "Agent reached maximum iterations without completing the task.",
		"iterations": iterations,
		"tool_calls": allToolCalls,
		"model":      model,
		"provider":   provider.String(),
		"usage": map[string]interface{}{
			"input_tokens":  totalUsage.InputTokens,
			"output_tokens": totalUsage.OutputTokens,
			"total_tokens":  totalUsage.TotalTokens,
		},
		"cost_usd": totalCost,
		"status":   "max_iterations_reached",
	}, nil
}

func (n *AgentNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.agent",
		Name:        "AI Agent",
		Description: "Autonomous AI agent that uses tools to accomplish tasks",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "bot",
		Color:       "#6366F1",
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
					{Name: "GPT-4o (OpenAI)", Value: "gpt-4o"},
					{Name: "GPT-4o Mini (OpenAI)", Value: "gpt-4o-mini"},
					{Name: "Claude 3.5 Sonnet (Anthropic)", Value: "claude-3-5-sonnet-20241022"},
				},
			},
			{
				Name:        "task",
				DisplayName: "Task",
				Type:        "string",
				Required:    true,
				Description: "The task for the agent to accomplish",
			},
			{
				Name:        "tools",
				DisplayName: "Tools",
				Type:        "json",
				Required:    true,
				Description: "Array of tool definitions the agent can use",
			},
			{
				Name:        "system_message",
				DisplayName: "System Message",
				Type:        "string",
				Required:    false,
				Description: "Custom instructions for the agent",
			},
			{
				Name:        "max_iterations",
				DisplayName: "Max Iterations",
				Type:        "number",
				Required:    false,
				Default:     10,
				Description: "Maximum number of tool-use iterations",
			},
			{
				Name:        "max_tokens",
				DisplayName: "Max Tokens",
				Type:        "number",
				Required:    false,
				Default:     4096,
			},
		},
		Credentials: []string{"openai", "anthropic"},
	}
}
