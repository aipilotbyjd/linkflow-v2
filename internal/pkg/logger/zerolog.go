package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(environment string, debug bool) {
	zerolog.TimeFieldFormat = time.RFC3339

	var output io.Writer = os.Stdout

	if environment == "development" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}
	}

	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}

	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(level)
}

func WithContext(ctx map[string]interface{}) zerolog.Logger {
	logger := log.With()
	for k, v := range ctx {
		logger = logger.Interface(k, v)
	}
	return logger.Logger()
}

func WithRequestID(requestID string) zerolog.Logger {
	return log.With().Str("request_id", requestID).Logger()
}

func WithUserID(userID string) zerolog.Logger {
	return log.With().Str("user_id", userID).Logger()
}

func WithWorkspaceID(workspaceID string) zerolog.Logger {
	return log.With().Str("workspace_id", workspaceID).Logger()
}

func WithWorkflowID(workflowID string) zerolog.Logger {
	return log.With().Str("workflow_id", workflowID).Logger()
}

func WithExecutionID(executionID string) zerolog.Logger {
	return log.With().Str("execution_id", executionID).Logger()
}
