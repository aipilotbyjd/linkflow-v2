package nodes

import (
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/actions/code"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/actions/email"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/actions/http"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/actions/transform"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/integrations/ai"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/integrations/cloud"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/integrations/communication"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/integrations/crm"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/integrations/database"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/integrations/payment"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/logic"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes/triggers"
)

// LoadAllNodes registers all available node types
func LoadAllNodes(registry *Registry) error {
	// Triggers
	if err := registry.Register("trigger.manual", func() Node { return triggers.NewManualTrigger() }); err != nil {
		return err
	}
	if err := registry.Register("trigger.webhook", func() Node { return triggers.NewWebhookTrigger() }); err != nil {
		return err
	}
	if err := registry.Register("trigger.schedule", func() Node { return triggers.NewScheduleTrigger() }); err != nil {
		return err
	}

	// Logic nodes
	if err := registry.Register("logic.if", func() Node { return logic.NewIfNode() }); err != nil {
		return err
	}
	if err := registry.Register("logic.switch", func() Node { return logic.NewSwitchNode() }); err != nil {
		return err
	}
	if err := registry.Register("logic.merge", func() Node { return logic.NewMergeNode() }); err != nil {
		return err
	}
	if err := registry.Register("logic.filter", func() Node { return logic.NewFilterNode() }); err != nil {
		return err
	}
	if err := registry.Register("logic.sort", func() Node { return logic.NewSortNode() }); err != nil {
		return err
	}
	if err := registry.Register("logic.limit", func() Node { return logic.NewLimitNode() }); err != nil {
		return err
	}
	if err := registry.Register("logic.loop", func() Node { return logic.NewLoopNode() }); err != nil {
		return err
	}
	if err := registry.Register("logic.wait", func() Node { return logic.NewWaitNode() }); err != nil {
		return err
	}
	if err := registry.Register("logic.aggregate", func() Node { return logic.NewAggregateNode() }); err != nil {
		return err
	}
	if err := registry.Register("logic.noop", func() Node { return logic.NewNoopNode() }); err != nil {
		return err
	}

	// Action nodes
	if err := registry.Register("action.http", func() Node { return http.NewHTTPRequestNode() }); err != nil {
		return err
	}
	if err := registry.Register("action.set", func() Node { return transform.NewSetNode() }); err != nil {
		return err
	}
	if err := registry.Register("action.email", func() Node { return email.NewSendEmailNode() }); err != nil {
		return err
	}
	if err := registry.Register("action.code", func() Node { return code.NewJavaScriptNode() }); err != nil {
		return err
	}

	// Cloud integrations
	if err := registry.Register("integration.aws_s3", func() Node { return cloud.NewAWSS3Node() }); err != nil {
		return err
	}
	if err := registry.Register("integration.google_drive", func() Node { return cloud.NewGoogleDriveNode() }); err != nil {
		return err
	}

	// Communication integrations
	if err := registry.Register("integration.slack", func() Node { return communication.NewSlackNode() }); err != nil {
		return err
	}
	if err := registry.Register("integration.discord", func() Node { return communication.NewDiscordNode() }); err != nil {
		return err
	}
	if err := registry.Register("integration.telegram", func() Node { return communication.NewTelegramNode() }); err != nil {
		return err
	}
	if err := registry.Register("integration.twilio", func() Node { return communication.NewTwilioNode() }); err != nil {
		return err
	}

	// Database integrations
	if err := registry.Register("integration.postgres", func() Node { return database.NewPostgresNode() }); err != nil {
		return err
	}
	if err := registry.Register("integration.mysql", func() Node { return database.NewMySQLNode() }); err != nil {
		return err
	}
	if err := registry.Register("integration.mongodb", func() Node { return database.NewMongoDBNode() }); err != nil {
		return err
	}
	if err := registry.Register("integration.redis", func() Node { return database.NewRedisNode() }); err != nil {
		return err
	}

	// AI integrations
	if err := registry.Register("integration.openai", func() Node { return ai.NewOpenAINode() }); err != nil {
		return err
	}
	if err := registry.Register("integration.anthropic", func() Node { return ai.NewAnthropicNode() }); err != nil {
		return err
	}

	// CRM integrations
	if err := registry.Register("integration.salesforce", func() Node { return crm.NewSalesforceNode() }); err != nil {
		return err
	}
	if err := registry.Register("integration.hubspot", func() Node { return crm.NewHubSpotNode() }); err != nil {
		return err
	}
	if err := registry.Register("integration.airtable", func() Node { return crm.NewAirtableNode() }); err != nil {
		return err
	}
	if err := registry.Register("integration.notion", func() Node { return crm.NewNotionNode() }); err != nil {
		return err
	}

	// Payment integrations
	if err := registry.Register("integration.stripe", func() Node { return payment.NewStripeNode() }); err != nil {
		return err
	}

	return nil
}

// LoadCoreNodes registers only core node types (triggers, logic, basic actions)
func LoadCoreNodes(registry *Registry) error {
	// Triggers
	if err := registry.Register("trigger.manual", func() Node { return triggers.NewManualTrigger() }); err != nil {
		return err
	}
	if err := registry.Register("trigger.webhook", func() Node { return triggers.NewWebhookTrigger() }); err != nil {
		return err
	}
	if err := registry.Register("trigger.schedule", func() Node { return triggers.NewScheduleTrigger() }); err != nil {
		return err
	}

	// Core logic
	if err := registry.Register("logic.if", func() Node { return logic.NewIfNode() }); err != nil {
		return err
	}
	if err := registry.Register("logic.noop", func() Node { return logic.NewNoopNode() }); err != nil {
		return err
	}

	// Core actions
	if err := registry.Register("action.http", func() Node { return http.NewHTTPRequestNode() }); err != nil {
		return err
	}
	if err := registry.Register("action.set", func() Node { return transform.NewSetNode() }); err != nil {
		return err
	}

	return nil
}
