package aibuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/core/domain/aibuilder"
)

// Service handles AI workflow generation
type Service struct {
	aiProvider    ai.ProviderAdapter
	nodeTemplates []aibuilder.NodeTemplate
}

// NewService creates a new AI builder service
func NewService(aiProvider ai.ProviderAdapter) *Service {
	return &Service{
		aiProvider:    aiProvider,
		nodeTemplates: defaultNodeTemplates(),
	}
}

// GenerateWorkflow generates a workflow from natural language
func (s *Service) GenerateWorkflow(ctx context.Context, workspaceID, userID uuid.UUID, prompt string, genCtx *aibuilder.Context) (*aibuilder.Result, error) {
	systemPrompt := s.buildSystemPrompt()
	userPrompt := s.buildUserPrompt(prompt, genCtx)

	temp := 0.7
	req := &ai.ChatRequest{
		Model: "gpt-4o",
		Messages: []ai.Message{
			ai.NewSystemMessage(systemPrompt),
			ai.NewUserMessage(userPrompt),
		},
		Temperature:    &temp,
		MaxTokens:      4096,
		ResponseFormat: &ai.ResponseFormat{Type: "json_object"},
	}

	resp, err := s.aiProvider.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	responseText := resp.GetText()
	if responseText == "" {
		return nil, fmt.Errorf("no response from AI")
	}

	var result aibuilder.Result
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	s.assignNodeIDs(&result)
	s.validateAndFixConnections(&result)

	return &result, nil
}

// SuggestImprovements suggests improvements for an existing workflow
func (s *Service) SuggestImprovements(ctx context.Context, workflowJSON string) ([]string, error) {
	systemPrompt := `You are a workflow automation expert. Analyze the workflow and suggest improvements.
Return a JSON object with a "suggestions" array of strings.`

	temp := 0.5
	req := &ai.ChatRequest{
		Model: "gpt-4o",
		Messages: []ai.Message{
			ai.NewSystemMessage(systemPrompt),
			ai.NewUserMessage(fmt.Sprintf("Analyze this workflow and suggest improvements:\n%s", workflowJSON)),
		},
		Temperature:    &temp,
		MaxTokens:      1024,
		ResponseFormat: &ai.ResponseFormat{Type: "json_object"},
	}

	resp, err := s.aiProvider.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	var result struct {
		Suggestions []string `json:"suggestions"`
	}
	if err := json.Unmarshal([]byte(resp.GetText()), &result); err != nil {
		return nil, err
	}

	return result.Suggestions, nil
}

// ExplainWorkflow generates a natural language explanation of a workflow
func (s *Service) ExplainWorkflow(ctx context.Context, workflowJSON string) (string, error) {
	systemPrompt := `You are a workflow automation expert. Explain what the workflow does in simple terms.
Return a JSON object with an "explanation" string field.`

	temp := 0.3
	req := &ai.ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []ai.Message{
			ai.NewSystemMessage(systemPrompt),
			ai.NewUserMessage(fmt.Sprintf("Explain this workflow:\n%s", workflowJSON)),
		},
		Temperature:    &temp,
		MaxTokens:      512,
		ResponseFormat: &ai.ResponseFormat{Type: "json_object"},
	}

	resp, err := s.aiProvider.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	var result struct {
		Explanation string `json:"explanation"`
	}
	if err := json.Unmarshal([]byte(resp.GetText()), &result); err != nil {
		return "", err
	}

	return result.Explanation, nil
}

