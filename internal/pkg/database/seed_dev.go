package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// DevSeedConfig holds configuration for development seeding
type DevSeedConfig struct {
	AdminEmail    string
	AdminPassword string
	CleanFirst    bool // If true, deletes existing dev data before seeding
}

// DefaultDevSeedConfig returns default configuration
func DefaultDevSeedConfig() DevSeedConfig {
	return DevSeedConfig{
		AdminEmail:    "admin@linkflow.dev",
		AdminPassword: "Admin123!",
		CleanFirst:    true,
	}
}

// SeedDevelopment seeds comprehensive test data for development/testing
// Creates: 5 users, 3 workspaces, folders, 30+ workflows, executions, credentials, schedules
func SeedDevelopment(db *gorm.DB, cfg DevSeedConfig) error {
	log.Info().Msg("Starting development data seeding...")

	if cfg.CleanFirst {
		if err := cleanDevData(db); err != nil {
			log.Warn().Err(err).Msg("Failed to clean dev data, continuing...")
		}
	}

	// 1. Create Users
	users, err := seedUsers(db, cfg)
	if err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}
	log.Info().Int("count", len(users)).Msg("Users seeded")

	// 2. Create Workspaces
	workspaces, err := seedWorkspaces(db, users)
	if err != nil {
		return fmt.Errorf("failed to seed workspaces: %w", err)
	}
	log.Info().Int("count", len(workspaces)).Msg("Workspaces seeded")

	// 3. Create Workspace Members
	if err := seedWorkspaceMembers(db, workspaces, users); err != nil {
		return fmt.Errorf("failed to seed workspace members: %w", err)
	}
	log.Info().Msg("Workspace members seeded")

	// 4. Create Folders
	folders, err := seedFolders(db, workspaces)
	if err != nil {
		return fmt.Errorf("failed to seed folders: %w", err)
	}
	log.Info().Int("count", len(folders)).Msg("Folders seeded")

	// 5. Create Credentials
	credentials, err := seedCredentials(db, workspaces, users)
	if err != nil {
		return fmt.Errorf("failed to seed credentials: %w", err)
	}
	log.Info().Int("count", len(credentials)).Msg("Credentials seeded")

	// 6. Create Workflows
	workflows, err := seedWorkflows(db, workspaces, folders, users)
	if err != nil {
		return fmt.Errorf("failed to seed workflows: %w", err)
	}
	log.Info().Int("count", len(workflows)).Msg("Workflows seeded")

	// 7. Create Schedules
	schedules, err := seedSchedules(db, workflows)
	if err != nil {
		return fmt.Errorf("failed to seed schedules: %w", err)
	}
	log.Info().Int("count", len(schedules)).Msg("Schedules seeded")

	// 8. Create Executions
	executions, err := seedExecutions(db, workflows, workspaces)
	if err != nil {
		return fmt.Errorf("failed to seed executions: %w", err)
	}
	log.Info().Int("count", len(executions)).Msg("Executions seeded")

	// 9. Create Environment Variables
	if err := seedEnvVars(db, workspaces, users); err != nil {
		return fmt.Errorf("failed to seed env vars: %w", err)
	}
	log.Info().Msg("Environment variables seeded")

	log.Info().Msg("Development data seeding completed successfully!")
	printSeedSummary(users, workspaces, folders, workflows, credentials, schedules, executions)

	return nil
}

func cleanDevData(db *gorm.DB) error {
	log.Info().Msg("Cleaning existing development data...")
	
	// Delete in reverse dependency order
	tables := []string{
		"node_executions",
		"execution_logs", 
		"executions",
		"schedules",
		"webhook_endpoints",
		"workflow_versions",
		"workflows",
		"credentials",
		"environment_variables",
		"workspace_members",
		"workspaces",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE created_at > NOW() - INTERVAL '30 days'", table)).Error; err != nil {
			log.Debug().Str("table", table).Err(err).Msg("Could not clean table")
		}
	}

	// Don't delete users - they might have real accounts
	return nil
}

