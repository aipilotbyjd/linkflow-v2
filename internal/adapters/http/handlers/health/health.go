package health

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type HealthChecker interface {
	Check(ctx context.Context) error
}

type Handler struct {
	dbChecker    HealthChecker
	redisChecker HealthChecker
	version      string
	startTime    time.Time
}

func NewHandler(checkers ...interface{}) *Handler {
	h := &Handler{
		version:   "1.0.0",
		startTime: time.Now(),
	}

	for i, checker := range checkers {
		if hc, ok := checker.(HealthChecker); ok {
			if i == 0 {
				h.dbChecker = hc
			} else if i == 1 {
				h.redisChecker = hc
			}
		} else if v, ok := checker.(string); ok {
			h.version = v
		}
	}

	return h
}

func NewHandlerWithCheckers(dbChecker, redisChecker HealthChecker, version string) *Handler {
	return &Handler{
		dbChecker:    dbChecker,
		redisChecker: redisChecker,
		version:      version,
		startTime:    time.Now(),
	}
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services,omitempty"`
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	overallStatus := "healthy"

	checks := map[string]HealthChecker{
		"database": h.dbChecker,
		"redis":    h.redisChecker,
	}

	for name, checker := range checks {
		if checker == nil {
			continue
		}
		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()
			status := "healthy"
			if err := checker.Check(ctx); err != nil {
				status = "unhealthy"
				mu.Lock()
				overallStatus = "degraded"
				mu.Unlock()
			}
			mu.Lock()
			services[name] = status
			mu.Unlock()
		}(name, checker)
	}
	wg.Wait()

	resp := HealthResponse{
		Status:    overallStatus,
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
		Timestamp: time.Now(),
		Services:  services,
	}

	if overallStatus != "healthy" {
		common.JSON(w, http.StatusServiceUnavailable, resp)
		return
	}

	common.Success(w, resp)
}

func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	common.Success(w, map[string]string{
		"status": "alive",
	})
}

func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if h.dbChecker != nil {
		if err := h.dbChecker.Check(ctx); err != nil {
			common.JSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not ready",
				"reason": "database unavailable",
			})
			return
		}
	}

	if h.redisChecker != nil {
		if err := h.redisChecker.Check(ctx); err != nil {
			common.JSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not ready",
				"reason": "redis unavailable",
			})
			return
		}
	}

	common.Success(w, map[string]string{
		"status": "ready",
	})
}
