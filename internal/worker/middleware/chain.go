package middleware

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/worker/processor"
)

// NextFunc is the function to call the next middleware
type NextFunc func(ctx context.Context) (*processor.NodeResult, error)

// Middleware is the interface for node execution middleware (matches processor.Middleware)
type Middleware interface {
	Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error)
}

// Ensure Chain implements processor.MiddlewareChain
var _ processor.MiddlewareChain = (*Chain)(nil)

// Chain manages a chain of middleware
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
	}
}

// Use adds middleware to the chain
func (c *Chain) Use(m Middleware) *Chain {
	c.middlewares = append(c.middlewares, m)
	return c
}

// Execute runs the middleware chain
func (c *Chain) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, handler func(ctx context.Context) (*processor.NodeResult, error)) (map[string]interface{}, error) {
	// Build the chain from the end
	final := handler

	// Wrap in reverse order so first middleware executes first
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		middleware := c.middlewares[i]
		next := final // Capture current final

		final = func(ctx context.Context) (*processor.NodeResult, error) {
			output, err := middleware.Execute(ctx, rctx, node, func(ctx context.Context) (*processor.NodeResult, error) {
				return next(ctx)
			})
			if err != nil {
				return nil, err
			}
			return &processor.NodeResult{Output: output}, nil
		}
	}

	// Execute the chain
	result, err := final(ctx)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.Output, nil
}

// Len returns the number of middlewares
func (c *Chain) Len() int {
	return len(c.middlewares)
}

// MiddlewareFunc is a function that implements Middleware
type MiddlewareFunc func(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error)

// Execute implements Middleware
func (f MiddlewareFunc) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	return f(ctx, rctx, node, next)
}

// Wrap wraps a function as middleware
func Wrap(fn MiddlewareFunc) Middleware {
	return fn
}

// Compose combines multiple middleware into one
func Compose(middlewares ...Middleware) Middleware {
	return MiddlewareFunc(func(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
		chain := NewChain(middlewares...)
		return chain.Execute(ctx, rctx, node, next)
	})
}

// Conditional creates middleware that only executes if condition is true
func Conditional(condition func(*processor.NodeDefinition) bool, m Middleware) Middleware {
	return MiddlewareFunc(func(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
		if condition(node) {
			return m.Execute(ctx, rctx, node, next)
		}
		result, err := next(ctx)
		if err != nil {
			return nil, err
		}
		return result.Output, nil
	})
}

// ForNodeTypes creates middleware that only executes for specific node types
func ForNodeTypes(types []string, m Middleware) Middleware {
	typeSet := make(map[string]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	return Conditional(func(node *processor.NodeDefinition) bool {
		return typeSet[node.Type]
	}, m)
}

// ForCategories creates middleware that only executes for specific categories
func ForCategories(categories []string, m Middleware) Middleware {
	return Conditional(func(node *processor.NodeDefinition) bool {
		for _, cat := range categories {
			if len(node.Type) >= len(cat) && node.Type[:len(cat)] == cat {
				return true
			}
		}
		return false
	}, m)
}