func seedUsers(db *gorm.DB, cfg DevSeedConfig) ([]models.User, error) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	
	users := []models.User{
		{
			ID:            uuid.New(),
			Email:         cfg.AdminEmail,
			PasswordHash:  string(hashedPassword),
			FirstName:     "Admin",
			LastName:      "User",
			Status:        "active",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			Email:         "john@linkflow.dev",
			PasswordHash:  string(hashedPassword),
			FirstName:     "John",
			LastName:      "Developer",
			Status:        "active",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			Email:         "jane@linkflow.dev",
			PasswordHash:  string(hashedPassword),
			FirstName:     "Jane",
			LastName:      "Designer",
			Status:        "active",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			Email:         "bob@linkflow.dev",
			PasswordHash:  string(hashedPassword),
			FirstName:     "Bob",
			LastName:      "Manager",
			Status:        "active",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			Email:         "alice@linkflow.dev",
			PasswordHash:  string(hashedPassword),
			FirstName:     "Alice",
			LastName:      "Analyst",
			Status:        "active",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	var result []models.User
	for _, u := range users {
		var existing models.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&u).Error; err != nil {
				return nil, err
			}
			result = append(result, u)
		} else {
			result = append(result, existing)
		}
	}

	return result, nil
}

func seedWorkspaces(db *gorm.DB, users []models.User) ([]models.Workspace, error) {
	workspaces := []models.Workspace{
		{
			ID:        uuid.New(),
			Name:      "Acme Corporation",
			Slug:      "acme-corp",
			OwnerID:   users[0].ID,
			PlanID:    models.PlanPro,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Marketing Team",
			Slug:      "marketing-team",
			OwnerID:   users[1].ID,
			PlanID:    models.PlanStarter,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "DevOps Squad",
			Slug:      "devops-squad",
			OwnerID:   users[2].ID,
			PlanID:    models.PlanBusiness,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	var result []models.Workspace
	for _, w := range workspaces {
		var existing models.Workspace
		if err := db.Where("slug = ?", w.Slug).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&w).Error; err != nil {
				return nil, err
			}
			result = append(result, w)
		} else {
			result = append(result, existing)
		}
	}

	return result, nil
}

func seedWorkspaceMembers(db *gorm.DB, workspaces []models.Workspace, users []models.User) error {
	now := time.Now()
	
	members := []models.WorkspaceMember{
		// Acme Corp members
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, UserID: users[0].ID, Role: "owner", JoinedAt: &now},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, UserID: users[1].ID, Role: "admin", JoinedAt: &now},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, UserID: users[2].ID, Role: "member", JoinedAt: &now},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, UserID: users[3].ID, Role: "member", JoinedAt: &now},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, UserID: users[4].ID, Role: "viewer", JoinedAt: &now},
		// Marketing Team members
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, UserID: users[1].ID, Role: "owner", JoinedAt: &now},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, UserID: users[3].ID, Role: "admin", JoinedAt: &now},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, UserID: users[4].ID, Role: "member", JoinedAt: &now},
		// DevOps Squad members
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, UserID: users[2].ID, Role: "owner", JoinedAt: &now},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, UserID: users[0].ID, Role: "admin", JoinedAt: &now},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, UserID: users[1].ID, Role: "member", JoinedAt: &now},
	}

	for _, m := range members {
		var existing models.WorkspaceMember
		if err := db.Where("workspace_id = ? AND user_id = ?", m.WorkspaceID, m.UserID).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&m).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func seedFolders(db *gorm.DB, workspaces []models.Workspace) ([]models.Folder, error) {
	colors := []string{"#3B82F6", "#10B981", "#F59E0B", "#EF4444", "#8B5CF6", "#EC4899"}
	
	folders := []models.Folder{
		// Acme Corp folders
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, Name: "Email Automations", Color: &colors[0], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, Name: "CRM Integrations", Color: &colors[1], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, Name: "Data Pipelines", Color: &colors[2], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, Name: "Notifications", Color: &colors[3], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// Marketing Team folders
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, Name: "Lead Generation", Color: &colors[4], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, Name: "Social Media", Color: &colors[5], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, Name: "Analytics", Color: &colors[0], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// DevOps Squad folders
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, Name: "CI/CD Pipelines", Color: &colors[1], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, Name: "Monitoring", Color: &colors[2], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, Name: "Infrastructure", Color: &colors[3], CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	var result []models.Folder
	for _, f := range folders {
		var existing models.Folder
		if err := db.Where("workspace_id = ? AND name = ?", f.WorkspaceID, f.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&f).Error; err != nil {
				return nil, err
			}
			result = append(result, f)
		} else {
			result = append(result, existing)
		}
	}

	// Create sub-folders
	subFolders := []models.Folder{
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, ParentID: &result[0].ID, Name: "Welcome Emails", Color: &colors[0], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, ParentID: &result[0].ID, Name: "Newsletter", Color: &colors[0], CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, ParentID: &result[7].ID, Name: "GitHub Actions", Color: &colors[1], CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, f := range subFolders {
		var existing models.Folder
		if err := db.Where("workspace_id = ? AND name = ? AND parent_id = ?", f.WorkspaceID, f.Name, f.ParentID).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&f).Error; err != nil {
				return nil, err
			}
			result = append(result, f)
		} else {
			result = append(result, existing)
		}
	}

	return result, nil
}

