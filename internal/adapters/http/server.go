package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/routes"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	config *config.ServerConfig
	logger logger.Logger
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.ServerConfig, router http.Handler, log logger.Logger) *Server {
	return &Server{
		server: &http.Server{
			Addr:         cfg.GetAddress(),
			Handler:      router,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		config: cfg,
		logger: log,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info().
		Str("address", s.config.GetAddress()).
		Msg("Starting HTTP server")

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// ShutdownWithTimeout shuts down the server with a timeout
func (s *Server) ShutdownWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}

// ServerDeps holds all dependencies needed for the server
type ServerDeps struct {
	Config        *config.Config
	Logger        logger.Logger
	RouteConfig   routes.Config
	RouteHandlers routes.Handlers
}

// NewServerWithDeps creates a server with all dependencies
func NewServerWithDeps(deps ServerDeps) *Server {
	router := routes.NewRouter(deps.RouteConfig, deps.RouteHandlers)
	return NewServer(&deps.Config.Server, router, deps.Logger)
}
