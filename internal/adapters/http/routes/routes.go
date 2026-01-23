package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/auth/jwt"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/rs/cors"
)

// Config holds router configuration
type Config struct {
	JWTManager    *jwt.Manager
	JWTBlacklist  *jwt.Blacklist
	MemberRepo    workspace.MemberRepository
	WorkspaceRepo workspace.Repository
	Logger        logger.Logger
	CorsOrigins   []string
	RateLimit     int
	RateBurst     int
}

// Handlers holds all HTTP handlers
type Handlers struct {
	Auth          AuthHandlers
	User          UserHandlers
	APIKey        APIKeyHandlers
	Workflow      WorkflowHandlers
	Execution     ExecutionHandlers
	Workspace     WorkspaceHandlers
	Credential    CredentialHandlers
	Schedule      ScheduleHandlers
	Webhook       WebhookHandlers
	Folder        FolderHandlers
	Dashboard     DashboardHandlers
	NodeType      NodeTypeHandlers
	Health        HealthHandlers
	Billing       BillingHandlers
	OAuth         OAuthHandlers
	Template      TemplateHandlers
	PinnedData    PinnedDataHandlers
	WorkflowShare WorkflowShareHandlers
	Marketplace   MarketplaceHandlers
	BinaryData    BinaryDataHandlers
	Admin         AdminHandlers
	Analytics     AnalyticsHandlers
	AIBuilder     AIBuilderHandlers
	Variable      VariableHandlers
}

// AuthHandlers holds auth-related handlers
type AuthHandlers struct {
	Register       http.HandlerFunc
	Login          http.HandlerFunc
	Refresh        http.HandlerFunc
	Logout         http.HandlerFunc
	ForgotPassword http.HandlerFunc
	ResetPassword  http.HandlerFunc
	SetupMFA       http.HandlerFunc
	VerifyMFA      http.HandlerFunc
	DisableMFA     http.HandlerFunc
	OAuthRedirect  http.HandlerFunc
	OAuthCallback  http.HandlerFunc
}

// UserHandlers holds user-related handlers
type UserHandlers struct {
	GetCurrentUser    http.HandlerFunc
	UpdateCurrentUser http.HandlerFunc
}

// APIKeyHandlers holds API key handlers
type APIKeyHandlers struct {
	List   http.HandlerFunc
	Create http.HandlerFunc
	Revoke http.HandlerFunc
}

// WorkflowHandlers holds workflow-related handlers
type WorkflowHandlers struct {
	Create          http.HandlerFunc
	Get             http.HandlerFunc
	List            http.HandlerFunc
	Search          http.HandlerFunc
	AdvancedSearch  http.HandlerFunc
	SearchFilters   http.HandlerFunc
	Update          http.HandlerFunc
	Delete          http.HandlerFunc
	Activate        http.HandlerFunc
	Deactivate      http.HandlerFunc
	Clone           http.HandlerFunc
	Duplicate       http.HandlerFunc
	Export          http.HandlerFunc
	Import          http.HandlerFunc
	Validate        http.HandlerFunc
	TestNode        http.HandlerFunc
	GetVersions     http.HandlerFunc
	GetVersion      http.HandlerFunc
	Rollback        http.HandlerFunc
	CompareVersions http.HandlerFunc
}

// ExecutionHandlers holds execution-related handlers
type ExecutionHandlers struct {
	Start          http.HandlerFunc
	Get            http.HandlerFunc
	List           http.HandlerFunc
	Cancel         http.HandlerFunc
	Retry          http.HandlerFunc
	Search         http.HandlerFunc
	BulkDelete     http.HandlerFunc
	Replay         http.HandlerFunc
	ReplayFromNode http.HandlerFunc
	GetNodes       http.HandlerFunc
	GetNode        http.HandlerFunc
	Stats          http.HandlerFunc
	GetWaiting     http.HandlerFunc
	ListWaiting    http.HandlerFunc
	Resume         http.HandlerFunc
	ResumeStatus   http.HandlerFunc
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
	UpdateMember http.HandlerFunc
	RemoveMember http.HandlerFunc
}

// CredentialHandlers holds credential-related handlers
type CredentialHandlers struct {
	Create  http.HandlerFunc
	Get     http.HandlerFunc
	List    http.HandlerFunc
	Update  http.HandlerFunc
	Delete  http.HandlerFunc
	Test    http.HandlerFunc
	Refresh http.HandlerFunc
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
	Trigger          http.HandlerFunc
	Create           http.HandlerFunc
	List             http.HandlerFunc
	RegenerateSecret http.HandlerFunc
	Activate         http.HandlerFunc
	Deactivate       http.HandlerFunc
}

