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

	// New integrations
	core.Register(&FTPNode{}, core.NodeMeta{
		Name:        "FTP",
		Description: "FTP file operations",
		Category:    "integrations",
		Icon:        "folder",
		Version:     "1.0.0",
		Tags:        []string{"files", "storage"},
	})

	core.Register(&SFTPNode{}, core.NodeMeta{
		Name:        "SFTP",
		Description: "Secure FTP file operations",
		Category:    "integrations",
		Icon:        "folder-lock",
		Version:     "1.0.0",
		Tags:        []string{"files", "storage"},
	})

	core.Register(&GraphQLNode{}, core.NodeMeta{
		Name:        "GraphQL",
		Description: "Execute GraphQL queries and mutations",
		Category:    "integrations",
		Icon:        "graphql",
		Version:     "1.0.0",
		Tags:        []string{"api"},
	})

	core.Register(&AWSS3Node{}, core.NodeMeta{
		Name:        "AWS S3",
		Description: "AWS S3 storage operations",
		Category:    "integrations",
		Icon:        "aws",
		Version:     "1.0.0",
		Tags:        []string{"cloud", "storage"},
	})

	core.Register(&TwilioNode{}, core.NodeMeta{
		Name:        "Twilio",
		Description: "Send SMS, MMS and make calls via Twilio",
		Category:    "integrations",
		Icon:        "phone",
		Version:     "1.0.0",
		Tags:        []string{"messaging", "communication"},
	})

	core.Register(&GoogleDriveNode{}, core.NodeMeta{
		Name:        "Google Drive",
		Description: "Google Drive file operations",
		Category:    "integrations",
		Icon:        "google-drive",
		Version:     "1.0.0",
		Tags:        []string{"cloud", "storage"},
	})

	core.Register(&JiraNode{}, core.NodeMeta{
		Name:        "Jira",
		Description: "Jira issue and project management",
		Category:    "integrations",
		Icon:        "jira",
		Version:     "1.0.0",
		Tags:        []string{"dev-tools", "project-management"},
	})

	core.Register(&SalesforceNode{}, core.NodeMeta{
		Name:        "Salesforce",
		Description: "Salesforce CRM operations",
		Category:    "integrations",
		Icon:        "salesforce",
		Version:     "1.0.0",
		Tags:        []string{"crm"},
	})

	core.Register(&SendGridNode{}, core.NodeMeta{
		Name:        "SendGrid",
		Description: "Send emails via SendGrid",
		Category:    "integrations",
		Icon:        "mail",
		Version:     "1.0.0",
		Tags:        []string{"messaging", "email"},
	})
}
