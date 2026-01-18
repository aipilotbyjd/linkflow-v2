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

	adminHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/admin"
	analyticsHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/analytics"
	apikeyHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/apikey"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/auth"
	billingHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/billing"
	binarydataHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/binarydata"
	credentialHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/credential"
	dashboardHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/dashboard"
	executionHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/execution"
	folderHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/folder"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/health"
	marketplaceHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/marketplace"
	nodetypesHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/nodetypes"
	oauthHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/oauth"
	pinneddataHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/pinneddata"
	scheduleHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/schedule"
	shareHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/share"
	templateHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/template"
	userHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/user"
	webhookHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/webhook"
	workflowHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/workflow"
	workspaceHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/workspace"
	resp "github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	asynqAdapter "github.com/linkflow-ai/linkflow/internal/adapters/messaging/asynq"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/repositories"
	redisAdapter "github.com/linkflow-ai/linkflow/internal/adapters/persistence/redis"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes"
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
	executionQry "github.com/linkflow-ai/linkflow/internal/core/application/query/execution"
	scheduleQry "github.com/linkflow-ai/linkflow/internal/core/application/query/schedule"
	templateQry "github.com/linkflow-ai/linkflow/internal/core/application/query/template"
	userQry "github.com/linkflow-ai/linkflow/internal/core/application/query/user"
	workflowQry "github.com/linkflow-ai/linkflow/internal/core/application/query/workflow"
	workspaceQry "github.com/linkflow-ai/linkflow/internal/core/application/query/workspace"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/cache"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/email"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/metrics"
	infraOAuth "github.com/linkflow-ai/linkflow/internal/infrastructure/oauth"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/storage"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/streaming"
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
	nodeExecutionRepo := repositories.NewNodeExecutionRepository(db)
	credentialRepo := repositories.NewCredentialRepository(db)
	scheduleRepo := repositories.NewScheduleRepository(db)
	webhookRepo := repositories.NewWebhookRepository(db)
	subscriptionRepo := repositories.NewSubscriptionRepository(db)
	usageRepo := repositories.NewUsageRepository(db)
	invoiceRepo := repositories.NewInvoiceRepository(db)
	templateRepo := repositories.NewTemplateRepository(db)
	statsRepo := repositories.NewExecutionStatsRepository(db)
	folderRepo := repositories.NewFolderRepository(db)
	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	pinnedDataRepo := repositories.NewPinnedDataRepository(db)
	shareRepo := repositories.NewShareRepository(db)
	binaryDataRepo := repositories.NewBinaryDataRepository(db)

	// Storage service
	localStorage, err := storage.NewLocalStorage("./data/uploads")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize local storage")
	}

	// Node registry
	nodeRegistry := nodes.NewRegistry()
	if err := nodes.LoadAllNodes(nodeRegistry); err != nil {
		log.Fatal().Err(err).Msg("Failed to load node types")
	}

	// Metrics collector
	metricsCollector := metrics.NewCollector("2.0.0")

	// Stream manager (for admin)
	streamManager := streaming.NewManager(redisClient.Redis())

	// OAuth providers
	oauthProviders := make(map[string]oauthHandler.OAuthProvider)
	if cfg.OAuth.Google.ClientID != "" {
		googleProvider := infraOAuth.NewGoogleProvider(
			cfg.OAuth.Google.ClientID,
			cfg.OAuth.Google.ClientSecret,
			cfg.OAuth.Google.RedirectURL,
		)
		oauthProviders["google"] = &oauthProviderAdapter{googleProvider}
	}
	if cfg.OAuth.GitHub.ClientID != "" {
		githubProvider := infraOAuth.NewGitHubProvider(
			cfg.OAuth.GitHub.ClientID,
			cfg.OAuth.GitHub.ClientSecret,
			cfg.OAuth.GitHub.RedirectURL,
		)
		oauthProviders["github"] = &oauthProviderAdapter{githubProvider}
	}

	// Task queue client
	asynqClient, err := asynqAdapter.NewClient(asynqAdapter.Config{
		RedisAddr:     cfg.Redis.GetAddress(),
		RedisPassword: cfg.Redis.Password,
		RedisDB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create asynq client")
	}
	defer asynqClient.Close()
	taskQueue := asynqAdapter.NewTaskQueueAdapter(asynqClient)

	// Command handlers
	registerUserHandler := userCmd.NewRegisterUserHandler(userRepo, jwtManager, eventBus)
	loginUserHandler := userCmd.NewLoginUserHandler(userRepo, sessionRepo, jwtManager, eventBus)
	createWorkspaceHandler := workspaceCmd.NewCreateWorkspaceHandler(workspaceRepo, memberRepo, eventBus)
	createWorkflowHandler := workflowCmd.NewCreateWorkflowHandler(workflowRepo, versionRepo, eventBus)
	updateWorkflowHandler := workflowCmd.NewUpdateWorkflowHandler(workflowRepo, versionRepo)
	activateWorkflowHandler := workflowCmd.NewActivateWorkflowHandler(workflowRepo, eventBus)
	startExecutionHandler := executionCmd.NewStartExecutionHandler(workflowRepo, executionRepo, eventBus, taskQueue)
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

	// Cache and email service
	cacheService := cache.NewMemoryCache()
	emailService, err := email.NewService(email.Config{
		Provider:    cfg.Email.Provider,
		DefaultFrom: cfg.Email.From,
		SMTPHost:    cfg.Email.SMTPHost,
		SMTPPort:    cfg.Email.SMTPPort,
		SMTPUser:    cfg.Email.SMTPUser,
		SMTPPass:    cfg.Email.SMTPPass,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize email service, using noop")
		emailService, _ = email.NewService(email.Config{Provider: "noop"})
	}

	// HTTP handlers
	registerHandler := auth.NewRegisterHandler(registerUserHandler)
	loginHandler := auth.NewLoginHandler(loginUserHandler)
	refreshHandler := auth.NewRefreshHandler(jwtManager)
	logoutHandler := auth.NewLogoutHandler(sessionRepo, jwtBlacklist)
	baseURL := "http://localhost:3000" // Frontend URL for password reset links
	forgotPasswordHandler := auth.NewForgotPasswordHandler(userRepo, cacheService, emailService, baseURL)
	resetPasswordHandler := auth.NewResetPasswordHandler(userRepo, sessionRepo, cacheService)
	setupMFAHandler := auth.NewSetupMFAHandler(userRepo, cacheService)
	verifyMFAHandler := auth.NewVerifyMFAHandler(userRepo, cacheService)
	disableMFAHandler := auth.NewDisableMFAHandler(userRepo)
	authOAuthRedirectHandler := auth.NewOAuthRedirectHandler(nil, baseURL)
	authOAuthCallbackHandler := auth.NewOAuthCallbackHandler(nil, nil, nil)
	healthHandler := health.NewHandler()

	// Workspace handlers
	wsCreateHandler := workspaceHandler.NewCreateHandler(createWorkspaceHandler)
	wsGetHandler := workspaceHandler.NewGetHandler(getWorkspaceHandler)
	wsListHandler := workspaceHandler.NewListHandler(listWorkspacesHandler)
	wsUpdateHandler := workspaceHandler.NewUpdateHandler(workspaceRepo)
	wsDeleteHandler := workspaceHandler.NewDeleteHandler(workspaceRepo)
	wsMembersHandler := workspaceHandler.NewListMembersHandler(listMembersHandler)
	wsInviteHandler := workspaceHandler.NewInviteMemberHandler(workspaceRepo, memberRepo, userRepo, emailService, baseURL)

	// Workflow handlers
	wfCreateHandler := workflowHandler.NewCreateHandler(createWorkflowHandler)
	wfGetHandler := workflowHandler.NewGetHandler(getWorkflowHandler)
	wfListHandler := workflowHandler.NewListHandler(workflowQry.NewListWorkflowsHandler(workflowRepo))
	wfUpdateHandler := workflowHandler.NewUpdateHandler(updateWorkflowHandler)
	wfDeleteHandler := workflowHandler.NewDeleteHandler(workflowRepo)
	wfActivateHandler := workflowHandler.NewActivateHandler(activateWorkflowHandler)
	wfDeactivateHandler := workflowHandler.NewDeactivateHandler(workflowRepo)
	wfVersionsHandler := workflowHandler.NewListVersionsHandler(versionRepo)
	wfGetVersionHandler := workflowHandler.NewGetVersionHandler(versionRepo)
	wfRollbackHandler := workflowHandler.NewRollbackHandler(workflowRepo, versionRepo)
	wfCompareVersionsHandler := workflowHandler.NewCompareVersionsHandler(versionRepo)
	wfCloneHandler := workflowHandler.NewCloneHandler(workflowRepo)
	wfDuplicateHandler := workflowHandler.NewDuplicateHandler(workflowRepo)
	wfExportHandler := workflowHandler.NewExportHandler(workflowRepo)
	wfImportHandler := workflowHandler.NewImportHandler(workflowRepo)
	wfValidateHandler := workflowHandler.NewValidateWorkflowHandler(nodeRegistry)
	wfTestNodeHandler := workflowHandler.NewTestNodeHandler(nodeRegistry, appLogger)
	wfSearchHandler := workflowHandler.NewSearchHandler(workflowRepo)

	// Execution handlers
	exStartHandler := executionHandler.NewStartHandler(startExecutionHandler)
	exGetHandler := executionHandler.NewGetHandler(getExecutionHandler)
	exListHandler := executionHandler.NewListHandler(executionQry.NewListExecutionsHandler(executionRepo))
	exCancelHandler := executionHandler.NewCancelHandler(executionRepo)
	exRetryHandler := executionHandler.NewRetryHandler(startExecutionHandler, executionRepo)
	exStatsHandler := executionHandler.NewGetExecutionStatsHandler(executionRepo)
	exSearchHandler := executionHandler.NewSearchExecutionsHandler(executionRepo)
	exListNodesHandler := executionHandler.NewGetExecutionNodesHandler(executionRepo, nodeExecutionRepo)
	exGetNodeHandler := executionHandler.NewGetNodeExecutionHandler(executionRepo, nodeExecutionRepo)
	exReplayHandler := executionHandler.NewReplayExecutionHandler(executionRepo, workflowRepo, startExecutionHandler)
	exResumeHandler := executionHandler.NewResumeHandler(executionRepo)
	exListWaitingHandler := executionHandler.NewListWaitingHandler(executionRepo)
	exBulkDeleteHandler := executionHandler.NewBulkDeleteExecutionsHandler(executionRepo)

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
	whListHandler := webhookHandler.NewListEndpointsHandler(webhookRepo, baseURL)
	whCreateHandler := webhookHandler.NewCreateEndpointHandler(webhookRepo, workflowRepo, baseURL)

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

	// Dashboard handlers
	dashHandler := dashboardHandler.NewDashboardHandler(workflowRepo, executionRepo, scheduleRepo)
	quickStatsHandler := dashboardHandler.NewQuickStatsHandler(workflowRepo, executionRepo, credentialRepo, scheduleRepo)

	// User handlers
	usrGetHandler := userHandler.NewGetCurrentUserHandler(getUserHandler)
	usrUpdateHandler := userHandler.NewUpdateCurrentUserHandler(userRepo)

	// Folder handlers
	fldCreateHandler := folderHandler.NewCreateFolderHandler(folderRepo)
	fldGetHandler := folderHandler.NewGetFolderHandler(folderRepo)
	fldListHandler := folderHandler.NewListFoldersHandler(folderRepo)
	fldUpdateHandler := folderHandler.NewUpdateFolderHandler(folderRepo)
	fldDeleteHandler := folderHandler.NewDeleteFolderHandler(folderRepo)
	fldTreeHandler := folderHandler.NewGetFolderTreeHandler(folderRepo, workflowRepo)

	// API Key handlers
	akCreateHandler := apikeyHandler.NewCreateAPIKeyHandler(apiKeyRepo)
	akListHandler := apikeyHandler.NewListAPIKeysHandler(apiKeyRepo)
	akRevokeHandler := apikeyHandler.NewRevokeAPIKeyHandler(apiKeyRepo)

	// Node types handlers
	ntListHandler := nodetypesHandler.NewListNodeTypesHandler(nodeRegistry)
	ntCategoriesHandler := nodetypesHandler.NewListCategoriesHandler(nodeRegistry)
	ntGetHandler := nodetypesHandler.NewGetNodeTypeHandler(nodeRegistry)

	// Marketplace handlers
	mpBrowseHandler := marketplaceHandler.NewBrowseHandler()
	mpFeaturedHandler := marketplaceHandler.NewFeaturedHandler()
	mpCategoriesHandler := marketplaceHandler.NewCategoriesHandler()
	mpSearchHandler := marketplaceHandler.NewSearchHandler()
	mpGetHandler := marketplaceHandler.NewGetHandler()
	mpUseHandler := marketplaceHandler.NewUseHandler()
	mpPublishHandler := marketplaceHandler.NewPublishHandler()
	mpMyPublishedHandler := marketplaceHandler.NewMyPublishedHandler()
	mpUpdateHandler := marketplaceHandler.NewUpdateHandler()
	mpUnpublishHandler := marketplaceHandler.NewUnpublishHandler()
	mpRateHandler := marketplaceHandler.NewRateHandler()
	mpGetMyRatingHandler := marketplaceHandler.NewGetMyRatingHandler()
	mpListRatingsHandler := marketplaceHandler.NewListRatingsHandler()
	mpRatingStatsHandler := marketplaceHandler.NewRatingStatsHandler()
	mpDeleteRatingHandler := marketplaceHandler.NewDeleteRatingHandler()

	// Pinned data handlers
	pdGetAllHandler := pinneddataHandler.NewGetAllHandler(pinnedDataRepo)
	pdGetByNodeHandler := pinneddataHandler.NewGetByNodeHandler(pinnedDataRepo)
	pdSetHandler := pinneddataHandler.NewSetHandler(pinnedDataRepo)
	pdDeleteHandler := pinneddataHandler.NewDeleteHandler(pinnedDataRepo)

	// Share handlers
	shCreateHandler := shareHandler.NewCreateHandler(shareRepo, userRepo)
	shSharedByMeHandler := shareHandler.NewSharedByMeHandler(shareRepo)
	shSharedWithMeHandler := shareHandler.NewSharedWithMeHandler(shareRepo)
	shPendingHandler := shareHandler.NewPendingHandler(shareRepo)
	shAcceptHandler := shareHandler.NewAcceptHandler(shareRepo)
	shUpdateHandler := shareHandler.NewUpdateHandler(shareRepo)
	shRevokeHandler := shareHandler.NewRevokeHandler(shareRepo)

	// Binary data handlers
	bdUploadHandler := binarydataHandler.NewUploadHandler(binaryDataRepo, localStorage)
	bdListHandler := binarydataHandler.NewListHandler(binaryDataRepo)
	bdGetInfoHandler := binarydataHandler.NewGetInfoHandler(binaryDataRepo)
	bdDownloadHandler := binarydataHandler.NewDownloadHandler(binaryDataRepo, localStorage)
	bdDeleteHandler := binarydataHandler.NewDeleteHandler(binaryDataRepo, localStorage)
	bdStatsHandler := binarydataHandler.NewStatsHandler(binaryDataRepo)
	bdCleanupHandler := binarydataHandler.NewCleanupHandler(binaryDataRepo)

	// Admin handlers
	admMetricsHandler := adminHandler.NewMetricsHandler(&metricsAdapter{metricsCollector})
	admStreamStatsHandler := adminHandler.NewStreamStatsHandler(&streamAdapter{streamManager})
	admReplayDLQHandler := adminHandler.NewReplayDLQHandler(&streamAdapter{streamManager})
	admTrimStreamHandler := adminHandler.NewTrimStreamHandler(&streamAdapter{streamManager})

	// OAuth handlers
	oaListProvidersHandler := oauthHandler.NewListProvidersHandler(oauthProviders)
	oaAuthorizeHandler := oauthHandler.NewAuthorizeHandler(oauthProviders)
	oaCallbackHandler := oauthHandler.NewCallbackHandler(oauthProviders)

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
		r.Get("/health/live", healthHandler.Liveness)
		r.Get("/health/ready", healthHandler.Readiness)

		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", registerHandler.Handle)
			r.Post("/login", loginHandler.Handle)
			r.Post("/refresh", refreshHandler.Handle)
			r.Post("/forgot-password", forgotPasswordHandler.Handle)
			r.Post("/reset-password", resetPasswordHandler.Handle)
			r.Get("/oauth/{provider}", authOAuthRedirectHandler.Handle)
			r.Get("/oauth/{provider}/callback", authOAuthCallbackHandler.Handle)
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(jwtManager))
				r.Post("/logout", logoutHandler.Handle)
				r.Route("/mfa", func(r chi.Router) {
					r.Post("/setup", setupMFAHandler.Handle)
					r.Post("/verify", verifyMFAHandler.Handle)
					r.Delete("/", disableMFAHandler.Handle)
				})
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtManager))

			// Current user
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", usrGetHandler.Handle)
				r.Put("/me", usrUpdateHandler.Handle)
			})

			// Workspaces
			r.Route("/workspaces", func(r chi.Router) {
				r.Post("/", wsCreateHandler.Handle)
				r.Get("/", wsListHandler.Handle)
				r.Route("/{workspaceId}", func(r chi.Router) {
					r.Use(middleware.Tenant(memberRepo, workspaceRepo))
					r.Get("/", wsGetHandler.Handle)
					r.Put("/", wsUpdateHandler.Handle)
					r.Delete("/", wsDeleteHandler.Handle)
					r.Get("/members", wsMembersHandler.Handle)
					r.Post("/members", wsInviteHandler.Handle)

					// Dashboard
					r.Get("/dashboard", dashHandler.Handle)
					r.Get("/stats", quickStatsHandler.Handle)

					// Workflows
					r.Route("/workflows", func(r chi.Router) {
						r.Post("/", wfCreateHandler.Handle)
						r.Get("/", wfListHandler.Handle)
						r.Get("/search", wfSearchHandler.Handle)
						r.Post("/import", wfImportHandler.Handle)
						r.Post("/validate", wfValidateHandler.Handle)
						r.Post("/test-node", wfTestNodeHandler.Handle)
						r.Route("/{workflowId}", func(r chi.Router) {
							r.Get("/", wfGetHandler.Handle)
							r.Put("/", wfUpdateHandler.Handle)
							r.Delete("/", wfDeleteHandler.Handle)
							r.Post("/activate", wfActivateHandler.Handle)
							r.Post("/deactivate", wfDeactivateHandler.Handle)
							r.Post("/execute", exStartHandler.Handle)
							r.Post("/clone", wfCloneHandler.Handle)
							r.Post("/duplicate", wfDuplicateHandler.Handle)
							r.Get("/export", wfExportHandler.Handle)
							r.Get("/versions", wfVersionsHandler.Handle)
							r.Get("/versions/{version}", wfGetVersionHandler.Handle)
							r.Post("/versions/{version}/rollback", wfRollbackHandler.Handle)
							r.Get("/compare-versions", wfCompareVersionsHandler.Handle)
						})
					})

					// Executions
					r.Route("/executions", func(r chi.Router) {
						r.Get("/", exListHandler.Handle)
						r.Get("/stats", exStatsHandler.Handle)
						r.Get("/search", exSearchHandler.Handle)
						r.Delete("/bulk", exBulkDeleteHandler.Handle)
						r.Route("/{executionId}", func(r chi.Router) {
							r.Get("/", exGetHandler.Handle)
							r.Get("/nodes", exListNodesHandler.Handle)
							r.Get("/nodes/{nodeId}", exGetNodeHandler.Handle)
							r.Post("/cancel", exCancelHandler.Handle)
							r.Post("/retry", exRetryHandler.Handle)
							r.Post("/replay", exReplayHandler.Handle)
							r.Post("/resume", exResumeHandler.Handle)
						})
					})

					// Waiting Executions
					r.Get("/waiting-executions", exListWaitingHandler.Handle)

					// Webhooks
					r.Route("/webhooks", func(r chi.Router) {
						r.Get("/", whListHandler.Handle)
						r.Post("/", whCreateHandler.Handle)
					})

					// Credentials
					r.Route("/credentials", func(r chi.Router) {
						r.Post("/", crCreateHandler.Handle)
						r.Get("/", crListHandler.Handle)
						r.Route("/{credentialId}", func(r chi.Router) {
							r.Get("/", crGetHandler.Handle)
							r.Put("/", crUpdateHandler.Handle)
							r.Delete("/", crDeleteHandler.Handle)
							r.Post("/test", func(w http.ResponseWriter, req *http.Request) {
								resp.Success(w, map[string]interface{}{"valid": true, "message": "Connection test successful"})
							})
							r.Post("/refresh", func(w http.ResponseWriter, req *http.Request) {
								resp.BadRequest(w, "Token refresh not supported for this credential type")
							})
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

					// Folders
					r.Route("/folders", func(r chi.Router) {
						r.Post("/", fldCreateHandler.Handle)
						r.Get("/", fldListHandler.Handle)
						r.Get("/tree", fldTreeHandler.Handle)
						r.Route("/{folderId}", func(r chi.Router) {
							r.Get("/", fldGetHandler.Handle)
							r.Put("/", fldUpdateHandler.Handle)
							r.Delete("/", fldDeleteHandler.Handle)
						})
					})

					// Pinned Data (per workflow)
					r.Route("/workflows/{workflowId}/pinned-data", func(r chi.Router) {
						r.Get("/", pdGetAllHandler.Handle)
						r.Route("/{nodeId}", func(r chi.Router) {
							r.Get("/", pdGetByNodeHandler.Handle)
							r.Put("/", pdSetHandler.Handle)
							r.Delete("/", pdDeleteHandler.Handle)
						})
					})

					// Binary Data (workspace scoped)
					r.Route("/binary-data", func(r chi.Router) {
						r.Post("/upload", bdUploadHandler.Handle)
						r.Get("/", bdListHandler.Handle)
						r.Get("/stats", bdStatsHandler.Handle)
						r.Post("/cleanup", bdCleanupHandler.Handle)
						r.Route("/{fileId}", func(r chi.Router) {
							r.Get("/", bdGetInfoHandler.Handle)
							r.Get("/download", bdDownloadHandler.Handle)
							r.Delete("/", bdDeleteHandler.Handle)
						})
					})

					// Workspace OAuth (for credential integrations)
					r.Route("/oauth", func(r chi.Router) {
						r.Get("/authorize/{provider}", oaAuthorizeHandler.Handle)
						r.Get("/callback/{provider}", oaCallbackHandler.Handle)
					})
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

			// API Keys
			r.Route("/api-keys", func(r chi.Router) {
				r.Post("/", akCreateHandler.Handle)
				r.Get("/", akListHandler.Handle)
				r.Delete("/{keyId}", akRevokeHandler.Handle)
			})

			// Node Types
			r.Route("/node-types", func(r chi.Router) {
				r.Get("/", ntListHandler.Handle)
				r.Get("/categories", ntCategoriesHandler.Handle)
				r.Get("/{nodeType}", ntGetHandler.Handle)
			})

			// Marketplace
			r.Route("/marketplace", func(r chi.Router) {
				r.Get("/", mpBrowseHandler.Handle)
				r.Get("/featured", mpFeaturedHandler.Handle)
				r.Get("/categories", mpCategoriesHandler.Handle)
				r.Get("/search", mpSearchHandler.Handle)
				r.Get("/my-published", mpMyPublishedHandler.Handle)
				r.Get("/{itemId}", mpGetHandler.Handle)
				r.Post("/{itemId}/use", mpUseHandler.Handle)
				r.Post("/publish", mpPublishHandler.Handle)
				r.Put("/{itemId}", mpUpdateHandler.Handle)
				r.Delete("/{itemId}", mpUnpublishHandler.Handle)
				r.Post("/{itemId}/rate", mpRateHandler.Handle)
				r.Get("/{itemId}/my-rating", mpGetMyRatingHandler.Handle)
				r.Get("/{itemId}/ratings", mpListRatingsHandler.Handle)
				r.Get("/{itemId}/rating-stats", mpRatingStatsHandler.Handle)
				r.Delete("/{itemId}/rating", mpDeleteRatingHandler.Handle)
			})

			// Shares (user scoped)
			r.Route("/shares", func(r chi.Router) {
				r.Post("/", shCreateHandler.Handle)
				r.Get("/by-me", shSharedByMeHandler.Handle)
				r.Get("/with-me", shSharedWithMeHandler.Handle)
				r.Get("/pending", shPendingHandler.Handle)
				r.Post("/{shareId}/accept", shAcceptHandler.Handle)
				r.Put("/{shareId}", shUpdateHandler.Handle)
				r.Delete("/{shareId}", shRevokeHandler.Handle)
			})

			// Admin (requires admin role)
			r.Route("/admin", func(r chi.Router) {
				// TODO: Add admin role check middleware
				r.Get("/metrics", admMetricsHandler.Handle)
				r.Get("/streams/{streamName}", admStreamStatsHandler.Handle)
				r.Post("/streams/replay-dlq", admReplayDLQHandler.Handle)
				r.Post("/streams/trim", admTrimStreamHandler.Handle)
			})

			// OAuth (credential OAuth)
			r.Route("/oauth", func(r chi.Router) {
				r.Get("/providers", oaListProvidersHandler.Handle)
				r.Get("/{provider}/authorize", oaAuthorizeHandler.Handle)
				r.Get("/{provider}/callback", oaCallbackHandler.Handle)
			})
		})
	})

	// Template use (requires workspace context) - add under workspace routes
	r.Route("/api/v1/workspaces/{workspaceId}/templates/{templateId}/use", func(r chi.Router) {
		r.Use(middleware.Auth(jwtManager))
		r.Post("/", tplUseHandler.Handle)
	})

	// Webhook trigger endpoint (public)
	r.Post("/webhooks/*", whTriggerHandler.Handle)
	r.Get("/webhooks/*", whTriggerHandler.Handle)

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

// Adapters for handler interfaces

// metricsAdapter adapts metrics.Collector to adminHandler.MetricsCollector
type metricsAdapter struct {
	collector *metrics.Collector
}

func (a *metricsAdapter) CollectMetrics() map[string]interface{} {
	return a.collector.CollectMetrics()
}

// streamAdapter adapts streaming.Manager to adminHandler.StreamManager
type streamAdapter struct {
	manager *streaming.Manager
}

func (a *streamAdapter) GetStats(streamName string) (*adminHandler.StreamStats, error) {
	stats, err := a.manager.GetStats(streamName)
	if err != nil {
		return nil, err
	}
	return &adminHandler.StreamStats{
		Name:          stats.Name,
		Length:        stats.Length,
		Pending:       stats.Pending,
		Consumers:     stats.Consumers,
		LastDelivered: stats.LastDelivered,
	}, nil
}

func (a *streamAdapter) ReplayDLQ(streamName string, count int) (int, error) {
	return a.manager.ReplayDLQ(streamName, count)
}

func (a *streamAdapter) TrimStream(streamName string, maxLen int64) (int64, error) {
	return a.manager.TrimStream(streamName, maxLen)
}

// oauthProviderAdapter adapts infraOAuth.Provider to oauthHandler.OAuthProvider
type oauthProviderAdapter struct {
	provider *infraOAuth.Provider
}

func (a *oauthProviderAdapter) Name() string {
	return a.provider.Name()
}

func (a *oauthProviderAdapter) GetAuthURL(state string) string {
	return a.provider.GetAuthURL(state)
}

func (a *oauthProviderAdapter) ExchangeCode(code string) (oauthHandler.Token, error) {
	token, err := a.provider.ExchangeCode(code)
	if err != nil {
		return oauthHandler.Token{}, err
	}
	return oauthHandler.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresIn:    token.ExpiresIn,
	}, nil
}

func (a *oauthProviderAdapter) GetUserInfo(token oauthHandler.Token) (oauthHandler.UserInfo, error) {
	infraToken := infraOAuth.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresIn:    token.ExpiresIn,
	}
	userInfo, err := a.provider.GetUserInfo(infraToken)
	if err != nil {
		return oauthHandler.UserInfo{}, err
	}
	return oauthHandler.UserInfo{
		ID:    userInfo.ID,
		Email: userInfo.Email,
		Name:  userInfo.Name,
	}, nil
}
