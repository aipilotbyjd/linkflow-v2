package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/rbac"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
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
	APIKeyRepo    user.APIKeyRepository
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
	RBAC          RBACHandlers
	Analytics     AnalyticsHandlers
	Invitation    InvitationHandlers
	AIBuilder     AIBuilderHandlers
	Variable      VariableHandlers
	Comment       CommentHandlers
	Replay        ReplayHandlers
	WebSocket     http.HandlerFunc
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
	MyPermissions     http.HandlerFunc
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
	UpdateSecurity   http.HandlerFunc
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
	GetDashboard       http.HandlerFunc
	GetUsageStatus     http.HandlerFunc
	// Usage Alerts
	ListAlerts         http.HandlerFunc
	CreateAlert        http.HandlerFunc
	UpdateAlert        http.HandlerFunc
	DeleteAlert        http.HandlerFunc
	GetAlertHistory    http.HandlerFunc
	AcknowledgeAlert   http.HandlerFunc
	// BYOK
	ListBYOKProviders  http.HandlerFunc
	ListBYOK           http.HandlerFunc
	CreateBYOK         http.HandlerFunc
	DeleteBYOK         http.HandlerFunc
	// Credit Top-Up
	ListCreditPacks    http.HandlerFunc
	GetTopUpSettings   http.HandlerFunc
	UpdateTopUpSettings http.HandlerFunc
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
	StreamStats         http.HandlerFunc
	ReplayDLQ           http.HandlerFunc
	TrimStream          http.HandlerFunc
	GetDisabledNodes    http.HandlerFunc
	UpdateDisabledNodes http.HandlerFunc
}

// RBACHandlers holds RBAC handlers
type RBACHandlers struct {
	ListRoles       http.HandlerFunc
	CreateRole      http.HandlerFunc
	UpdateRole      http.HandlerFunc
	DeleteRole      http.HandlerFunc
	ListPermissions http.HandlerFunc
}

// AnalyticsHandlers holds analytics handlers
type AnalyticsHandlers struct {
	WorkflowAnalytics  http.HandlerFunc
	WorkspaceAnalytics http.HandlerFunc
}

