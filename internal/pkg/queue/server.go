package queue

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/linkflow-ai/linkflow/internal/pkg/config"
	"github.com/rs/zerolog/log"
)

type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func NewServer(cfg *config.RedisConfig, concurrency int) *Server {
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Addr(),
			Password: cfg.Password,
			DB:       cfg.DB,
		},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				QueueCritical: 6,
				QueueDefault:  3,
				QueueLow:      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				taskID := "unknown"
				if rw := task.ResultWriter(); rw != nil {
					taskID = rw.TaskID()
				}
				log.Error().
					Str("task_type", task.Type()).
					Str("task_id", taskID).
					Err(err).
					Msg("Task failed")
			}),
			Logger: &asynqLogger{},
		},
	)

	return &Server{
		server: server,
		mux:    asynq.NewServeMux(),
	}
}

func (s *Server) HandleFunc(pattern string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) Start() error {
	log.Info().Msg("Starting queue server...")
	return s.server.Start(s.mux)
}

func (s *Server) Shutdown() {
	log.Info().Msg("Shutting down queue server...")
	s.server.Shutdown()
}

// asynqLogger implements asynq.Logger interface
type asynqLogger struct{}

func (l *asynqLogger) Debug(args ...interface{}) {
	log.Debug().Msgf("%v", args)
}

func (l *asynqLogger) Info(args ...interface{}) {
	log.Info().Msgf("%v", args)
}

func (l *asynqLogger) Warn(args ...interface{}) {
	log.Warn().Msgf("%v", args)
}

func (l *asynqLogger) Error(args ...interface{}) {
	log.Error().Msgf("%v", args)
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	log.Fatal().Msgf("%v", args)
}
