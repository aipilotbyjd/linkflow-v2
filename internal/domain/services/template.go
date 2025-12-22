package services

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

type TemplateService struct {
	workflowSvc *WorkflowService
}

func NewTemplateService(workflowSvc *WorkflowService) *TemplateService {
	return &TemplateService{
		workflowSvc: workflowSvc,
	}
}

type WorkflowTemplate struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    string           `json:"category"`
	Icon        string           `json:"icon"`
	Tags        []string         `json:"tags"`
	Nodes       models.JSONArray `json:"nodes"`
	Connections models.JSONArray `json:"connections"`
	Settings    models.JSON      `json:"settings"`
}

func (s *TemplateService) List(ctx context.Context, category string) []WorkflowTemplate {
	templates := getBuiltInTemplates()
	
	if category != "" {
		filtered := make([]WorkflowTemplate, 0)
		for _, t := range templates {
			if t.Category == category {
				filtered = append(filtered, t)
			}
		}
		return filtered
	}
	
	return templates
}

func (s *TemplateService) Get(ctx context.Context, templateID string) (*WorkflowTemplate, error) {
	templates := getBuiltInTemplates()
	for _, t := range templates {
		if t.ID == templateID {
			return &t, nil
		}
	}
	return nil, ErrWorkflowNotFound
}

func (s *TemplateService) CreateFromTemplate(ctx context.Context, templateID string, workspaceID, userID uuid.UUID, name string) (*models.Workflow, error) {
	template, err := s.Get(ctx, templateID)
	if err != nil {
		return nil, err
	}

	workflowName := name
	if workflowName == "" {
		workflowName = template.Name
	}

	return s.workflowSvc.Create(ctx, CreateWorkflowInput{
		WorkspaceID: workspaceID,
		CreatedBy:   userID,
		Name:        workflowName,
		Description: &template.Description,
		Nodes:       template.Nodes,
		Connections: template.Connections,
		Settings:    template.Settings,
		Tags:        template.Tags,
	})
}

