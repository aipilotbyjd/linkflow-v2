package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/config"
)

// Init initializes Sentry with the given configuration
func Init(cfg config.SentryConfig, serviceName string) error {
	if !cfg.Enabled || cfg.DSN == "" {
		return nil
	}

	environment := cfg.Environment
	if environment == "" {
		environment = "development"
	}

	sampleRate := cfg.TracesSampleRate
	if sampleRate == 0 {
		sampleRate = 0.1 // Default 10% tracing
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      environment,
		Debug:            cfg.Debug,
		TracesSampleRate: sampleRate,
		ServerName:       serviceName,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Add custom processing here if needed
			return event
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// Flush flushes any buffered events to Sentry
func Flush(timeout time.Duration) {
	sentry.Flush(timeout)
}

// CaptureException captures an exception to Sentry
func CaptureException(err error) {
	sentry.CaptureException(err)
}

// CaptureMessage captures a message to Sentry
func CaptureMessage(message string) {
	sentry.CaptureMessage(message)
}

// WithScope runs a function with a new Sentry scope
func WithScope(fn func(scope *sentry.Scope)) {
	sentry.WithScope(fn)
}

// SetUser sets the current user for Sentry events
func SetUser(id, email, username string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       id,
			Email:    email,
			Username: username,
		})
	})
}

// SetTag sets a tag for all future Sentry events
func SetTag(key, value string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
}

// SetContext sets additional context for Sentry events
func SetContext(key string, context map[string]interface{}) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetContext(key, context)
	})
}

// RecoverWithContext recovers from a panic and reports to Sentry
func RecoverWithContext(ctx map[string]interface{}) {
	if r := recover(); r != nil {
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetContext("panic_context", ctx)
			if err, ok := r.(error); ok {
				sentry.CaptureException(err)
			} else {
				sentry.CaptureMessage("Panic recovered")
			}
		})
		sentry.Flush(2 * time.Second)
		panic(r) // Re-panic after reporting
	}
}