// FolderHandlers holds folder-related handlers
type FolderHandlers struct {
	Create http.HandlerFunc
	Get    http.HandlerFunc
	List   http.HandlerFunc
	Tree   http.HandlerFunc
	Update http.HandlerFunc
	Delete http.HandlerFunc
}

// DashboardHandlers holds dashboard-related handlers
type DashboardHandlers struct {
	GetDashboard  http.HandlerFunc
	GetQuickStats http.HandlerFunc
}

// NodeTypeHandlers holds node type handlers
type NodeTypeHandlers struct {
	List          http.HandlerFunc
	GetCategories http.HandlerFunc
	Get           http.HandlerFunc
}

// HealthHandlers holds health check handlers
type HealthHandlers struct {
	Health    http.HandlerFunc
	Liveness  http.HandlerFunc
	Readiness http.HandlerFunc
}

// BillingHandlers holds billing-related handlers
type BillingHandlers struct {
	GetPlans           http.HandlerFunc
	GetSubscription    http.HandlerFunc
	CreateSubscription http.HandlerFunc
	CancelSubscription http.HandlerFunc
	GetUsage           http.HandlerFunc
	GetInvoices        http.HandlerFunc
	StripeWebhook      http.HandlerFunc
}

// OAuthHandlers holds OAuth-related handlers
type OAuthHandlers struct {
	ListProviders http.HandlerFunc
	Authorize     http.HandlerFunc
	Callback      http.HandlerFunc
}

// TemplateHandlers holds template-related handlers
type TemplateHandlers struct {
	List          http.HandlerFunc
	GetFeatured   http.HandlerFunc
	GetCategories http.HandlerFunc
	GetByCategory http.HandlerFunc
	Search        http.HandlerFunc
	Get           http.HandlerFunc
	Use           http.HandlerFunc
}

// PinnedDataHandlers holds pinned data handlers
type PinnedDataHandlers struct {
	GetAll    http.HandlerFunc
	GetByNode http.HandlerFunc
	Set       http.HandlerFunc
	Delete    http.HandlerFunc
}

// WorkflowShareHandlers holds workflow sharing handlers
type WorkflowShareHandlers struct {
	Create       http.HandlerFunc
	SharedByMe   http.HandlerFunc
	SharedWithMe http.HandlerFunc
	Pending      http.HandlerFunc
	Accept       http.HandlerFunc
	Update       http.HandlerFunc
	Revoke       http.HandlerFunc
}

// MarketplaceHandlers holds marketplace handlers
type MarketplaceHandlers struct {
	Browse       http.HandlerFunc
	Featured     http.HandlerFunc
	Categories   http.HandlerFunc
	Search       http.HandlerFunc
	Get          http.HandlerFunc
	Use          http.HandlerFunc
	Publish      http.HandlerFunc
	MyPublished  http.HandlerFunc
	Update       http.HandlerFunc
	Sync         http.HandlerFunc
	Unpublish    http.HandlerFunc
	Rate         http.HandlerFunc
	GetMyRating  http.HandlerFunc
	ListRatings  http.HandlerFunc
	RatingStats  http.HandlerFunc
	DeleteRating http.HandlerFunc
}

// BinaryDataHandlers holds binary data handlers
type BinaryDataHandlers struct {
	Upload   http.HandlerFunc
	List     http.HandlerFunc
	GetInfo  http.HandlerFunc
	Download http.HandlerFunc
	Delete   http.HandlerFunc
	GetStats http.HandlerFunc
	Cleanup  http.HandlerFunc
}

// AdminHandlers holds admin handlers
type AdminHandlers struct {
	Metrics     http.HandlerFunc
	StreamStats http.HandlerFunc
	ReplayDLQ   http.HandlerFunc
	TrimStream  http.HandlerFunc
}

// AnalyticsHandlers holds analytics handlers
type AnalyticsHandlers struct {
	WorkflowAnalytics  http.HandlerFunc
	WorkspaceAnalytics http.HandlerFunc
}

// AIBuilderHandlers holds AI workflow builder handlers
type AIBuilderHandlers struct {
	Generate http.HandlerFunc
	Suggest  http.HandlerFunc
	Explain  http.HandlerFunc
}

