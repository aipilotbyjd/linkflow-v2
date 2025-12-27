package database

import (
	"fmt"
	"time"

	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewGormDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	gormConfig := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true,
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	log.Info().Msg("Database connected successfully")

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	log.Info().Msg("Running database migrations...")

	err := db.AutoMigrate(
		// Users & Auth
		&models.User{},
		&models.Session{},
		&models.APIKey{},
		&models.OAuthConnection{},
		&models.PasswordResetToken{},

		// Workspaces
		&models.Workspace{},
		&models.WorkspaceMember{},
		&models.WorkspaceInvitation{},

		// Workflows
		&models.Workflow{},
		&models.WorkflowVersion{},
		&models.WorkflowFolder{},

		// Executions
		&models.Execution{},
		&models.NodeExecution{},
		&models.ExecutionLog{},

		// Credentials
		&models.Credential{},

		// Schedules
		&models.Schedule{},

		// Webhooks
		&models.WebhookEndpoint{},
		&models.WebhookLog{},

		// Billing
		&models.Plan{},
		&models.Subscription{},
		&models.Usage{},
		&models.Invoice{},
		&models.OperationLog{},

		// Workflow Sharing & Marketplace
		&models.WorkflowShare{},
		&models.TemplateMarketplace{},
		&models.TemplateRating{},
		&models.WorkflowVariable{},

		// New Features
		&models.Template{},
		&models.WaitingExecution{},
		&models.PinnedData{},
		&models.BinaryData{},
		&models.OAuthState{},

		// Execution Features
		&models.ExecutionQueue{},
		&models.ExecutionShare{},
		&models.SubWorkflowExecution{},

		// Alerts & Monitoring
		&models.Alert{},
		&models.AlertLog{},
		&models.WorkspaceAnalytics{},
		&models.WorkflowAnalytics{},

		// Team & Collaboration
		&models.WorkflowComment{},
		&models.AuditLog{},
		&models.Permission{},
		&models.Role{},
		&models.RolePermission{},

		// Environment Variables
		&models.EnvironmentVariable{},

		// Webhook Features
		&models.WebhookSignatureConfig{},

		// Rate Limiting
		&models.CredentialRateLimit{},

		// Import/Export
		&models.WorkflowExport{},
		&models.WorkflowImport{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info().Msg("Database migrations completed")
	return nil
}

func SeedPlans(db *gorm.DB) error {
	plans := models.DefaultPlans()

	for _, plan := range plans {
		var existing models.Plan
		if err := db.First(&existing, "id = ?", plan.ID).Error; err == gorm.ErrRecordNotFound {
			plan.CreatedAt = time.Now()
			plan.UpdatedAt = time.Now()
			if err := db.Create(&plan).Error; err != nil {
				return fmt.Errorf("failed to seed plan %s: %w", plan.ID, err)
			}
			log.Info().Str("plan", plan.ID).Msg("Created plan")
		} else if err == nil {
			// Update existing plan with new fields
			plan.CreatedAt = existing.CreatedAt
			plan.UpdatedAt = time.Now()
			if err := db.Model(&existing).Updates(&plan).Error; err != nil {
				return fmt.Errorf("failed to update plan %s: %w", plan.ID, err)
			}
			log.Info().Str("plan", plan.ID).Msg("Updated plan")
		}
	}

	return nil
}

// SeedPermissions seeds default permissions
func SeedPermissions(db *gorm.DB) error {
	perms := models.DefaultPermissions()

	for _, perm := range perms {
		var existing models.Permission
		if err := db.First(&existing, "name = ?", perm.Name).Error; err == gorm.ErrRecordNotFound {
			perm.CreatedAt = time.Now()
			if err := db.Create(&perm).Error; err != nil {
				return fmt.Errorf("failed to seed permission %s: %w", perm.Name, err)
			}
			log.Debug().Str("permission", perm.Name).Msg("Created permission")
		}
	}

	log.Info().Int("count", len(perms)).Msg("Permissions seeded")
	return nil
}

// SeedRoles seeds default system roles
func SeedRoles(db *gorm.DB) error {
	roles := models.DefaultRoles()

	for _, role := range roles {
		var existing models.Role
		if err := db.First(&existing, "name = ? AND workspace_id IS NULL", role.Name).Error; err == gorm.ErrRecordNotFound {
			role.CreatedAt = time.Now()
			role.UpdatedAt = time.Now()
			if err := db.Create(&role).Error; err != nil {
				return fmt.Errorf("failed to seed role %s: %w", role.Name, err)
			}
			log.Debug().Str("role", role.Name).Msg("Created role")
		}
	}

	log.Info().Int("count", len(roles)).Msg("Roles seeded")
	return nil
}

// SeedAll runs all seed functions
func SeedAll(db *gorm.DB) error {
	if err := SeedPlans(db); err != nil {
		return err
	}
	if err := SeedPermissions(db); err != nil {
		return err
	}
	if err := SeedRoles(db); err != nil {
		return err
	}
	return nil
}