func (s *Service) buildSystemPrompt() string {
	nodeTypesJSON, _ := json.Marshal(s.nodeTemplates)

	return fmt.Sprintf(`You are a workflow automation expert that generates workflow definitions.

AVAILABLE NODE TYPES:
%s

RESPONSE FORMAT (JSON):
{
  "name": "Workflow name",
  "description": "What the workflow does",
  "nodes": [
    {
      "id": "node_1",
      "type": "trigger.webhook",
      "name": "Webhook Trigger",
      "position": {"x": 0, "y": 0},
      "parameters": {}
    }
  ],
  "connections": [
    {
      "source": "node_1",
      "sourceOutput": "main",
      "target": "node_2",
      "targetInput": "main"
    }
  ],
  "settings": {},
  "explanation": "Step by step explanation",
  "suggestions": ["Optional suggestions for improvements"]
}

RULES:
1. Every workflow MUST start with a trigger node (trigger.manual, trigger.webhook, or trigger.schedule)
2. All nodes must be connected
3. Use appropriate node types for each task
4. Position nodes in a logical left-to-right flow (increment x by 250 for each step)
5. Include helpful parameter values based on the user's request
6. Be creative but practical`, nodeTypesJSON)
}

func (s *Service) buildUserPrompt(prompt string, ctx *aibuilder.Context) string {
	var sb strings.Builder
	sb.WriteString("Create a workflow for: ")
	sb.WriteString(prompt)

	if ctx != nil {
		if ctx.PreferredTrigger != nil {
			sb.WriteString(fmt.Sprintf("\nPreferred trigger: %s", *ctx.PreferredTrigger))
		}
		if len(ctx.AvailableCredTypes) > 0 {
			sb.WriteString(fmt.Sprintf("\nAvailable credentials: %s", strings.Join(ctx.AvailableCredTypes, ", ")))
		}
		for k, v := range ctx.Constraints {
			sb.WriteString(fmt.Sprintf("\n%s: %s", k, v))
		}
	}

	return sb.String()
}

func (s *Service) assignNodeIDs(result *aibuilder.Result) {
	for i := range result.Nodes {
		if result.Nodes[i]["id"] == nil || result.Nodes[i]["id"] == "" {
			result.Nodes[i]["id"] = fmt.Sprintf("node_%d", i+1)
		}
	}
}

func (s *Service) validateAndFixConnections(result *aibuilder.Result) {
	nodeIDs := make(map[string]bool)
	for _, node := range result.Nodes {
		if id, ok := node["id"].(string); ok {
			nodeIDs[id] = true
		}
	}

	validConnections := make([]map[string]interface{}, 0)
	for _, conn := range result.Connections {
		source, _ := conn["source"].(string)
		target, _ := conn["target"].(string)
		if nodeIDs[source] && nodeIDs[target] {
			if conn["sourceOutput"] == nil {
				conn["sourceOutput"] = "main"
			}
			if conn["targetInput"] == nil {
				conn["targetInput"] = "main"
			}
			validConnections = append(validConnections, conn)
		}
	}
	result.Connections = validConnections
}

