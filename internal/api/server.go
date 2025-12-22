package api

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
	"github.com/linkflow-ai/linkflow/internal/api/handlers"
	"github.com/linkflow-ai/linkflow/internal/api/middleware"
	"github.com/linkflow-ai/linkflow/internal/api/websocket"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Server struct {
	cfg        *config.Config
	router     *chi.Mux
	httpServer *http.Server
	wsHub      *websocket.Hub
}

type Services struct {
	Auth       *services.AuthService
	User       *services.UserService
	Workspace  *services.WorkspaceService
	Workflow   *services.WorkflowService
	Execution  *services.ExecutionService
	Credential *services.CredentialService
	Schedule   *services.ScheduleService
	Billing    *services.BillingService
}

func NewServer(
	cfg *config.Config,
	svc *Services,
	jwtManager *crypto.JWTManager,
	redisClient *pkgredis.Client,
	queueClient *queue.Client,
	db *gorm.DB,
) *Server {
	router := chi.NewRouter()

	// WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// WebSocket subscriber (listens to Redis events and broadcasts to clients)
	wsSubscriber := websocket.NewSubscriber(redisClient.Client, wsHub)
	wsSubscriber.Start()

	// Global middleware
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(middleware.Logger())
	router.Use(middleware.Recoverer())
	router.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.App.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Workspace-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	router.Use(corsHandler.Handler)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(svc.Auth, jwtManager, redisClient)
	userHandler := handlers.NewUserHandler(svc.User)
	workspaceHandler := handlers.NewWorkspaceHandler(svc.Workspace, svc.Billing)
	workflowHandler := handlers.NewWorkflowHandler(svc.Workflow, svc.Billing, queueClient)
	executionHandler := handlers.NewExecutionHandler(svc.Execution, queueClient)
	credentialHandler := handlers.NewCredentialHandler(svc.Credential)
	scheduleHandler := handlers.NewScheduleHandler(svc.Schedule)
	billingHandler := handlers.NewBillingHandler(svc.Billing)
	healthHandler := handlers.NewHealthHandlerWithDeps(db, redisClient.Client)
	webhookHandler := handlers.NewWebhookHandler(svc.Workflow, svc.Execution, queueClient)
	wsHandler := handlers.NewWebSocketHandler(wsHub, jwtManager)
	nodeTypeHandler := handlers.NewNodeTypeHandler(svc.Workflow, svc.Execution)

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, redisClient)
	tenantMiddleware := middleware.NewTenantMiddleware(svc.Workspace)
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// Routes
	router.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Use(rateLimiter.Limit(100, time.Minute)) // 100 requests per minute

			// Auth
			r.Post("/auth/register", authHandler.Register)
			r.Post("/auth/login", authHandler.Login)
			r.Post("/auth/refresh", authHandler.RefreshToken)
			r.Post("/auth/forgot-password", authHandler.ForgotPassword)
			r.Post("/auth/reset-password", authHandler.ResetPassword)
			r.Get("/auth/oauth/{provider}", authHandler.OAuthRedirect)
			r.Get("/auth/oauth/{provider}/callback", authHandler.OAuthCallback)

			// Health
			r.Get("/health", healthHandler.Health)
			r.Get("/health/live", healthHandler.Live)
			r.Get("/health/ready", healthHandler.Ready)

			// Plans (public)
			r.Get("/billing/plans", billingHandler.GetPlans)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Use(rateLimiter.Limit(1000, time.Minute)) // 1000 requests per minute

			// Auth
			r.Post("/auth/logout", authHandler.Logout)
			r.Post("/auth/mfa/setup", authHandler.SetupMFA)
			r.Post("/auth/mfa/verify", authHandler.VerifyMFA)
			r.Delete("/auth/mfa", authHandler.DisableMFA)

			// User
			r.Get("/users/me", userHandler.GetCurrentUser)
			r.Put("/users/me", userHandler.UpdateCurrentUser)

			// Node Types (for workflow editor)
			r.Get("/node-types", nodeTypeHandler.ListNodeTypes)
			r.Get("/node-types/categories", nodeTypeHandler.GetNodeCategories)
			r.Get("/node-types/{nodeType}", nodeTypeHandler.GetNodeType)

			// Workspaces
			r.Get("/workspaces", workspaceHandler.List)
			r.Post("/workspaces", workspaceHandler.Create)

			// Workspace-scoped routes
			r.Route("/workspaces/{workspaceID}", func(r chi.Router) {
				r.Use(tenantMiddleware.RequireMembership)

				r.Get("/", workspaceHandler.Get)
				r.Put("/", workspaceHandler.Update)
				r.Delete("/", workspaceHandler.Delete)
				r.Get("/members", workspaceHandler.GetMembers)
				r.Post("/members", workspaceHandler.InviteMember)
				r.Put("/members/{userID}", workspaceHandler.UpdateMemberRole)
				r.Delete("/members/{userID}", workspaceHandler.RemoveMember)

				// Workflows
				r.Get("/workflows", workflowHandler.List)
				r.Post("/workflows", workflowHandler.Create)
				r.Get("/workflows/{workflowID}", workflowHandler.Get)
				r.Put("/workflows/{workflowID}", workflowHandler.Update)
				r.Delete("/workflows/{workflowID}", workflowHandler.Delete)
				r.Post("/workflows/{workflowID}/execute", workflowHandler.Execute)
				r.Post("/workflows/{workflowID}/clone", workflowHandler.Clone)
				r.Post("/workflows/{workflowID}/activate", workflowHandler.Activate)
				r.Post("/workflows/{workflowID}/deactivate", workflowHandler.Deactivate)
				r.Get("/workflows/{workflowID}/versions", workflowHandler.GetVersions)
				r.Get("/workflows/{workflowID}/versions/{version}", workflowHandler.GetVersion)
				r.Get("/workflows/{workflowID}/export", workflowHandler.Export)
				r.Post("/workflows/{workflowID}/duplicate", workflowHandler.Duplicate)
				r.Post("/workflows/import", workflowHandler.Import)
				r.Post("/workflows/validate", nodeTypeHandler.ValidateWorkflow)
				r.Post("/workflows/test-node", nodeTypeHandler.TestNode)

				// Executions
				r.Get("/executions", executionHandler.List)
				r.Get("/executions/{executionID}", executionHandler.Get)
				r.Post("/executions/{executionID}/cancel", executionHandler.Cancel)
				r.Post("/executions/{executionID}/retry", executionHandler.Retry)
				r.Get("/executions/{executionID}/nodes", executionHandler.GetNodes)

				// Credentials
				r.Get("/credentials", credentialHandler.List)
				r.Post("/credentials", credentialHandler.Create)
				r.Get("/credentials/{credentialID}", credentialHandler.Get)
				r.Put("/credentials/{credentialID}", credentialHandler.Update)
				r.Delete("/credentials/{credentialID}", credentialHandler.Delete)
				r.Post("/credentials/{credentialID}/test", credentialHandler.Test)

				// Schedules
				r.Get("/schedules", scheduleHandler.List)
				r.Post("/schedules", scheduleHandler.Create)
				r.Get("/schedules/{scheduleID}", scheduleHandler.Get)
				r.Put("/schedules/{scheduleID}", scheduleHandler.Update)
				r.Delete("/schedules/{scheduleID}", scheduleHandler.Delete)
				r.Post("/schedules/{scheduleID}/pause", scheduleHandler.Pause)
				r.Post("/schedules/{scheduleID}/resume", scheduleHandler.Resume)

				// Billing
				r.Get("/billing/subscription", billingHandler.GetSubscription)
				r.Post("/billing/subscription", billingHandler.CreateSubscription)
				r.Delete("/billing/subscription", billingHandler.CancelSubscription)
				r.Get("/billing/usage", billingHandler.GetUsage)
				r.Get("/billing/invoices", billingHandler.GetInvoices)
			})
		})
	})

	// Webhooks (separate from API)
	router.Route("/webhooks", func(r chi.Router) {
		r.Post("/{endpointID}", webhookHandler.Handle)
		r.Get("/{endpointID}", webhookHandler.Handle)
		r.Post("/stripe", billingHandler.HandleStripeWebhook)
	})

	// WebSocket
	router.Get("/ws", wsHandler.HandleConnection)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		cfg:        cfg,
		router:     router,
		httpServer: httpServer,
		wsHub:      wsHub,
	}
}

func (s *Server) Start() error {
	log.Info().Str("addr", s.httpServer.Addr).Msg("Starting HTTP server")

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Info().Msg("Server stopped")
	return nil
}

func (s *Server) Router() *chi.Mux {
	return s.router
}

func (s *Server) Hub() *websocket.Hub {
	return s.wsHub
}