// VariableHandlers holds variable and environment handlers
type VariableHandlers struct {
	List             http.HandlerFunc
	Create           http.HandlerFunc
	Update           http.HandlerFunc
	Delete           http.HandlerFunc
	ListEnvironments http.HandlerFunc
	Resolve          http.HandlerFunc
}

// NewRouter creates a new HTTP router
func NewRouter(cfg Config, handlers Handlers) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Recovery(cfg.Logger))
	r.Use(middleware.Logging(cfg.Logger))

	// Security middleware
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.MaxBodySize(10 * 1024 * 1024)) // 10MB max body

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

	// Health endpoints
	r.Get("/health", handlers.Health.Health)
	r.Get("/metrics", handlers.Admin.Metrics)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health endpoints
		r.Get("/health", handlers.Health.Health)
		r.Get("/health/live", handlers.Health.Liveness)
		r.Get("/health/ready", handlers.Health.Readiness)

		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.Auth.Register)
			r.Post("/login", handlers.Auth.Login)
			r.Post("/refresh", handlers.Auth.Refresh)
			r.Post("/forgot-password", handlers.Auth.ForgotPassword)
			r.Post("/reset-password", handlers.Auth.ResetPassword)
			r.Get("/oauth/{provider}", handlers.Auth.OAuthRedirect)
			r.Get("/oauth/{provider}/callback", handlers.Auth.OAuthCallback)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthWithBlacklist(cfg.JWTManager, cfg.JWTBlacklist))
				r.Post("/logout", handlers.Auth.Logout)
				r.Post("/mfa/setup", handlers.Auth.SetupMFA)
				r.Post("/mfa/verify", handlers.Auth.VerifyMFA)
				r.Delete("/mfa", handlers.Auth.DisableMFA)
			})
		})

		// Node types (public - for workflow editor)
		r.Get("/node-types", handlers.NodeType.List)
		r.Get("/node-types/categories", handlers.NodeType.GetCategories)
		r.Get("/node-types/{nodeType}", handlers.NodeType.Get)

		// Billing plans (public)
		r.Get("/billing/plans", handlers.Billing.GetPlans)

		// OAuth providers (public)
		r.Get("/oauth/providers", handlers.OAuth.ListProviders)

		// Templates (public)
		r.Route("/templates", func(r chi.Router) {
			r.Get("/", handlers.Template.List)
			r.Get("/featured", handlers.Template.GetFeatured)
			r.Get("/categories", handlers.Template.GetCategories)
			r.Get("/categories/{category}", handlers.Template.GetByCategory)
			r.Get("/search", handlers.Template.Search)
			r.Get("/{templateId}", handlers.Template.Get)
		})

		// Marketplace (public browse)
		r.Route("/marketplace", func(r chi.Router) {
			r.Get("/", handlers.Marketplace.Browse)
			r.Get("/featured", handlers.Marketplace.Featured)
			r.Get("/categories", handlers.Marketplace.Categories)
			r.Get("/search", handlers.Marketplace.Search)
			r.Get("/{templateId}", handlers.Marketplace.Get)
			r.Get("/{templateId}/ratings", handlers.Marketplace.ListRatings)
			r.Get("/{templateId}/ratings/stats", handlers.Marketplace.RatingStats)

			// Protected marketplace routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthWithBlacklist(cfg.JWTManager, cfg.JWTBlacklist))
				r.Post("/{templateId}/ratings", handlers.Marketplace.Rate)
				r.Get("/{templateId}/ratings/me", handlers.Marketplace.GetMyRating)
				r.Delete("/{templateId}/ratings", handlers.Marketplace.DeleteRating)
			})
		})

		// Resume execution (public with token)
		r.Post("/resume/{token}", handlers.Execution.Resume)
		r.Get("/resume/{token}/status", handlers.Execution.ResumeStatus)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthWithBlacklist(cfg.JWTManager, cfg.JWTBlacklist))

			// User profile
			r.Get("/users/me", handlers.User.GetCurrentUser)
			r.Put("/users/me", handlers.User.UpdateCurrentUser)

			// API Keys
			r.Get("/api-keys", handlers.APIKey.List)
			r.Post("/api-keys", handlers.APIKey.Create)
			r.Delete("/api-keys/{keyId}", handlers.APIKey.Revoke)

			// User workspaces
			r.Get("/workspaces", handlers.Workspace.List)
			r.Post("/workspaces", handlers.Workspace.Create)

			// Admin routes
			r.Route("/admin", func(r chi.Router) {
				r.Get("/streams/webhooks/stats", handlers.Admin.StreamStats)
				r.Post("/streams/webhooks/replay", handlers.Admin.ReplayDLQ)
				r.Post("/streams/webhooks/trim", handlers.Admin.TrimStream)
			})

			// Workspace-scoped routes
			r.Route("/workspaces/{workspaceId}", func(r chi.Router) {
				r.Use(middleware.Tenant(cfg.MemberRepo, cfg.WorkspaceRepo))

				r.Get("/", handlers.Workspace.Get)
				r.Put("/", handlers.Workspace.Update)
				r.Delete("/", handlers.Workspace.Delete)

				// Dashboard
				r.Get("/dashboard", handlers.Dashboard.GetDashboard)
				r.Get("/stats", handlers.Dashboard.GetQuickStats)

				// Analytics
				r.Get("/analytics", handlers.Analytics.WorkspaceAnalytics)

				// Members
				r.Get("/members", handlers.Workspace.ListMembers)
				r.Post("/members", handlers.Workspace.InviteMember)
				r.Post("/members/invite", handlers.Workspace.InviteMember)
				r.Put("/members/{memberId}", handlers.Workspace.UpdateMember)
				r.Delete("/members/{memberId}", handlers.Workspace.RemoveMember)

				// Billing
				r.Route("/billing", func(r chi.Router) {
					r.Get("/subscription", handlers.Billing.GetSubscription)
					r.Post("/subscription", handlers.Billing.CreateSubscription)
					r.Delete("/subscription", handlers.Billing.CancelSubscription)
					r.Get("/usage", handlers.Billing.GetUsage)
					r.Get("/invoices", handlers.Billing.GetInvoices)
				})

				// OAuth
				r.Get("/oauth/authorize/{provider}", handlers.OAuth.Authorize)
				r.Get("/oauth/callback/{provider}", handlers.OAuth.Callback)

				// Folders
				r.Route("/folders", func(r chi.Router) {
					r.Get("/", handlers.Folder.List)
					r.Get("/tree", handlers.Folder.Tree)
					r.Post("/", handlers.Folder.Create)

					r.Route("/{folderId}", func(r chi.Router) {
						r.Get("/", handlers.Folder.Get)
						r.Put("/", handlers.Folder.Update)
						r.Delete("/", handlers.Folder.Delete)
					})
				})

				// AI Workflow Builder
				r.Route("/ai-builder", func(r chi.Router) {
					r.Post("/generate", handlers.AIBuilder.Generate)
					r.Post("/suggest", handlers.AIBuilder.Suggest)
					r.Post("/explain", handlers.AIBuilder.Explain)
				})

				// Variables & Environments
				r.Route("/variables", func(r chi.Router) {
					r.Get("/", handlers.Variable.List)
					r.Post("/", handlers.Variable.Create)
					r.Get("/resolve", handlers.Variable.Resolve)
					r.Put("/{variableId}", handlers.Variable.Update)
					r.Delete("/{variableId}", handlers.Variable.Delete)
				})
				r.Get("/environments", handlers.Variable.ListEnvironments)

				// Admin routes
				r.Route("/admin", func(r chi.Router) {
					// Admin functionality can be added here
				})

				// Workflows
				r.Route("/workflows", func(r chi.Router) {
					r.Get("/", handlers.Workflow.List)
					r.Post("/", handlers.Workflow.Create)
					r.Get("/search/advanced", handlers.Workflow.AdvancedSearch)
					r.Post("/search/advanced", handlers.Workflow.AdvancedSearch)
					r.Get("/search/filters", handlers.Workflow.SearchFilters)
					r.Post("/validate", handlers.Workflow.Validate)
					r.Post("/test-node", handlers.Workflow.TestNode)
					r.Post("/import", handlers.Workflow.Import)

					r.Route("/{workflowId}", func(r chi.Router) {
						r.Get("/", handlers.Workflow.Get)
						r.Put("/", handlers.Workflow.Update)
						r.Delete("/", handlers.Workflow.Delete)
						r.Post("/activate", handlers.Workflow.Activate)
						r.Post("/deactivate", handlers.Workflow.Deactivate)
						r.Post("/execute", handlers.Execution.Start)
						r.Post("/clone", handlers.Workflow.Clone)
						r.Post("/duplicate", handlers.Workflow.Duplicate)
						r.Get("/export", handlers.Workflow.Export)
						r.Get("/versions", handlers.Workflow.GetVersions)
						r.Get("/versions/{version}", handlers.Workflow.GetVersion)
						r.Post("/versions/{version}/rollback", handlers.Workflow.Rollback)
						r.Get("/compare-versions", handlers.Workflow.CompareVersions)
						r.Get("/analytics", handlers.Analytics.WorkflowAnalytics)

						// Webhooks for workflow
						r.Post("/webhooks", handlers.Webhook.Create)
						r.Get("/webhooks", handlers.Webhook.List)

						// Pinned data
						r.Get("/pinned-data", handlers.PinnedData.GetAll)
						r.Post("/pinned-data", handlers.PinnedData.Set)
						r.Get("/pinned-data/{nodeId}", handlers.PinnedData.GetByNode)
						r.Delete("/pinned-data/{nodeId}", handlers.PinnedData.Delete)
					})
				})

				// Webhooks management
				r.Route("/webhooks", func(r chi.Router) {
					r.Route("/{webhookId}", func(r chi.Router) {
						r.Post("/regenerate-secret", handlers.Webhook.RegenerateSecret)
						r.Post("/activate", handlers.Webhook.Activate)
						r.Post("/deactivate", handlers.Webhook.Deactivate)
					})
				})

				// Executions
				r.Route("/executions", func(r chi.Router) {
					r.Get("/", handlers.Execution.List)
					r.Get("/search", handlers.Execution.Search)
					r.Get("/stats", handlers.Execution.Stats)
					r.Delete("/bulk", handlers.Execution.BulkDelete)

					r.Route("/{executionId}", func(r chi.Router) {
						r.Get("/", handlers.Execution.Get)
						r.Post("/cancel", handlers.Execution.Cancel)
						r.Post("/retry", handlers.Execution.Retry)
						r.Post("/replay", handlers.Execution.Replay)
						r.Post("/replay-from-node", handlers.Execution.ReplayFromNode)
						r.Get("/nodes", handlers.Execution.GetNodes)
						r.Get("/nodes/{nodeId}", handlers.Execution.GetNode)
						r.Get("/waiting", handlers.Execution.GetWaiting)

						// Binary data for execution
						r.Post("/binary", handlers.BinaryData.Upload)
						r.Get("/binary", handlers.BinaryData.List)
						r.Delete("/binary/cleanup", handlers.BinaryData.Cleanup)
					})
				})

				// Waiting executions
				r.Get("/waiting-executions", handlers.Execution.ListWaiting)

				// Binary data
				r.Route("/binary", func(r chi.Router) {
					r.Get("/stats", handlers.BinaryData.GetStats)
					r.Route("/{binaryId}", func(r chi.Router) {
						r.Get("/", handlers.BinaryData.GetInfo)
						r.Get("/download", handlers.BinaryData.Download)
						r.Delete("/", handlers.BinaryData.Delete)
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
						r.Post("/test", handlers.Credential.Test)
						r.Post("/refresh", handlers.Credential.Refresh)
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

				// Templates usage
				r.Post("/templates/{templateId}/use", handlers.Template.Use)

				// Workflow sharing
				r.Route("/workflow-shares", func(r chi.Router) {
					r.Post("/", handlers.WorkflowShare.Create)
					r.Get("/shared-by-me", handlers.WorkflowShare.SharedByMe)
					r.Get("/shared-with-me", handlers.WorkflowShare.SharedWithMe)
					r.Get("/pending", handlers.WorkflowShare.Pending)
					r.Post("/{shareId}/accept", handlers.WorkflowShare.Accept)
					r.Put("/{shareId}", handlers.WorkflowShare.Update)
					r.Delete("/{shareId}", handlers.WorkflowShare.Revoke)
				})

				// Marketplace publishing
				r.Route("/marketplace", func(r chi.Router) {
					r.Post("/", handlers.Marketplace.Publish)
					r.Get("/my-published", handlers.Marketplace.MyPublished)
					r.Post("/{templateId}/use", handlers.Marketplace.Use)
					r.Put("/{templateId}", handlers.Marketplace.Update)
					r.Post("/{templateId}/sync", handlers.Marketplace.Sync)
					r.Delete("/{templateId}", handlers.Marketplace.Unpublish)
				})
			})
		})
	})

	// Webhook trigger routes (public, outside /api/v1)
	r.Route("/webhooks", func(r chi.Router) {
		r.HandleFunc("/*", handlers.Webhook.Trigger)
	})

	// Stripe webhook
	r.Post("/stripe/webhook", handlers.Billing.StripeWebhook)

	return r
}
