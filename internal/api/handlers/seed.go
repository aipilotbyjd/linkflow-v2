package handlers

import (
	"net/http"
	"os"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/pkg/database"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type SeedHandler struct {
	db *gorm.DB
}

func NewSeedHandler(db *gorm.DB) *SeedHandler {
	return &SeedHandler{db: db}
}

// Seed handles GET /api/v1/seed
// Protected by SEED_SECRET environment variable
func (h *SeedHandler) Seed(w http.ResponseWriter, r *http.Request) {
	// Check for seed secret in query param or header
	secret := r.URL.Query().Get("secret")
	if secret == "" {
		secret = r.Header.Get("X-Seed-Secret")
	}

	expectedSecret := os.Getenv("SEED_SECRET")
	
	// If SEED_SECRET is not set, check if we're in development
	env := os.Getenv("APP_ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	// Require secret in non-development environments
	if env != "development" && env != "local" {
		if expectedSecret == "" {
			dto.ErrorResponse(w, http.StatusForbidden, "SEED_SECRET environment variable not set")
			return
		}
		if secret != expectedSecret {
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid seed secret")
			return
		}
	}

	// Get optional parameters
	adminEmail := r.URL.Query().Get("admin_email")
	if adminEmail == "" {
		adminEmail = "admin@linkflow.dev"
	}
	
	adminPassword := r.URL.Query().Get("admin_password")
	if adminPassword == "" {
		adminPassword = "Admin123!"
	}

	cleanFirst := r.URL.Query().Get("clean") != "false"

	log.Info().
		Str("admin_email", adminEmail).
		Bool("clean_first", cleanFirst).
		Str("environment", env).
		Msg("Starting database seeding via API")

	// Run base seeders first
	if err := database.SeedAll(h.db); err != nil {
		log.Error().Err(err).Msg("Failed to run base seeders")
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to run base seeders: "+err.Error())
		return
	}

	// Run development seeder
	cfg := database.DevSeedConfig{
		AdminEmail:    adminEmail,
		AdminPassword: adminPassword,
		CleanFirst:    cleanFirst,
	}

	if err := database.SeedDevelopment(h.db, cfg); err != nil {
		log.Error().Err(err).Msg("Failed to seed development data")
		dto.ErrorResponse(w, http.StatusInternalServerError, "failed to seed data: "+err.Error())
		return
	}

	log.Info().Msg("Database seeding completed successfully via API")

	dto.NewResponse(map[string]interface{}{
		"message": "Database seeded successfully",
		"config": map[string]interface{}{
			"admin_email":  adminEmail,
			"clean_first":  cleanFirst,
			"environment":  env,
		},
		"data_created": map[string]int{
			"users":                 5,
			"workspaces":            3,
			"workspace_members":     11,
			"folders":               13,
			"workflows":             30,
			"credentials":           12,
			"schedules":             7,
			"executions":            60,
			"environment_variables": 7,
		},
		"test_accounts": []map[string]string{
			{"email": "admin@linkflow.dev", "password": adminPassword},
			{"email": "john@linkflow.dev", "password": adminPassword},
			{"email": "jane@linkflow.dev", "password": adminPassword},
			{"email": "bob@linkflow.dev", "password": adminPassword},
			{"email": "alice@linkflow.dev", "password": adminPassword},
		},
	}).Send(w)
}

// SeedStatus handles GET /api/v1/seed/status
// Returns current seed status without requiring authentication
func (h *SeedHandler) SeedStatus(w http.ResponseWriter, r *http.Request) {
	var counts struct {
		Users       int64
		Workspaces  int64
		Workflows   int64
		Executions  int64
		Credentials int64
	}

	h.db.Table("users").Count(&counts.Users)
	h.db.Table("workspaces").Count(&counts.Workspaces)
	h.db.Table("workflows").Count(&counts.Workflows)
	h.db.Table("executions").Count(&counts.Executions)
	h.db.Table("credentials").Count(&counts.Credentials)

	env := os.Getenv("APP_ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	seedSecretSet := os.Getenv("SEED_SECRET") != ""

	dto.NewResponse(map[string]interface{}{
		"seeded": counts.Users > 0,
		"counts": map[string]int64{
			"users":       counts.Users,
			"workspaces":  counts.Workspaces,
			"workflows":   counts.Workflows,
			"executions":  counts.Executions,
			"credentials": counts.Credentials,
		},
		"environment":     env,
		"seed_secret_set": seedSecretSet,
		"seed_endpoint":   "/api/v1/seed",
	}).Send(w)
}
