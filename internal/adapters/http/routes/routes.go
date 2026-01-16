package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/rs/cors"
)

// Config holds router configuration
type Config struct {
	JWTManager    *jwt.Manager
	MemberRepo    workspace.MemberRepository
	WorkspaceRepo workspace.Repository
	Logger        logger.Logger
	CorsOrigins   []string
	RateLimit     int
	RateBurst     int
}

// Handlers holds all HTTP handlers
type Handlers struct {
	Auth       AuthHandlers
	Workflow   WorkflowHandlers
	Execution  ExecutionHandlers
	Workspace  WorkspaceHandlers
	Credential CredentialHandlers
	Schedule   ScheduleHandlers
	Webhook    WebhookHandlers
}

// AuthHandlers holds auth-related handlers
type AuthHandlers struct {
	Register http.HandlerFunc
	Login    http.HandlerFunc
	Refresh  http.HandlerFunc
	Logout   http.HandlerFunc
}

// WorkflowHandlers holds workflow-related handlers
type WorkflowHandlers struct {
	Create     http.HandlerFunc
	Get        http.HandlerFunc
	List       http.HandlerFunc
	Update     http.HandlerFunc
	Delete     http.HandlerFunc
	Activate   http.HandlerFunc
	Deactivate http.HandlerFunc
}

// ExecutionHandlers holds execution-related handlers
type ExecutionHandlers struct {
	Start  http.HandlerFunc
	Get    http.HandlerFunc
	List   http.HandlerFunc
	Cancel http.HandlerFunc
	Retry  http.HandlerFunc
}

// WorkspaceHandlers holds workspace-related handlers
type WorkspaceHandlers struct {
	Create       http.HandlerFunc
	Get          http.HandlerFunc
	List         http.HandlerFunc
	Update       http.HandlerFunc
	Delete       http.HandlerFunc
	ListMembers  http.HandlerFunc
	InviteMember http.HandlerFunc
	RemoveMember http.HandlerFunc
}

// CredentialHandlers holds credential-related handlers
type CredentialHandlers struct {
	Create http.HandlerFunc
	Get    http.HandlerFunc
	List   http.HandlerFunc
	Update http.HandlerFunc
	Delete http.HandlerFunc
}

// ScheduleHandlers holds schedule-related handlers
type ScheduleHandlers struct {
	Create http.HandlerFunc
	Get    http.HandlerFunc
	List   http.HandlerFunc
	Update http.HandlerFunc
	Delete http.HandlerFunc
	Pause  http.HandlerFunc
	Resume http.HandlerFunc
}

// WebhookHandlers holds webhook-related handlers
type WebhookHandlers struct {
	Trigger http.HandlerFunc
}

// NewRouter creates a new HTTP router
func NewRouter(cfg Config, handlers Handlers) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Recovery(cfg.Logger))
	r.Use(middleware.Logging(cfg.Logger))

	// CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   cfg.CorsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(corsHandler.Handler)

	// Rate limiting
	if cfg.RateLimit > 0 {
		r.Use(middleware.RateLimit(cfg.RateLimit, cfg.RateBurst))
	}

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		common.Success(w, map[string]string{"status": "healthy"})
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.Auth.Register)
			r.Post("/login", handlers.Auth.Login)
			r.Post("/refresh", handlers.Auth.Refresh)
			
			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(cfg.JWTManager))
				r.Post("/logout", handlers.Auth.Logout)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWTManager))

			// User workspaces
			r.Get("/workspaces", handlers.Workspace.List)
			r.Post("/workspaces", handlers.Workspace.Create)

			// Workspace-scoped routes
			r.Route("/workspaces/{workspaceId}", func(r chi.Router) {
				r.Use(middleware.Tenant(cfg.MemberRepo, cfg.WorkspaceRepo))

				r.Get("/", handlers.Workspace.Get)
				r.Put("/", handlers.Workspace.Update)
				r.Delete("/", handlers.Workspace.Delete)

				// Members
				r.Get("/members", handlers.Workspace.ListMembers)
				r.Post("/members/invite", handlers.Workspace.InviteMember)
				r.Delete("/members/{memberId}", handlers.Workspace.RemoveMember)

				// Workflows
				r.Route("/workflows", func(r chi.Router) {
					r.Get("/", handlers.Workflow.List)
					r.Post("/", handlers.Workflow.Create)

					r.Route("/{workflowId}", func(r chi.Router) {
						r.Get("/", handlers.Workflow.Get)
						r.Put("/", handlers.Workflow.Update)
						r.Delete("/", handlers.Workflow.Delete)
						r.Post("/activate", handlers.Workflow.Activate)
						r.Post("/deactivate", handlers.Workflow.Deactivate)
						r.Post("/execute", handlers.Execution.Start)
					})
				})

				// Executions
				r.Route("/executions", func(r chi.Router) {
					r.Get("/", handlers.Execution.List)
					
					r.Route("/{executionId}", func(r chi.Router) {
						r.Get("/", handlers.Execution.Get)
						r.Post("/cancel", handlers.Execution.Cancel)
						r.Post("/retry", handlers.Execution.Retry)
					})
				})

				// Credentials
				r.Route("/credentials", func(r chi.Router) {
					r.Get("/", handlers.Credential.List)
					r.Post("/", handlers.Credential.Create)

					r.Route("/{credentialId}", func(r chi.Router) {
						r.Get("/", handlers.Credential.Get)
						r.Put("/", handlers.Credential.Update)
						r.Delete("/", handlers.Credential.Delete)
					})
				})

				// Schedules
				r.Route("/schedules", func(r chi.Router) {
					r.Get("/", handlers.Schedule.List)
					r.Post("/", handlers.Schedule.Create)

					r.Route("/{scheduleId}", func(r chi.Router) {
						r.Get("/", handlers.Schedule.Get)
						r.Put("/", handlers.Schedule.Update)
						r.Delete("/", handlers.Schedule.Delete)
						r.Post("/pause", handlers.Schedule.Pause)
						r.Post("/resume", handlers.Schedule.Resume)
					})
				})
			})
		})
	})

	// Webhook trigger routes (public, outside /api/v1)
	r.Route("/webhooks", func(r chi.Router) {
		r.HandleFunc("/*", handlers.Webhook.Trigger)
	})

	return r
}
