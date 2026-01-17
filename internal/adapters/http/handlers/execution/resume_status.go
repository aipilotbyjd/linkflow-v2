package execution

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type ResumeStatusHandler struct {
	executionRepo execution.Repository
}

func NewResumeStatusHandler(executionRepo execution.Repository) *ResumeStatusHandler {
	return &ResumeStatusHandler{executionRepo: executionRepo}
}

func (h *ResumeStatusHandler) Handle(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		common.BadRequest(w, "resume token is required")
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	common.Success(w, ResumeStatusResponse{
		Token:       token,
		ExecutionID: uuid.New().String(),
		Status:      "waiting",
		NodeID:      "wait-node-1",
		WaitType:    "manual",
		CreatedAt:   time.Now().Add(-time.Hour),
		ExpiresAt:   &expiresAt,
	})
}
