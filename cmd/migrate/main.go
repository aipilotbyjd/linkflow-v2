package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/mappers"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
	"github.com/linkflow-ai/linkflow/internal/core/domain/sitesettings"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "configs/config.yaml", "Path to config file")
	direction := flag.String("direction", "up", "Migration direction: up or down")
	steps := flag.Int("steps", 0, "Number of migrations to run (0 = all)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Str("app", cfg.App.Name).
		Str("direction", *direction).
		Msg("Running database migrations")

	// Initialize database
	db, err := postgres.NewClient(postgres.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		Database:     cfg.Database.Name,
		SSLMode:      cfg.Database.SSLMode,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		MaxLifetime:  cfg.Database.MaxLifetime,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}()

	// Run migrations
	switch *direction {
	case "up":
		if err := runMigrationsUp(db, *steps); err != nil {
			log.Fatal().Err(err).Msg("Migration up failed")
		}
		log.Info().Msg("Migrations completed successfully")

	case "down":
		if err := runMigrationsDown(db, *steps); err != nil {
			log.Fatal().Err(err).Msg("Migration down failed")
		}
		log.Info().Msg("Rollback completed successfully")

	case "status":
		if err := showMigrationStatus(db); err != nil {
			log.Fatal().Err(err).Msg("Failed to get migration status")
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown direction: %s\n", *direction)
		fmt.Fprintln(os.Stderr, "Usage: migrate -direction=[up|down|status] [-steps=N]")
		os.Exit(1)
	}
}

func runMigrationsUp(db *gorm.DB, steps int) error {
	log.Info().Int("steps", steps).Msg("Running GORM AutoMigrate")

	// Enable UUID extension
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	// AutoMigrate all models from persistence layer
	if err := db.AutoMigrate(
		// User & Auth
		&models.User{},
		&models.UserSession{},
		&user.APIKey{},
		&user.OAuthConnection{},
		&user.PasswordResetToken{},

		// Workspace
		&models.Workspace{},
		&models.WorkspaceMember{},
		&models.Role{},
		&models.Permission{},

		// Workflow
		&models.Workflow{},
		&models.WorkflowVersion{},

		// Execution
		&models.Execution{},
		&models.NodeExecution{},

		// Folder
		&models.Folder{},

		// Additional models
		&models.PinnedData{},
		&models.Share{},
		&models.BinaryData{},
		&models.Variable{},
		&models.Environment{},
		&models.EnvironmentVariable{},

		// Billing
		&billing.Subscription{},
		&billing.Usage{},
		&billing.Invoice{},

		// AI Integration
		&models.AIUsage{},
		&models.PromptTemplate{},
		&models.AICache{},

		// Site Settings
		&sitesettings.SiteSettings{},
	); err != nil {
		return fmt.Errorf("auto-migrate failed: %w", err)
	}

	// Seed RBAC
	if err := seedRBAC(db); err != nil {
		return fmt.Errorf("rbac seeding failed: %w", err)
	}

	// Migrate Member Roles
	if err := migrateMemberRoles(db); err != nil {
		// Log error but don't fail migration, as this might be partial or redundant
		log.Error().Err(err).Msg("Member role migration failed")
	}

	log.Info().Msg("All tables created/updated successfully")
	return nil
}

func migrateMemberRoles(db *gorm.DB) error {
	log.Info().Msg("Migrating member roles to RBAC")

	// Get system roles map
	systemRoles := make(map[string]uuid.UUID)
	var roles []models.Role
	if err := db.Where("workspace_id IS NULL").Find(&roles).Error; err != nil {
		return err
	}
	for _, r := range roles {
		systemRoles[r.Name] = r.ID
	}

	// Find members with NULL role_id
	var members []models.WorkspaceMember
	if err := db.Where("role_id IS NULL").Find(&members).Error; err != nil {
		return err
	}

	for _, m := range members {
		roleName := ""
		// Map existing lowercase role string to System Role Name (Capitalized)
		switch m.Role {
		case "owner":
			roleName = rbac.RoleOwner
		case "admin":
			roleName = rbac.RoleAdmin
		case "member":
			roleName = rbac.RoleEditor // Map member -> Editor
		case "viewer":
			roleName = rbac.RoleViewer
		default:
			roleName = rbac.RoleViewer // Default fallback
		}

		if roleID, ok := systemRoles[roleName]; ok {
			// Update role_id
			if err := db.Model(&m).Update("role_id", roleID).Error; err != nil {
				log.Error().Err(err).Str("member_id", m.ID.String()).Msg("Failed to update member role_id")
			}
		} else {
			log.Warn().Str("role", m.Role).Msg("Could not map legacy role to system role")
		}
	}
	return nil
}

func seedRBAC(db *gorm.DB) error {
	log.Info().Msg("Seeding RBAC permissions and roles")

	// 1. Seed Permissions
	for _, p := range rbac.AllPermissions {
		model := mappers.ToModelPermission(p)
		// Use FirstOrCreate to ensure we don't duplicate, but also update fields if needed?
		// For permissions, ID is key. If name/desc changes, we might want to update.
		// For now FirstOrCreate on ID is enough.
		if err := db.FirstOrCreate(&model, models.Permission{ID: p.ID}).Error; err != nil {
			return err
		}
	}

	// 2. Seed System Roles
	systemRoles := rbac.GetSystemRoles()
	for _, r := range systemRoles {
		// Check if role exists by Name and WorkspaceID (NULL)
		var existing models.Role
		err := db.Where("name = ? AND workspace_id IS NULL", r.Name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			// Create
			roleModel := mappers.ToModelRole(r)
			if err := db.Create(roleModel).Error; err != nil {
				return err
			}
			// Assign permissions
			perms := make([]models.Permission, len(r.Permissions))
			for i, p := range r.Permissions {
				perms[i] = models.Permission{ID: p.ID}
			}
			if err := db.Model(roleModel).Association("Permissions").Replace(perms); err != nil {
				return err
			}
			log.Info().Str("role", r.Name).Msg("Created system role")
		} else if err != nil {
			return err
		} else {
			// Update permissions for existing system role
			perms := make([]models.Permission, len(r.Permissions))
			for i, p := range r.Permissions {
				perms[i] = models.Permission{ID: p.ID}
			}
			if err := db.Model(&existing).Association("Permissions").Replace(perms); err != nil {
				return err
			}
		}
	}

	return nil
}

func runMigrationsDown(db *gorm.DB, steps int) error {
	log.Info().Int("steps", steps).Msg("Running migrations down")
	log.Warn().Msg("Rollback not implemented - GORM AutoMigrate only adds columns, doesn't remove")
	return nil
}

func showMigrationStatus(db *gorm.DB) error {
	log.Info().Msg("Migration status - checking tables")

	tables := []string{
		"users", "user_sessions", "api_keys", "oauth_connections", "password_reset_tokens",
		"workspaces", "workspace_members",
		"workflows", "workflow_versions",
		"executions", "node_executions",
		"credentials", "schedules", "webhook_endpoints",
		"templates", "folders", "shares", "pinned_data", "binary_data",
		"variables", "environments", "environment_variables",
		"plans", "subscriptions", "usage_records", "invoices",
		"site_settings",
	}

	for _, table := range tables {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			log.Warn().Str("table", table).Msg("NOT EXISTS")
		} else {
			log.Info().Str("table", table).Int64("rows", count).Msg("EXISTS")
		}
	}

	return nil
}
