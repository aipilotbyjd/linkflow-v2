package main

import (
	"github.com/linkflow-ai/linkflow/internal/api"
	"github.com/linkflow-ai/linkflow/internal/pkg/logger"
	"github.com/rs/zerolog/log"

	// Import node packages to register them via init()
	_ "github.com/linkflow-ai/linkflow/internal/worker/nodes/actions"
	_ "github.com/linkflow-ai/linkflow/internal/worker/nodes/integrations"
	_ "github.com/linkflow-ai/linkflow/internal/worker/nodes/logic"
	_ "github.com/linkflow-ai/linkflow/internal/worker/nodes/triggers"
)

func main() {
	// Initialize app with all dependencies (wire-generated)
	app, err := InitializeApp()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize application")
	}

	// Initialize logger
	logger.Init(app.Config.App.Environment, app.Config.App.Debug)

	log.Info().
		Str("app", app.Config.App.Name).
		Str("env", app.Config.App.Environment).
		Msg("Starting API server")

	// Create server with all dependencies
	server := api.NewServer(
		app.Config,
		app.Services,
		app.Repos,
		app.JWTManager,
		app.Redis,
		app.Queue,
		app.DB,
	)

	// Start server
	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server error")
	}
}