func defaultNodeTemplates() []aibuilder.NodeTemplate {
	return []aibuilder.NodeTemplate{
		{Type: "trigger.manual", Name: "Manual Trigger", Description: "Start workflow manually", Category: "trigger", Keywords: []string{"manual", "start", "run"}},
		{Type: "trigger.webhook", Name: "Webhook Trigger", Description: "Start workflow from HTTP webhook", Category: "trigger", Keywords: []string{"webhook", "http", "api", "post"}},
		{Type: "trigger.schedule", Name: "Schedule Trigger", Description: "Start workflow on a schedule", Category: "trigger", Keywords: []string{"schedule", "cron", "timer", "daily", "hourly"}},
		{Type: "logic.if", Name: "If", Description: "Conditional branching", Category: "logic", Keywords: []string{"if", "condition", "branch", "check"}},
		{Type: "logic.switch", Name: "Switch", Description: "Multiple condition routing", Category: "logic", Keywords: []string{"switch", "route", "multiple"}},
		{Type: "logic.loop", Name: "Loop", Description: "Iterate over items", Category: "logic", Keywords: []string{"loop", "iterate", "foreach", "each"}},
		{Type: "logic.filter", Name: "Filter", Description: "Filter items by condition", Category: "logic", Keywords: []string{"filter", "where", "select"}},
		{Type: "logic.merge", Name: "Merge", Description: "Merge multiple branches", Category: "logic", Keywords: []string{"merge", "combine", "join"}},
		{Type: "logic.wait", Name: "Wait", Description: "Wait for a duration", Category: "logic", Keywords: []string{"wait", "delay", "sleep", "pause"}},
		{Type: "action.http", Name: "HTTP Request", Description: "Make HTTP API calls", Category: "action", Keywords: []string{"http", "api", "request", "fetch", "get", "post"}},
		{Type: "action.set", Name: "Set", Description: "Set or transform data", Category: "action", Keywords: []string{"set", "transform", "map", "data"}},
		{Type: "action.email", Name: "Send Email", Description: "Send an email", Category: "action", Keywords: []string{"email", "send", "mail", "notify"}},
		{Type: "action.code", Name: "Code", Description: "Run JavaScript code", Category: "action", Keywords: []string{"code", "javascript", "script", "custom"}},
		{Type: "integration.slack", Name: "Slack", Description: "Send Slack messages", Category: "integration", Keywords: []string{"slack", "message", "channel", "notify"}},
		{Type: "integration.discord", Name: "Discord", Description: "Send Discord messages", Category: "integration", Keywords: []string{"discord", "message", "webhook"}},
		{Type: "integration.telegram", Name: "Telegram", Description: "Send Telegram messages", Category: "integration", Keywords: []string{"telegram", "bot", "message"}},
		{Type: "integration.twilio", Name: "Twilio", Description: "Send SMS via Twilio", Category: "integration", Keywords: []string{"twilio", "sms", "text", "phone"}},
		{Type: "integration.postgres", Name: "PostgreSQL", Description: "Query PostgreSQL database", Category: "integration", Keywords: []string{"postgres", "database", "sql", "query"}},
		{Type: "integration.mysql", Name: "MySQL", Description: "Query MySQL database", Category: "integration", Keywords: []string{"mysql", "database", "sql", "query"}},
		{Type: "integration.mongodb", Name: "MongoDB", Description: "Query MongoDB database", Category: "integration", Keywords: []string{"mongodb", "nosql", "document"}},
		{Type: "integration.redis", Name: "Redis", Description: "Redis operations", Category: "integration", Keywords: []string{"redis", "cache", "key-value"}},
		{Type: "integration.aws_s3", Name: "AWS S3", Description: "AWS S3 file operations", Category: "integration", Keywords: []string{"s3", "aws", "file", "storage", "bucket"}},
		{Type: "integration.google_drive", Name: "Google Drive", Description: "Google Drive operations", Category: "integration", Keywords: []string{"google", "drive", "file", "storage"}},
		{Type: "integration.stripe", Name: "Stripe", Description: "Stripe payment operations", Category: "integration", Keywords: []string{"stripe", "payment", "charge", "customer"}},
		{Type: "integration.salesforce", Name: "Salesforce", Description: "Salesforce CRM operations", Category: "integration", Keywords: []string{"salesforce", "crm", "lead", "opportunity"}},
		{Type: "integration.hubspot", Name: "HubSpot", Description: "HubSpot CRM operations", Category: "integration", Keywords: []string{"hubspot", "crm", "contact", "deal"}},
		{Type: "integration.notion", Name: "Notion", Description: "Notion operations", Category: "integration", Keywords: []string{"notion", "page", "database", "note"}},
		{Type: "integration.airtable", Name: "Airtable", Description: "Airtable operations", Category: "integration", Keywords: []string{"airtable", "spreadsheet", "base", "record"}},
		{Type: "ai.chat", Name: "AI Chat", Description: "AI chat completion", Category: "ai", Keywords: []string{"ai", "chat", "gpt", "llm", "generate"}},
		{Type: "ai.vision", Name: "AI Vision", Description: "AI image analysis", Category: "ai", Keywords: []string{"ai", "vision", "image", "analyze"}},
		{Type: "ai.image", Name: "AI Image", Description: "AI image generation", Category: "ai", Keywords: []string{"ai", "image", "generate", "dalle"}},
	}
}