func getBuiltInTemplates() []WorkflowTemplate {
	return []WorkflowTemplate{
		{
			ID:          "webhook-to-slack",
			Name:        "Webhook to Slack",
			Description: "Receive webhook and send message to Slack channel",
			Category:    "communication",
			Icon:        "slack",
			Tags:        []string{"webhook", "slack", "notification"},
			Nodes:       parseJSONArray(`[{"id":"webhook","type":"trigger.webhook","name":"Webhook Trigger","position":{"x":100,"y":200},"config":{"method":"POST","path":"/webhook"}},{"id":"slack","type":"integration.slack","name":"Send Slack Message","position":{"x":400,"y":200},"config":{"operation":"postMessage","channel":"#general","text":"{{$json.message}}"}}]`),
			Connections: parseJSONArray(`[{"source":"webhook","target":"slack"}]`),
			Settings:    parseJSON(`{}`),
		},
		{
			ID:          "scheduled-report",
			Name:        "Scheduled Report",
			Description: "Generate and email a report on schedule",
			Category:    "automation",
			Icon:        "clock",
			Tags:        []string{"schedule", "email", "report"},
			Nodes:       parseJSONArray(`[{"id":"schedule","type":"trigger.schedule","name":"Daily Schedule","position":{"x":100,"y":200},"config":{"cron":"0 9 * * *","timezone":"UTC"}},{"id":"http","type":"action.http","name":"Fetch Data","position":{"x":300,"y":200},"config":{"method":"GET","url":"https://api.example.com/report"}},{"id":"email","type":"integration.email","name":"Send Report","position":{"x":500,"y":200},"config":{"to":"team@example.com","subject":"Daily Report","body":"{{$json}}"}}]`),
			Connections: parseJSONArray(`[{"source":"schedule","target":"http"},{"source":"http","target":"email"}]`),
			Settings:    parseJSON(`{}`),
		},
		{
			ID:          "github-pr-notify",
			Name:        "GitHub PR Notification",
			Description: "Notify on new pull requests",
			Category:    "development",
			Icon:        "github",
			Tags:        []string{"github", "pr", "notification", "slack"},
			Nodes:       parseJSONArray(`[{"id":"webhook","type":"trigger.webhook","name":"GitHub Webhook","position":{"x":100,"y":200},"config":{"method":"POST","path":"/github"}},{"id":"filter","type":"logic.condition","name":"Is PR Event","position":{"x":300,"y":200},"config":{"conditions":[{"field":"$json.action","operator":"equals","value":"opened"}]}},{"id":"slack","type":"integration.slack","name":"Notify Slack","position":{"x":500,"y":200},"config":{"operation":"postMessage","channel":"#dev","text":"New PR: {{$json.pull_request.title}} by {{$json.pull_request.user.login}}"}}]`),
			Connections: parseJSONArray(`[{"source":"webhook","target":"filter"},{"source":"filter","target":"slack","sourceHandle":"true"}]`),
			Settings:    parseJSON(`{}`),
		},
		{
			ID:          "form-to-database",
			Name:        "Form Submission to Database",
			Description: "Store form submissions in PostgreSQL",
			Category:    "data",
			Icon:        "database",
			Tags:        []string{"form", "database", "postgresql"},
			Nodes:       parseJSONArray(`[{"id":"webhook","type":"trigger.webhook","name":"Form Webhook","position":{"x":100,"y":200},"config":{"method":"POST","path":"/form"}},{"id":"transform","type":"action.code","name":"Transform Data","position":{"x":300,"y":200},"config":{"code":"return { ...items[0].json, submitted_at: new Date().toISOString() }"}},{"id":"postgres","type":"integration.postgresql","name":"Insert Record","position":{"x":500,"y":200},"config":{"operation":"insert","table":"submissions","data":"{{$json}}"}}]`),
			Connections: parseJSONArray(`[{"source":"webhook","target":"transform"},{"source":"transform","target":"postgres"}]`),
			Settings:    parseJSON(`{}`),
		},
		{
			ID:          "ai-content-generator",
			Name:        "AI Content Generator",
			Description: "Generate content with OpenAI and save to Notion",
			Category:    "ai",
			Icon:        "brain",
			Tags:        []string{"openai", "notion", "ai", "content"},
			Nodes:       parseJSONArray(`[{"id":"trigger","type":"trigger.manual","name":"Manual Trigger","position":{"x":100,"y":200},"config":{}},{"id":"openai","type":"integration.openai","name":"Generate Content","position":{"x":300,"y":200},"config":{"operation":"chatCompletion","model":"gpt-4","messages":[{"role":"user","content":"Write a blog post about: {{$json.topic}}"}]}},{"id":"notion","type":"integration.notion","name":"Create Page","position":{"x":500,"y":200},"config":{"operation":"createPage","parentId":"{{$json.notionParentId}}","title":"{{$json.topic}}","content":"{{$node.openai.json.choices[0].message.content}}"}}]`),
			Connections: parseJSONArray(`[{"source":"trigger","target":"openai"},{"source":"openai","target":"notion"}]`),
			Settings:    parseJSON(`{}`),
		},
		{
			ID:          "data-sync",
			Name:        "Data Sync Pipeline",
			Description: "Sync data between two systems with transformation",
			Category:    "data",
			Icon:        "sync",
			Tags:        []string{"sync", "etl", "data"},
			Nodes:       parseJSONArray(`[{"id":"schedule","type":"trigger.schedule","name":"Hourly Sync","position":{"x":100,"y":200},"config":{"cron":"0 * * * *"}},{"id":"source","type":"action.http","name":"Fetch Source","position":{"x":300,"y":200},"config":{"method":"GET","url":"{{$env.SOURCE_API_URL}}"}},{"id":"transform","type":"action.code","name":"Transform","position":{"x":500,"y":200},"config":{"code":"return items.map(item => ({ ...item.json, synced_at: new Date() }))"}},{"id":"loop","type":"logic.loop","name":"Process Each","position":{"x":700,"y":200},"config":{"mode":"forEach"}},{"id":"dest","type":"action.http","name":"Send to Dest","position":{"x":900,"y":200},"config":{"method":"POST","url":"{{$env.DEST_API_URL}}","body":"{{$json}}"}}]`),
			Connections: parseJSONArray(`[{"source":"schedule","target":"source"},{"source":"source","target":"transform"},{"source":"transform","target":"loop"},{"source":"loop","target":"dest","sourceHandle":"loop"}]`),
			Settings:    parseJSON(`{}`),
		},
		{
			ID:          "error-alerting",
			Name:        "Error Alerting System",
			Description: "Monitor for errors and alert via multiple channels",
			Category:    "monitoring",
			Icon:        "alert",
			Tags:        []string{"monitoring", "alert", "error"},
			Nodes:       parseJSONArray(`[{"id":"webhook","type":"trigger.webhook","name":"Error Webhook","position":{"x":100,"y":200},"config":{"method":"POST","path":"/errors"}},{"id":"check","type":"logic.condition","name":"Is Critical","position":{"x":300,"y":200},"config":{"conditions":[{"field":"$json.severity","operator":"equals","value":"critical"}]}},{"id":"slack","type":"integration.slack","name":"Slack Alert","position":{"x":500,"y":100},"config":{"operation":"postMessage","channel":"#alerts","text":"🚨 Critical Error: {{$json.message}}"}},{"id":"email","type":"integration.email","name":"Email Alert","position":{"x":500,"y":300},"config":{"to":"oncall@example.com","subject":"Critical Error Alert","body":"Error: {{$json.message}}\n\nStack: {{$json.stack}}"}}]`),
			Connections: parseJSONArray(`[{"source":"webhook","target":"check"},{"source":"check","target":"slack","sourceHandle":"true"},{"source":"check","target":"email","sourceHandle":"true"}]`),
			Settings:    parseJSON(`{}`),
		},
		{
			ID:          "lead-enrichment",
			Name:        "Lead Enrichment Pipeline",
			Description: "Enrich leads with external data and save to CRM",
			Category:    "sales",
			Icon:        "users",
			Tags:        []string{"leads", "enrichment", "crm"},
			Nodes:       parseJSONArray(`[{"id":"webhook","type":"trigger.webhook","name":"New Lead","position":{"x":100,"y":200},"config":{"method":"POST","path":"/leads"}},{"id":"enrich","type":"action.http","name":"Enrich Data","position":{"x":300,"y":200},"config":{"method":"GET","url":"https://api.clearbit.com/v2/people/find?email={{$json.email}}"}},{"id":"merge","type":"logic.merge","name":"Merge Data","position":{"x":500,"y":200},"config":{"mode":"combine"}},{"id":"crm","type":"action.http","name":"Save to CRM","position":{"x":700,"y":200},"config":{"method":"POST","url":"{{$env.CRM_API_URL}}/leads","body":"{{$json}}"}}]`),
			Connections: parseJSONArray(`[{"source":"webhook","target":"enrich"},{"source":"enrich","target":"merge"},{"source":"webhook","target":"merge"},{"source":"merge","target":"crm"}]`),
			Settings:    parseJSON(`{}`),
		},
		{
			ID:          "backup-workflow",
			Name:        "Automated Backup",
			Description: "Backup data to S3 on schedule",
			Category:    "operations",
			Icon:        "archive",
			Tags:        []string{"backup", "s3", "automation"},
			Nodes:       parseJSONArray(`[{"id":"schedule","type":"trigger.schedule","name":"Daily Backup","position":{"x":100,"y":200},"config":{"cron":"0 2 * * *","timezone":"UTC"}},{"id":"fetch","type":"action.http","name":"Export Data","position":{"x":300,"y":200},"config":{"method":"GET","url":"{{$env.API_URL}}/export"}},{"id":"compress","type":"action.code","name":"Compress","position":{"x":500,"y":200},"config":{"code":"const zlib = require('zlib'); return { data: zlib.gzipSync(JSON.stringify(items[0].json)).toString('base64'), filename: 'backup-' + new Date().toISOString().split('T')[0] + '.json.gz' }"}},{"id":"notify","type":"integration.slack","name":"Notify","position":{"x":700,"y":200},"config":{"operation":"postMessage","channel":"#ops","text":"✅ Backup completed: {{$json.filename}}"}}]`),
			Connections: parseJSONArray(`[{"source":"schedule","target":"fetch"},{"source":"fetch","target":"compress"},{"source":"compress","target":"notify"}]`),
			Settings:    parseJSON(`{}`),
		},
		{
			ID:          "chatbot-integration",
			Name:        "AI Chatbot",
			Description: "AI-powered chatbot with memory",
			Category:    "ai",
			Icon:        "message-circle",
			Tags:        []string{"chatbot", "ai", "openai"},
			Nodes:       parseJSONArray(`[{"id":"webhook","type":"trigger.webhook","name":"Chat Message","position":{"x":100,"y":200},"config":{"method":"POST","path":"/chat"}},{"id":"history","type":"action.http","name":"Get History","position":{"x":300,"y":200},"config":{"method":"GET","url":"{{$env.REDIS_URL}}/history/{{$json.sessionId}}"}},{"id":"ai","type":"integration.openai","name":"Generate Response","position":{"x":500,"y":200},"config":{"operation":"chatCompletion","model":"gpt-4","messages":"{{$node.history.json.messages.concat([{role:'user',content:$json.message}])}}"}},"id":"save","type":"action.http","name":"Save Message","position":{"x":700,"y":200},"config":{"method":"POST","url":"{{$env.REDIS_URL}}/history/{{$json.sessionId}}","body":{"message":"{{$json.message}}","response":"{{$node.ai.json.choices[0].message.content}}"}}},{"id":"respond","type":"logic.respond","name":"Send Response","position":{"x":900,"y":200},"config":{"body":"{{$node.ai.json.choices[0].message.content}}"}}]`),
			Connections: parseJSONArray(`[{"source":"webhook","target":"history"},{"source":"history","target":"ai"},{"source":"ai","target":"save"},{"source":"save","target":"respond"}]`),
			Settings:    parseJSON(`{}`),
		},
	}
}

func parseJSON(s string) models.JSON {
	var result models.JSON
	_ = json.Unmarshal([]byte(s), &result)
	return result
}

func parseJSONArray(s string) models.JSONArray {
	var result models.JSONArray
	_ = json.Unmarshal([]byte(s), &result)
	return result
}
