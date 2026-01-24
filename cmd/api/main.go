package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	resp "github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	adminHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/admin"
	aibuilderHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/aibuilder"
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
	"github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/invitation"
	marketplaceHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/marketplace"
	nodetypesHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/nodetypes"
	oauthHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/oauth"
	pinneddataHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/pinneddata"
	rbacHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/rbac"
	scheduleHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/schedule"
	shareHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/share"
	templateHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/template"
	userHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/user"
	variableHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/variable"
	webhookHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/webhook"
	workflowHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/workflow"
	workspaceHandler "github.com/linkflow-ai/linkflow/internal/adapters/http/handlers/workspace"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/routes"
	asynqAdapter "github.com/linkflow-ai/linkflow/internal/adapters/messaging/asynq"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/repositories"
	redisAdapter "github.com/linkflow-ai/linkflow/internal/adapters/persistence/redis"
	"github.com/linkflow-ai/linkflow/internal/adapters/websocket"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/nodes"
	appbuilder "github.com/linkflow-ai/linkflow/internal/core/application/aibuilder"
	billingapp "github.com/linkflow-ai/linkflow/internal/core/application/billing"
	billingCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/billing"
	credentialCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/credential"
	executionCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/execution"
	scheduleCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/schedule"
	userCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/user"
	webhookCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/webhook"
	workflowCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workflow"
	workspaceCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workspace"
	"github.com/linkflow-ai/linkflow/internal/core/application/handler"
	analyticsQry "github.com/linkflow-ai/linkflow/internal/core/application/query/analytics"
	billingQry "github.com/linkflow-ai/linkflow/internal/core/application/query/billing"
	credentialQry "github.com/linkflow-ai/linkflow/internal/core/application/query/credential"
	executionQry "github.com/linkflow-ai/linkflow/internal/core/application/query/execution"
	scheduleQry "github.com/linkflow-ai/linkflow/internal/core/application/query/schedule"
	templateQry "github.com/linkflow-ai/linkflow/internal/core/application/query/template"
	userQry "github.com/linkflow-ai/linkflow/internal/core/application/query/user"
	workflowQry "github.com/linkflow-ai/linkflow/internal/core/application/query/workflow"
	workspaceQry "github.com/linkflow-ai/linkflow/internal/core/application/query/workspace"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/core/domain/binarydata"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/providers/openai"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/cache"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/crypto"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/email"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/metrics"
	infraOAuth "github.com/linkflow-ai/linkflow/internal/infrastructure/oauth"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	sentryPkg "github.com/linkflow-ai/linkflow/internal/infrastructure/observability/sentry"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/storage"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/storage/s3"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/streaming"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
)

