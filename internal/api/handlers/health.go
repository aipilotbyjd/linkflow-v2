package handlers

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	dto.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "linkflow-api",
	})
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	dto.JSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// TODO: Check database and Redis connections
	dto.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
