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

		// Workflow Sharing & Marketplace
		&models.WorkflowShare{},
		&models.TemplateMarketplace{},
		&models.TemplateRating{},
		&models.WorkflowVariable{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info().Msg("Database migrations completed")
	return nil
}

func SeedPlans(db *gorm.DB) error {
	plans := []models.Plan{
		{
			ID:               "free",
			Name:             "Free",
			Description:      strPtr("Perfect for getting started"),
			PriceMonthly:     0,
			PriceYearly:      0,
			ExecutionsLimit:  100,
			WorkflowsLimit:   5,
			MembersLimit:     1,
			CredentialsLimit: 5,
			RetentionDays:    7,
			IsActive:         true,
			SortOrder:        1,
		},
		{
			ID:               "starter",
			Name:             "Starter",
			Description:      strPtr("For small teams"),
			PriceMonthly:     1900, // $19
			PriceYearly:      19000,
			ExecutionsLimit:  1000,
			WorkflowsLimit:   25,
			MembersLimit:     5,
			CredentialsLimit: 25,
			RetentionDays:    30,
			IsActive:         true,
			SortOrder:        2,
		},
		{
			ID:               "pro",
			Name:             "Pro",
			Description:      strPtr("For growing businesses"),
			PriceMonthly:     4900, // $49
			PriceYearly:      49000,
			ExecutionsLimit:  10000,
			WorkflowsLimit:   100,
			MembersLimit:     20,
			CredentialsLimit: 100,
			RetentionDays:    90,
			IsActive:         true,
			SortOrder:        3,
		},
		{
			ID:               "enterprise",
			Name:             "Enterprise",
			Description:      strPtr("For large organizations"),
			PriceMonthly:     19900, // $199
			PriceYearly:      199000,
			ExecutionsLimit:  100000,
			WorkflowsLimit:   -1, // unlimited
			MembersLimit:     -1,
			CredentialsLimit: -1,
			RetentionDays:    365,
			IsActive:         true,
			SortOrder:        4,
		},
	}

	for _, plan := range plans {
		var existing models.Plan
		if err := db.First(&existing, "id = ?", plan.ID).Error; err == gorm.ErrRecordNotFound {
			plan.CreatedAt = time.Now()
			plan.UpdatedAt = time.Now()
			if err := db.Create(&plan).Error; err != nil {
				return fmt.Errorf("failed to seed plan %s: %w", plan.ID, err)
			}
			log.Info().Str("plan", plan.ID).Msg("Created plan")
		}
	}

	return nil
}

func strPtr(s string) *string {
	return &s
}
