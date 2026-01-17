package execution

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type ResumeRequest struct {
	Data map[string]interface{} `json:"data,omitempty"`
}

type ResumeResponse struct {
	ExecutionID string    `json:"executionId"`
	Status      string    `json:"status"`
	ResumedAt   time.Time `json:"resumedAt"`
}

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
			common.BadRequest(w, "invalid request body")
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

type ResumeStatusResponse struct {
	Token       string    `json:"token"`
	ExecutionID string    `json:"executionId"`
	Status      string    `json:"status"`
	NodeID      string    `json:"nodeId"`
	WaitType    string    `json:"waitType"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

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
