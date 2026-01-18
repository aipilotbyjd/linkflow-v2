package execution

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type ResumeHandler struct {
	executionRepo execution.Repository
}

func NewResumeHandler(executionRepo execution.Repository) *ResumeHandler {
	return &ResumeHandler{executionRepo: executionRepo}
}

func (h *ResumeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		common.BadRequest(w, "resume token is required")
		return
	}

	var req ResumeRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.BadRequest(w, "Invalid request body")
			return
		}

		if errors := validation.Validate(req); len(errors) > 0 {
			details := make([]common.ValidationDetail, len(errors))
			for i, e := range errors {
				details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
			}
			common.ValidationErrors(w, details)
			return
		}
	}

	executionID := uuid.New()
	_ = req.Data

	common.Success(w, ResumeResponse{
		ExecutionID: executionID.String(),
		Status:      "running",
		ResumedAt:   time.Now(),
	})
}
