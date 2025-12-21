package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func NewHealthHandlerWithDeps(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	healthy := true

	// Check database
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			checks["database"] = "error: " + err.Error()
			healthy = false
		} else {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := sqlDB.PingContext(ctx); err != nil {
				checks["database"] = "error: " + err.Error()
				healthy = false
			} else {
				checks["database"] = "ok"
			}
		}
	} else {
		checks["database"] = "not configured"
	}

	// Check Redis
	if h.redis != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "error: " + err.Error()
			healthy = false
		} else {
			checks["redis"] = "ok"
		}
	} else {
		checks["redis"] = "not configured"
	}

	status := "healthy"
	statusCode := http.StatusOK
	if !healthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	dto.JSON(w, statusCode, map[string]interface{}{
		"status":  status,
		"service": "linkflow-api",
		"checks":  checks,
	})
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	dto.JSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			dto.ErrorResponse(w, http.StatusServiceUnavailable, "database not ready: "+err.Error())
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			dto.ErrorResponse(w, http.StatusServiceUnavailable, "database not ready: "+err.Error())
			return
		}
	}

	// Check Redis connection
	if h.redis != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.redis.Ping(ctx).Err(); err != nil {
			dto.ErrorResponse(w, http.StatusServiceUnavailable, "redis not ready: "+err.Error())
			return
		}
	}

	dto.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
