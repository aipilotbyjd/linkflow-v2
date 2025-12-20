package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/worker/processor"
)

// TracingMiddleware adds distributed tracing to node execution
type TracingMiddleware struct {
	serviceName string
	tracer      Tracer
}

// Tracer is the interface for distributed tracing
type Tracer interface {
	StartSpan(ctx context.Context, operationName string, opts ...SpanOption) (context.Context, Span)
}

// Span represents a trace span
type Span interface {
	SetTag(key string, value interface{})
	SetBaggageItem(key, value string)
	LogFields(fields map[string]interface{})
	SetError(err error)
	Finish()
}

// SpanOption configures a span
type SpanOption func(*SpanConfig)

// SpanConfig holds span configuration
type SpanConfig struct {
	Tags     map[string]interface{}
	ParentID string
}

// WithTag adds a tag to the span
func WithTag(key string, value interface{}) SpanOption {
	return func(cfg *SpanConfig) {
		if cfg.Tags == nil {
			cfg.Tags = make(map[string]interface{})
		}
		cfg.Tags[key] = value
	}
}

// WithParent sets the parent span ID
func WithParent(parentID string) SpanOption {
	return func(cfg *SpanConfig) {
		cfg.ParentID = parentID
	}
}

// NewTracingMiddleware creates a new tracing middleware
func NewTracingMiddleware(serviceName string, tracer Tracer) *TracingMiddleware {
	return &TracingMiddleware{
		serviceName: serviceName,
		tracer:      tracer,
	}
}

// Execute implements Middleware
func (m *TracingMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	if m.tracer == nil {
		result, err := next(ctx)
		if err != nil {
			return nil, err
		}
		return result.Output, nil
	}

	// Create span
	operationName := fmt.Sprintf("%s.%s", m.serviceName, node.Type)
	ctx, span := m.tracer.StartSpan(ctx, operationName,
		WithTag("execution.id", rctx.ExecutionID.String()),
		WithTag("workflow.id", rctx.WorkflowID.String()),
		WithTag("workspace.id", rctx.WorkspaceID.String()),
		WithTag("node.id", node.ID),
		WithTag("node.type", node.Type),
		WithTag("node.name", node.Name),
	)
	defer span.Finish()

	// Execute
	result, err := next(ctx)
	if err != nil {
		span.SetError(err)
		span.SetTag("error", true)
		return nil, err
	}

	span.SetTag("success", true)
	return result.Output, nil
}

// SimpleTracer is a basic in-memory tracer for development
type SimpleTracer struct {
	spans []SimpleSpan
}

// SimpleSpan is a basic span implementation
type SimpleSpan struct {
	TraceID       string
	SpanID        string
	ParentID      string
	OperationName string
	Tags          map[string]interface{}
	Logs          []SpanLog
	StartTime     time.Time
	EndTime       time.Time
	Error         error
}

// SpanLog represents a log entry in a span
type SpanLog struct {
	Timestamp time.Time
	Fields    map[string]interface{}
}

// NewSimpleTracer creates a simple tracer
func NewSimpleTracer() *SimpleTracer {
	return &SimpleTracer{
		spans: make([]SimpleSpan, 0),
	}
}

// StartSpan starts a new span
func (t *SimpleTracer) StartSpan(ctx context.Context, operationName string, opts ...SpanOption) (context.Context, Span) {
	cfg := &SpanConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	span := &simpleSpanImpl{
		tracer: t,
		span: SimpleSpan{
			TraceID:       uuid.New().String(),
			SpanID:        uuid.New().String()[:8],
			ParentID:      cfg.ParentID,
			OperationName: operationName,
			Tags:          cfg.Tags,
			Logs:          make([]SpanLog, 0),
			StartTime:     time.Now(),
		},
	}

	return ctx, span
}

// GetSpans returns all recorded spans
func (t *SimpleTracer) GetSpans() []SimpleSpan {
	return t.spans
}

type simpleSpanImpl struct {
	tracer *SimpleTracer
	span   SimpleSpan
}

func (s *simpleSpanImpl) SetTag(key string, value interface{}) {
	if s.span.Tags == nil {
		s.span.Tags = make(map[string]interface{})
	}
	s.span.Tags[key] = value
}

func (s *simpleSpanImpl) SetBaggageItem(key, value string) {
	s.SetTag("baggage."+key, value)
}

func (s *simpleSpanImpl) LogFields(fields map[string]interface{}) {
	s.span.Logs = append(s.span.Logs, SpanLog{
		Timestamp: time.Now(),
		Fields:    fields,
	})
}

func (s *simpleSpanImpl) SetError(err error) {
	s.span.Error = err
	s.SetTag("error", true)
	s.SetTag("error.message", err.Error())
}

func (s *simpleSpanImpl) Finish() {
	s.span.EndTime = time.Now()
	s.tracer.spans = append(s.tracer.spans, s.span)
}

// NoopTracer is a no-op tracer
type NoopTracer struct{}

// StartSpan returns a no-op span
func (t *NoopTracer) StartSpan(ctx context.Context, operationName string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &noopSpan{}
}

type noopSpan struct{}

func (s *noopSpan) SetTag(key string, value interface{})      {}
func (s *noopSpan) SetBaggageItem(key, value string)          {}
func (s *noopSpan) LogFields(fields map[string]interface{})   {}
func (s *noopSpan) SetError(err error)                        {}
func (s *noopSpan) Finish()                                   {}

// TracingContext holds trace context for propagation
type TracingContext struct {
	TraceID  string
	SpanID   string
	ParentID string
	Baggage  map[string]string
}

// ExtractTracingContext extracts tracing context from input
func ExtractTracingContext(input map[string]interface{}) *TracingContext {
	tc := &TracingContext{
		Baggage: make(map[string]string),
	}

	if traceData, ok := input["$trace"].(map[string]interface{}); ok {
		if v, ok := traceData["traceId"].(string); ok {
			tc.TraceID = v
		}
		if v, ok := traceData["spanId"].(string); ok {
			tc.SpanID = v
		}
		if v, ok := traceData["parentId"].(string); ok {
			tc.ParentID = v
		}
		if baggage, ok := traceData["baggage"].(map[string]interface{}); ok {
			for k, v := range baggage {
				if vs, ok := v.(string); ok {
					tc.Baggage[k] = vs
				}
			}
		}
	}

	return tc
}

// InjectTracingContext injects tracing context into output
func InjectTracingContext(output map[string]interface{}, tc *TracingContext) {
	output["$trace"] = map[string]interface{}{
		"traceId":  tc.TraceID,
		"spanId":   tc.SpanID,
		"parentId": tc.ParentID,
		"baggage":  tc.Baggage,
	}
}

// WorkflowTraceMiddleware traces entire workflow execution
type WorkflowTraceMiddleware struct {
	tracer Tracer
}

// NewWorkflowTraceMiddleware creates a workflow trace middleware
func NewWorkflowTraceMiddleware(tracer Tracer) *WorkflowTraceMiddleware {
	return &WorkflowTraceMiddleware{tracer: tracer}
}

// StartWorkflowTrace starts a trace for a workflow execution
func (m *WorkflowTraceMiddleware) StartWorkflowTrace(ctx context.Context, rctx *processor.RuntimeContext) (context.Context, Span) {
	if m.tracer == nil {
		return ctx, &noopSpan{}
	}

	return m.tracer.StartSpan(ctx, "workflow.execute",
		WithTag("execution.id", rctx.ExecutionID.String()),
		WithTag("workflow.id", rctx.WorkflowID.String()),
		WithTag("workspace.id", rctx.WorkspaceID.String()),
	)
}
