package health

import (
	"context"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// Checker interface for health checks
type Checker interface {
	Check(ctx context.Context) error
}

// Handler handles health check endpoints
type Handler struct {
	checkers map[string]Checker
}

// NewHandler creates a new health handler
func NewHandler() *Handler {
	return &Handler{
		checkers: make(map[string]Checker),
	}
}

// RegisterChecker registers a health checker
func (h *Handler) RegisterChecker(name string, checker Checker) {
	h.checkers[name] = checker
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Services  map[string]ServiceStatus `json:"services,omitempty"`
}

// ServiceStatus represents individual service status
type ServiceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Health handles basic health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	common.Success(w, HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// Live handles liveness probe
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	common.Success(w, map[string]string{
		"status": "alive",
	})
}

// Ready handles readiness probe
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]ServiceStatus)
	allHealthy := true

	for name, checker := range h.checkers {
		if err := checker.Check(ctx); err != nil {
			services[name] = ServiceStatus{
				Status:  "unhealthy",
				Message: err.Error(),
			}
			allHealthy = false
		} else {
			services[name] = ServiceStatus{
				Status: "healthy",
			}
		}
	}

	response := HealthResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	}

	if allHealthy {
		response.Status = "ready"
		common.Success(w, response)
	} else {
		response.Status = "not_ready"
		common.JSON(w, http.StatusServiceUnavailable, response)
	}
}
