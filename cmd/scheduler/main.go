package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/repositories"
	redisAdapter "github.com/linkflow-ai/linkflow/internal/adapters/persistence/redis"
	"github.com/linkflow-ai/linkflow/internal/adapters/scheduler"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
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

	// Initialize logger
	var appLogger logger.Logger
	if cfg.App.IsDevelopment() {
		appLogger = logger.NewDevelopment()
	} else {
		appLogger = logger.NewDefault()
	}

	appLogger.Info().
		Str("app", cfg.App.Name).
		Str("service", "scheduler").
		Str("environment", cfg.App.Environment).
		Msg("Starting scheduler service")

	// Initialize database
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

	// Initialize Redis
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

	// Initialize Asynq client for dispatching tasks
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.GetAddress(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer asynqClient.Close()

	// Initialize repositories
	scheduleRepo := repositories.NewScheduleRepository(db)

	// Initialize scheduler components
	dispatcher := scheduler.NewDispatcher(asynqClient, scheduleRepo)
	leaderElection := scheduler.NewLeaderElection(redisClient.Redis(), 30*time.Second)

	// Configure scheduler
	schedulerConfig := scheduler.Config{
		PollInterval:   cfg.Scheduler.PollInterval,
		DispatchBuffer: 100,
		LeaderLease:    30 * time.Second,
		InstanceID:     getInstanceID(),
	}

	// Create scheduler server
	server := scheduler.NewServer(scheduleRepo, dispatcher, leaderElection, schedulerConfig)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Start health check server
	healthServer := startHealthServer(appLogger, server, cfg.Scheduler.HealthPort)

	// Start metrics server
	var metricsServer *http.Server
	if cfg.Metrics.Enabled {
		mux := http.NewServeMux()
		mux.Handle(cfg.Metrics.Path, promhttp.Handler())
		metricsServer = &http.Server{
			Addr:               cfg.Metrics.GetAddress(),
			Handler:            mux,
			ReadHeaderTimeout:  10 * time.Second,
			ReadTimeout:        30 * time.Second,
			WriteTimeout:       30 * time.Second,
		}
		go func() {
			appLogger.Info().
				Str("address", cfg.Metrics.GetAddress()).
				Str("path", cfg.Metrics.Path).
				Msg("Starting scheduler metrics server")
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				appLogger.Error().Err(err).Msg("Metrics server error")
			}
		}()
	}

	// Handle shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		appLogger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		cancel()
	}()

	// Start scheduler
	appLogger.Info().
		Str("instance_id", schedulerConfig.InstanceID).
		Dur("poll_interval", schedulerConfig.PollInterval).
		Msg("Starting scheduler server")

	if err := server.Start(ctx); err != nil {
		appLogger.Error().Err(err).Msg("Scheduler error")
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	appLogger.Info().Msg("Shutting down scheduler...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Stop leader election
	leaderElection.Stop()

	// Stop scheduler server
	server.Stop()

	// Stop health server
	if healthServer != nil {
		_ = healthServer.Shutdown(shutdownCtx)
	}

	// Stop metrics server
	if metricsServer != nil {
		_ = metricsServer.Shutdown(shutdownCtx)
	}

	appLogger.Info().Msg("Scheduler stopped gracefully")
}

// getInstanceID returns a unique instance identifier
func getInstanceID() string {
	// Try to get from environment (useful in Kubernetes)
	if id := os.Getenv("HOSTNAME"); id != "" {
		return id
	}
	if id := os.Getenv("POD_NAME"); id != "" {
		return id
	}
	// Generate a unique ID
	hostname, _ := os.Hostname()
	return hostname + "-" + time.Now().Format("20060102150405")
}

// startHealthServer starts a health check HTTP server
func startHealthServer(appLogger logger.Logger, server *scheduler.Server, port int) *http.Server {
	if port <= 0 {
		port = 8091
	}

	mux := http.NewServeMux()

	// Liveness probe
	mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"alive"}`))
	})

	// Readiness probe
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		if server.IsLeader() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ready","leader":true}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ready","leader":false}`))
		}
	})

	// Metrics endpoint (Legacy JSON) - keep for now but consider deprecating
	mux.HandleFunc("/metrics/json", func(w http.ResponseWriter, r *http.Request) {
		metrics := server.Metrics().Snapshot()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simple JSON encoding
		w.Write([]byte(`{`))
		first := true
		for k, v := range metrics {
			if !first {
				w.Write([]byte(`,`))
			}
			first = false
			w.Write([]byte(`"` + k + `":`))
			switch val := v.(type) {
			case int64:
				w.Write([]byte(intToStr(val)))
			case string:
				w.Write([]byte(`"` + val + `"`))
			default:
				w.Write([]byte(`null`))
			}
		}
		w.Write([]byte(`}`))
	})

	httpServer := &http.Server{
		Addr:         ":" + intToStr(int64(port)),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		appLogger.Info().Int("port", port).Msg("Starting health check server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error().Err(err).Msg("Health server error")
		}
	}()

	return httpServer
}

func intToStr(n int64) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
