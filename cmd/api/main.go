package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"

	analyticsHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/analytics"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/auth"
	billingHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/billing"
	credentialHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/credential"
	dashboardHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/dashboard"
	executionHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/execution"
	folderHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/folder"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/health"
	userHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/user"
	scheduleHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/schedule"
	templateHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/template"
	webhookHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/webhook"
	workflowHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/workflow"
	workspaceHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/workspace"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	asynqAdapter "github.com/linkflow-ai/linkflow/internal/adapters/messaging/asynq"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/repositories"
	redisAdapter "github.com/linkflow-ai/linkflow/internal/adapters/persistence/redis"
	billingCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/billing"
	credentialCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/credential"
	executionCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/execution"
	scheduleCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/schedule"
	userCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/user"
	webhookCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/webhook"
	workflowCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workflow"
	workspaceCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workspace"
	analyticsQry "github.com/linkflow-ai/linkflow/internal/core/application/query/analytics"
	billingQry "github.com/linkflow-ai/linkflow/internal/core/application/query/billing"
	credentialQry "github.com/linkflow-ai/linkflow/internal/core/application/query/credential"
	templateQry "github.com/linkflow-ai/linkflow/internal/core/application/query/template"
	userQry "github.com/linkflow-ai/linkflow/internal/core/application/query/user"
	executionQry "github.com/linkflow-ai/linkflow/internal/core/application/query/execution"
	scheduleQry "github.com/linkflow-ai/linkflow/internal/core/application/query/schedule"
	workflowQry "github.com/linkflow-ai/linkflow/internal/core/application/query/workflow"
	workspaceQry "github.com/linkflow-ai/linkflow/internal/core/application/query/workspace"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "local"
		}
		configPath = fmt.Sprintf("configs/config.%s.yaml", env)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	var appLogger logger.Logger
	if cfg.App.IsDevelopment() {
		appLogger = logger.NewDevelopment()
	} else {
		appLogger = logger.NewDefault()
	}

	appLogger.Info().Str("app", cfg.App.Name).Str("environment", cfg.App.Environment).Msg("Starting API server")

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
		appLogger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer postgres.Close(db)

	redisClient, err := redisAdapter.NewClient(redisAdapter.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	queueClient, err := asynqAdapter.NewClient(asynqAdapter.Config{
		RedisAddr:     cfg.Redis.GetAddress(),
		RedisPassword: cfg.Redis.Password,
		RedisDB:       cfg.Redis.DB,
	})
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to create queue client")
	}
	defer queueClient.Close()

	jwtManager := jwt.NewManager(jwt.Config{
		Secret:        cfg.JWT.Secret,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
		Issuer:        cfg.JWT.Issuer,
	})
	jwtBlacklist := jwt.NewBlacklist(redisClient.Redis())
	eventBus := events.NewInMemoryBus()

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	workspaceRepo := repositories.NewWorkspaceRepository(db)
	memberRepo := repositories.NewMemberRepository(db)
	workflowRepo := repositories.NewWorkflowRepository(db)
	versionRepo := repositories.NewVersionRepository(db)
	executionRepo := repositories.NewExecutionRepository(db)
	credentialRepo := repositories.NewCredentialRepository(db)
	scheduleRepo := repositories.NewScheduleRepository(db)
	webhookRepo := repositories.NewWebhookRepository(db)
	subscriptionRepo := repositories.NewSubscriptionRepository(db)
	usageRepo := repositories.NewUsageRepository(db)
	invoiceRepo := repositories.NewInvoiceRepository(db)
	templateRepo := repositories.NewTemplateRepository(db)
	statsRepo := repositories.NewExecutionStatsRepository(db)
	folderRepo := repositories.NewFolderRepository(db)

	// Command handlers
	registerUserHandler := userCmd.NewRegisterUserHandler(userRepo, jwtManager, eventBus)
	loginUserHandler := userCmd.NewLoginUserHandler(userRepo, sessionRepo, jwtManager, eventBus)
	createWorkspaceHandler := workspaceCmd.NewCreateWorkspaceHandler(workspaceRepo, memberRepo, eventBus)
	createWorkflowHandler := workflowCmd.NewCreateWorkflowHandler(workflowRepo, versionRepo, eventBus)
	updateWorkflowHandler := workflowCmd.NewUpdateWorkflowHandler(workflowRepo, versionRepo)
	activateWorkflowHandler := workflowCmd.NewActivateWorkflowHandler(workflowRepo, eventBus)
	startExecutionHandler := executionCmd.NewStartExecutionHandler(workflowRepo, executionRepo, eventBus)
	createCredentialHandler := credentialCmd.NewCreateCredentialHandler(credentialRepo, eventBus)
	updateCredentialHandler := credentialCmd.NewUpdateCredentialHandler(credentialRepo)
	deleteCredentialHandler := credentialCmd.NewDeleteCredentialHandler(credentialRepo)
	createScheduleHandler := scheduleCmd.NewCreateScheduleHandler(scheduleRepo, eventBus)
	triggerWebhookHandler := webhookCmd.NewTriggerWebhookHandler(webhookRepo, nil)
	createSubscriptionHandler := billingCmd.NewCreateSubscriptionHandler(subscriptionRepo, eventBus)
	cancelSubscriptionHandler := billingCmd.NewCancelSubscriptionHandler(subscriptionRepo)

	// Query handlers
	getWorkspaceHandler := workspaceQry.NewGetWorkspaceHandler(workspaceRepo)
	listWorkspacesHandler := workspaceQry.NewListWorkspacesHandler(memberRepo)
	listMembersHandler := workspaceQry.NewListMembersHandler(memberRepo)
	getWorkflowHandler := workflowQry.NewGetWorkflowHandler(workflowRepo)
	getExecutionHandler := executionQry.NewGetExecutionHandler(executionRepo)
	getCredentialHandler := credentialQry.NewGetCredentialHandler(credentialRepo)
	listCredentialsHandler := credentialQry.NewListCredentialsHandler(credentialRepo)
	getScheduleHandler := scheduleQry.NewGetScheduleHandler(scheduleRepo)
	listSchedulesHandler := scheduleQry.NewListSchedulesHandler(scheduleRepo)
	getPlansHandler := billingQry.NewGetPlansHandler()
	getSubscriptionHandler := billingQry.NewGetSubscriptionHandler(subscriptionRepo)
	getUsageHandler := billingQry.NewGetUsageHandler(usageRepo)
	getInvoicesHandler := billingQry.NewGetInvoicesHandler(invoiceRepo)
	listTemplatesHandler := templateQry.NewListTemplatesHandler(templateRepo)
	getTemplateHandler := templateQry.NewGetTemplateHandler(templateRepo)
	getFeaturedTemplatesHandler := templateQry.NewGetFeaturedHandler(templateRepo)
	getByCategoryHandler := templateQry.NewGetByCategoryHandler(templateRepo)
	searchTemplatesHandler := templateQry.NewSearchTemplatesHandler(templateRepo)
	getWorkflowAnalyticsHandler := analyticsQry.NewGetWorkflowAnalyticsHandler(statsRepo)
	getWorkspaceAnalyticsHandler := analyticsQry.NewGetWorkspaceAnalyticsHandler(statsRepo)
	getUserHandler := userQry.NewGetUserHandler(userRepo)

	// HTTP handlers
	registerHandler := auth.NewRegisterHandler(registerUserHandler)
	loginHandler := auth.NewLoginHandler(loginUserHandler)
	refreshHandler := auth.NewRefreshHandler(jwtManager)
	logoutHandler := auth.NewLogoutHandler(sessionRepo, jwtBlacklist)
	healthHandler := health.NewHandler()

	// Workspace handlers
	wsCreateHandler := workspaceHandler.NewCreateHandler(createWorkspaceHandler)
	wsGetHandler := workspaceHandler.NewGetHandler(getWorkspaceHandler)
	wsListHandler := workspaceHandler.NewListHandler(listWorkspacesHandler)
	wsUpdateHandler := workspaceHandler.NewUpdateHandler(workspaceRepo)
	wsDeleteHandler := workspaceHandler.NewDeleteHandler(workspaceRepo)
	wsMembersHandler := workspaceHandler.NewListMembersHandler(listMembersHandler)

	// Workflow handlers
	wfCreateHandler := workflowHandler.NewCreateHandler(createWorkflowHandler)
	wfGetHandler := workflowHandler.NewGetHandler(getWorkflowHandler)
	wfListHandler := workflowHandler.NewListHandler(workflowQry.NewListWorkflowsHandler(workflowRepo))
	wfUpdateHandler := workflowHandler.NewUpdateHandler(updateWorkflowHandler)
	wfDeleteHandler := workflowHandler.NewDeleteHandler(workflowRepo)
	wfActivateHandler := workflowHandler.NewActivateHandler(activateWorkflowHandler)
	wfDeactivateHandler := workflowHandler.NewDeactivateHandler(workflowRepo)
	wfVersionsHandler := workflowHandler.NewListVersionsHandler(versionRepo)

	// Execution handlers
	exStartHandler := executionHandler.NewStartHandler(startExecutionHandler)
	exGetHandler := executionHandler.NewGetHandler(getExecutionHandler)
	exListHandler := executionHandler.NewListHandler(executionQry.NewListExecutionsHandler(executionRepo))
	exCancelHandler := executionHandler.NewCancelHandler(executionRepo)
	exRetryHandler := executionHandler.NewRetryHandler(startExecutionHandler, executionRepo)

	// Credential handlers
	crCreateHandler := credentialHandler.NewCreateHandler(createCredentialHandler)
	crGetHandler := credentialHandler.NewGetHandler(getCredentialHandler)
	crListHandler := credentialHandler.NewListHandler(listCredentialsHandler)
	crUpdateHandler := credentialHandler.NewUpdateHandler(updateCredentialHandler)
	crDeleteHandler := credentialHandler.NewDeleteHandler(deleteCredentialHandler)

	// Schedule handlers
	schCreateHandler := scheduleHandler.NewCreateHandler(createScheduleHandler)
	schGetHandler := scheduleHandler.NewGetHandler(getScheduleHandler)
	schListHandler := scheduleHandler.NewListHandler(listSchedulesHandler)
	schUpdateHandler := scheduleHandler.NewUpdateHandler(scheduleRepo)
	schPauseHandler := scheduleHandler.NewPauseHandler(scheduleRepo)
	schResumeHandler := scheduleHandler.NewResumeHandler(scheduleRepo)
	schDeleteHandler := scheduleHandler.NewDeleteHandler(scheduleRepo)

	// Webhook handlers
	whTriggerHandler := webhookHandler.NewTriggerHandler(triggerWebhookHandler)

	// Billing handlers
	blGetPlansHandler := billingHandler.NewGetPlansHandler(getPlansHandler)
	blGetSubscriptionHandler := billingHandler.NewGetSubscriptionHandler(getSubscriptionHandler)
	blCreateSubscriptionHandler := billingHandler.NewCreateSubscriptionHandler(createSubscriptionHandler)
	blCancelSubscriptionHandler := billingHandler.NewCancelSubscriptionHandler(cancelSubscriptionHandler)
	blGetUsageHandler := billingHandler.NewGetUsageHandler(getUsageHandler)
	blGetInvoicesHandler := billingHandler.NewGetInvoicesHandler(getInvoicesHandler)

	// Template handlers
	tplListHandler := templateHandler.NewListHandler(listTemplatesHandler)
	tplGetHandler := templateHandler.NewGetHandler(getTemplateHandler)
	tplFeaturedHandler := templateHandler.NewFeaturedHandler(getFeaturedTemplatesHandler)
	tplCategoriesHandler := templateHandler.NewCategoriesHandler()
	tplByCategoryHandler := templateHandler.NewByCategoryHandler(getByCategoryHandler)
	tplSearchHandler := templateHandler.NewSearchHandler(searchTemplatesHandler)
	tplUseHandler := templateHandler.NewUseHandler(templateRepo, workflowRepo)

	// Analytics handlers
	anWorkflowHandler := analyticsHandler.NewWorkflowAnalyticsHandler(getWorkflowAnalyticsHandler)
	anWorkspaceHandler := analyticsHandler.NewWorkspaceAnalyticsHandler(getWorkspaceAnalyticsHandler)

	// Dashboard handler
	dashHandler := dashboardHandler.NewDashboardHandler(workflowRepo, executionRepo, scheduleRepo)

	// User handler
	usrGetHandler := userHandler.NewGetCurrentUserHandler(getUserHandler)

	// Folder handlers
	fldCreateHandler := folderHandler.NewCreateFolderHandler(folderRepo)
	fldGetHandler := folderHandler.NewGetFolderHandler(folderRepo)
	fldListHandler := folderHandler.NewListFoldersHandler(folderRepo)
	fldUpdateHandler := folderHandler.NewUpdateFolderHandler(folderRepo)
	fldDeleteHandler := folderHandler.NewDeleteFolderHandler(folderRepo)
	fldTreeHandler := folderHandler.NewGetTreeHandler(folderRepo)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logging(appLogger))
	r.Use(middleware.Recovery(appLogger))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   cfg.App.CorsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(corsHandler.Handler)

	r.Get("/health", healthHandler.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler.Health)

		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", registerHandler.Handle)
			r.Post("/login", loginHandler.Handle)
			r.Post("/refresh", refreshHandler.Handle)
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(jwtManager))
				r.Post("/logout", logoutHandler.Handle)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtManager))

			// Current user
			r.Get("/me", usrGetHandler.Handle)

			// Workspaces
			r.Route("/workspaces", func(r chi.Router) {
				r.Post("/", wsCreateHandler.Handle)
				r.Get("/", wsListHandler.Handle)
				r.Route("/{workspaceId}", func(r chi.Router) {
					r.Get("/", wsGetHandler.Handle)
					r.Put("/", wsUpdateHandler.Handle)
					r.Delete("/", wsDeleteHandler.Handle)
					r.Get("/members", wsMembersHandler.Handle)

					// Dashboard
					r.Get("/dashboard", dashHandler.Handle)

					// Workflows
					r.Route("/workflows", func(r chi.Router) {
						r.Post("/", wfCreateHandler.Handle)
						r.Get("/", wfListHandler.Handle)
						r.Route("/{workflowId}", func(r chi.Router) {
							r.Get("/", wfGetHandler.Handle)
							r.Put("/", wfUpdateHandler.Handle)
							r.Delete("/", wfDeleteHandler.Handle)
							r.Post("/activate", wfActivateHandler.Handle)
							r.Post("/deactivate", wfDeactivateHandler.Handle)
							r.Get("/versions", wfVersionsHandler.Handle)
						})
					})

					// Executions
					r.Route("/executions", func(r chi.Router) {
						r.Post("/", exStartHandler.Handle)
						r.Get("/", exListHandler.Handle)
						r.Route("/{executionId}", func(r chi.Router) {
							r.Get("/", exGetHandler.Handle)
							r.Post("/cancel", exCancelHandler.Handle)
							r.Post("/retry", exRetryHandler.Handle)
						})
					})

					// Credentials
					r.Route("/credentials", func(r chi.Router) {
						r.Post("/", crCreateHandler.Handle)
						r.Get("/", crListHandler.Handle)
						r.Route("/{credentialId}", func(r chi.Router) {
							r.Get("/", crGetHandler.Handle)
							r.Put("/", crUpdateHandler.Handle)
							r.Delete("/", crDeleteHandler.Handle)
						})
					})

					// Schedules
					r.Route("/schedules", func(r chi.Router) {
						r.Post("/", schCreateHandler.Handle)
						r.Get("/", schListHandler.Handle)
						r.Route("/{scheduleId}", func(r chi.Router) {
							r.Get("/", schGetHandler.Handle)
							r.Put("/", schUpdateHandler.Handle)
							r.Delete("/", schDeleteHandler.Handle)
							r.Post("/pause", schPauseHandler.Handle)
							r.Post("/resume", schResumeHandler.Handle)
						})
					})

					// Billing (workspace scoped)
					r.Route("/billing", func(r chi.Router) {
						r.Get("/subscription", blGetSubscriptionHandler.Handle)
						r.Post("/subscription", blCreateSubscriptionHandler.Handle)
						r.Delete("/subscription", blCancelSubscriptionHandler.Handle)
						r.Get("/usage", blGetUsageHandler.Handle)
						r.Get("/invoices", blGetInvoicesHandler.Handle)
					})

					// Analytics (workspace scoped)
					r.Get("/analytics", anWorkspaceHandler.Handle)
					r.Get("/workflows/{workflowId}/analytics", anWorkflowHandler.Handle)
				})
			})

			// Billing plans (not workspace scoped)
			r.Get("/billing/plans", blGetPlansHandler.Handle)

			// Templates (not workspace scoped for browsing)
			r.Route("/templates", func(r chi.Router) {
				r.Get("/", tplListHandler.Handle)
				r.Get("/featured", tplFeaturedHandler.Handle)
				r.Get("/categories", tplCategoriesHandler.Handle)
				r.Get("/categories/{category}", tplByCategoryHandler.Handle)
				r.Get("/search", tplSearchHandler.Handle)
				r.Get("/{templateId}", tplGetHandler.Handle)
			})
		})
	})

	// Template use (requires workspace context) - add under workspace routes
	r.Route("/api/v1/workspaces/{workspaceId}/templates/{templateId}/use", func(r chi.Router) {
		r.Use(middleware.Auth(jwtManager))
		r.Post("/", tplUseHandler.Handle)
	})

	// Webhook trigger endpoint (public)
	r.Post("/webhooks/{endpointId}", whTriggerHandler.Handle)
	r.Get("/webhooks/{endpointId}", whTriggerHandler.Handle)

	server := &http.Server{
		Addr:         cfg.Server.GetAddress(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		appLogger.Info().Str("address", cfg.Server.GetAddress()).Msg("HTTP server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error().Err(err).Msg("Server shutdown error")
	}
	appLogger.Info().Msg("Server stopped")
}
