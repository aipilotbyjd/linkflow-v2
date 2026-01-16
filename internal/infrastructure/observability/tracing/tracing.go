package tracing

import (
	"context"
	"sync"
)

// Config holds tracing configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerEndpoint string
	Enabled        bool
	SampleRate     float64
}

// Span represents a tracing span (no-op implementation)
type Span struct {
	name       string
	attributes map[string]interface{}
	mu         sync.Mutex
}

// End ends the span
func (s *Span) End() {}

// SetAttribute sets an attribute on the span
func (s *Span) SetAttribute(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.attributes == nil {
		s.attributes = make(map[string]interface{})
	}
	s.attributes[key] = value
}

// RecordError records an error on the span
func (s *Span) RecordError(err error) {}

// Tracer provides tracing functionality (no-op implementation)
type Tracer struct {
	config Config
}

// New creates a new tracer
func New(config Config) (*Tracer, error) {
	// This is a no-op implementation
	// For actual OpenTelemetry integration, add the otel packages
	return &Tracer{config: config}, nil
}

// Start starts a new span
func (t *Tracer) Start(ctx context.Context, name string) (context.Context, *Span) {
	span := &Span{name: name}
	return ctx, span
}

// Shutdown shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	return nil
}

// StartWorkflowSpan starts a span for workflow execution
func (t *Tracer) StartWorkflowSpan(ctx context.Context, workflowID, workflowName string) (context.Context, *Span) {
	ctx, span := t.Start(ctx, "workflow.execute")
	span.SetAttribute("workflow.id", workflowID)
	span.SetAttribute("workflow.name", workflowName)
	return ctx, span
}

// StartNodeSpan starts a span for node execution
func (t *Tracer) StartNodeSpan(ctx context.Context, nodeID, nodeType, nodeName string) (context.Context, *Span) {
	ctx, span := t.Start(ctx, "node.execute")
	span.SetAttribute("node.id", nodeID)
	span.SetAttribute("node.type", nodeType)
	span.SetAttribute("node.name", nodeName)
	return ctx, span
}

// StartHTTPSpan starts a span for an HTTP request
func (t *Tracer) StartHTTPSpan(ctx context.Context, method, path string) (context.Context, *Span) {
	ctx, span := t.Start(ctx, "http.request")
	span.SetAttribute("http.method", method)
	span.SetAttribute("http.path", path)
	return ctx, span
}

// StartDBSpan starts a span for a database operation
func (t *Tracer) StartDBSpan(ctx context.Context, operation, table string) (context.Context, *Span) {
	ctx, span := t.Start(ctx, "db."+operation)
	span.SetAttribute("db.operation", operation)
	span.SetAttribute("db.table", table)
	return ctx, span
}
