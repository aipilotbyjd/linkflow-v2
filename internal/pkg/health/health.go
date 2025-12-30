package health

import (
	"context"
	"sync"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name      string        `json:"name"`
	Status    Status        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Duration  time.Duration `json:"duration_ms"`
	Timestamp time.Time     `json:"timestamp"`
}

// Checker is an interface for health check implementations
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// Manager manages multiple health checkers
type Manager struct {
	checkers []Checker
	timeout  time.Duration
	mu       sync.RWMutex
}

// NewManager creates a new health check manager
func NewManager(timeout time.Duration) *Manager {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Manager{
		checkers: make([]Checker, 0),
		timeout:  timeout,
	}
}

// Register adds a health checker to the manager
func (m *Manager) Register(checker Checker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers = append(m.checkers, checker)
}

// CheckAll runs all health checks and returns results
func (m *Manager) CheckAll(ctx context.Context) ([]CheckResult, Status) {
	m.mu.RLock()
	checkers := make([]Checker, len(m.checkers))
	copy(checkers, m.checkers)
	m.mu.RUnlock()

	results := make([]CheckResult, len(checkers))
	overallStatus := StatusHealthy

	var wg sync.WaitGroup
	for i, checker := range checkers {
		wg.Add(1)
		go func(idx int, c Checker) {
			defer wg.Done()
			results[idx] = m.runCheck(ctx, c)
		}(i, checker)
	}
	wg.Wait()

	// Determine overall status
	for _, result := range results {
		if result.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		}
		if result.Status == StatusDegraded {
			overallStatus = StatusDegraded
		}
	}

	return results, overallStatus
}

func (m *Manager) runCheck(ctx context.Context, checker Checker) CheckResult {
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	start := time.Now()
	err := checker.Check(ctx)
	duration := time.Since(start)

	result := CheckResult{
		Name:      checker.Name(),
		Duration:  duration,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = err.Error()
	} else {
		result.Status = StatusHealthy
	}

	return result
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status    Status        `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Checks    []CheckResult `json:"checks,omitempty"`
}

// GetHealthResponse returns a formatted health response
func (m *Manager) GetHealthResponse(ctx context.Context, includeDetails bool) HealthResponse {
	checks, status := m.CheckAll(ctx)
	
	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
	}
	
	if includeDetails {
		response.Checks = checks
	}
	
	return response
}
