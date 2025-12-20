package integrations

import "github.com/linkflow-ai/linkflow/internal/worker/core"

func init() {
	// Register all integration nodes
	core.Register(&SlackNode{}, core.NodeMeta{
		Name:        "Slack",
		Description: "Send messages and interact with Slack",
		Category:    "integrations",
		Icon:        "slack",
		Version:     "1.0.0",
		Tags:        []string{"messaging", "communication"},
	})

	core.Register(&EmailNode{}, core.NodeMeta{
		Name:        "Email",
		Description: "Send emails via SMTP",
		Category:    "integrations",
		Icon:        "mail",
		Version:     "1.0.0",
		Tags:        []string{"messaging", "communication"},
	})

	core.Register(&OpenAINode{}, core.NodeMeta{
		Name:        "OpenAI",
		Description: "Use OpenAI GPT models for AI tasks",
		Category:    "integrations",
		Icon:        "openai",
		Version:     "1.0.0",
		Tags:        []string{"ai", "llm"},
	})

	core.Register(&GitHubNode{}, core.NodeMeta{
		Name:        "GitHub",
		Description: "Interact with GitHub repositories",
		Category:    "integrations",
		Icon:        "github",
		Version:     "1.0.0",
		Tags:        []string{"dev-tools", "vcs"},
	})

	core.Register(&DiscordNode{}, core.NodeMeta{
		Name:        "Discord",
		Description: "Send messages to Discord webhooks",
		Category:    "integrations",
		Icon:        "discord",
		Version:     "1.0.0",
		Tags:        []string{"messaging", "communication"},
	})

	core.Register(&TelegramNode{}, core.NodeMeta{
		Name:        "Telegram",
		Description: "Send messages via Telegram bot",
		Category:    "integrations",
		Icon:        "telegram",
		Version:     "1.0.0",
		Tags:        []string{"messaging", "communication"},
	})

	core.Register(&PostgresNode{}, core.NodeMeta{
		Name:        "PostgreSQL",
		Description: "Execute PostgreSQL queries",
		Category:    "integrations",
		Icon:        "database",
		Version:     "1.0.0",
		Tags:        []string{"database"},
	})

	core.Register(&NotionNode{}, core.NodeMeta{
		Name:        "Notion",
		Description: "Interact with Notion pages and databases",
		Category:    "integrations",
		Icon:        "notion",
		Version:     "1.0.0",
		Tags:        []string{"productivity"},
	})

	core.Register(&AirtableNode{}, core.NodeMeta{
		Name:        "Airtable",
		Description: "Interact with Airtable bases",
		Category:    "integrations",
		Icon:        "airtable",
		Version:     "1.0.0",
		Tags:        []string{"database", "productivity"},
	})

	core.Register(&AnthropicNode{}, core.NodeMeta{
		Name:        "Anthropic",
		Description: "Use Claude AI models",
		Category:    "integrations",
		Icon:        "anthropic",
		Version:     "1.0.0",
		Tags:        []string{"ai", "llm"},
	})
}
