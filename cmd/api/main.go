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

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/auth"
	credentialHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/credential"
	executionHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/execution"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/health"
	scheduleHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/schedule"
	webhookHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/webhook"
	workflowHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/workflow"
	workspaceHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/workspace"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	asynqAdapter "github.com/linkflow-ai/linkflow/internal/adapters/messaging/asynq"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/repositories"
	redisAdapter "github.com/linkflow-ai/linkflow/internal/adapters/persistence/redis"
	credentialCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/credential"
	executionCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/execution"
	scheduleCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/schedule"
	userCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/user"
	webhookCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/webhook"
	workflowCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workflow"
	workspaceCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workspace"
	credentialQry "github.com/linkflow-ai/linkflow/internal/core/application/query/credential"
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
			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				claims := middleware.GetUserFromContext(r.Context())
				if claims == nil {
					common.Unauthorized(w, "not authenticated")
					return
				}
				common.Success(w, map[string]interface{}{
					"user_id": claims.UserID.String(),
					"email":   claims.Email,
				})
			})

			// Workspaces
			r.Route("/workspaces", func(r chi.Router) {
				r.Post("/", wsCreateHandler.Handle)
				r.Get("/", wsListHandler.Handle)
				r.Route("/{workspaceId}", func(r chi.Router) {
					r.Get("/", wsGetHandler.Handle)
					r.Put("/", wsUpdateHandler.Handle)
					r.Delete("/", wsDeleteHandler.Handle)
					r.Get("/members", wsMembersHandler.Handle)

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
				})
			})
		})
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
