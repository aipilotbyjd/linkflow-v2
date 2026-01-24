package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres"
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
	"gorm.io/gorm"
)

type InvitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

func (r *InvitationRepository) Create(ctx context.Context, invitation *workspace.Invitation) error {
	now := time.Now()
	model := &models.WorkspaceInvitation{
		ID:          invitation.ID,
		WorkspaceID: invitation.WorkspaceID,
		Email:       invitation.Email,
		Role:        string(invitation.Role),
		RoleID:      invitation.RoleID,
		Token:       invitation.Token,
		InvitedBy:   invitation.InvitedBy,
		ExpiresAt:   invitation.ExpiresAt,
		CreatedAt:   now,
	}
	return postgres.GetTx(ctx, r.db).Create(model).Error
}

func (r *InvitationRepository) Update(ctx context.Context, invitation *workspace.Invitation) error {
	model := &models.WorkspaceInvitation{
		ID:         invitation.ID,
		AcceptedAt: invitation.AcceptedAt,
	}
	return postgres.GetTx(ctx, r.db).Model(model).Updates(model).Error
}

func (r *InvitationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return postgres.GetTx(ctx, r.db).Delete(&models.WorkspaceInvitation{}, "id = ?", id).Error
}

func (r *InvitationRepository) FindByID(ctx context.Context, id uuid.UUID) (*workspace.Invitation, error) {
	var model models.WorkspaceInvitation
	if err := postgres.GetTx(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil if not found, or specific error
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *InvitationRepository) FindByToken(ctx context.Context, token string) (*workspace.Invitation, error) {
	var model models.WorkspaceInvitation
	if err := postgres.GetTx(ctx, r.db).First(&model, "token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *InvitationRepository) FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]workspace.Invitation, error) {
	var models []models.WorkspaceInvitation
	if err := postgres.GetTx(ctx, r.db).Where("workspace_id = ?", workspaceID).Find(&models).Error; err != nil {
		return nil, err
	}
	invitations := make([]workspace.Invitation, len(models))
	for i, m := range models {
		invitations[i] = *r.toDomain(&m)
	}
	return invitations, nil
}

func (r *InvitationRepository) FindByEmail(ctx context.Context, email string) ([]workspace.Invitation, error) {
	var models []models.WorkspaceInvitation
	if err := postgres.GetTx(ctx, r.db).Where("email = ?", email).Find(&models).Error; err != nil {
		return nil, err
	}
	invitations := make([]workspace.Invitation, len(models))
	for i, m := range models {
		invitations[i] = *r.toDomain(&m)
	}
	return invitations, nil
}

func (r *InvitationRepository) FindPendingByWorkspaceAndEmail(ctx context.Context, workspaceID uuid.UUID, email string) (*workspace.Invitation, error) {
	var model models.WorkspaceInvitation
	if err := postgres.GetTx(ctx, r.db).Where("workspace_id = ? AND email = ? AND accepted_at IS NULL AND expires_at > ?", workspaceID, email, time.Now()).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomain(&model), nil
}

func (r *InvitationRepository) CleanupExpired(ctx context.Context) (int64, error) {
	result := postgres.GetTx(ctx, r.db).Where("expires_at < ? AND accepted_at IS NULL", time.Now()).Delete(&models.WorkspaceInvitation{})
	return result.RowsAffected, result.Error
}

func (r *InvitationRepository) toDomain(m *models.WorkspaceInvitation) *workspace.Invitation {
	return &workspace.Invitation{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		Email:       m.Email,
		Role:        workspace.Role(m.Role),
		RoleID:      m.RoleID,
		Token:       m.Token,
		InvitedBy:   m.InvitedBy,
		ExpiresAt:   m.ExpiresAt,
		AcceptedAt:  m.AcceptedAt,
		CreatedAt:   m.CreatedAt,
	}
}
