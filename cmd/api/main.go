package main

import (
	"github.com/linkflow-ai/linkflow/internal/api"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	"github.com/linkflow-ai/linkflow/internal/pkg/database"
	"github.com/linkflow-ai/linkflow/internal/pkg/logger"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger
	logger.Init(cfg.App.Environment, cfg.App.Debug)

	log.Info().
		Str("app", cfg.App.Name).
		Str("env", cfg.App.Environment).
		Msg("Starting API server")

	// Connect to database
	db, err := database.NewGormDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	// Seed plans
	if err := database.SeedPlans(db); err != nil {
		log.Fatal().Err(err).Msg("Failed to seed plans")
	}

	// Connect to Redis
	redisClient, err := pkgredis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	// Initialize queue client
	queueClient := queue.NewClient(&cfg.Redis)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	workspaceRepo := repositories.NewWorkspaceRepository(db)
	memberRepo := repositories.NewWorkspaceMemberRepository(db)
	invitationRepo := repositories.NewWorkspaceInvitationRepository(db)
	workflowRepo := repositories.NewWorkflowRepository(db)
	versionRepo := repositories.NewWorkflowVersionRepository(db)
	executionRepo := repositories.NewExecutionRepository(db)
	nodeExecutionRepo := repositories.NewNodeExecutionRepository(db)
	credentialRepo := repositories.NewCredentialRepository(db)
	scheduleRepo := repositories.NewScheduleRepository(db)
	planRepo := repositories.NewPlanRepository(db)
	subscriptionRepo := repositories.NewSubscriptionRepository(db)
	usageRepo := repositories.NewUsageRepository(db)
	invoiceRepo := repositories.NewInvoiceRepository(db)

	// Initialize crypto
	jwtManager := crypto.NewJWTManager(crypto.JWTConfig{
		Secret:        cfg.JWT.Secret,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
		Issuer:        cfg.JWT.Issuer,
	})

	encryptor, err := crypto.NewEncryptor(cfg.JWT.Secret[:32])
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create encryptor")
	}

	otpManager := crypto.NewOTPManager(cfg.App.Name)

	// Initialize services
	authSvc := services.NewAuthService(userRepo, sessionRepo, jwtManager, otpManager, encryptor)
	userSvc := services.NewUserService(userRepo)
	workspaceSvc := services.NewWorkspaceService(workspaceRepo, memberRepo, invitationRepo)
	workflowSvc := services.NewWorkflowService(workflowRepo, versionRepo)
	executionSvc := services.NewExecutionService(executionRepo, nodeExecutionRepo, workflowRepo)
	credentialSvc := services.NewCredentialService(credentialRepo, encryptor)
	scheduleSvc := services.NewScheduleService(scheduleRepo)
	billingSvc := services.NewBillingService(planRepo, subscriptionRepo, usageRepo, invoiceRepo, workspaceRepo)

	// Create server
	server := api.NewServer(
		cfg,
		&api.Services{
			Auth:       authSvc,
			User:       userSvc,
			Workspace:  workspaceSvc,
			Workflow:   workflowSvc,
			Execution:  executionSvc,
			Credential: credentialSvc,
			Schedule:   scheduleSvc,
			Billing:    billingSvc,
		},
		jwtManager,
		redisClient,
		queueClient,
	)

	// Start server
	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server error")
	}
}
