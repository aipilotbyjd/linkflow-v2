package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger interface for structured logging
type Logger interface {
	Debug() Event
	Info() Event
	Warn() Event
	Error() Event
	Fatal() Event
	WithContext(ctx context.Context) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

// Event interface for building log events
type Event interface {
	Str(key, value string) Event
	Int(key string, value int) Event
	Int64(key string, value int64) Event
	Float64(key string, value float64) Event
	Bool(key string, value bool) Event
	Err(err error) Event
	Interface(key string, value interface{}) Event
	Dur(key string, d time.Duration) Event
	Time(key string, t time.Time) Event
	Msg(msg string)
	Msgf(format string, args ...interface{})
	Send()
}

// ZerologLogger implements Logger using zerolog
type ZerologLogger struct {
	logger zerolog.Logger
}

// ZerologEvent implements Event using zerolog
type ZerologEvent struct {
	event *zerolog.Event
}

// Config holds logger configuration
type Config struct {
	Level      string
	Format     string // json or console
	TimeFormat string
	Output     io.Writer
}

// New creates a new zerolog-based logger
func New(cfg Config) *ZerologLogger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	// Set time format
	if cfg.TimeFormat != "" {
		zerolog.TimeFieldFormat = cfg.TimeFormat
	} else {
		zerolog.TimeFieldFormat = time.RFC3339
	}

	// Create output
	var output io.Writer = cfg.Output
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        cfg.Output,
			TimeFormat: time.RFC3339,
		}
	}

	// Parse level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	logger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &ZerologLogger{logger: logger}
}

// NewDefault creates a default logger
func NewDefault() *ZerologLogger {
	return New(Config{
		Level:  "info",
		Format: "json",
		Output: os.Stdout,
	})
}

// NewDevelopment creates a development logger with console output
func NewDevelopment() *ZerologLogger {
	return New(Config{
		Level:  "debug",
		Format: "console",
		Output: os.Stdout,
	})
}

// Debug starts a debug level event
func (l *ZerologLogger) Debug() Event {
	return &ZerologEvent{event: l.logger.Debug()}
}

// Info starts an info level event
func (l *ZerologLogger) Info() Event {
	return &ZerologEvent{event: l.logger.Info()}
}

// Warn starts a warn level event
func (l *ZerologLogger) Warn() Event {
	return &ZerologEvent{event: l.logger.Warn()}
}

// Error starts an error level event
func (l *ZerologLogger) Error() Event {
	return &ZerologEvent{event: l.logger.Error()}
}

// Fatal starts a fatal level event
func (l *ZerologLogger) Fatal() Event {
	return &ZerologEvent{event: l.logger.Fatal()}
}

// WithContext returns a logger with context
func (l *ZerologLogger) WithContext(ctx context.Context) Logger {
	return &ZerologLogger{logger: l.logger.With().Logger()}
}

// WithField returns a logger with an additional field
func (l *ZerologLogger) WithField(key string, value interface{}) Logger {
	return &ZerologLogger{logger: l.logger.With().Interface(key, value).Logger()}
}

// WithFields returns a logger with additional fields
func (l *ZerologLogger) WithFields(fields map[string]interface{}) Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &ZerologLogger{logger: ctx.Logger()}
}

// WithError returns a logger with an error field
func (l *ZerologLogger) WithError(err error) Logger {
	return &ZerologLogger{logger: l.logger.With().Err(err).Logger()}
}

// Event methods

func (e *ZerologEvent) Str(key, value string) Event {
	e.event = e.event.Str(key, value)
	return e
}

func (e *ZerologEvent) Int(key string, value int) Event {
	e.event = e.event.Int(key, value)
	return e
}

func (e *ZerologEvent) Int64(key string, value int64) Event {
	e.event = e.event.Int64(key, value)
	return e
}

func (e *ZerologEvent) Float64(key string, value float64) Event {
	e.event = e.event.Float64(key, value)
	return e
}

func (e *ZerologEvent) Bool(key string, value bool) Event {
	e.event = e.event.Bool(key, value)
	return e
}

func (e *ZerologEvent) Err(err error) Event {
	e.event = e.event.Err(err)
	return e
}

func (e *ZerologEvent) Interface(key string, value interface{}) Event {
	e.event = e.event.Interface(key, value)
	return e
}

func (e *ZerologEvent) Dur(key string, d time.Duration) Event {
	e.event = e.event.Dur(key, d)
	return e
}

func (e *ZerologEvent) Time(key string, t time.Time) Event {
	e.event = e.event.Time(key, t)
	return e
}

func (e *ZerologEvent) Msg(msg string) {
	e.event.Msg(msg)
}

func (e *ZerologEvent) Msgf(format string, args ...interface{}) {
	e.event.Msgf(format, args...)
}

func (e *ZerologEvent) Send() {
	e.event.Send()
}

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

func (l *NoOpLogger) Debug() Event                                    { return &NoOpEvent{} }
func (l *NoOpLogger) Info() Event                                     { return &NoOpEvent{} }
func (l *NoOpLogger) Warn() Event                                     { return &NoOpEvent{} }
func (l *NoOpLogger) Error() Event                                    { return &NoOpEvent{} }
func (l *NoOpLogger) Fatal() Event                                    { return &NoOpEvent{} }
func (l *NoOpLogger) WithContext(ctx context.Context) Logger          { return l }
func (l *NoOpLogger) WithField(key string, value interface{}) Logger  { return l }
func (l *NoOpLogger) WithFields(fields map[string]interface{}) Logger { return l }
func (l *NoOpLogger) WithError(err error) Logger                      { return l }

type NoOpEvent struct{}

func (e *NoOpEvent) Str(key, value string) Event                   { return e }
func (e *NoOpEvent) Int(key string, value int) Event               { return e }
func (e *NoOpEvent) Int64(key string, value int64) Event           { return e }
func (e *NoOpEvent) Float64(key string, value float64) Event       { return e }
func (e *NoOpEvent) Bool(key string, value bool) Event             { return e }
func (e *NoOpEvent) Err(err error) Event                           { return e }
func (e *NoOpEvent) Interface(key string, value interface{}) Event { return e }
func (e *NoOpEvent) Dur(key string, d time.Duration) Event         { return e }
func (e *NoOpEvent) Time(key string, t time.Time) Event            { return e }
func (e *NoOpEvent) Msg(msg string)                                {}
func (e *NoOpEvent) Msgf(format string, args ...interface{})       {}
func (e *NoOpEvent) Send()                                         {}
