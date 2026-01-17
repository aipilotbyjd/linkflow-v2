package credential

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
)

type CredentialTester interface {
	Test(ctx context.Context, credType credential.Type, encryptedData string) error
}

type TestHandler struct {
	credentialRepo credential.Repository
	tester         CredentialTester
}

func NewTestHandler(credentialRepo credential.Repository, tester CredentialTester) *TestHandler {
	return &TestHandler{
		credentialRepo: credentialRepo,
		tester:         tester,
	}
}

type TestResponse struct {
	Success  bool         `json:"success"`
	Message  string       `json:"message"`
	TestedAt time.Time    `json:"testedAt"`
	Details  *TestDetails `json:"details,omitempty"`
}

type TestDetails struct {
	ConnectionTime string `json:"connectionTime,omitempty"`
	Endpoint       string `json:"endpoint,omitempty"`
	Version        string `json:"version,omitempty"`
}

func (h *TestHandler) Handle(w http.ResponseWriter, r *http.Request) {
	credentialIDStr := chi.URLParam(r, "credentialId")
	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		common.BadRequest(w, "invalid credential ID")
		return
	}

	workspaceID := middleware.GetWorkspaceID(r.Context())

	cred, err := h.credentialRepo.FindByID(r.Context(), credentialID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if cred.WorkspaceID.String() != workspaceID.String() {
		common.Forbidden(w, "access denied")
		return
	}

	startTime := time.Now()

	var testErr error
	if h.tester != nil {
		testErr = h.tester.Test(r.Context(), cred.Type, cred.Data)
	}

	connectionTime := time.Since(startTime)

	if testErr != nil {
		common.Success(w, TestResponse{
			Success:  false,
			Message:  "Connection test failed: " + testErr.Error(),
			TestedAt: time.Now(),
		})
		return
	}

	common.Success(w, TestResponse{
		Success:  true,
		Message:  "Connection test successful",
		TestedAt: time.Now(),
		Details: &TestDetails{
			ConnectionTime: connectionTime.String(),
		},
	})
}
