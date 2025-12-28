package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/linkflow-ai/linkflow/internal/pkg/email"
	"github.com/linkflow-ai/linkflow/internal/pkg/logger"
	"github.com/linkflow-ai/linkflow/internal/pkg/streams"
	"github.com/linkflow-ai/linkflow/internal/worker"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize app with all dependencies (wire-generated)
	app, err := InitializeWorkerApp()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize worker application")
	}

	// Initialize logger
	logger.Init(app.Config.App.Environment, app.Config.App.Debug)

	log.Info().
		Str("app", app.Config.App.Name).
		Str("service", "worker").
		Msg("Starting worker service")

	// Initialize Asynq client for email queue
	asynqOpts := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", app.Config.Redis.Host, app.Config.Redis.Port),
		Password: app.Config.Redis.Password,
		DB:       app.Config.Redis.DB,
	}
	if app.Config.Redis.TLS {
		asynqOpts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	asynqClient := asynq.NewClient(asynqOpts)
	defer asynqClient.Close()

	// Initialize email service
	emailCfg := &email.Config{
		SMTPHost:     app.Config.SMTP.Host,
		SMTPPort:     app.Config.SMTP.Port,
		SMTPUser:     app.Config.SMTP.Username,
		SMTPPassword: app.Config.SMTP.Password,
		FromEmail:    app.Config.SMTP.From,
		FromName:     app.Config.SMTP.FromName,
		QueueEnabled: true,
	}
	emailSvc := email.NewService(emailCfg, asynqClient)

	defer app.Queue.Close()

	// Initialize webhook stream consumers if enabled
	var webhookConsumers []*streams.WebhookConsumer
	ctx, cancel := context.WithCancel(context.Background())

	if app.Config.Features.WebhookStream.Enabled {
		webhookStream := streams.NewWebhookStream(app.Redis.Client)

		// Start multiple consumers based on config
		consumerCount := app.Config.Features.WebhookStream.ConsumerCount
		if consumerCount < 1 {
			consumerCount = 2
		}

		for i := 0; i < consumerCount; i++ {
			consumerName := fmt.Sprintf("worker-%d-consumer-%d", os.Getpid(), i)
			consumer := streams.NewWebhookConsumer(webhookStream, app.WorkflowSvc, app.Queue, consumerName)

			if err := consumer.Start(ctx); err != nil {
				log.Error().Err(err).Int("consumer", i).Msg("Failed to start webhook consumer")
				continue
			}

			webhookConsumers = append(webhookConsumers, consumer)
			log.Info().Str("consumer", consumerName).Msg("Webhook stream consumer started")
		}

		log.Info().Int("count", len(webhookConsumers)).Msg("Webhook stream consumers running")
	}

	// Create worker
	w := worker.New(app.Config, app.ExecutionSvc, app.CredentialSvc, app.WorkflowSvc, app.BillingSvc, app.Redis.Client, emailSvc)

	// Handle shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Info().Msg("Shutting down worker...")

		// Stop webhook consumers first
		cancel()
		for _, consumer := range webhookConsumers {
			consumer.Stop()
		}

		w.Shutdown()
	}()

	// Start worker
	if err := w.Start(); err != nil {
		log.Fatal().Err(err).Msg("Worker error")
	}
}