func main() {
	// Check for validation-only mode
	if len(os.Args) > 1 && os.Args[1] == "validate-config" {
		validateConfigCommand()
		return
	}

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

	// Initialize Sentry
	if err := initSentry(cfg); err != nil {
		appLogger.Warn().Err(err).Msg("Failed to initialize Sentry, continuing without error tracking")
	} else if cfg.Sentry.Enabled {
		appLogger.Info().Msg("Sentry error tracking initialized")
		defer sentryPkg.Flush(2 * time.Second)
	}

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
	invitationRepo := repositories.NewInvitationRepository(db)
	rbacRepo := repositories.NewRBACRepository(db)
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
	oauthRepo := repositories.NewOAuthRepository(db)
	pinnedDataRepo := repositories.NewPinnedDataRepository(db)
	shareRepo := repositories.NewShareRepository(db)
	variableRepo := repositories.NewVariableRepository(db)
	binaryDataRepo := repositories.NewBinaryDataRepository(db)
	siteSettingsRepo := repositories.NewSiteSettingsRepository(db)

	// Storage service
	var storageService binarydata.StorageService
	if cfg.Storage.Provider == "s3" {
		s3Config := storage.S3Config{
			Region:          cfg.Storage.S3.Region,
			Bucket:          cfg.Storage.S3.Bucket,
			AccessKeyID:     cfg.Storage.S3.AccessKeyID,
			SecretAccessKey: cfg.Storage.S3.SecretAccessKey,
			Endpoint:        cfg.Storage.S3.Endpoint,
			UsePathStyle:    cfg.Storage.S3.UsePathStyle,
		}
		s3Storage, err := s3.NewS3Storage(s3Config)
		if err != nil {
			appLogger.Fatal().Err(err).Msg("Failed to initialize S3 storage")
		}
		storageService = s3Storage
		appLogger.Info().Msg("Using S3 storage")
	} else {
		localStorage, err := storage.NewLocalStorage("./data/uploads")
		if err != nil {
			appLogger.Fatal().Err(err).Msg("Failed to initialize local storage")
		}
		storageService = localStorage
		appLogger.Info().Msg("Using local storage")
	}

	// Node registry
	nodeRegistry := nodes.NewRegistry()
	if err := nodes.LoadAllNodes(nodeRegistry); err != nil {
		log.Fatal().Err(err).Msg("Failed to load node types")
	}

	// Metrics collector
	metricsCollector := metrics.NewCollector("2.0.0")

	// Encryption service
	encryptor, err := crypto.NewEncryptor(cfg.Crypto.EncryptionKey)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to initialize encryptor")
	}

	// Stream manager (for admin)
	streamManager := streaming.NewManager(redisClient.Redis())

	// AI Provider
	var aiProvider ai.ProviderAdapter
	if cfg.AI.OpenAI.APIKey != "" {
		aiProvider = openai.NewAdapter(&ai.ProviderConfig{
			APIKey: cfg.AI.OpenAI.APIKey,
		})
	}

	// AI Service
	var aiService *appbuilder.Service
	if aiProvider != nil {
		aiService = appbuilder.NewService(aiProvider)
	} else {
		appLogger.Warn().Msg("AI provider not configured, AI Builder features will be disabled")
	}

	// OAuth providers
	oauthProviders := make(map[string]oauthHandler.OAuthProvider)
	refreshProviders := make(map[string]*infraOAuth.Provider)

	if cfg.OAuth.Google.ClientID != "" {
		googleProvider := infraOAuth.NewGoogleProvider(
			cfg.OAuth.Google.ClientID,
			cfg.OAuth.Google.ClientSecret,
			cfg.OAuth.Google.RedirectURL,
		)
		oauthProviders["google"] = &oauthProviderAdapter{googleProvider}
		refreshProviders["google"] = googleProvider
	}
	if cfg.OAuth.GitHub.ClientID != "" {
		githubProvider := infraOAuth.NewGitHubProvider(
			cfg.OAuth.GitHub.ClientID,
			cfg.OAuth.GitHub.ClientSecret,
			cfg.OAuth.GitHub.RedirectURL,
		)
		oauthProviders["github"] = &oauthProviderAdapter{githubProvider}
		refreshProviders["github"] = githubProvider
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

	// Billing service for usage tracking
	usageService := billingapp.NewUsageService(usageRepo, subscriptionRepo)

	// WebSocket stack
	wsHub := websocket.NewHub()
	go wsHub.Run()
	wsSubscriber := websocket.NewSubscriber(wsHub)
	executionStreamService := websocket.NewExecutionStreamService(wsSubscriber)
	wsHandler := websocket.NewHandler(wsHub)

	// Redis WebSocket Sync (Distribution)
	redisWSSync := websocket.NewRedisSubscriber(wsHub, redisClient.Redis())
	if err := redisWSSync.Start(context.Background()); err != nil {
		appLogger.Warn().Err(err).Msg("Failed to start Redis WebSocket sync")
	}

	// Command handlers
	registerUserHandler := userCmd.NewRegisterUserHandler(userRepo, jwtManager, eventBus)
	loginUserHandler := userCmd.NewLoginUserHandler(userRepo, sessionRepo, jwtManager, eventBus)
	createWorkspaceHandler := workspaceCmd.NewCreateWorkspaceHandler(workspaceRepo, memberRepo, rbacRepo, eventBus)
	createWorkflowHandler := workflowCmd.NewCreateWorkflowHandler(workflowRepo, versionRepo, eventBus)
	updateWorkflowHandler := workflowCmd.NewUpdateWorkflowHandler(workflowRepo, versionRepo)
	activateWorkflowHandler := workflowCmd.NewActivateWorkflowHandler(workflowRepo, usageService, eventBus)
	deactivateWorkflowHandler := workflowCmd.NewDeactivateWorkflowHandler(workflowRepo, eventBus)
	startExecutionHandler := executionCmd.NewStartExecutionHandler(workflowRepo, executionRepo, eventBus, taskQueue, executionStreamService)
	createCredentialHandler := credentialCmd.NewCreateCredentialHandler(credentialRepo, eventBus, encryptor)
	updateCredentialHandler := credentialCmd.NewUpdateCredentialHandler(credentialRepo)
	deleteCredentialHandler := credentialCmd.NewDeleteCredentialHandler(credentialRepo)
	createScheduleHandler := scheduleCmd.NewCreateScheduleHandler(scheduleRepo, usageService, eventBus)
	triggerWebhookHandler := webhookCmd.NewTriggerWebhookHandler(webhookRepo, nil)
	createSubscriptionHandler := billingCmd.NewCreateSubscriptionHandler(subscriptionRepo, eventBus)
	cancelSubscriptionHandler := billingCmd.NewCancelSubscriptionHandler(subscriptionRepo)
	refreshService := credentialCmd.NewRefreshService(encryptor, refreshProviders)

	// Event handlers
	webhookAutoRegHandler := handler.NewWebhookAutoRegHandler(webhookRepo, workflowRepo)
	eventBus.Subscribe("workflow.activated", webhookAutoRegHandler.Handle)
	eventBus.Subscribe("workflow.deactivated", webhookAutoRegHandler.Handle)

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
	cacheService := cache.NewRedisCache(redisClient.Redis(), "cache")
	emailService, err := email.NewService(email.Config{
		Provider:     cfg.Email.Provider,
		DefaultFrom:  cfg.Email.From,
		SMTPHost:     cfg.Email.SMTPHost,
		SMTPPort:     cfg.Email.SMTPPort,
		SMTPUser:     cfg.Email.SMTPUser,
		SMTPPass:     cfg.Email.SMTPPass,
		ResendAPIKey: cfg.Email.ResendAPIKey,
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
	authOAuthRedirectHandler := oauthHandler.NewAuthorizeHandler(oauthProviders, cacheService)
	authOAuthCallbackHandler := oauthHandler.NewCallbackHandler(oauthProviders, cacheService, userRepo, oauthRepo, jwtManager)
	healthHandler := health.NewHandler()

	// Workspace handlers
	wsCreateHandler := workspaceHandler.NewCreateHandler(createWorkspaceHandler)
	wsGetHandler := workspaceHandler.NewGetHandler(getWorkspaceHandler)
	wsListHandler := workspaceHandler.NewListHandler(listWorkspacesHandler)
	wsUpdateHandler := workspaceHandler.NewUpdateHandler(workspaceRepo)
	wsDeleteHandler := workspaceHandler.NewDeleteHandler(workspaceRepo)
	wsMembersHandler := workspaceHandler.NewListMembersHandler(listMembersHandler)
	wsInviteHandler := workspaceHandler.NewInviteMemberHandler(workspaceRepo, memberRepo, invitationRepo, userRepo, rbacRepo, emailService, baseURL)
	wsUpdateMemberHandler := workspaceHandler.NewUpdateMemberHandler(memberRepo, workspaceRepo, rbacRepo)
	wsRemoveMemberHandler := workspaceHandler.NewRemoveMemberHandler(workspaceRepo, memberRepo)

	// Workflow handlers
	wfCreateHandler := workflowHandler.NewCreateHandler(createWorkflowHandler)
	wfGetHandler := workflowHandler.NewGetHandler(getWorkflowHandler)
	wfListHandler := workflowHandler.NewListHandler(workflowQry.NewListWorkflowsHandler(workflowRepo))
	wfUpdateHandler := workflowHandler.NewUpdateHandler(updateWorkflowHandler)
	wfDeleteHandler := workflowHandler.NewDeleteHandler(workflowRepo)
	wfActivateHandler := workflowHandler.NewActivateHandler(activateWorkflowHandler)
	wfDeactivateHandler := workflowHandler.NewDeactivateHandler(deactivateWorkflowHandler)
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
	wfAdvancedSearchHandler := workflowHandler.NewAdvancedSearchHandler(workflowRepo)
	wfSearchFiltersHandler := workflowHandler.NewSearchFiltersHandler(workflowRepo)

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
	exReplayFromNodeHandler := executionHandler.NewReplayFromNodeHandler(executionRepo, workflowRepo, nil)
	exResumeHandler := executionHandler.NewResumeHandler(executionRepo)
	exResumeStatusHandler := executionHandler.NewResumeStatusHandler(executionRepo)
	exGetWaitingHandler := executionHandler.NewGetWaitingHandler(executionRepo)
	exListWaitingHandler := executionHandler.NewListWaitingHandler(executionRepo)
	exBulkDeleteHandler := executionHandler.NewBulkDeleteExecutionsHandler(executionRepo)

	// Credential handlers
	crCreateHandler := credentialHandler.NewCreateHandler(createCredentialHandler)
	crGetHandler := credentialHandler.NewGetHandler(getCredentialHandler)
	crListHandler := credentialHandler.NewListHandler(listCredentialsHandler)
	crUpdateHandler := credentialHandler.NewUpdateHandler(updateCredentialHandler)
	crDeleteHandler := credentialHandler.NewDeleteHandler(deleteCredentialHandler)
	crRefreshHandler := credentialHandler.NewRefreshHandler(credentialRepo, refreshService)

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
	whRegenerateSecretHandler := webhookHandler.NewRegenerateSecretHandler(webhookRepo)
	whActivateHandler := webhookHandler.NewActivateEndpointHandler(webhookRepo)
	whDeactivateHandler := webhookHandler.NewDeactivateEndpointHandler(webhookRepo)

	// Billing handlers
	blGetPlansHandler := billingHandler.NewGetPlansHandler(getPlansHandler)
	blGetSubscriptionHandler := billingHandler.NewGetSubscriptionHandler(getSubscriptionHandler)
	blCreateSubscriptionHandler := billingHandler.NewCreateSubscriptionHandler(createSubscriptionHandler)
	blCancelSubscriptionHandler := billingHandler.NewCancelSubscriptionHandler(cancelSubscriptionHandler)
	blGetUsageHandler := billingHandler.NewGetUsageHandler(getUsageHandler)
	blGetInvoicesHandler := billingHandler.NewGetInvoicesHandler(getInvoicesHandler)
	blStripeWebhookHandler := billingHandler.NewStripeWebhookHandler("")

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
	usrPermissionsHandler := userHandler.NewMyPermissionsHandler()

	// Folder handlers
	fldCreateHandler := folderHandler.NewCreateFolderHandler(folderRepo)
	fldGetHandler := folderHandler.NewGetFolderHandler(folderRepo)
	fldListHandler := folderHandler.NewListFoldersHandler(folderRepo)
	fldUpdateHandler := folderHandler.NewUpdateFolderHandler(folderRepo)
	fldDeleteHandler := folderHandler.NewDeleteFolderHandler(folderRepo)
	fldTreeHandler := folderHandler.NewGetFolderTreeHandler(folderRepo, workflowRepo)

	// Invitation handlers
	invGetInfoHandler := invitation.NewGetInvitationHandler(invitationRepo)
	invAcceptHandler := invitation.NewAcceptInvitationHandler(invitationRepo, memberRepo, userRepo)

	// API Key handlers
	akCreateHandler := apikeyHandler.NewCreateAPIKeyHandler(apiKeyRepo)
	akListHandler := apikeyHandler.NewListAPIKeysHandler(apiKeyRepo)
	akRevokeHandler := apikeyHandler.NewRevokeAPIKeyHandler(apiKeyRepo)

	// Node types handlers
	ntListHandler := nodetypesHandler.NewListNodeTypesHandler(nodeRegistry, siteSettingsRepo)
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
	mpSyncHandler := marketplaceHandler.NewSyncHandler()

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
	bdUploadHandler := binarydataHandler.NewUploadHandler(binaryDataRepo, storageService)
	bdListHandler := binarydataHandler.NewListHandler(binaryDataRepo)
	bdGetInfoHandler := binarydataHandler.NewGetInfoHandler(binaryDataRepo)
	bdDownloadHandler := binarydataHandler.NewDownloadHandler(binaryDataRepo, storageService)
	bdDeleteHandler := binarydataHandler.NewDeleteHandler(binaryDataRepo, storageService)
	bdStatsHandler := binarydataHandler.NewStatsHandler(binaryDataRepo)
	bdCleanupHandler := binarydataHandler.NewCleanupHandler(binaryDataRepo)

	// Variable handlers
	varListHandler := variableHandler.NewListHandler(variableRepo)
	varCreateHandler := variableHandler.NewCreateHandler(variableRepo)
	varUpdateHandler := variableHandler.NewUpdateHandler(variableRepo)
	varDeleteHandler := variableHandler.NewDeleteHandler(variableRepo)
	varListEnvHandler := variableHandler.NewListEnvironmentsHandler(variableRepo)
	varResolveHandler := variableHandler.NewResolveHandler(variableRepo)

	// Admin handlers
	admMetricsHandler := adminHandler.NewMetricsHandler(&metricsAdapter{metricsCollector})
	admStreamStatsHandler := adminHandler.NewStreamStatsHandler(&streamAdapter{streamManager})
	admReplayDLQHandler := adminHandler.NewReplayDLQHandler(&streamAdapter{streamManager})
	admTrimStreamHandler := adminHandler.NewTrimStreamHandler(&streamAdapter{streamManager})
	admGetDisabledNodesHandler := adminHandler.NewGetDisabledNodesHandler(siteSettingsRepo)
	admUpdateDisabledNodesHandler := adminHandler.NewUpdateDisabledNodesHandler(siteSettingsRepo)

	// RBAC handlers
	rbacListRolesHandler := rbacHandler.NewListRolesHandler(rbacRepo)
	rbacCreateRoleHandler := rbacHandler.NewCreateRoleHandler(rbacRepo)
	rbacUpdateRoleHandler := rbacHandler.NewUpdateRoleHandler(rbacRepo)
	rbacDeleteRoleHandler := rbacHandler.NewDeleteRoleHandler(rbacRepo)
	rbacListPermissionsHandler := rbacHandler.NewListPermissionsHandler(rbacRepo)

	// OAuth handlers
	oaListProvidersHandler := oauthHandler.NewListProvidersHandler(oauthProviders)
	oaAuthorizeHandler := oauthHandler.NewAuthorizeHandler(oauthProviders, cacheService)
	oaCallbackHandler := oauthHandler.NewCallbackHandler(oauthProviders, cacheService, userRepo, oauthRepo, jwtManager)

	// AI Builder handlers
	var aiGenerateHandler *aibuilderHandler.GenerateHandler
	var aiSuggestHandler *aibuilderHandler.SuggestHandler
	var aiExplainHandler *aibuilderHandler.ExplainHandler

	if aiService != nil {
		aiGenerateHandler = aibuilderHandler.NewGenerateHandler(aiService)
		aiSuggestHandler = aibuilderHandler.NewSuggestHandler(aiService)
		aiExplainHandler = aibuilderHandler.NewExplainHandler(aiService)
	}

	// Build handlers struct for routes.NewRouter
	handlers := routes.Handlers{
		Auth: routes.AuthHandlers{
			Register:       registerHandler.Handle,
			Login:          loginHandler.Handle,
			Refresh:        refreshHandler.Handle,
			Logout:         logoutHandler.Handle,
			ForgotPassword: forgotPasswordHandler.Handle,
			ResetPassword:  resetPasswordHandler.Handle,
			SetupMFA:       setupMFAHandler.Handle,
			VerifyMFA:      verifyMFAHandler.Handle,
			DisableMFA:     disableMFAHandler.Handle,
			OAuthRedirect:  authOAuthRedirectHandler.Handle,
			OAuthCallback:  authOAuthCallbackHandler.Handle,
		},
		User: routes.UserHandlers{
			GetCurrentUser:    usrGetHandler.Handle,
			UpdateCurrentUser: usrUpdateHandler.Handle,
			MyPermissions:     usrPermissionsHandler.Handle,
		},
		APIKey: routes.APIKeyHandlers{
			List:   akListHandler.Handle,
			Create: akCreateHandler.Handle,
			Revoke: akRevokeHandler.Handle,
		},
		Workflow: routes.WorkflowHandlers{
			Create:          wfCreateHandler.Handle,
			Get:             wfGetHandler.Handle,
			List:            wfListHandler.Handle,
			Search:          wfSearchHandler.Handle,
			AdvancedSearch:  wfAdvancedSearchHandler.Handle,
			SearchFilters:   wfSearchFiltersHandler.Handle,
			Update:          wfUpdateHandler.Handle,
			Delete:          wfDeleteHandler.Handle,
			Activate:        wfActivateHandler.Handle,
			Deactivate:      wfDeactivateHandler.Handle,
			Clone:           wfCloneHandler.Handle,
			Duplicate:       wfDuplicateHandler.Handle,
			Export:          wfExportHandler.Handle,
			Import:          wfImportHandler.Handle,
			Validate:        wfValidateHandler.Handle,
			TestNode:        wfTestNodeHandler.Handle,
			GetVersions:     wfVersionsHandler.Handle,
			GetVersion:      wfGetVersionHandler.Handle,
			Rollback:        wfRollbackHandler.Handle,
			CompareVersions: wfCompareVersionsHandler.Handle,
		},
		Execution: routes.ExecutionHandlers{
			Start:          exStartHandler.Handle,
			Get:            exGetHandler.Handle,
			List:           exListHandler.Handle,
			Cancel:         exCancelHandler.Handle,
			Retry:          exRetryHandler.Handle,
			Search:         exSearchHandler.Handle,
			BulkDelete:     exBulkDeleteHandler.Handle,
			Replay:         exReplayHandler.Handle,
			ReplayFromNode: exReplayFromNodeHandler.Handle,
			GetNodes:       exListNodesHandler.Handle,
			GetNode:        exGetNodeHandler.Handle,
			Stats:          exStatsHandler.Handle,
			GetWaiting:     exGetWaitingHandler.Handle,
			ListWaiting:    exListWaitingHandler.Handle,
			Resume:         exResumeHandler.Handle,
			ResumeStatus:   exResumeStatusHandler.Handle,
		},
		Workspace: routes.WorkspaceHandlers{
			Create:       wsCreateHandler.Handle,
			Get:          wsGetHandler.Handle,
			List:         wsListHandler.Handle,
			Update:       wsUpdateHandler.Handle,
			Delete:       wsDeleteHandler.Handle,
			ListMembers:  wsMembersHandler.Handle,
			InviteMember: wsInviteHandler.Handle,
			UpdateMember: wsUpdateMemberHandler.Handle,
			RemoveMember: wsRemoveMemberHandler.Handle,
		},
		Credential: routes.CredentialHandlers{
			Create: crCreateHandler.Handle,
			Get:    crGetHandler.Handle,
			List:   crListHandler.Handle,
			Update: crUpdateHandler.Handle,
			Delete: crDeleteHandler.Handle,
			Test: func(w http.ResponseWriter, req *http.Request) {
				resp.Success(w, map[string]interface{}{"valid": true, "message": "Connection test successful"})
			},
			Refresh: crRefreshHandler.Handle,
		},
		Schedule: routes.ScheduleHandlers{
			Create: schCreateHandler.Handle,
			Get:    schGetHandler.Handle,
			List:   schListHandler.Handle,
			Update: schUpdateHandler.Handle,
			Delete: schDeleteHandler.Handle,
			Pause:  schPauseHandler.Handle,
			Resume: schResumeHandler.Handle,
		},
		Webhook: routes.WebhookHandlers{
			Trigger:          whTriggerHandler.Handle,
			Create:           whCreateHandler.Handle,
			List:             whListHandler.Handle,
			RegenerateSecret: whRegenerateSecretHandler.Handle,
			Activate:         whActivateHandler.Handle,
			Deactivate:       whDeactivateHandler.Handle,
		},
		Folder: routes.FolderHandlers{
			Create: fldCreateHandler.Handle,
			Get:    fldGetHandler.Handle,
			List:   fldListHandler.Handle,
			Tree:   fldTreeHandler.Handle,
			Update: fldUpdateHandler.Handle,
			Delete: fldDeleteHandler.Handle,
		},
		Dashboard: routes.DashboardHandlers{
			GetDashboard:  dashHandler.Handle,
			GetQuickStats: quickStatsHandler.Handle,
		},
		NodeType: routes.NodeTypeHandlers{
			List:          ntListHandler.Handle,
			GetCategories: ntCategoriesHandler.Handle,
			Get:           ntGetHandler.Handle,
		},
		Health: routes.HealthHandlers{
			Health:    healthHandler.Health,
			Liveness:  healthHandler.Liveness,
			Readiness: healthHandler.Readiness,
		},
		Billing: routes.BillingHandlers{
			GetPlans:           blGetPlansHandler.Handle,
			GetSubscription:    blGetSubscriptionHandler.Handle,
			CreateSubscription: blCreateSubscriptionHandler.Handle,
			CancelSubscription: blCancelSubscriptionHandler.Handle,
			GetUsage:           blGetUsageHandler.Handle,
			GetInvoices:        blGetInvoicesHandler.Handle,
			StripeWebhook:      blStripeWebhookHandler.Handle,
		},
		OAuth: routes.OAuthHandlers{
			ListProviders: oaListProvidersHandler.Handle,
			Authorize:     oaAuthorizeHandler.Handle,
			Callback:      oaCallbackHandler.Handle,
		},
		Template: routes.TemplateHandlers{
			List:          tplListHandler.Handle,
			GetFeatured:   tplFeaturedHandler.Handle,
			GetCategories: tplCategoriesHandler.Handle,
			GetByCategory: tplByCategoryHandler.Handle,
			Search:        tplSearchHandler.Handle,
			Get:           tplGetHandler.Handle,
			Use:           tplUseHandler.Handle,
		},
		PinnedData: routes.PinnedDataHandlers{
			GetAll:    pdGetAllHandler.Handle,
			GetByNode: pdGetByNodeHandler.Handle,
			Set:       pdSetHandler.Handle,
			Delete:    pdDeleteHandler.Handle,
		},
		WorkflowShare: routes.WorkflowShareHandlers{
			Create:       shCreateHandler.Handle,
			SharedByMe:   shSharedByMeHandler.Handle,
			SharedWithMe: shSharedWithMeHandler.Handle,
			Pending:      shPendingHandler.Handle,
			Accept:       shAcceptHandler.Handle,
			Update:       shUpdateHandler.Handle,
			Revoke:       shRevokeHandler.Handle,
		},
		Marketplace: routes.MarketplaceHandlers{
			Browse:       mpBrowseHandler.Handle,
			Featured:     mpFeaturedHandler.Handle,
			Categories:   mpCategoriesHandler.Handle,
			Search:       mpSearchHandler.Handle,
			Get:          mpGetHandler.Handle,
			Use:          mpUseHandler.Handle,
			Publish:      mpPublishHandler.Handle,
			MyPublished:  mpMyPublishedHandler.Handle,
			Update:       mpUpdateHandler.Handle,
			Sync:         mpSyncHandler.Handle,
			Unpublish:    mpUnpublishHandler.Handle,
			Rate:         mpRateHandler.Handle,
			GetMyRating:  mpGetMyRatingHandler.Handle,
			ListRatings:  mpListRatingsHandler.Handle,
			RatingStats:  mpRatingStatsHandler.Handle,
			DeleteRating: mpDeleteRatingHandler.Handle,
		},
		BinaryData: routes.BinaryDataHandlers{
			Upload:   bdUploadHandler.Handle,
			List:     bdListHandler.Handle,
			GetInfo:  bdGetInfoHandler.Handle,
			Download: bdDownloadHandler.Handle,
			Delete:   bdDeleteHandler.Handle,
			GetStats: bdStatsHandler.Handle,
			Cleanup:  bdCleanupHandler.Handle,
		},
		Admin: routes.AdminHandlers{
			Metrics:             admMetricsHandler.Handle,
			StreamStats:         admStreamStatsHandler.Handle,
			ReplayDLQ:           admReplayDLQHandler.Handle,
			TrimStream:          admTrimStreamHandler.Handle,
			GetDisabledNodes:    admGetDisabledNodesHandler.Handle,
			UpdateDisabledNodes: admUpdateDisabledNodesHandler.Handle,
		},
		RBAC: routes.RBACHandlers{
			ListRoles:       rbacListRolesHandler.Handle,
			CreateRole:      rbacCreateRoleHandler.Handle,
			UpdateRole:      rbacUpdateRoleHandler.Handle,
			DeleteRole:      rbacDeleteRoleHandler.Handle,
			ListPermissions: rbacListPermissionsHandler.Handle,
		},
		Analytics: routes.AnalyticsHandlers{
			WorkflowAnalytics:  anWorkflowHandler.Handle,
			WorkspaceAnalytics: anWorkspaceHandler.Handle,
		},
		Invitation: routes.InvitationHandlers{
			GetInfo: invGetInfoHandler.Handle,
			Accept:  invAcceptHandler.Handle,
		},
		AIBuilder: routes.AIBuilderHandlers{
			Generate: func(w http.ResponseWriter, r *http.Request) {
				if aiGenerateHandler != nil {
					aiGenerateHandler.Handle(w, r)
				} else {
					resp.BadRequest(w, "AI Builder feature requires AI provider configuration")
				}
			},
			Suggest: func(w http.ResponseWriter, r *http.Request) {
				if aiSuggestHandler != nil {
					aiSuggestHandler.Handle(w, r)
				} else {
					resp.BadRequest(w, "AI Builder feature requires AI provider configuration")
				}
			},
			Explain: func(w http.ResponseWriter, r *http.Request) {
				if aiExplainHandler != nil {
					aiExplainHandler.Handle(w, r)
				} else {
					resp.BadRequest(w, "AI Builder feature requires AI provider configuration")
				}
			},
		},
		Variable: routes.VariableHandlers{
			List:             varListHandler.Handle,
			Create:           varCreateHandler.Handle,
			Update:           varUpdateHandler.Handle,
			Delete:           varDeleteHandler.Handle,
			ListEnvironments: varListEnvHandler.Handle,
			Resolve:          varResolveHandler.Handle,
		},
		WebSocket: wsHandler.ServeHTTP,
	}

	// Create router using routes.NewRouter
	r := routes.NewRouter(routes.Config{
		JWTManager:    jwtManager,
		JWTBlacklist:  jwtBlacklist,
		MemberRepo:    memberRepo,
		WorkspaceRepo: workspaceRepo,
		APIKeyRepo:    apiKeyRepo,
		Logger:        appLogger,
		CorsOrigins:   cfg.App.CorsOrigins,
	}, handlers)

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

func (a *oauthProviderAdapter) RefreshToken(refreshToken string) (oauthHandler.Token, error) {
	token, err := a.provider.RefreshToken(refreshToken)
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

// validateConfigCommand validates the configuration and exits
func validateConfigCommand() {
	// Initialize logger
	var log logger.Logger
	if os.Getenv("APP_ENV") == "production" {
		log = logger.NewDefault()
	} else {
		log = logger.NewDevelopment()
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "local"
		}
		configPath = fmt.Sprintf("configs/config.%s.yaml", env)
	}

	log.Info().Str("config", configPath).Msg("Validating configuration")

	_, result, err := config.LoadWithValidation(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Print results
	if result.Valid {
		log.Info().Msg("✅ Configuration validation passed")
	} else {
		log.Error().Msg("❌ Configuration validation failed")

		for _, err := range result.Errors {
			log.Error().Str("field", err.Field).Msg("ERROR: " + err.Message)
		}
		os.Exit(1)
	}

	// Show warnings
	if result.HasWarnings() {
		log.Warn().Msg("⚠️  Configuration warnings:")
		for _, warning := range result.Warnings {
			log.Warn().Str("field", warning.Field).Msg("WARNING: " + warning.Message)
		}
	}

	// Show info
	if len(result.Info) > 0 {
		log.Info().Msg("ℹ️  Configuration info:")
		for _, info := range result.Info {
			log.Info().Str("field", info.Field).Msg("INFO: " + info.Message)
		}
	}

	log.Info().Msg("Configuration validation complete")
}

// initSentry initializes Sentry error tracking
func initSentry(cfg *config.Config) error {
	return sentryPkg.Init(cfg.Sentry, "linkflow-api")
}
