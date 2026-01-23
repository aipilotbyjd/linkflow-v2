package byok

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

// KeyEncryptor interface for encrypting API keys
type KeyEncryptor interface {
	Encrypt(plaintext string) (string, error)
}

// CreateHandler handles creating BYOK configurations
type CreateHandler struct {
	repo      BYOKRepository
	encryptor KeyEncryptor
}

func NewCreateHandler(repo BYOKRepository, encryptor KeyEncryptor) *CreateHandler {
	return &CreateHandler{repo: repo, encryptor: encryptor}
}

func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)
	userID := middleware.GetUserID(ctx)

	var req CreateBYOKRequest
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

	// Validate provider
	provider := billing.AIProvider(req.Provider)
	if billing.GetProviderInfo(provider) == nil {
		common.BadRequest(w, "Invalid provider. Supported: openai, anthropic, google, azure_openai, groq, mistral")
		return
	}

	// Check if already exists for this provider
	existing, _ := h.repo.FindByWorkspaceAndProvider(ctx, workspaceID, provider)
	if existing != nil {
		common.Error(w, http.StatusConflict, "ALREADY_EXISTS", "A configuration for this provider already exists")
		return
	}

	// Encrypt the API key
	encryptedKey, err := h.encryptor.Encrypt(req.APIKey)
	if err != nil {
		common.Error(w, http.StatusInternalServerError, "ENCRYPTION_ERROR", "Failed to secure API key")
		return
	}

	config := &billing.BYOKConfig{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		Provider:        provider,
		Name:            req.Name,
		APIKeyEncrypted: encryptedKey,
		APIKeyMasked:    billing.MaskAPIKey(req.APIKey),
		OrganizationID:  req.OrganizationID,
		BaseURL:         req.BaseURL,
		IsActive:        true,
		IsValid:         false, // Will be validated separately
		CreatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.repo.Create(ctx, config); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToBYOKResponse(config))
}
