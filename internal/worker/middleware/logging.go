package middleware

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/processor"
	"github.com/rs/zerolog/log"
)

// LoggingMiddleware logs node execution details
type LoggingMiddleware struct {
	logInput    bool
	logOutput   bool
	logDuration bool
}

// LoggingOptions configures logging middleware
type LoggingOptions struct {
	LogInput    bool
	LogOutput   bool
	LogDuration bool
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(opts ...LoggingOptions) *LoggingMiddleware {
	m := &LoggingMiddleware{
		logInput:    false,
		logOutput:   false,
		logDuration: true,
	}

	if len(opts) > 0 {
		m.logInput = opts[0].LogInput
		m.logOutput = opts[0].LogOutput
		m.logDuration = opts[0].LogDuration
	}

	return m
}

// Execute implements Middleware
func (m *LoggingMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	logger := log.With().
		Str("execution_id", rctx.ExecutionID.String()).
		Str("workflow_id", rctx.WorkflowID.String()).
		Str("node_id", node.ID).
		Str("node_type", node.Type).
		Str("node_name", node.Name).
		Logger()

	// Log start
	startLog := logger.Debug()
	if m.logInput {
		input := rctx.PrepareNodeInput(node)
		startLog = startLog.Interface("input", truncateForLog(input, 500))
	}
	startLog.Msg("Node execution started")

	startTime := time.Now()

	// Execute
	result, err := next(ctx)
	duration := time.Since(startTime)

	// Log completion
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", duration).
			Msg("Node execution failed")
		return nil, err
	}

	endLog := logger.Debug().Dur("duration", duration)
	if m.logOutput && result != nil {
		endLog = endLog.Interface("output", truncateForLog(result.Output, 500))
	}
	endLog.Msg("Node execution completed")

	return result.Output, nil
}

// truncateForLog truncates large values for logging
func truncateForLog(data interface{}, maxLen int) interface{} {
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case string:
		if len(v) > maxLen {
			return v[:maxLen] + "...(truncated)"
		}
		return v

	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = truncateForLog(val, maxLen)
		}
		return result

	case []interface{}:
		if len(v) > 10 {
			truncated := make([]interface{}, 10)
			copy(truncated, v[:10])
			return map[string]interface{}{
				"_truncated": true,
				"_total":     len(v),
				"items":      truncated,
			}
		}
		return v

	default:
		return v
	}
}

// StructuredLogEntry represents a structured log entry
type StructuredLogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	ExecutionID string                 `json:"execution_id"`
	WorkflowID  string                 `json:"workflow_id"`
	NodeID      string                 `json:"node_id"`
	NodeType    string                 `json:"node_type"`
	Message     string                 `json:"message"`
	Duration    int64                  `json:"duration_ms,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NodeLogger provides node-specific logging
type NodeLogger struct {
	rctx *processor.RuntimeContext
	node *processor.NodeDefinition
}

// NewNodeLogger creates a node logger
func NewNodeLogger(rctx *processor.RuntimeContext, node *processor.NodeDefinition) *NodeLogger {
	return &NodeLogger{rctx: rctx, node: node}
}

// Debug logs a debug message
func (l *NodeLogger) Debug(msg string, fields ...map[string]interface{}) {
	logger := log.Debug().
		Str("execution_id", l.rctx.ExecutionID.String()).
		Str("node_id", l.node.ID)

	if len(fields) > 0 {
		for k, v := range fields[0] {
			logger = logger.Interface(k, v)
		}
	}

	logger.Msg(msg)
}

// Info logs an info message
func (l *NodeLogger) Info(msg string, fields ...map[string]interface{}) {
	logger := log.Info().
		Str("execution_id", l.rctx.ExecutionID.String()).
		Str("node_id", l.node.ID)

	if len(fields) > 0 {
		for k, v := range fields[0] {
			logger = logger.Interface(k, v)
		}
	}

	logger.Msg(msg)
}

// Warn logs a warning message
func (l *NodeLogger) Warn(msg string, fields ...map[string]interface{}) {
	logger := log.Warn().
		Str("execution_id", l.rctx.ExecutionID.String()).
		Str("node_id", l.node.ID)

	if len(fields) > 0 {
		for k, v := range fields[0] {
			logger = logger.Interface(k, v)
		}
	}

	logger.Msg(msg)
}

// Error logs an error message
func (l *NodeLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	logger := log.Error().
		Err(err).
		Str("execution_id", l.rctx.ExecutionID.String()).
		Str("node_id", l.node.ID)

	if len(fields) > 0 {
		for k, v := range fields[0] {
			logger = logger.Interface(k, v)
		}
	}

	logger.Msg(msg)
}