// InvitationHandlers holds invitation handlers
type InvitationHandlers struct {
	GetInfo http.HandlerFunc
	Accept  http.HandlerFunc
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

// CommentHandlers holds comment handlers
type CommentHandlers struct {
	List     http.HandlerFunc
	Create   http.HandlerFunc
	Update   http.HandlerFunc
	Delete   http.HandlerFunc
	Resolve  http.HandlerFunc
}

// ReplayHandlers holds debug replay handlers
type ReplayHandlers struct {
	Create    http.HandlerFunc
	GetEvents http.HandlerFunc
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

	// Health endpoints (public)
	r.Get("/health", handlers.Health.Health)

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

		// Invitations (public with token)
		r.Get("/invitations/{token}", handlers.Invitation.GetInfo)
		r.Post("/invitations/accept", handlers.Invitation.Accept) // Authenticated usually, but might handle login flow

		// WebSocket (authenticated)
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthWithBlacklist(cfg.JWTManager, cfg.JWTBlacklist))
			r.Get("/ws", handlers.WebSocket)
		})

		// Protected routes (Mixed Auth)
		r.Group(func(r chi.Router) {
			r.Use(middleware.APIKey(cfg.APIKeyRepo))
			r.Use(middleware.OptionalAuth(cfg.JWTManager))
			r.Use(middleware.RequireAuth)

			// User workspaces
			r.Get("/workspaces", handlers.Workspace.List)
			r.Post("/workspaces", handlers.Workspace.Create)

			// Admin routes
			r.Route("/admin", func(r chi.Router) {
				r.Get("/streams/webhooks/stats", handlers.Admin.StreamStats)
				r.Post("/streams/webhooks/replay", handlers.Admin.ReplayDLQ)
				r.Post("/streams/webhooks/trim", handlers.Admin.TrimStream)
				r.Get("/disabled-nodes", handlers.Admin.GetDisabledNodes)
				r.Put("/disabled-nodes", handlers.Admin.UpdateDisabledNodes)
			})

			// Workspace-scoped routes
			r.Route("/workspaces/{workspaceId}", func(r chi.Router) {
				r.Use(middleware.Tenant(cfg.MemberRepo, cfg.WorkspaceRepo))

				// User Context in Workspace
				r.Get("/users/me/permissions", handlers.User.MyPermissions)

				// Workspace Management
				r.With(middleware.RequirePermission(rbac.PermWorkspaceRead)).Get("/", handlers.Workspace.Get)
				r.With(middleware.RequirePermission(rbac.PermWorkspaceWrite)).Put("/", handlers.Workspace.Update)
				r.With(middleware.RequirePermission(rbac.PermWorkspaceDelete)).Delete("/", handlers.Workspace.Delete)

				// Dashboard
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequirePermission(rbac.PermWorkspaceRead))
					r.Get("/dashboard", handlers.Dashboard.GetDashboard)
					r.Get("/stats", handlers.Dashboard.GetQuickStats)
				})

				// Analytics
				r.With(middleware.RequirePermission(rbac.PermWorkspaceRead)).Get("/analytics", handlers.Analytics.WorkspaceAnalytics)

				// Members
				r.Route("/members", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermMemberRead)).Get("/", handlers.Workspace.ListMembers)
					r.With(middleware.RequirePermission(rbac.PermMemberWrite)).Post("/", handlers.Workspace.InviteMember)
					r.With(middleware.RequirePermission(rbac.PermMemberWrite)).Post("/invite", handlers.Workspace.InviteMember)
					r.With(middleware.RequirePermission(rbac.PermMemberWrite)).Put("/{memberId}", handlers.Workspace.UpdateMember)
					r.With(middleware.RequirePermission(rbac.PermMemberDelete)).Delete("/{memberId}", handlers.Workspace.RemoveMember)
				})

				// RBAC Roles
				r.Route("/roles", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermWorkspaceWrite)).Get("/", handlers.RBAC.ListRoles)
					r.With(middleware.RequirePermission(rbac.PermWorkspaceWrite)).Post("/", handlers.RBAC.CreateRole)
					r.Route("/{roleId}", func(r chi.Router) {
						r.With(middleware.RequirePermission(rbac.PermWorkspaceWrite)).Put("/", handlers.RBAC.UpdateRole)
						r.With(middleware.RequirePermission(rbac.PermWorkspaceWrite)).Delete("/", handlers.RBAC.DeleteRole)
					})
				})

				// Billing
				r.Route("/billing", func(r chi.Router) {
					r.Use(middleware.RequirePermission(rbac.PermBillingRead))
					r.Get("/subscription", handlers.Billing.GetSubscription)
					r.Get("/usage", handlers.Billing.GetUsage)
					r.Get("/invoices", handlers.Billing.GetInvoices)
					r.With(middleware.RequirePermission(rbac.PermBillingWrite)).Post("/subscription", handlers.Billing.CreateSubscription)
					r.With(middleware.RequirePermission(rbac.PermBillingWrite)).Delete("/subscription", handlers.Billing.CancelSubscription)
					r.Get("/dashboard", handlers.Billing.GetDashboard)
					r.Get("/usage-status", handlers.Billing.GetUsageStatus)
					
					// Usage Alerts
					r.Route("/alerts", func(r chi.Router) {
						r.Get("/", handlers.Billing.ListAlerts)
						r.With(middleware.RequirePermission(rbac.PermBillingWrite)).Post("/", handlers.Billing.CreateAlert)
						r.Get("/history", handlers.Billing.GetAlertHistory)
						r.With(middleware.RequirePermission(rbac.PermBillingWrite)).Put("/{alertId}", handlers.Billing.UpdateAlert)
						r.With(middleware.RequirePermission(rbac.PermBillingWrite)).Delete("/{alertId}", handlers.Billing.DeleteAlert)
						r.With(middleware.RequirePermission(rbac.PermBillingWrite)).Post("/{alertLogId}/acknowledge", handlers.Billing.AcknowledgeAlert)
					})
					
					// BYOK (Bring Your Own Key)
					r.Route("/byok", func(r chi.Router) {
						r.Get("/providers", handlers.Billing.ListBYOKProviders)
						r.Get("/", handlers.Billing.ListBYOK)
						r.With(middleware.RequirePermission(rbac.PermBillingWrite)).Post("/", handlers.Billing.CreateBYOK)
						r.With(middleware.RequirePermission(rbac.PermBillingWrite)).Delete("/{byokId}", handlers.Billing.DeleteBYOK)
					})
					
					// Credit Top-Up
					r.Route("/topup", func(r chi.Router) {
						r.Get("/packs", handlers.Billing.ListCreditPacks)
						r.Get("/settings", handlers.Billing.GetTopUpSettings)
						r.With(middleware.RequirePermission(rbac.PermBillingWrite)).Put("/settings", handlers.Billing.UpdateTopUpSettings)
					})
				})

				// OAuth
				r.Get("/oauth/authorize/{provider}", handlers.OAuth.Authorize)
				r.Get("/oauth/callback/{provider}", handlers.OAuth.Callback)

				// Folders
				r.Route("/folders", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/", handlers.Folder.List)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/tree", handlers.Folder.Tree)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/", handlers.Folder.Create)
					r.Route("/{folderId}", func(r chi.Router) {
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/", handlers.Folder.Get)
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Put("/", handlers.Folder.Update)
						r.With(middleware.RequirePermission(rbac.PermWorkflowDelete)).Delete("/", handlers.Folder.Delete)
					})
				})

				// AI Workflow Builder
				r.Route("/ai-builder", func(r chi.Router) {
					r.Use(middleware.RequirePermission(rbac.PermWorkflowWrite))
					r.Post("/generate", handlers.AIBuilder.Generate)
					r.Post("/suggest", handlers.AIBuilder.Suggest)
					r.Post("/explain", handlers.AIBuilder.Explain)
				})

				// Variables & Environments
				r.Route("/variables", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/", handlers.Variable.List)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/resolve", handlers.Variable.Resolve)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/", handlers.Variable.Create)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Put("/{variableId}", handlers.Variable.Update)
					r.With(middleware.RequirePermission(rbac.PermWorkflowDelete)).Delete("/{variableId}", handlers.Variable.Delete)
				})
				r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/environments", handlers.Variable.ListEnvironments)

				// Comments
				r.Route("/workflows/{workflowId}/comments", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/", handlers.Comment.List)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/", handlers.Comment.Create)
					r.Route("/{commentId}", func(r chi.Router) {
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Put("/", handlers.Comment.Update)
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Delete("/", handlers.Comment.Delete)
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/resolve", handlers.Comment.Resolve)
					})
				})

				// Debug Replay
				r.Route("/replay", func(r chi.Router) {
					r.Use(middleware.RequirePermission(rbac.PermWorkflowExecute))
					r.Post("/", handlers.Replay.Create)
					r.Get("/{sessionId}/events", handlers.Replay.GetEvents)
				})

				// Workflows
				r.Route("/workflows", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/", handlers.Workflow.List)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/", handlers.Workflow.Create)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/search/advanced", handlers.Workflow.AdvancedSearch)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Post("/search/advanced", handlers.Workflow.AdvancedSearch)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/search/filters", handlers.Workflow.SearchFilters)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/validate", handlers.Workflow.Validate)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/test-node", handlers.Workflow.TestNode)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/import", handlers.Workflow.Import)

					r.Route("/{workflowId}", func(r chi.Router) {
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/", handlers.Workflow.Get)
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Put("/", handlers.Workflow.Update)
						r.With(middleware.RequirePermission(rbac.PermWorkflowDelete)).Delete("/", handlers.Workflow.Delete)
						r.With(middleware.RequirePermission(rbac.PermWorkflowPublish)).Post("/activate", handlers.Workflow.Activate)
						r.With(middleware.RequirePermission(rbac.PermWorkflowPublish)).Post("/deactivate", handlers.Workflow.Deactivate)
						r.With(middleware.RequirePermission(rbac.PermWorkflowExecute)).Post("/execute", handlers.Execution.Start)
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/clone", handlers.Workflow.Clone)
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/duplicate", handlers.Workflow.Duplicate)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/export", handlers.Workflow.Export)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/versions", handlers.Workflow.GetVersions)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/versions/{version}", handlers.Workflow.GetVersion)
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/versions/{version}/rollback", handlers.Workflow.Rollback)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/compare-versions", handlers.Workflow.CompareVersions)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/analytics", handlers.Analytics.WorkflowAnalytics)

						// Webhooks for workflow
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/webhooks", handlers.Webhook.Create)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/webhooks", handlers.Webhook.List)

						// Pinned data
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/pinned-data", handlers.PinnedData.GetAll)
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/pinned-data", handlers.PinnedData.Set)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/pinned-data/{nodeId}", handlers.PinnedData.GetByNode)
						r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Delete("/pinned-data/{nodeId}", handlers.PinnedData.Delete)
					})
				})

				// Webhooks management
				r.Route("/webhooks", func(r chi.Router) {
					r.Use(middleware.RequirePermission(rbac.PermWorkflowWrite))
					r.Route("/{webhookId}", func(r chi.Router) {
						r.Post("/regenerate-secret", handlers.Webhook.RegenerateSecret)
						r.Post("/activate", handlers.Webhook.Activate)
						r.Post("/deactivate", handlers.Webhook.Deactivate)
						r.Put("/security", handlers.Webhook.UpdateSecurity)
					})
				})

				// Executions
				r.Route("/executions", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/", handlers.Execution.List)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/search", handlers.Execution.Search)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/stats", handlers.Execution.Stats)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/waiting", handlers.Execution.ListWaiting)
					r.With(middleware.RequirePermission(rbac.PermWorkflowDelete)).Delete("/bulk", handlers.Execution.BulkDelete)

					r.Route("/{executionId}", func(r chi.Router) {
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/", handlers.Execution.Get)
						r.With(middleware.RequirePermission(rbac.PermWorkflowExecute)).Post("/cancel", handlers.Execution.Cancel)
						r.With(middleware.RequirePermission(rbac.PermWorkflowExecute)).Post("/retry", handlers.Execution.Retry)
						r.With(middleware.RequirePermission(rbac.PermWorkflowExecute)).Post("/replay", handlers.Execution.Replay)
						r.With(middleware.RequirePermission(rbac.PermWorkflowExecute)).Post("/replay-from-node", handlers.Execution.ReplayFromNode)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/nodes", handlers.Execution.GetNodes)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/nodes/{nodeId}", handlers.Execution.GetNode)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/waiting", handlers.Execution.GetWaiting)

						// Binary data for execution
						r.With(middleware.RequirePermission(rbac.PermWorkflowExecute)).Post("/binary", handlers.BinaryData.Upload)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/binary", handlers.BinaryData.List)
						r.With(middleware.RequirePermission(rbac.PermWorkflowDelete)).Delete("/binary/cleanup", handlers.BinaryData.Cleanup)
					})
				})

				// Binary data
				r.Route("/binary", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/stats", handlers.BinaryData.GetStats)
					r.Route("/{binaryId}", func(r chi.Router) {
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/", handlers.BinaryData.GetInfo)
						r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/download", handlers.BinaryData.Download)
						r.With(middleware.RequirePermission(rbac.PermWorkflowDelete)).Delete("/", handlers.BinaryData.Delete)
					})
				})

				// Credentials
				r.Route("/credentials", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermCredentialRead)).Get("/", handlers.Credential.List)
					r.With(middleware.RequirePermission(rbac.PermCredentialWrite)).Post("/", handlers.Credential.Create)
					r.Route("/{credentialId}", func(r chi.Router) {
						r.With(middleware.RequirePermission(rbac.PermCredentialRead)).Get("/", handlers.Credential.Get)
						r.With(middleware.RequirePermission(rbac.PermCredentialWrite)).Put("/", handlers.Credential.Update)
						r.With(middleware.RequirePermission(rbac.PermCredentialDelete)).Delete("/", handlers.Credential.Delete)
						r.With(middleware.RequirePermission(rbac.PermCredentialWrite)).Post("/test", handlers.Credential.Test)
						r.With(middleware.RequirePermission(rbac.PermCredentialWrite)).Post("/refresh", handlers.Credential.Refresh)
					})
				})

				// Schedules
				r.Route("/schedules", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermScheduleRead)).Get("/", handlers.Schedule.List)
					r.With(middleware.RequirePermission(rbac.PermScheduleWrite)).Post("/", handlers.Schedule.Create)
					r.Route("/{scheduleId}", func(r chi.Router) {
						r.With(middleware.RequirePermission(rbac.PermScheduleRead)).Get("/", handlers.Schedule.Get)
						r.With(middleware.RequirePermission(rbac.PermScheduleWrite)).Put("/", handlers.Schedule.Update)
						r.With(middleware.RequirePermission(rbac.PermScheduleDelete)).Delete("/", handlers.Schedule.Delete)
						r.With(middleware.RequirePermission(rbac.PermScheduleWrite)).Post("/pause", handlers.Schedule.Pause)
						r.With(middleware.RequirePermission(rbac.PermScheduleWrite)).Post("/resume", handlers.Schedule.Resume)
					})
				})

				// Workflow sharing
				r.Route("/workflow-shares", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/", handlers.WorkflowShare.Create)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/shared-by-me", handlers.WorkflowShare.SharedByMe)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/shared-with-me", handlers.WorkflowShare.SharedWithMe)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/pending", handlers.WorkflowShare.Pending)
					r.Post("/{shareId}/accept", handlers.WorkflowShare.Accept)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Put("/{shareId}", handlers.WorkflowShare.Update)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Delete("/{shareId}", handlers.WorkflowShare.Revoke)
				})

				// Marketplace publishing
				r.Route("/marketplace", func(r chi.Router) {
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/", handlers.Marketplace.Publish)
					r.With(middleware.RequirePermission(rbac.PermWorkflowRead)).Get("/my-published", handlers.Marketplace.MyPublished)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/{templateId}/use", handlers.Marketplace.Use)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Put("/{templateId}", handlers.Marketplace.Update)
					r.With(middleware.RequirePermission(rbac.PermWorkflowWrite)).Post("/{templateId}/sync", handlers.Marketplace.Sync)
					r.With(middleware.RequirePermission(rbac.PermWorkflowDelete)).Delete("/{templateId}", handlers.Marketplace.Unpublish)
				})
			})
		})

		// User-only routes (JWT required)
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthWithBlacklist(cfg.JWTManager, cfg.JWTBlacklist))
			r.Get("/users/me", handlers.User.GetCurrentUser)
			r.Put("/users/me", handlers.User.UpdateCurrentUser)
			r.Get("/api-keys", handlers.APIKey.List)
			r.Post("/api-keys", handlers.APIKey.Create)
			r.Delete("/api-keys/{keyId}", handlers.APIKey.Revoke)
			r.Get("/permissions", handlers.RBAC.ListPermissions)
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
