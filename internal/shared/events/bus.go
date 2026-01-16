package events

import (
	"context"
	"sync"
)

// Handler processes events
type Handler func(ctx context.Context, event Event) error

// Bus is the event bus interface
type Bus interface {
	Publish(ctx context.Context, events ...Event) error
	Subscribe(eventName string, handler Handler)
	Unsubscribe(eventName string, handler Handler)
}

// InMemoryBus is an in-memory event bus implementation
type InMemoryBus struct {
	handlers map[string][]Handler
	mu       sync.RWMutex
}

func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		handlers: make(map[string][]Handler),
	}
}

func (b *InMemoryBus) Publish(ctx context.Context, events ...Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, event := range events {
		handlers := b.handlers[event.EventName()]
		for _, handler := range handlers {
			if err := handler(ctx, event); err != nil {
				return err
			}
		}
		// Also notify wildcard handlers
		wildcardHandlers := b.handlers["*"]
		for _, handler := range wildcardHandlers {
			if err := handler(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *InMemoryBus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

func (b *InMemoryBus) Unsubscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	handlers := b.handlers[eventName]
	for i, h := range handlers {
		// Compare function pointers
		if &h == &handler {
			b.handlers[eventName] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// AsyncBus wraps a Bus and publishes events asynchronously
type AsyncBus struct {
	bus     Bus
	queue   chan eventWrapper
	workers int
	wg      sync.WaitGroup
}

type eventWrapper struct {
	ctx   context.Context
	event Event
}

func NewAsyncBus(bus Bus, queueSize, workers int) *AsyncBus {
	if queueSize <= 0 {
		queueSize = 1000
	}
	if workers <= 0 {
		workers = 4
	}
	ab := &AsyncBus{
		bus:     bus,
		queue:   make(chan eventWrapper, queueSize),
		workers: workers,
	}
	ab.start()
	return ab
}

func (b *AsyncBus) start() {
	for i := 0; i < b.workers; i++ {
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			for ew := range b.queue {
				_ = b.bus.Publish(ew.ctx, ew.event)
			}
		}()
	}
}

func (b *AsyncBus) Publish(ctx context.Context, events ...Event) error {
	for _, event := range events {
		select {
		case b.queue <- eventWrapper{ctx: ctx, event: event}:
		default:
			// Queue full, publish synchronously
			if err := b.bus.Publish(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *AsyncBus) Subscribe(eventName string, handler Handler) {
	b.bus.Subscribe(eventName, handler)
}

func (b *AsyncBus) Unsubscribe(eventName string, handler Handler) {
	b.bus.Unsubscribe(eventName, handler)
}

func (b *AsyncBus) Close() {
	close(b.queue)
	b.wg.Wait()
}

// NoOpBus is a no-operation event bus (for testing)
type NoOpBus struct{}

func NewNoOpBus() *NoOpBus {
	return &NoOpBus{}
}

func (b *NoOpBus) Publish(ctx context.Context, events ...Event) error {
	return nil
}

func (b *NoOpBus) Subscribe(eventName string, handler Handler) {}

func (b *NoOpBus) Unsubscribe(eventName string, handler Handler) {}
