package credential

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type CreateCredentialCommand struct {
	WorkspaceID  uuid.UUID
	CreatedBy    uuid.UUID
	Name         string
	Description  *string
	Type         credential.Type
	Provider     string
	Data         types.JSON
	SharingScope credential.SharingScope
}

// EncryptionService interface for encrypting credential data
type EncryptionService interface {
	Encrypt(plaintext string) (string, error)
}

type CreateCredentialHandler struct {
	credentialRepo credential.Repository
	eventBus       events.Bus
	encryptor      EncryptionService
}

func NewCreateCredentialHandler(credentialRepo credential.Repository, eventBus events.Bus, encryptor EncryptionService) *CreateCredentialHandler {
	return &CreateCredentialHandler{
		credentialRepo: credentialRepo,
		eventBus:       eventBus,
		encryptor:      encryptor,
	}
}

func (h *CreateCredentialHandler) Handle(ctx context.Context, cmd CreateCredentialCommand) (*credential.Credential, error) {
	// Check name uniqueness first
	exists, err := h.credentialRepo.ExistsByName(ctx, cmd.WorkspaceID, cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check credential name: %w", err)
	}
	if exists {
		return nil, credential.ErrCredentialNameExists
	}

	// Encrypt credential data
	dataJSON, err := json.Marshal(cmd.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize credential data: %w", err)
	}

	var encryptedData string
	if h.encryptor != nil {
		encryptedData, err = h.encryptor.Encrypt(string(dataJSON))
		if err != nil {
			return nil, credential.ErrEncryptionFailed
		}
	} else {
		// Fallback for testing - in production, encryptor should always be set
		encryptedData = string(dataJSON)
	}

	// Create credential (includes validation)
	cred, err := credential.NewCredential(cmd.WorkspaceID, cmd.CreatedBy, cmd.Name, cmd.Type, encryptedData)
	if err != nil {
		return nil, err
	}

	cred.Description = cmd.Description
	if cmd.Provider != "" {
		cred.SetProvider(cmd.Provider, nil)
	}
	if cmd.SharingScope != "" {
		cred.SetSharingScope(cmd.SharingScope)
	}

	if err := h.credentialRepo.Create(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	if h.eventBus != nil {
		_ = h.eventBus.Publish(ctx, events.CredentialCreated{
			BaseEvent:    events.NewBaseEvent("credential.created", cred.ID, "credential"),
			CredentialID: cred.ID,
			WorkspaceID:  cred.WorkspaceID,
			Type:         string(cred.Type),
			CreatedBy:    cred.CreatedBy,
		})
	}

	return cred, nil
}