func seedCredentials(db *gorm.DB, workspaces []models.Workspace, users []models.User) ([]models.Credential, error) {
	credentials := []models.Credential{
		// Acme Corp credentials
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, CreatedBy: users[0].ID, Name: "Slack Bot Token", Type: "slack", Data: `{"token":"xoxb-xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, CreatedBy: users[0].ID, Name: "SendGrid API Key", Type: "sendgrid", Data: `{"api_key":"SG.xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, CreatedBy: users[0].ID, Name: "PostgreSQL Production", Type: "postgres", Data: `{"host":"db.example.com","port":5432}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, CreatedBy: users[1].ID, Name: "AWS S3 Bucket", Type: "aws_s3", Data: `{"access_key":"xxx","secret_key":"xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, CreatedBy: users[1].ID, Name: "Stripe API", Type: "stripe", Data: `{"secret_key":"sk_xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// Marketing Team credentials
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, CreatedBy: users[1].ID, Name: "HubSpot API", Type: "hubspot", Data: `{"api_key":"xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, CreatedBy: users[1].ID, Name: "Mailchimp", Type: "mailchimp", Data: `{"api_key":"xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, CreatedBy: users[3].ID, Name: "Google Analytics", Type: "google", Data: `{"client_id":"xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// DevOps Squad credentials
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, CreatedBy: users[2].ID, Name: "GitHub Token", Type: "github", Data: `{"token":"ghp_xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, CreatedBy: users[2].ID, Name: "Docker Hub", Type: "docker", Data: `{"username":"xxx","password":"xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, CreatedBy: users[0].ID, Name: "PagerDuty API", Type: "pagerduty", Data: `{"api_key":"xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, CreatedBy: users[0].ID, Name: "Datadog API", Type: "datadog", Data: `{"api_key":"xxx","app_key":"xxx"}`, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	var result []models.Credential
	for _, c := range credentials {
		var existing models.Credential
		if err := db.Where("workspace_id = ? AND name = ?", c.WorkspaceID, c.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&c).Error; err != nil {
				return nil, err
			}
			result = append(result, c)
		} else {
			result = append(result, existing)
		}
	}

	return result, nil
}

func seedWorkflows(db *gorm.DB, workspaces []models.Workspace, folders []models.Folder, users []models.User) ([]models.Workflow, error) {
	statuses := []string{"active", "active", "active", "inactive", "draft", "archived"}
	
	// Basic workflow template
	basicNodes := models.JSONArray{
		map[string]interface{}{
			"id": "trigger_1", "type": "trigger.manual", "name": "Manual Trigger",
			"position": map[string]int{"x": 100, "y": 100}, "parameters": map[string]interface{}{},
		},
		map[string]interface{}{
			"id": "http_1", "type": "action.http", "name": "HTTP Request",
			"position": map[string]int{"x": 300, "y": 100},
			"parameters": map[string]interface{}{"url": "https://api.example.com/data", "method": "GET"},
		},
	}
	basicConnections := models.JSONArray{
		map[string]interface{}{"id": "conn_1", "source_node_id": "trigger_1", "target_node_id": "http_1"},
	}

	// Email workflow template
	emailNodes := models.JSONArray{
		map[string]interface{}{
			"id": "trigger_1", "type": "trigger.webhook", "name": "Webhook Trigger",
			"position": map[string]int{"x": 100, "y": 100}, "parameters": map[string]interface{}{},
		},
		map[string]interface{}{
			"id": "email_1", "type": "action.email", "name": "Send Email",
			"position": map[string]int{"x": 300, "y": 100},
			"parameters": map[string]interface{}{"to": "{{trigger.email}}", "subject": "Welcome!"},
		},
	}

	workflows := []models.Workflow{
		// Acme Corp - Email Automations folder (index 0)
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: &folders[0].ID, CreatedBy: users[0].ID, Name: "Welcome Email Sequence", Status: "active", Version: 1, Nodes: emailNodes, Connections: basicConnections, Tags: models.StringArray{"email", "onboarding"}, ExecutionCount: 150, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: &folders[0].ID, CreatedBy: users[0].ID, Name: "Newsletter Sender", Status: "active", Version: 2, Nodes: emailNodes, Connections: basicConnections, Tags: models.StringArray{"email", "newsletter"}, ExecutionCount: 89, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: &folders[0].ID, CreatedBy: users[1].ID, Name: "Abandoned Cart Reminder", Status: "active", Version: 1, Nodes: emailNodes, Connections: basicConnections, Tags: models.StringArray{"email", "ecommerce"}, ExecutionCount: 234, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// Acme Corp - CRM Integrations folder (index 1)
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: &folders[1].ID, CreatedBy: users[1].ID, Name: "Salesforce Lead Sync", Status: "active", Version: 3, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"crm", "salesforce"}, ExecutionCount: 567, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: &folders[1].ID, CreatedBy: users[2].ID, Name: "HubSpot Contact Update", Status: "inactive", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"crm", "hubspot"}, ExecutionCount: 45, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// Acme Corp - Data Pipelines folder (index 2)
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: &folders[2].ID, CreatedBy: users[0].ID, Name: "Daily Data Export", Status: "active", Version: 5, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"data", "export"}, ExecutionCount: 730, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: &folders[2].ID, CreatedBy: users[1].ID, Name: "S3 Backup Pipeline", Status: "active", Version: 2, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"data", "backup", "aws"}, ExecutionCount: 365, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: &folders[2].ID, CreatedBy: users[2].ID, Name: "ETL Customer Data", Status: "draft", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"data", "etl"}, ExecutionCount: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// Acme Corp - Notifications folder (index 3)
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: &folders[3].ID, CreatedBy: users[0].ID, Name: "Slack Alert on Error", Status: "active", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"notification", "slack"}, ExecutionCount: 123, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: &folders[3].ID, CreatedBy: users[3].ID, Name: "PagerDuty Incident", Status: "active", Version: 2, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"notification", "pagerduty"}, ExecutionCount: 67, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// Acme Corp - Root level workflows
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: nil, CreatedBy: users[0].ID, Name: "Quick Test Workflow", Status: "draft", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"test"}, ExecutionCount: 5, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, FolderID: nil, CreatedBy: users[1].ID, Name: "API Health Check", Status: "active", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"monitoring", "health"}, ExecutionCount: 8760, CreatedAt: time.Now(), UpdatedAt: time.Now()},

		// Marketing Team - Lead Generation folder (index 4)
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, FolderID: &folders[4].ID, CreatedBy: users[1].ID, Name: "Form Submission Handler", Status: "active", Version: 2, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"leads", "forms"}, ExecutionCount: 456, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, FolderID: &folders[4].ID, CreatedBy: users[3].ID, Name: "Lead Scoring Automation", Status: "active", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"leads", "scoring"}, ExecutionCount: 234, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, FolderID: &folders[4].ID, CreatedBy: users[4].ID, Name: "LinkedIn Lead Import", Status: "inactive", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"leads", "linkedin"}, ExecutionCount: 12, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// Marketing Team - Social Media folder (index 5)
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, FolderID: &folders[5].ID, CreatedBy: users[1].ID, Name: "Twitter Auto Post", Status: "active", Version: 3, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"social", "twitter"}, ExecutionCount: 180, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, FolderID: &folders[5].ID, CreatedBy: users[3].ID, Name: "Instagram Scheduler", Status: "draft", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"social", "instagram"}, ExecutionCount: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// Marketing Team - Analytics folder (index 6)
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, FolderID: &folders[6].ID, CreatedBy: users[4].ID, Name: "Weekly Report Generator", Status: "active", Version: 2, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"analytics", "reports"}, ExecutionCount: 52, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, FolderID: &folders[6].ID, CreatedBy: users[4].ID, Name: "Campaign Performance Tracker", Status: "active", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"analytics", "campaigns"}, ExecutionCount: 89, CreatedAt: time.Now(), UpdatedAt: time.Now()},

		// DevOps Squad - CI/CD Pipelines folder (index 7)
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, FolderID: &folders[7].ID, CreatedBy: users[2].ID, Name: "GitHub PR Notifier", Status: "active", Version: 4, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"ci", "github"}, ExecutionCount: 1234, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, FolderID: &folders[7].ID, CreatedBy: users[0].ID, Name: "Auto Deploy to Staging", Status: "active", Version: 2, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"cd", "deploy"}, ExecutionCount: 456, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, FolderID: &folders[7].ID, CreatedBy: users[1].ID, Name: "Docker Image Builder", Status: "active", Version: 3, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"ci", "docker"}, ExecutionCount: 789, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// DevOps Squad - Monitoring folder (index 8)
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, FolderID: &folders[8].ID, CreatedBy: users[2].ID, Name: "Uptime Monitor", Status: "active", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"monitoring", "uptime"}, ExecutionCount: 43200, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, FolderID: &folders[8].ID, CreatedBy: users[0].ID, Name: "Error Rate Alert", Status: "active", Version: 2, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"monitoring", "alerts"}, ExecutionCount: 156, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, FolderID: &folders[8].ID, CreatedBy: users[1].ID, Name: "Log Aggregator", Status: "inactive", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"monitoring", "logs"}, ExecutionCount: 34, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// DevOps Squad - Infrastructure folder (index 9)
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, FolderID: &folders[9].ID, CreatedBy: users[2].ID, Name: "AWS Cost Reporter", Status: "active", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"infra", "aws", "cost"}, ExecutionCount: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, FolderID: &folders[9].ID, CreatedBy: users[0].ID, Name: "SSL Certificate Monitor", Status: "active", Version: 1, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"infra", "ssl"}, ExecutionCount: 365, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		// DevOps Squad - Root level
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, FolderID: nil, CreatedBy: users[2].ID, Name: "Incident Response Bot", Status: "active", Version: 5, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"incident", "automation"}, ExecutionCount: 89, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, FolderID: nil, CreatedBy: users[1].ID, Name: "Database Backup Checker", Status: "archived", Version: 2, Nodes: basicNodes, Connections: basicConnections, Tags: models.StringArray{"backup", "database"}, ExecutionCount: 180, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	// Randomize statuses
	for i := range workflows {
		workflows[i].Status = statuses[i%len(statuses)]
	}

	var result []models.Workflow
	for _, w := range workflows {
		var existing models.Workflow
		if err := db.Where("workspace_id = ? AND name = ?", w.WorkspaceID, w.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&w).Error; err != nil {
				return nil, err
			}
			result = append(result, w)
		} else {
			result = append(result, existing)
		}
	}

	return result, nil
}

func seedSchedules(db *gorm.DB, workflows []models.Workflow) ([]models.Schedule, error) {
	schedules := []models.Schedule{
		{ID: uuid.New(), WorkflowID: workflows[5].ID, WorkspaceID: workflows[5].WorkspaceID, CreatedBy: workflows[5].CreatedBy, CronExpression: "0 0 * * *", Timezone: "UTC", IsActive: true, Name: "Daily Data Export", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkflowID: workflows[6].ID, WorkspaceID: workflows[6].WorkspaceID, CreatedBy: workflows[6].CreatedBy, CronExpression: "0 2 * * *", Timezone: "UTC", IsActive: true, Name: "Nightly S3 Backup", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkflowID: workflows[11].ID, WorkspaceID: workflows[11].WorkspaceID, CreatedBy: workflows[11].CreatedBy, CronExpression: "*/5 * * * *", Timezone: "UTC", IsActive: true, Name: "Health Check Every 5 Min", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkflowID: workflows[18].ID, WorkspaceID: workflows[18].WorkspaceID, CreatedBy: workflows[18].CreatedBy, CronExpression: "0 9 * * 1", Timezone: "America/New_York", IsActive: true, Name: "Weekly Report Monday 9AM", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkflowID: workflows[23].ID, WorkspaceID: workflows[23].WorkspaceID, CreatedBy: workflows[23].CreatedBy, CronExpression: "* * * * *", Timezone: "UTC", IsActive: true, Name: "Uptime Check Every Minute", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkflowID: workflows[27].ID, WorkspaceID: workflows[27].WorkspaceID, CreatedBy: workflows[27].CreatedBy, CronExpression: "0 6 1 * *", Timezone: "UTC", IsActive: true, Name: "Monthly Cost Report", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkflowID: workflows[28].ID, WorkspaceID: workflows[28].WorkspaceID, CreatedBy: workflows[28].CreatedBy, CronExpression: "0 0 * * *", Timezone: "UTC", IsActive: true, Name: "Daily SSL Check", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	var result []models.Schedule
	for _, s := range schedules {
		var existing models.Schedule
		if err := db.Where("workflow_id = ?", s.WorkflowID).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&s).Error; err != nil {
				return nil, err
			}
			result = append(result, s)
		} else {
			result = append(result, existing)
		}
	}

	return result, nil
}

func seedExecutions(db *gorm.DB, workflows []models.Workflow, workspaces []models.Workspace) ([]models.Execution, error) {
	statuses := []string{"completed", "completed", "completed", "completed", "failed", "running", "queued"}
	triggerTypes := []string{"manual", "webhook", "schedule"}

	var executions []models.Execution
	
	// Create 3-5 executions per workflow for first 15 workflows
	for i := 0; i < 15 && i < len(workflows); i++ {
		numExecs := 3 + (i % 3) // 3, 4, or 5 executions
		for j := 0; j < numExecs; j++ {
			startTime := time.Now().Add(-time.Duration(j*24) * time.Hour)
			endTime := startTime.Add(time.Duration(100+(j*50)) * time.Millisecond)
			status := statuses[(i+j)%len(statuses)]
			
			exec := models.Execution{
				ID:              uuid.New(),
				WorkflowID:      workflows[i].ID,
				WorkspaceID:     workflows[i].WorkspaceID,
				WorkflowVersion: workflows[i].Version,
				Status:          status,
				TriggerType:     triggerTypes[j%len(triggerTypes)],
				InputData:       models.JSON{"test": true, "iteration": j},
				StartedAt:       &startTime,
				QueuedAt:        startTime,
			}
			
			if status == "completed" || status == "failed" {
				exec.CompletedAt = &endTime
				exec.OutputData = models.JSON{"success": status == "completed", "result": "test output"}
			}
			if status == "failed" {
				errMsg := "Test error: something went wrong"
				exec.ErrorMessage = &errMsg
			}

			executions = append(executions, exec)
		}
	}

	var result []models.Execution
	for _, e := range executions {
		if err := db.Create(&e).Error; err != nil {
			log.Debug().Err(err).Msg("Could not create execution")
			continue
		}
		result = append(result, e)
	}

	return result, nil
}

func seedEnvVars(db *gorm.DB, workspaces []models.Workspace, users []models.User) error {
	envVars := []models.EnvironmentVariable{
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, CreatedBy: users[0].ID, Name: "API_BASE_URL", Value: "https://api.acme.com", IsSecret: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, CreatedBy: users[0].ID, Name: "DEBUG_MODE", Value: "false", IsSecret: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[0].ID, CreatedBy: users[0].ID, Name: "SECRET_KEY", Value: "encrypted_secret_value", IsSecret: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, CreatedBy: users[1].ID, Name: "MARKETING_API_KEY", Value: "mkt_xxx", IsSecret: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[1].ID, CreatedBy: users[1].ID, Name: "CAMPAIGN_PREFIX", Value: "MKT2024", IsSecret: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, CreatedBy: users[2].ID, Name: "DEPLOY_ENV", Value: "staging", IsSecret: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), WorkspaceID: workspaces[2].ID, CreatedBy: users[2].ID, Name: "SLACK_WEBHOOK", Value: "https://hooks.slack.com/xxx", IsSecret: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, ev := range envVars {
		var existing models.EnvironmentVariable
		if err := db.Where("workspace_id = ? AND name = ?", ev.WorkspaceID, ev.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&ev).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func printSeedSummary(users []models.User, workspaces []models.Workspace, folders []models.Folder, workflows []models.Workflow, credentials []models.Credential, schedules []models.Schedule, executions []models.Execution) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("DEVELOPMENT SEED DATA SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Users:       %d\n", len(users))
	fmt.Printf("Workspaces:  %d\n", len(workspaces))
	fmt.Printf("Folders:     %d\n", len(folders))
	fmt.Printf("Workflows:   %d\n", len(workflows))
	fmt.Printf("Credentials: %d\n", len(credentials))
	fmt.Printf("Schedules:   %d\n", len(schedules))
	fmt.Printf("Executions:  %d\n", len(executions))
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("\nTEST ACCOUNTS (all use same password):")
	fmt.Println("Password: Admin123!")
	fmt.Println("")
	for _, u := range users {
		fmt.Printf("  - %s (%s %s)\n", u.Email, u.FirstName, u.LastName)
	}
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("\nWORKSPACES:")
	for _, w := range workspaces {
		fmt.Printf("  - %s (slug: %s, plan: %s)\n", w.Name, w.Slug, w.PlanID)
	}
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
