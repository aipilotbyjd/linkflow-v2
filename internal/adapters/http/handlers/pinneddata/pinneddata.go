package pinneddata

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

type PinnedData struct {
	ID         string          `json:"id"`
	WorkflowID string          `json:"workflowId"`
	NodeID     string          `json:"nodeId"`
	Data       json.RawMessage `json:"data"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

type PinnedDataRepository interface {
	GetByWorkflow(workflowID string) ([]PinnedData, error)
	GetByNode(workflowID, nodeID string) (*PinnedData, error)
	Set(workflowID, nodeID string, data json.RawMessage) (*PinnedData, error)
	Delete(workflowID, nodeID string) error
}

type Handler struct {
	repo PinnedDataRepository
}

func NewHandler(repo PinnedDataRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowId")
	_ = middleware.GetWorkspaceID(r.Context())

	pinnedData := []PinnedData{
		{
			ID:         uuid.New().String(),
			WorkflowID: workflowID,
			NodeID:     "node-1",
			Data:       json.RawMessage(`{"key": "sample data"}`),
			CreatedAt:  time.Now().AddDate(0, 0, -1),
			UpdatedAt:  time.Now(),
		},
	}

	common.Success(w, map[string]interface{}{
		"pinnedData": pinnedData,
		"workflowId": workflowID,
	})
}

func (h *Handler) GetByNode(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowId")
	nodeID := chi.URLParam(r, "nodeId")

	pinnedData := PinnedData{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		NodeID:     nodeID,
		Data:       json.RawMessage(`{"key": "sample data for node"}`),
		CreatedAt:  time.Now().AddDate(0, 0, -1),
		UpdatedAt:  time.Now(),
	}

	common.Success(w, pinnedData)
}

type SetPinnedDataRequest struct {
	NodeID string          `json:"nodeId"`
	Data   json.RawMessage `json:"data"`
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowId")

	var req SetPinnedDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.NodeID == "" {
		common.Error(w, http.StatusBadRequest, "MISSING_NODE_ID", "Node ID is required")
		return
	}

	if len(req.Data) == 0 {
		common.Error(w, http.StatusBadRequest, "MISSING_DATA", "Data is required")
		return
	}

	pinnedData := PinnedData{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		NodeID:     req.NodeID,
		Data:       req.Data,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	common.JSON(w, http.StatusCreated, pinnedData)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowId")
	nodeID := chi.URLParam(r, "nodeId")

	_ = workflowID
	_ = nodeID

	w.WriteHeader(http.StatusNoContent)
}
